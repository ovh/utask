package builtin

import (
	"github.com/ovh/utask/engine/step"
	pluginapiovh "github.com/ovh/utask/pkg/plugins/builtin/apiovh"
	pluginecho "github.com/ovh/utask/pkg/plugins/builtin/echo"
	pluginemail "github.com/ovh/utask/pkg/plugins/builtin/email"
	pluginhttp "github.com/ovh/utask/pkg/plugins/builtin/http"
	pluginnotify "github.com/ovh/utask/pkg/plugins/builtin/notify"
	pluginssh "github.com/ovh/utask/pkg/plugins/builtin/ssh"
	pluginsubtask "github.com/ovh/utask/pkg/plugins/builtin/subtask"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// Register takes all builtin plugins and registers them as step executors
func Register() error {
	for _, p := range []taskplugin.PluginExecutor{
		pluginssh.Plugin,
		pluginhttp.Plugin,
		pluginapiovh.Plugin,
		pluginsubtask.Plugin,
		pluginnotify.Plugin,
		pluginecho.Plugin,
		pluginemail.Plugin,
	} {
		if err := step.RegisterRunner(p.PluginName(), p); err != nil {
			return err
		}
	}
	return nil
}
