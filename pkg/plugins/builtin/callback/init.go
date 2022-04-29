package plugincallback

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/ovh/configstore"
	"github.com/ovh/utask"
	"github.com/ovh/utask/api"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/pkg/plugins"
	"github.com/wI2L/fizz"
)

const (
	configAlias               = "callback-config"
	defaultCallbackPathPrefix = "/unsecured/callback"
)

var (
	Init = NewCallbackInit()
)

type CallbackConfig struct {
	BaseURL    string `json:"base_url"`
	PathPrefix string `json:"path_prefix,omitempty"`
}

type CallbackInit struct {
	cfg CallbackConfig
}

func NewCallbackInit() *CallbackInit {
	return &CallbackInit{}
}

func (ci *CallbackInit) Init(s *plugins.Service) error {
	if err := ci.loadConfig(s.Store); err != nil {
		return fmt.Errorf("unable to load configuration: %s", err)
	}

	db.RegisterTableModel(callback{}, "callback", []string{"id"}, true)
	db.RegisterKeyRotations(RotateEncryptionKeys)

	group := api.PluginRouterGroup{
		Path:        defaultCallbackPathPrefix,
		Name:        "callback",
		Description: "Callback plugin routes.",
		Routes: []api.PluginRoute{
			{
				Path:   "/:id",
				Method: "POST",
				Infos: []fizz.OperationOption{
					fizz.ID("HandleCallback"),
					fizz.Summary("Call waiting callback"),
					fizz.Description("This action updates a waiting callback to resolves it."),
				},
				Handlers: []gin.HandlerFunc{
					tonic.Handler(HandleCallback, 200),
				},
				Maintenance: true,
			},
		},
	}
	if err := s.Server.RegisterPluginRoutes(group); err != nil {
		return err
	}

	return nil
}

func (ci *CallbackInit) Description() string {
	return `This plugin will init the callback task plugin.`
}

func (ci *CallbackInit) loadConfig(store *configstore.Store) error {
	var ret CallbackConfig
	var notFound configstore.ErrItemNotFound

	cbFilter := configstore.Filter().Store(store).Slice(configAlias).Squash()
	cbItems, err := cbFilter.GetItemList()
	if err != nil {
		return err
	}

	if cbItems.Len() > 0 {
		jsonStr, err := cbItems.Items[0].Value()
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(jsonStr), &ret); err != nil {
			return err
		}

		if ret.BaseURL == "" {
			return fmt.Errorf("\"base_url\" key not defined in %q", configAlias)
		}
	} else {
		utaskFilter := configstore.Filter().Store(store).Slice(utask.UtaskCfgSecretAlias).Squash()
		utaskItem, err := utaskFilter.GetFirstItem()
		if err != nil {
			if errors.As(err, &notFound) {
				return fmt.Errorf("configstore: get %q: no item found", configAlias)
			}
			return err
		}

		var utaskRet utask.Cfg

		jsonStr, err := utaskItem.Value()
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(jsonStr), &utaskRet); err != nil {
			return err
		}

		if utaskRet.BaseURL == "" {
			return fmt.Errorf("configstore: get %q: no item found", configAlias)
		}

		ret.BaseURL = utaskRet.BaseURL
	}

	ci.cfg.BaseURL = strings.TrimSuffix(ret.BaseURL, "/")
	ci.cfg.PathPrefix = ret.PathPrefix

	return nil
}
