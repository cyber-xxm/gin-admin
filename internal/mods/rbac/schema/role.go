package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

const (
	RoleStatusEnabled  = 1 // Enabled
	RoleStatusDisabled = 2 // Disabled

	RoleResultTypeSelect = "select" // Select
)

// Role management for RBAC
type Role struct {
	ID          string    `json:"id" gorm:"size:20;primarykey;"` // Unique ID
	Name        string    `json:"name" gorm:"size:128;index"`    // Display name of role
	Description string    `json:"description" gorm:"size:1024"`  // Details about role
	Status      int       `json:"status"`                        // Status of role (disabled, enabled)
	CreatedAt   time.Time `json:"created_at" gorm:"index;"`      // Create time
	UpdatedAt   time.Time `json:"updated_at" gorm:"index;"`      // Update time
	Menus       RoleMenus `json:"menus" gorm:"-"`                // Role menu list
}

func (a *Role) TableName() string {
	return config.C.FormatTableName("role")
}

// Defining the query parameters for the `Role` struct.
type RoleQueryParam struct {
	util.PaginationParam
	LikeName    string     `form:"name"`       // Display name of role
	Status      int        `form:"status"`     // Status of role (disabled, enabled)
	ResultType  string     `form:"resultType"` // Result type (options: select)
	GtUpdatedAt *time.Time `form:"-"`          // Update time is greater than
}

// Defining the query options for the `Role` struct.
type RoleQueryOptions struct {
	util.QueryOptions
}

// Defining the query result for the `Role` struct.
type RoleQueryResult struct {
	Data       Roles
	PageResult *util.PaginationResult
}

// Defining the slice of `Role` struct.
type Roles []*Role

// Defining the data structure for creating a `Role` struct.
type RoleForm struct {
	Name        string    `json:"name" binding:"required,max=128"`     // Display name of role
	Description string    `json:"description"`                         // Details about role
	Status      int       `json:"status" binding:"required,oneof=1 2"` // Status of role (enabled, disabled)
	Menus       RoleMenus `json:"menus"`                               // Role menu list
}

// A validation function for the `RoleForm` struct.
func (a *RoleForm) Validate() error {
	return nil
}

func (a *RoleForm) FillTo(role *Role) error {
	role.Name = a.Name
	role.Description = a.Description
	role.Status = a.Status
	return nil
}
