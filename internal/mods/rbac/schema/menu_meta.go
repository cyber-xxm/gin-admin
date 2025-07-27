package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// Menu resource management for RBAC
type MenuMeta struct {
	ID         string    `json:"id" gorm:"size:20;primarykey"` // Unique ID
	MenuID     string    `json:"menu_id" gorm:"size:20;index"` // From Menu.ID
	Title      string    `json:"title" gorm:"size:100;"`       // HTTP method
	I18nKey    string    `json:"i18n_key" gorm:"size:255;"`    // I18nKey
	Icon       string    `json:"icon" gorm:"size:255;"`        // Icon
	Order      int       `json:"order" gorm:"index;"`          // Order
	HideInMenu bool      `json:"hide_in_menu" gorm:"index;"`   // HideInMenu
	MultiTab   bool      `json:"multi_tab" gorm:"index;"`      // MultiTab
	ActiveMenu string    `json:"active_menu" gorm:"index;"`    // ActiveMenu
	CreatedAt  time.Time `json:"created_at" gorm:"index;"`     // Create time
	UpdatedAt  time.Time `json:"updated_at" gorm:"index;"`     // Update time
}

func (a *MenuMeta) TableName() string {
	return config.C.FormatTableName("menu_meta")
}

// Defining the query parameters for the `MenuMeta` struct.
type MenuMetaQueryParam struct {
	util.PaginationParam
	MenuID  string   `form:"-"` // From Menu.ID
	MenuIDs []string `form:"-"` // From Menu.ID
}

// Defining the query options for the `MenuMeta` struct.
type MenuMetaQueryOptions struct {
	util.QueryOptions
}

// Defining the query result for the `MenuMeta` struct.
type MenuMetaQueryResult struct {
	Data       *MenuMeta
	PageResult *util.PaginationResult
}

// Defining the slice of `MenuMeta` struct.
type MenuMetas []*MenuMeta

// Defining the data structure for creating a `MenuMeta` struct.
type MenuMetaForm struct {
}

// A validation function for the `MenuMetaForm` struct.
func (a *MenuMetaForm) Validate() error {
	return nil
}

func (a *MenuMetaForm) FillTo(menuMeta *MenuMeta) error {
	return nil
}
