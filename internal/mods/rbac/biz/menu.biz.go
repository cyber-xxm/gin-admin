package biz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/cachex"
	"github.com/LyricTian/gin-admin/v10/pkg/encoding/json"
	"github.com/LyricTian/gin-admin/v10/pkg/encoding/yaml"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"go.uber.org/zap"
)

// Menu management for RBAC
type Menu struct {
	Cache       cachex.Cacher
	Trans       *util.Trans
	MenuDAL     *dal.Menu
	MenuMetaDAL *dal.MenuMeta
	RoleMenuDAL *dal.RoleMenu
}

func (a *Menu) InitFromFile(ctx context.Context, menuFile string) error {
	f, err := os.ReadFile(menuFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logging.Context(ctx).Warn("Menu data file not found, skip init menu data from file", zap.String("file", menuFile))
			return nil
		}
		return err
	}

	var menus schema.Menus
	if ext := filepath.Ext(menuFile); ext == ".json" {
		if err := json.Unmarshal(f, &menus); err != nil {
			return errors.Wrapf(err, "Unmarshal JSON file '%s' failed", menuFile)
		}
	} else if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(f, &menus); err != nil {
			return errors.Wrapf(err, "Unmarshal YAML file '%s' failed", menuFile)
		}
	} else {
		return errors.Errorf("Unsupported file type '%s'", ext)
	}

	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		return a.createInBatchByParent(ctx, menus, nil)
	})
}

func (a *Menu) createInBatchByParent(ctx context.Context, items schema.Menus, parent *schema.Menu) error {
	total := len(items)

	for i, item := range items {
		var parentID string
		if parent != nil {
			parentID = parent.ID
		}

		var (
			menuItem *schema.Menu
			err      error
		)

		if item.ID != "" {
			menuItem, err = a.MenuDAL.Get(ctx, item.ID)
		} else if item.Name != "" {
			menuItem, err = a.MenuDAL.GetByNameAndParentID(ctx, item.Name, parentID)
		}

		if err != nil {
			return err
		}

		if menuItem != nil {
			changed := false
			if menuItem.Name != item.Name {
				menuItem.Name = item.Name
				changed = true
			}
			if menuItem.Description != item.Description {
				menuItem.Description = item.Description
				changed = true
			}
			if menuItem.Path != item.Path {
				menuItem.Path = item.Path
				changed = true
			}
			if menuItem.Status != item.Status {
				menuItem.Status = item.Status
				changed = true
			}
			if changed {
				menuItem.UpdatedAt = time.Now()
				if err := a.MenuDAL.Update(ctx, menuItem); err != nil {
					return err
				}
			}
		} else {
			if item.ID == "" {
				item.ID = util.NewXID()
			}
			if item.Meta.Order == 0 {
				item.Meta.Order = total - i
			}
			item.ParentID = parentID
			menuItem = item
			if err := a.MenuDAL.Create(ctx, item); err != nil {
				return err
			}
		}

		if item.Meta.ID != "" {
			exists, err := a.MenuMetaDAL.Exists(ctx, item.Meta.ID)
			if err != nil {
				return err
			} else if exists {
				continue
			}
		}

		if item.Meta.ID == "" {
			item.Meta.ID = util.NewXID()
		}
		item.Meta.MenuID = menuItem.ID
		if err := a.MenuMetaDAL.Create(ctx, item.Meta); err != nil {
			return err
		}

		if item.Children != nil {
			if err := a.createInBatchByParent(ctx, *item.Children, menuItem); err != nil {
				return err
			}
		}
	}
	return nil
}

// Query menus from the data access object based on the provided parameters and options.
func (a *Menu) Query(ctx context.Context, params schema.MenuQueryParam) (*schema.MenuQueryResult, error) {
	params.Pagination = false

	result, err := a.MenuDAL.Query(ctx, params, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: schema.MenusOrderParams,
		},
	})
	if err != nil {
		return nil, err
	}

	if params.LikeName != "" {
		result.Data, err = a.appendChildren(ctx, result.Data)
		if err != nil {
			return nil, err
		}
	}

	for i, item := range result.Data {
		resResult, err := a.MenuMetaDAL.Query(ctx, schema.MenuMetaQueryParam{
			MenuID: item.ID,
		})
		if err != nil {
			return nil, err
		}
		result.Data[i].Meta = resResult.Data
	}

	result.Data = result.Data.ToTree()
	return result, nil
}

func (a *Menu) appendChildren(ctx context.Context, data schema.Menus) (schema.Menus, error) {
	if len(data) == 0 {
		return data, nil
	}

	existsInData := func(id string) bool {
		for _, item := range data {
			if item.ID == id {
				return true
			}
		}
		return false
	}

	if parentIDs := data.SplitParentIDs(); len(parentIDs) > 0 {
		parentResult, err := a.MenuDAL.Query(ctx, schema.MenuQueryParam{
			InIDs: parentIDs,
		})
		if err != nil {
			return nil, err
		}
		for _, p := range parentResult.Data {
			if existsInData(p.ID) {
				continue
			}
			data = append(data, p)
		}
	}
	sort.Sort(data)

	return data, nil
}

// Get the specified menu from the data access object.
func (a *Menu) Get(ctx context.Context, id string) (*schema.Menu, error) {
	menu, err := a.MenuDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if menu == nil {
		return nil, errors.NotFound("", "Menu not found")
	}

	menuResResult, err := a.MenuMetaDAL.Query(ctx, schema.MenuMetaQueryParam{
		MenuID: menu.ID,
	})
	if err != nil {
		return nil, err
	}
	menu.Meta = menuResResult.Data

	return menu, nil
}

// Create a new menu in the data access object.
func (a *Menu) Create(ctx context.Context, formItem *schema.MenuForm) (*schema.Menu, error) {
	if config.C.General.DenyOperateMenu {
		return nil, errors.BadRequest("", "Menu creation is not allowed")
	}

	menu := &schema.Menu{
		ID:        util.NewXID(),
		CreatedAt: time.Now(),
	}

	if parentID := formItem.ParentID; parentID != "" {
		parent, err := a.MenuDAL.Get(ctx, parentID)
		if err != nil {
			return nil, err
		} else if parent == nil {
			return nil, errors.NotFound("", "Parent not found")
		}
	}

	if exists, err := a.MenuDAL.ExistsCodeByParentID(ctx, formItem.Name, formItem.ParentID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Menu code already exists at the same level")
	}

	if err := formItem.FillTo(menu); err != nil {
		return nil, err
	}

	err := a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.MenuDAL.Create(ctx, menu); err != nil {
			return err
		}

		menu.Meta.ID = util.NewXID()
		menu.Meta.MenuID = menu.ID
		menu.Meta.CreatedAt = time.Now()
		if err := a.MenuMetaDAL.Create(ctx, menu.Meta); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return menu, nil
}

// Update the specified menu in the data access object.
func (a *Menu) Update(ctx context.Context, id string, formItem *schema.MenuForm) error {
	if config.C.General.DenyOperateMenu {
		return errors.BadRequest("", "Menu update is not allowed")
	}

	menu, err := a.MenuDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if menu == nil {
		return errors.NotFound("", "Menu not found")
	}

	if menu.Name != formItem.Name {
		if exists, err := a.MenuDAL.ExistsCodeByParentID(ctx, formItem.Name, formItem.ParentID); err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Menu code already exists at the same level")
		}
	}

	if err := formItem.FillTo(menu); err != nil {
		return err
	}

	return a.Trans.Exec(ctx, func(ctx context.Context) error {

		if err := a.MenuDAL.Update(ctx, menu); err != nil {
			return err
		}

		if err := a.MenuMetaDAL.DeleteByMenuID(ctx, id); err != nil {
			return err
		}
		if formItem.Meta.ID == "" {
			formItem.Meta.ID = util.NewXID()
		}
		formItem.Meta.MenuID = id
		if formItem.Meta.CreatedAt.IsZero() {
			formItem.Meta.CreatedAt = time.Now()
		}
		formItem.Meta.UpdatedAt = time.Now()
		if err := a.MenuMetaDAL.Create(ctx, formItem.Meta); err != nil {
			return err
		}

		return a.syncToCasbin(ctx)
	})
}

// Delete the specified menu from the data access object.
func (a *Menu) Delete(ctx context.Context, id string) error {
	if config.C.General.DenyOperateMenu {
		return errors.BadRequest("", "Menu deletion is not allowed")
	}

	menu, err := a.MenuDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if menu == nil {
		return errors.NotFound("", "Menu not found")
	}

	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.delete(ctx, id); err != nil {
			return err
		}
		return a.syncToCasbin(ctx)
	})
}

func (a *Menu) delete(ctx context.Context, id string) error {
	if err := a.MenuDAL.Delete(ctx, id); err != nil {
		return err
	}
	if err := a.MenuMetaDAL.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	if err := a.RoleMenuDAL.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	return nil
}

func (a *Menu) syncToCasbin(ctx context.Context) error {
	return a.Cache.Set(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin, fmt.Sprintf("%d", time.Now().Unix()))
}
