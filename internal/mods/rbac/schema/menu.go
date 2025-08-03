package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

const (
	MenuStatusDisabled = 2
	MenuStatusEnabled  = 1
)

var (
	MenusOrderParams = []util.OrderByParam{
		{Field: "created_at", Direction: util.DESC},
	}
)

// Menu management for RBAC
type Menu struct {
	ID          string    `json:"id" gorm:"size:20;primarykey;"`   // Unique ID
	Code        string    `json:"code" gorm:"size:128;index"`      // Display code of menu
	Name        string    `json:"name" gorm:"size:128;index"`      // Display name of menu
	Path        string    `json:"path" gorm:"size:255;"`           // Access path of menu
	Component   string    `json:"component" gorm:"size:255;"`      // Code of menu (unique for each level)
	Description string    `json:"description" gorm:"size:1024"`    // Details about menu
	ParentID    string    `json:"parent_id" gorm:"size:20;index;"` // Parent ID (From Menu.ID)
	Children    *Menus    `json:"children" gorm:"-"`               // Child menus
	MenuType    int8      `json:"menu_type"`                       // MenuType of menu (1-dir, 2-menu)
	Status      int8      `json:"status"`                          // Status of menu (1-enabled, 2-disabled)
	CreatedAt   time.Time `json:"created_at" gorm:"index;"`        // Create time
	UpdatedAt   time.Time `json:"updated_at" gorm:"index;"`        // Update time
	Meta        *MenuMeta `json:"meta" gorm:"-"`                   // Meta of menu
}

func (a *Menu) TableName() string {
	return config.C.FormatTableName("menu")
}

// Defining the query parameters for the `Menu` struct.
type MenuQueryParam struct {
	util.PaginationParam
	CodePath         string   `form:"code"` // Code path (like xxx.xxx.xxx)
	LikeName         string   `form:"name"` // Display name of menu
	InIDs            []string `form:"-"`    // Include menu IDs
	Status           int8     `form:"-"`    // Status of menu (2-disabled, 1-enabled)
	ParentID         string   `form:"-"`    // Parent ID (From Menu.ID)
	ParentPathPrefix string   `form:"-"`    // Parent path (split by .)
	UserID           string   `form:"-"`    // User ID
	RoleID           string   `form:"-"`    // Role ID
}

// Defining the query options for the `Menu` struct.
type MenuQueryOptions struct {
	util.QueryOptions
}

// Defining the query result for the `Menu` struct.
type MenuQueryResult struct {
	Data       Menus
	PageResult *util.PaginationResult
}

// Defining the slice of `Menu` struct.
type Menus []*Menu

func (a Menus) Len() int {
	return len(a)
}

func (a Menus) Less(i, j int) bool {
	if a[i].Meta.Order == a[j].Meta.Order {
		return a[i].CreatedAt.Unix() > a[j].CreatedAt.Unix()
	}
	return a[i].Meta.Order > a[j].Meta.Order
}

func (a Menus) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Menus) ToMap() map[string]*Menu {
	m := make(map[string]*Menu)
	for _, item := range a {
		m[item.ID] = item
	}
	return m
}

func (a Menus) SplitParentIDs() []string {
	parentIDs := make([]string, 0, len(a))
	idMapper := make(map[string]struct{})
	for _, item := range a {
		if _, ok := idMapper[item.ID]; ok {
			continue
		}
		idMapper[item.ID] = struct{}{}
	}
	return parentIDs
}

func (a Menus) ToTree() Menus {
	var list Menus
	m := a.ToMap()
	for _, item := range a {
		if item.ParentID == "" {
			list = append(list, item)
			continue
		}
		if parent, ok := m[item.ParentID]; ok {
			if parent.Children == nil {
				children := Menus{item}
				parent.Children = &children
				continue
			}
			*parent.Children = append(*parent.Children, item)
		}
	}
	return list
}

// Defining the data structure for creating a `Menu` struct.
type MenuForm struct {
	Name        string    `json:"name" binding:"required,max=128"` // Display name of menu
	Code        string    `json:"code" binding:"required,max=128"` // Display code of menu
	Path        string    `json:"path" binding:"required"`         // Access path of menu
	Component   string    `json:"component"`                       // Details about menu
	Description string    `json:"description"`                     // Details about menu
	ParentID    string    `json:"parent_id"`                       // Parent ID (From Menu.ID)
	MenuType    int8      `json:"menu_type" binding:"required"`    // MenuType of menu (1-dir, 2-menu)
	Status      int8      `json:"status" binding:"required"`       // Status of menu (1-enabled, 2-disabled)
	Meta        *MenuMeta `json:"meta" binding:"required"`         // Meta of menu
}

func (a *MenuForm) FillTo(menu *Menu) error {
	menu.Name = a.Name
	menu.Code = a.Code
	menu.Path = a.Path
	menu.Component = a.Component
	menu.Description = a.Description
	menu.ParentID = a.ParentID
	menu.MenuType = a.MenuType
	menu.Status = a.Status
	return nil
}
