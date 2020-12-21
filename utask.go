package utask

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/ovh/configstore"
)

var (
	// Version holds the tag of current µTask release
	Version string
	// Commit is the current git commit hash
	Commit string
	// App name (from configuration)
	App string
	// InstanceID identifies this running instance of µTask, as registered in DB
	InstanceID uint64

	// FInitializersFolder is the path to a folder containing
	// .so plugins for µTask initialization
	FInitializersFolder string
	// FPluginFolder is the path to a folder containing
	// .so plugins to be registered as step action executors
	FPluginFolder string
	// FTemplatesFolder is the path to a folder containing
	// .yaml templates for tasks
	FTemplatesFolder string
	// FFunctionsFolder is the path to a folder containing
	// functions files used by script plugin
	FFunctionsFolder string
	// FScriptsFolder is the path to a folder containing
	// scripts files used by script plugin
	FScriptsFolder string
	// FRegion is the region in which this instance of µTask is running
	FRegion string
	// FPort is the port on which the http server listens
	FPort uint
	// FDebug is a flag to toggle debug log
	FDebug bool
	// FMaintenanceMode is a flag to prevent all write operations on the API,
	// except for admin actions (key rotation)
	FMaintenanceMode bool
)

// AppName returns the name of the application (from config)
func AppName() string { return App }

const (
	// DBName is the name of µTask DB, as registered on zesty
	DBName = "uservice_task"

	// MaxPageSize is the upper limit for the number of elements returned in a single page
	MaxPageSize = 10000
	// DefaultPageSize is the default number of elements returned in a single page
	DefaultPageSize = 1000
	// MinPageSize is the lower limit for the number of elements returned in a single page
	MinPageSize = 10

	// DefaultRetryMax is the default number of retries allowed for a task's execution
	DefaultRetryMax = 100

	// defaultInstanceCollectorWaitDuration is the default duration between two crashed tasks being resolved
	defaultInstanceCollectorWaitDuration = time.Second
	// defaultMaxConcurrentExecutions is the default maximum concurrent task executions in the instance
	defaultMaxConcurrentExecutions = 100
	// defaultMaxConcurrentExecutionsFromCrashed is the default maximum concurrent crashed task executions in the instance
	defaultMaxConcurrentExecutionsFromCrashed = 20

	// MaxTextSizeLong is the maximum number of characters accepted in a text-type field
	MaxTextSizeLong = 100000 // ~100 kB
	// MaxTextSize is the maximum number of characters accepted in a simple string field
	MaxTextSize = 1000 // ~1 kB
	// MinTextSize is the minimum number of characters accepted in any string-type field
	MinTextSize = 3

	defaultResourceAcquireTimeout = time.Minute

	// This is the key used in Values for a step to refer to itself
	This = "this"

	// UtaskCfgSecretAlias is the key for the config item containing global configuration data
	UtaskCfgSecretAlias = "utask-cfg"
)

// Cfg holds global configuration data
type Cfg struct {
	ApplicationName                            string                   `json:"application_name"`
	AdminUsernames                             []string                 `json:"admin_usernames"`
	CompletedTaskExpiration                    string                   `json:"completed_task_expiration"`
	NotifyConfig                               map[string]NotifyBackend `json:"notify_config"`
	NotifyActions                              NotifyActions            `json:"notify_actions"`
	DatabaseConfig                             *DatabaseConfig          `json:"database_config"`
	ConcealedSecrets                           []string                 `json:"concealed_secrets"`
	ResourceLimits                             map[string]uint          `json:"resource_limits"`
	ResourceAcquireTimeout                     string                   `json:"resource_acquire_timeout"`
	resourceAcquireTimeoutDuration             time.Duration            `json:"-"`
	MaxConcurrentExecutions                    *int                     `json:"max_concurrent_executions"`
	MaxConcurrentExecutionsFromCrashed         *int                     `json:"max_concurrent_executions_from_crashed"`
	MaxConcurrentExecutionsFromCrashedComputed int                      `json:"-"`
	DelayBetweenCrashedTasksResolution         string                   `json:"delay_between_crashed_tasks_resolution"`
	InstanceCollectorWaitDuration              time.Duration            `json:"-"`
	DashboardPathPrefix                        string                   `json:"dashboard_path_prefix"`
	DashboardAPIPathPrefix                     string                   `json:"dashboard_api_path_prefix"`
	DashboardSentryDSN                         string                   `json:"dashboard_sentry_dsn"`
	EditorPathPrefix                           string                   `json:"editor_path_prefix"`
	ServerOptions                              ServerOpt                `json:"server_options"`

	resourceSemaphores map[string]*semaphore.Weighted
	executionSemaphore *semaphore.Weighted
	deadResources      map[string]struct{}
}

// ServerOpt holds the configuration for the http server
type ServerOpt struct {
	MaxBodyBytes int64 `json:"max_body_bytes"`
}

// NotifyBackend holds configuration for instantiating a notify client
type NotifyBackend struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

// NotifyBackendTat holds configuration for instantiating a Tat notify client
type NotifyBackendTat struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
	Topic    string `json:"topic"`
}

// NotifyBackendSlack holds configuration for instantiating a Slack notify client
type NotifyBackendSlack struct {
	WebhookURL string `json:"webhook_url"`
}

// NotifyBackendWebhook holds configuration for instantiating a Webhook notify client
type NotifyBackendWebhook struct {
	WebhookURL string            `json:"webhook_url"`
	Username   string            `json:"username"`
	Password   string            `json:"password"`
	Headers    map[string]string `json:"headers"`
}

// NotifyActions holds configuration of each actions
// By default all the actions are enabled /w any config name registered
type NotifyActions struct {
	TaskStateAction NotifyActionsParameters `json:"task_state_action,omitempty"`
}

// NotifyActionsParameters holds configuration needed to define each Notify actions
// If NotifyBackends is empty, the default is any
type NotifyActionsParameters struct {
	Disabled       bool     `json:"disabled"`
	NotifyBackends []string `json:"notify_backends"`
}

// DatabaseConfig holds configuration to fine-tune DB connection
type DatabaseConfig struct {
	MaxOpenConns    *int   `json:"max_open_conns"`
	MaxIdleConns    *int   `json:"max_idle_conns"`
	ConnMaxLifetime *int   `json:"conn_max_lifetime"`
	ConfigName      string `json:"config_name"`
}

func (c *Cfg) buildLimits() {
	c.resourceSemaphores = make(map[string]*semaphore.Weighted)
	c.deadResources = make(map[string]struct{})

	for k, v := range c.ResourceLimits {
		if v <= 0 {
			c.deadResources[k] = struct{}{}
		} else {
			c.resourceSemaphores[k] = semaphore.NewWeighted(int64(v))
		}
	}

	if maxConcurrentExecutions := c.getMaxConcurrentExecutions(); maxConcurrentExecutions >= 0 {
		c.executionSemaphore = semaphore.NewWeighted(int64(maxConcurrentExecutions))
	}
}

func (c *Cfg) getMaxConcurrentExecutions() int {
	if c.MaxConcurrentExecutions != nil {
		return *c.MaxConcurrentExecutions
	}
	return defaultMaxConcurrentExecutions
}

var (
	// ErrDeadResource is returned when a resource will never be available, as the max concurrent execution is set to 0, and there is no reason to wait
	ErrDeadResource = errors.New("resource is not available, as configured with 0 concurrent execution")
	// ErrFailedAcquireResource is returned when tried to acquire a resource, but the resource is not available
	ErrFailedAcquireResource = errors.New("failed to acquire the requested resource")
)

// AcquireResource takes a semaphore slot for a named resource
// limiting the amount of concurrent actions runnable on said resource
func AcquireResource(ctx context.Context, name string) error {
	if global == nil {
		return nil
	}
	if _, ok := global.deadResources[name]; ok {
		return ErrDeadResource
	}

	s := global.resourceSemaphores[name]
	if s == nil {
		return nil
	}

	semaphoreCtx := ctx
	if global.resourceAcquireTimeoutDuration != 0 {
		ctx, cancelFunc := context.WithTimeout(ctx, global.resourceAcquireTimeoutDuration)
		defer cancelFunc()
		semaphoreCtx = ctx
	}
	return s.Acquire(semaphoreCtx, 1)
}

// TryAcquireResource takes a semaphore slot for a named resource
// limiting the amount of concurrent actions runnable on said resource
func TryAcquireResource(name string) error {
	if global == nil {
		return nil
	}
	if _, ok := global.deadResources[name]; ok {
		return ErrDeadResource
	}

	s := global.resourceSemaphores[name]
	if s == nil {
		return nil
	}

	if s.TryAcquire(1) {
		return nil
	}

	return ErrFailedAcquireResource
}

// AcquireResources is an helper to call AcquireResource with an array
// If failed to acquire a resource, because context is in error, already
// acquired resources will be released, and error will be returned.
func AcquireResources(ctx context.Context, names []string) error {
	acquiredList := []string{}
	var globalerr error
	for _, name := range names {
		if err := AcquireResource(ctx, name); err != nil {
			globalerr = err
			break
		}
		acquiredList = append(acquiredList, name)
	}
	if globalerr != nil {
		for _, name := range names {
			ReleaseResource(name)
		}
	}
	return globalerr
}

// ReleaseResource frees up a semaphore slot for a named resource
func ReleaseResource(name string) {
	if global == nil {
		return
	}
	s := global.resourceSemaphores[name]
	if s == nil {
		return
	}
	s.Release(1)
}

// ReleaseResources is an helper to call ReleaseResource with an array
func ReleaseResources(names []string) {
	for _, name := range names {
		ReleaseResource(name)
	}
}

// AcquireExecutionSlot takes a slot from a global semaphore
// putting a cap on the total amount of concurrent task executions
func AcquireExecutionSlot(ctx context.Context) error {
	if global == nil {
		return nil
	}
	if global.executionSemaphore == nil {
		return nil
	}
	return global.executionSemaphore.Acquire(ctx, 1)
}

// ReleaseExecutionSlot frees up a slot on the global execution semaphore
func ReleaseExecutionSlot() {
	if global == nil {
		return
	}
	if global.executionSemaphore == nil {
		return
	}
	global.executionSemaphore.Release(1)
}

var global *Cfg

// Config returns the global configuration data of this instance
// once lazy-loaded from configstore
func Config(store *configstore.Store) (*Cfg, error) {
	if global == nil {
		global = &Cfg{}

		cfgStr, err := configstore.Filter().Slice(UtaskCfgSecretAlias).Squash().Store(store).MustGetFirstItem().Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get utask configuration from store: %s", err)
		}

		if err := json.Unmarshal([]byte(cfgStr), &global); err != nil {
			return nil, fmt.Errorf("failed to unmarshal utask configuration: %s", err)
		}

		if global.DelayBetweenCrashedTasksResolution != "" {
			global.InstanceCollectorWaitDuration, err = time.ParseDuration(global.DelayBetweenCrashedTasksResolution)
			if err != nil {
				return nil, fmt.Errorf("failed to parse \"delay_between_crashed_tasks_resolution\": %s", err)
			}
		} else {
			global.InstanceCollectorWaitDuration = defaultInstanceCollectorWaitDuration
		}
		global.MaxConcurrentExecutionsFromCrashedComputed = defaultMaxConcurrentExecutionsFromCrashed
		if global.MaxConcurrentExecutionsFromCrashed != nil {
			global.MaxConcurrentExecutionsFromCrashedComputed = *global.MaxConcurrentExecutionsFromCrashed
		}

		if global.ResourceAcquireTimeout != "" {
			global.resourceAcquireTimeoutDuration, err = time.ParseDuration(global.ResourceAcquireTimeout)
			if err != nil {
				return nil, fmt.Errorf("failed to parse \"resource_acquire_timeout\": %s", err)
			}
		} else {
			global.resourceAcquireTimeoutDuration = defaultResourceAcquireTimeout
		}

		App = global.ApplicationName

		global.buildLimits()

		if global.MaxConcurrentExecutionsFromCrashedComputed > global.getMaxConcurrentExecutions() {
			return nil, errors.New("max_concurrent_executions_from_crashed can't be greater than max_concurrent_executions")
		}
	}

	return global, nil
}
