package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/configstore"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ovh/utask"
	"github.com/ovh/utask/api"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/engine"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	notify "github.com/ovh/utask/pkg/notify/init"
	"github.com/ovh/utask/pkg/plugins"
	"github.com/ovh/utask/pkg/plugins/builtin"
)

const (
	defaultInitializersFolder = "./init"
	defaultPluginFolder       = "./plugins"
	defaultTemplatesFolder    = "./templates"
	defaultScriptsFolder      = "./scripts"
	defaultRegion             = "default"
	defaultPort               = 8081

	envInit        = "INIT"
	envPlugins     = "PLUGINS"
	envTemplates   = "TEMPLATES"
	envScripts     = "SCRIPTS"
	envRegion      = "REGION"
	envHTTPPort    = "SERVER_PORT"
	envDebug       = "DEBUG"
	envMaintenance = "MAINTENANCE_MODE"

	basicAuthKey = "basic-auth"
)

var (
	store  *configstore.Store
	server *api.Server
)

func init() {
	viper.BindEnv(envInit)
	viper.BindEnv(envPlugins)
	viper.BindEnv(envTemplates)
	viper.BindEnv(envScripts)
	viper.BindEnv(envRegion)
	viper.BindEnv(envHTTPPort)
	viper.BindEnv(envDebug)
	viper.BindEnv(envMaintenance)

	// Logger
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = time.RFC3339
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)

	flags := rootCmd.Flags()

	flags.StringVar(&utask.FInitializersFolder, "init-path", defaultInitializersFolder, "Initializer folder absolute path")
	flags.StringVar(&utask.FPluginFolder, "plugins-path", defaultPluginFolder, "Plugins folder absolute path")
	flags.StringVar(&utask.FTemplatesFolder, "templates-path", defaultTemplatesFolder, "Templates folder absolute path")
	flags.StringVar(&utask.FScriptsFolder, "scripts-path", defaultScriptsFolder, "Scripts folder absolute path")
	flags.StringVar(&utask.FRegion, "region", defaultRegion, "Region in which instance is located")
	flags.UintVar(&utask.FPort, "http-port", defaultPort, "HTTP port to expose")
	flags.BoolVar(&utask.FDebug, "debug", false, "Run engine in debug mode")
	flags.BoolVar(&utask.FMaintenanceMode, "maintenance-mode", false, "Switch API to maintenance mode")

	viper.BindPFlag(envInit, rootCmd.Flags().Lookup("init-path"))
	viper.BindPFlag(envPlugins, rootCmd.Flags().Lookup("plugins-path"))
	viper.BindPFlag(envTemplates, rootCmd.Flags().Lookup("templates-path"))
	viper.BindPFlag(envScripts, rootCmd.Flags().Lookup("scripts-path"))
	viper.BindPFlag(envRegion, rootCmd.Flags().Lookup("region"))
	viper.BindPFlag(envHTTPPort, rootCmd.Flags().Lookup("http-port"))
	viper.BindPFlag(envDebug, rootCmd.Flags().Lookup("debug"))
	viper.BindPFlag(envMaintenance, rootCmd.Flags().Lookup("maintenance-mode"))
}

var rootCmd = &cobra.Command{
	Short: "µTask, the extensible automation engine\n\n",
	Long: "µTask is an extensible automation engine. It performs tasks\n" +
		"that follow statically declared scenarios, a sequence of actions\n" +
		"whose results build upon each other. It exposes an HTTP API to handle\n" +
		"the creation and resolution of tasks, oversight of intermediate states\n" +
		"and manual intervention if needed.\n",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		utask.FInitializersFolder = viper.GetString(envInit)
		utask.FPluginFolder = viper.GetString(envPlugins)
		utask.FTemplatesFolder = viper.GetString(envTemplates)
		utask.FScriptsFolder = viper.GetString(envScripts)
		utask.FRegion = viper.GetString(envRegion)
		utask.FPort = viper.GetUint(envHTTPPort)
		utask.FDebug = viper.GetBool(envDebug)
		utask.FMaintenanceMode = viper.GetBool(envMaintenance)

		store = configstore.NewStore()
		store.InitFromEnvironment()

		defaultAuthHandler, err := basicAuthHandler(store)
		if err != nil {
			return err
		}

		server = api.NewServer()
		server.WithAuth(defaultAuthHandler)

		for _, err := range []error{
			// register builtin executors
			builtin.Register(),
			// run custom initialization code built as *.so plugins
			plugins.InitializersFromFolder(utask.FInitializersFolder, &plugins.Service{Store: store, Server: server}),
			// load custom executors built as *.so plugins
			plugins.ExecutorsFromFolder(utask.FPluginFolder),
			// init authorization module (admin username list)
			auth.Init(store),
			// init notify module
			notify.Init(store),
		} {
			if err != nil {
				return err
			}
		}

		if utask.FDebug {
			log.SetLevel(log.DebugLevel)
		}

		if utask.FPort > 65535 || utask.FPort == 0 {
			return errors.New("Incorrect HTTP port range")
		}

		return db.Init(store)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		dbp, err := zesty.NewDBProvider(utask.DBName)
		if err != nil {
			return err
		}

		if err := tasktemplate.LoadFromDir(dbp, utask.FTemplatesFolder); err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			// Stop collectors
			cancel()

			// Grace period to commit still-running resolutions
			time.Sleep(time.Second)

			log.Info("Exiting...")
		}()

		if err := engine.Init(ctx, store); err != nil {
			return err
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

// if a map of user passwords is found in configstore
// use them as basic auth check on incoming requests
func basicAuthHandler(store *configstore.Store) (func(*http.Request) (string, error), error) {
	authMap := map[string]string{}
	basicAuthStr, err := configstore.Filter().Slice(basicAuthKey).Squash().Store(store).MustGetFirstItem().Value()
	if err == nil {
		userPasswords := map[string]string{}
		if err := json.Unmarshal([]byte(basicAuthStr), &userPasswords); err != nil {
			return nil, fmt.Errorf("failed to unmarshal utask configuration: %s", err)
		}
		for user, pass := range userPasswords {
			header := "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
			authMap[header] = user
		}
	}
	if len(authMap) > 0 {
		return func(r *http.Request) (string, error) {
			authHeader := r.Header.Get("Authorization")
			user, found := authMap[authHeader]
			if !found {
				return "", errors.Unauthorizedf("User not found")
			}
			return user, nil
		}, nil
	}
	// fallback to expecting a username in x-remote-user header
	return func(r *http.Request) (string, error) {
		return r.Header.Get("x-remote-user"), nil
	}, nil
}
