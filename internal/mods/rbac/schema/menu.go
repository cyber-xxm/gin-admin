package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

const (
	MenuStatusDisabled = "disabled"
	MenuStatusEnabled  = "enabled"
)

var (
	MenusOrderParams = []util.OrderByParam{
		{Field: "sequence", Direction: util.DESC},
		{Field: "created_at", Direction: util.DESC},
	}
)

// Menu management for RBAC
type Menu struct {
	ID          string    `json:"id" gorm:"size:20;primarykey;"`   // Unique ID
	Name        string    `json:"name" gorm:"size:128;index"`      // Display name of menu
	Path        string    `json:"path" gorm:"size:255;"`           // Access path of menu
	Component   string    `json:"component" gorm:"size:255;"`      // Code of menu (unique for each level)
	Description string    `json:"description" gorm:"size:1024"`    // Details about menu
	ParentID    string    `json:"parent_id" gorm:"size:20;index;"` // Parent ID (From Menu.ID)
	Children    *Menus    `json:"children" gorm:"-"`               // Child menus
	Status      string    `json:"status" gorm:"size:20;index"`     // Status of menu (enabled, disabled)
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
	CodePath         string   `form:"code"`             // Code path (like xxx.xxx.xxx)
	LikeName         string   `form:"name"`             // Display name of menu
	IncludeResources bool     `form:"includeResources"` // Include resources
	InIDs            []string `form:"-"`                // Include menu IDs
	Status           string   `form:"-"`                // Status of menu (disabled, enabled)
	ParentID         string   `form:"-"`                // Parent ID (From Menu.ID)
	ParentPathPrefix string   `form:"-"`                // Parent path (split by .)
	UserID           string   `form:"-"`                // User ID
	RoleID           string   `form:"-"`                // Role ID
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
	Name        string    `json:"name" binding:"required,max=128"`                  // Display name of menu
	Path        string    `json:"path" binding:"required"`                          // Access path of menu
	Component   string    `json:"component"`                                        // Details about menu
	Description string    `json:"description"`                                      // Details about menu
	ParentID    string    `json:"parent_id"`                                        // Parent ID (From Menu.ID)
	Status      string    `json:"status" binding:"required,oneof=disabled enabled"` // Status of menu (enabled, disabled)
	Meta        *MenuMeta `json:"meta" binding:"required"`                          // Meta of menu
}

func (a *MenuForm) FillTo(menu *Menu) error {
	menu.Name = a.Name
	menu.Path = a.Path
	menu.Component = a.Component
	menu.Description = a.Description
	menu.ParentID = a.ParentID
	menu.Status = a.Status
	return nil
}
