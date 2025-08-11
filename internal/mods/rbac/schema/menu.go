package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/datatypes"
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
	ID         string         `json:"id" gorm:"size:20;primarykey;comment:'主键'"`                              // Unique ID
	RouteName  string         `json:"route_name,omitempty" gorm:"size:128;index;comment:'路由名称'"`              // Display code of menu
	RoutePath  string         `json:"route_path,omitempty" gorm:"size:255;comment:'路由路径'"`                    // Access path of menu
	MenuName   string         `json:"menu_name,omitempty" gorm:"size:128;index;comment:'菜单名称'"`               // Display name of menu
	MenuType   int8           `json:"menu_type,omitempty" gorm:"default:1;comment:'菜单类型：1-目录，2-菜单'"`          // MenuType of menu (1-dir, 2-menu)
	Component  string         `json:"component,omitempty" gorm:"size:255;comment:'菜单组件'"`                     // Code of menu (unique for each level)
	ParentID   string         `json:"parent_id,omitempty" gorm:"size:20;index;comment:'父ID'"`                 // Parent ID (From Menu.ID)
	Order      int            `json:"order,omitempty" gorm:"index;comment:'菜单排序'"`                            // Order
	I18nKey    string         `json:"i18n_key,omitempty" gorm:"size:255;comment:'菜单国际化key'"`                  // I18nKey
	Href       string         `json:"href,omitempty" gorm:"size:255;comment:'菜单外链'"`                          // Href
	Query      datatypes.JSON `json:"query,omitempty" gorm:"size:255;comment:'路由查询参数'"`                       // Query
	Icon       string         `json:"icon,omitempty" gorm:"size:255;comment:'菜单图标'"`                          // Icon
	IconType   int8           `json:"icon_type,omitempty" gorm:"default:1;comment:'图标类型：1-iconify图标，2-本地图标'"` // IconType of menu (1-iconify, 2-local)
	Children   *Menus         `json:"children,omitempty" gorm:"-"`                                            // Child menus
	Status     int8           `json:"status,omitempty" gorm:"default:1;comment:'菜单状态：1-启用，2-禁用'"`             // Status of menu
	HideMenu   bool           `json:"hide_menu,omitempty" gorm:"index;comment:'隐藏菜单'"`                        // HideMenu
	MultiTab   bool           `json:"multi_tab,omitempty" gorm:"index;comment:'支持多页签'"`                       // MultiTab
	Constant   bool           `json:"constant,omitempty" gorm:"index;comment:'常量路由'"`                         // Constant
	KeepAlive  bool           `json:"keep_alive,omitempty" gorm:"index;comment:'是否缓存该路由'"`                    // KeepAlive
	ActiveMenu string         `json:"active_menu,omitempty" gorm:"index;comment:'进入该路由时激活的菜单键'"`              // ActiveMenu
	CreatedAt  time.Time      `json:"created_at,omitempty" gorm:"index;comment:'创建时间'"`                       // Create time
	UpdatedAt  time.Time      `json:"updated_at,omitempty" gorm:"index;comment:'更新时间'"`                       // Update time
}

func (a *Menu) TableName() string {
	return config.C.FormatTableName("menu")
}

type Route struct {
	ID        string  `json:",omitempty"`
	ParentID  string  `json:",omitempty"`
	Name      string  `json:"name,omitempty"`
	Path      string  `json:"path,omitempty"`
	Component string  `json:"component,omitempty"`
	Redirect  string  `json:"redirect,omitempty"`
	Children  *Routes `json:"children,omitempty"`
	Meta      struct {
		Title      string `json:"title,omitempty"`
		I18nKey    string `json:"i18n_key,omitempty"`
		Icon       string `json:"icon,omitempty"`
		Order      int    `json:"order,omitempty"`
		HideMenu   bool   `json:"hide_menu,omitempty"`
		ActiveMenu string `json:"active_menu,omitempty"`
		MultiTab   bool   `json:"multi_tab,omitempty"`
		KeepAlive  bool   `json:"keep_alive,omitempty"`
	} `json:"meta,omitempty"`
}

// Defining the query parameters for the `Menu` struct.
type MenuQueryParam struct {
	util.PaginationParam
	Status int8   `form:"-"` // Status of menu (2-disabled, 1-enabled)
	UserID string `form:"-"` // User ID
	RoleID string `form:"-"` // Role ID
}

// Defining the query options for the `Menu` struct.
type MenuQueryOptions struct {
	util.QueryOptions
}

// Defining the query result for the `Menu` struct.
type MenuQueryResult struct {
	Data       Menus
	Routes     Routes
	PageResult *util.PaginationResult
}

// Defining the slice of `Menu` struct.
type Menus []*Menu

type Routes []*Route

func (a Menus) Len() int {
	return len(a)
}

func (a Menus) Less(i, j int) bool {
	if a[i].Order == a[j].Order {
		return a[i].CreatedAt.Unix() > a[j].CreatedAt.Unix()
	}
	return a[i].Order > a[j].Order
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
func (a Routes) ToMap() map[string]*Route {
	m := make(map[string]*Route)
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

func (a Routes) ToTree() Routes {
	var list Routes
	m := a.ToMap()
	for _, item := range a {
		if item.ParentID == "" {
			list = append(list, item)
			continue
		}
		if parent, ok := m[item.ParentID]; ok {
			if parent.Children == nil {
				children := Routes{item}
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
	RouteName  string         `json:"route_name" binding:"required,max=128"` // RouteName of menu
	RoutePath  string         `json:"route_path" binding:"required,max=128"` // RoutePath of menu
	MenuName   string         `json:"menu_name" binding:"required,max=128"`  // MenuName of menu
	MenuType   int8           `json:"menu_type" binding:"required"`          // MenuType of menu (1-dir, 2-menu)
	Component  string         `json:"component"`                             // Component of menu
	ParentID   string         `json:"parent_id"`                             // Parent ID (From Menu.ID)
	Order      int            `json:"order"`                                 // Order of menu
	I18nKey    string         `json:"i18n_key"`                              // I18nKey of menu
	Href       string         `json:"href"`                                  // Href of menu
	Query      datatypes.JSON `json:"query"`                                 // Query of menu
	Icon       string         `json:"icon"`                                  // Icon of menu
	IconType   int8           `json:"icon_type"`                             // Details about menu
	Status     int8           `json:"status"  binding:"required"`            // IconType of menu
	HideMenu   bool           `json:"hide_menu" `                            // HideMenu of menu
	MultiTab   bool           `json:"multi_tab"`                             // MultiTab of menu
	Constant   bool           `json:"constant"`                              // Constant of menu
	ActiveMenu string         `json:"active_menu"`                           // ActiveMenu of menu
}

func (a *MenuForm) FillTo(menu *Menu) {
	menu.RouteName = a.RouteName
	menu.RoutePath = a.RoutePath
	menu.MenuName = a.MenuName
	menu.MenuType = a.MenuType
	menu.Component = a.Component
	menu.ParentID = a.ParentID
	menu.Order = a.Order
	menu.I18nKey = a.I18nKey
	menu.Href = a.Href
	menu.Query = a.Query
	menu.Icon = a.Icon
	menu.IconType = a.IconType
	menu.Status = a.Status
	menu.HideMenu = a.HideMenu
	menu.MultiTab = a.MultiTab
	menu.Constant = a.Constant
	menu.ActiveMenu = a.ActiveMenu
}
