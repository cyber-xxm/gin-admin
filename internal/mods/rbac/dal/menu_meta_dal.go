package dal

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// Get menu resource storage instance
func GetMenuMetaDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.MenuMeta))
}

// Menu resource management for RBAC
type MenuMeta struct {
	DB *gorm.DB
}

// Query menu resources from the database based on the provided parameters and options.
func (a *MenuMeta) Query(ctx context.Context, params schema.MenuMetaQueryParam, opts ...schema.MenuMetaQueryOptions) (*schema.MenuMetaQueryResult, error) {
	var opt schema.MenuMetaQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetMenuMetaDB(ctx, a.DB)
	if v := params.MenuID; len(v) > 0 {
		db = db.Where("menu_id = ?", v)
	}
	if v := params.MenuIDs; len(v) > 0 {
		db = db.Where("menu_id IN ?", v)
	}

	var list *schema.MenuMeta
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &schema.MenuMetaQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get the specified menu resource from the database.
func (a *MenuMeta) Get(ctx context.Context, id string, opts ...schema.MenuMetaQueryOptions) (*schema.MenuMeta, error) {
	var opt schema.MenuMetaQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.MenuMeta)
	ok, err := util.FindOne(ctx, GetMenuMetaDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exist checks if the specified menu resource exists in the database.
func (a *MenuMeta) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuMetaDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// ExistsMethodPathByMenuID checks if the specified menu resource exists in the database.
func (a *MenuMeta) ExistsMethodPathByMenuID(ctx context.Context, method, path, menuID string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuMetaDB(ctx, a.DB).Where("method=? AND path=? AND menu_id=?", method, path, menuID))
	return ok, errors.WithStack(err)
}

// Create a new menu resource.
func (a *MenuMeta) Create(ctx context.Context, item *schema.MenuMeta) error {
	result := GetMenuMetaDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update the specified menu resource in the database.
func (a *MenuMeta) Update(ctx context.Context, item *schema.MenuMeta) error {
	result := GetMenuMetaDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete the specified menu resource from the database.
func (a *MenuMeta) Delete(ctx context.Context, id string) error {
	result := GetMenuMetaDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.MenuMeta))
	return errors.WithStack(result.Error)
}

// Deletes the menu resource by menu id.
func (a *MenuMeta) DeleteByMenuID(ctx context.Context, menuID string) error {
	result := GetMenuMetaDB(ctx, a.DB).Where("menu_id=?", menuID).Delete(new(schema.MenuMeta))
	return errors.WithStack(result.Error)
}
