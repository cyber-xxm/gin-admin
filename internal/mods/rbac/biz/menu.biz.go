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

	for _, item := range items {
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
		} else if item.RouteName != "" {
			menuItem, err = a.MenuDAL.GetByRouteNameAndParentID(ctx, item.RouteName, parentID)
		}

		if err != nil {
			return err
		}

		if menuItem != nil {
			changed := false
			if menuItem.RouteName != item.RouteName {
				menuItem.RouteName = item.RouteName
				changed = true
			}
			if menuItem.RoutePath != item.RoutePath {
				menuItem.RoutePath = item.RoutePath
				changed = true
			}
			if menuItem.MenuName != item.MenuName {
				menuItem.MenuName = item.MenuName
				changed = true
			}
			if menuItem.MenuType != item.MenuType {
				menuItem.MenuType = item.MenuType
				changed = true
			}
			if menuItem.Component != item.Component {
				menuItem.Component = item.Component
				changed = true
			}
			if menuItem.Order != item.Order {
				menuItem.Order = item.Order
				changed = true
			}
			if menuItem.I18nKey != item.I18nKey {
				menuItem.I18nKey = item.I18nKey
				changed = true
			}
			if menuItem.Href != item.Href {
				menuItem.Href = item.Href
				changed = true
			}
			if menuItem.Icon != item.Icon {
				menuItem.Icon = item.Icon
				changed = true
			}
			if menuItem.IconType != item.IconType {
				menuItem.Icon = item.Icon
				changed = true
			}
			if menuItem.Status != item.Status {
				menuItem.Status = item.Status
				changed = true
			}
			if menuItem.HideMenu != item.HideMenu {
				menuItem.HideMenu = item.HideMenu
				changed = true
			}
			if menuItem.MultiTab != item.MultiTab {
				menuItem.MultiTab = item.MultiTab
				changed = true
			}
			if menuItem.Constant != item.Constant {
				menuItem.Constant = item.Constant
				changed = true
			}
			if menuItem.KeepAlive != item.KeepAlive {
				menuItem.KeepAlive = item.KeepAlive
				changed = true
			}
			if menuItem.ActiveMenu != item.ActiveMenu {
				menuItem.ActiveMenu = item.ActiveMenu
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
			item.ParentID = parentID
			menuItem = item
			if err := a.MenuDAL.Create(ctx, item); err != nil {
				return err
			}
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
	isRoot := util.FromIsRootUser(ctx)
	if !isRoot {
		params.UserID = util.FromUserID(ctx)
	}
	result, err := a.MenuDAL.Query(ctx, params, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: schema.MenusOrderParams,
		},
	})
	if err != nil {
		return nil, err
	}

	result.Data = result.Data.ToTree()
	sort.Sort(result.Data)
	return result, nil
}

// Get the specified menu from the data access object.
func (a *Menu) Get(ctx context.Context, id string) (*schema.Menu, error) {
	menu, err := a.MenuDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if menu == nil {
		return nil, errors.NotFound("", "Menu not found")
	}

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

	if exists, err := a.MenuDAL.ExistsRouteNameByParentID(ctx, formItem.RouteName, formItem.ParentID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Menu code already exists at the same level")
	}

	formItem.FillTo(menu)

	err := a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.MenuDAL.Create(ctx, menu); err != nil {
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

	if menu.RouteName != formItem.RouteName {
		if exists, err := a.MenuDAL.ExistsRouteNameByParentID(ctx, formItem.RouteName, formItem.ParentID); err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Menu code already exists at the same level")
		}
	}

	formItem.FillTo(menu)
	return a.Trans.Exec(ctx, func(ctx context.Context) error {

		if err := a.MenuDAL.Update(ctx, menu); err != nil {
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
	if err := a.RoleMenuDAL.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	return nil
}

func (a *Menu) syncToCasbin(ctx context.Context) error {
	return a.Cache.Set(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin, fmt.Sprintf("%d", time.Now().Unix()))
}

func (a *Menu) GetAllPages() []string {
	return []string{
		"home",
		"403",
		"404",
		"405",
		"function_multi-tab",
		"function_tab",
		"exception_403",
		"exception_404",
		"exception_500",
		"multi-menu_first_child",
		"multi-menu_second_child_home",
		"manage_log",
		"manage_user",
		"manage_role",
		"manage_menu",
		"manage_user-detail",
		"about",
	}
}
