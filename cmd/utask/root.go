package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	formatters "github.com/fabienm/go-logrus-formatters"
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
	"github.com/ovh/utask/engine/functions"
	functionsrunner "github.com/ovh/utask/engine/functions/runner"
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
	defaultFunctionsFolder    = "./functions"
	defaultScriptsFolder      = "./scripts"
	defaultRegion             = "default"
	defaultPort               = 8081
	defaultLogsFormat         = "text"

	envInit        = "INIT"
	envPlugins     = "PLUGINS"
	envTemplates   = "TEMPLATES"
	envFunctions   = "FUNCTIONS"
	envScripts     = "SCRIPTS"
	envRegion      = "REGION"
	envHTTPPort    = "SERVER_PORT"
	envDebug       = "DEBUG"
	envMaintenance = "MAINTENANCE_MODE"
	envLogsFormat  = "LOGS_FORMAT"

	basicAuthKey  = "basic-auth"
	groupsAuthKey = "groups-auth"
)

var (
	store  *configstore.Store
	server *api.Server
)

//nolint:errcheck
func init() {
	viper.BindEnv(envInit)
	viper.BindEnv(envPlugins)
	viper.BindEnv(envTemplates)
	viper.BindEnv(envFunctions)
	viper.BindEnv(envScripts)
	viper.BindEnv(envRegion)
	viper.BindEnv(envHTTPPort)
	viper.BindEnv(envDebug)
	viper.BindEnv(envMaintenance)
	viper.BindEnv(envLogsFormat)

	flags := rootCmd.Flags()

	flags.StringVar(&utask.FInitializersFolder, "init-path", defaultInitializersFolder, "Initializer folder absolute path")
	flags.StringVar(&utask.FPluginFolder, "plugins-path", defaultPluginFolder, "Plugins folder absolute path")
	flags.StringVar(&utask.FTemplatesFolder, "templates-path", defaultTemplatesFolder, "Templates folder absolute path")
	flags.StringVar(&utask.FFunctionsFolder, "functions-path", defaultFunctionsFolder, "Functions folder absolute path")
	flags.StringVar(&utask.FScriptsFolder, "scripts-path", defaultScriptsFolder, "Scripts folder absolute path")
	flags.StringVar(&utask.FRegion, "region", defaultRegion, "Region in which instance is located")
	flags.UintVar(&utask.FPort, "http-port", defaultPort, "HTTP port to expose")
	flags.BoolVar(&utask.FDebug, "debug", false, "Run engine in debug mode")
	flags.BoolVar(&utask.FMaintenanceMode, "maintenance-mode", false, "Switch API to maintenance mode")
	flags.StringVar(&utask.FLogsFormat, "logs-format", defaultLogsFormat, "Format of the logs (text or gelf)")

	viper.BindPFlag(envInit, rootCmd.Flags().Lookup("init-path"))
	viper.BindPFlag(envPlugins, rootCmd.Flags().Lookup("plugins-path"))
	viper.BindPFlag(envTemplates, rootCmd.Flags().Lookup("templates-path"))
	viper.BindPFlag(envFunctions, rootCmd.Flags().Lookup("functions-path"))
	viper.BindPFlag(envScripts, rootCmd.Flags().Lookup("scripts-path"))
	viper.BindPFlag(envRegion, rootCmd.Flags().Lookup("region"))
	viper.BindPFlag(envHTTPPort, rootCmd.Flags().Lookup("http-port"))
	viper.BindPFlag(envDebug, rootCmd.Flags().Lookup("debug"))
	viper.BindPFlag(envMaintenance, rootCmd.Flags().Lookup("maintenance-mode"))
	viper.BindPFlag(envLogsFormat, rootCmd.Flags().Lookup("logs-format"))
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
		utask.FLogsFormat = viper.GetString(envLogsFormat)

		// Logger.
		var formatter log.Formatter
		switch utask.FLogsFormat {
		case "text":
			textFormatter := new(log.TextFormatter)
			textFormatter.TimestampFormat = time.RFC3339
			textFormatter.FullTimestamp = true
			formatter = textFormatter
		case "gelf":
			hostname, _ := os.Hostname()
			formatter = formatters.NewGelf(hostname)
		}
		log.SetOutput(os.Stdout)
		log.SetFormatter(formatter)

		store = configstore.DefaultStore
		store.InitFromEnvironment()

		defaultAuthHandler, err := basicAuthHandler(store)
		if err != nil {
			return err
		}

		server = api.NewServer()
		server.WithGroupAuth(defaultAuthHandler)

		for _, err := range []error{
			// register builtin executors
			builtin.Register(),
			// run custom initialization code built as *.so plugins
			plugins.InitializersFromFolder(utask.FInitializersFolder, &plugins.Service{Store: store, Server: server}),
			// load custom executors built as *.so plugins
			plugins.ExecutorsFromFolder(utask.FPluginFolder),
			// load the functions
			functions.LoadFromDir(utask.FFunctionsFolder),
			// register functions as runners
			functionsrunner.Init(),
			// init authorization module (admin username list)
			auth.Init(store),
			// init notify module
			notify.Init(store),
		} {
			if err != nil {
				return err
			}
		}

		cfg, err := utask.Config(store)
		if err != nil {
			return err
		}
		server.SetDashboardPathPrefix(cfg.DashboardPathPrefix)
		server.SetDashboardAPIPathPrefix(cfg.DashboardAPIPathPrefix)
		server.SetEditorPathPrefix(cfg.EditorPathPrefix)
		server.SetDashboardSentryDSN(cfg.DashboardSentryDSN)
		server.SetMaxBodyBytes(cfg.ServerOptions.MaxBodyBytes)

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
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			// Stop collectors
			cancel()
			log.Info("Exiting...")

			gracePeriodWaitGroup := make(chan struct{})
			go func() {
				wg.Wait()
				close(gracePeriodWaitGroup)
			}()

			t := time.NewTicker(5 * time.Second)
			// Running steps have 3 seconds to stop running after context cancelation
			// Grace period of 2 seconds (3+2=5) to commit still-running resolutions
			select {
			case <-gracePeriodWaitGroup:
				// all important goroutines exited successfully, bye-bye!
			case <-t.C:
				// game over, exiting before everyone said bye :(
				log.Warn("5 seconds timeout for exiting expired")
			}

			log.Info("Bye!")
		}()

		if err := engine.Init(ctx, &wg, store); err != nil {
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

// basicAuthHandler handles user and groups authentication.
//
// How does it work?
// If a map of user passwords is found in configstore, use them as basic auth
// check on incoming requests. The groups of the user are determined from the
// configuration in configstore. If nothing is found, the zero value of a slice
// is returned (i.e. `nil`).
//
// It is a default implementation which can be overridden by Server.WithAuth or
// Server.WithGroupAuth functions in api package.
func basicAuthHandler(store *configstore.Store) (func(*http.Request) (string, []string, error), error) {
	groupsMap := map[string][]string{}
	groupsAuthStr, err := configstore.Filter().Slice(groupsAuthKey).Squash().Store(store).MustGetFirstItem().Value()
	if err == nil {
		if err = json.Unmarshal([]byte(groupsAuthStr), &groupsMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal utask configuration: %s", err)
		}
	}

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
		return func(r *http.Request) (string, []string, error) {
			authHeader := r.Header.Get("Authorization")
			user, found := authMap[authHeader]
			if !found {
				return "", nil, errors.Unauthorizedf("User not found")
			}
			return user, groupsMap[user], nil
		}, nil
	}
	// fallback to expecting a username in x-remote-user header
	return func(r *http.Request) (string, []string, error) {
		user := r.Header.Get("x-remote-user")
		return user, groupsMap[user], nil
	}, nil
}
