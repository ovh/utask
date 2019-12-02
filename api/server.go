package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/loopfz/gadgeto/tonic/utils/jujerr"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"

	"github.com/ovh/utask"
	"github.com/ovh/utask/api/handler"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/pkg/auth"
)

// Server wraps the http handler that exposes a REST API to control
// the task orchestration engine
type Server struct {
	httpHandler    *fizz.Fizz
	authMiddleware func(*gin.Context)
}

// NewServer returns a new Server
func NewServer() *Server {
	return &Server{
		authMiddleware: func(c *gin.Context) { c.Next() }, // default no-op middleware
	}
}

// WithAuth configures the Server's auth middleware
// it receives an authProvider function capable of extracting a caller's identity from an *http.Request
// the authProvider function also has discretion to deny authorization for a request by returning an error
func (s *Server) WithAuth(authProvider func(*http.Request) (string, error)) {
	if authProvider != nil {
		s.authMiddleware = authMiddleware(authProvider)
	}
}

// ListenAndServe launches an http server and stays blocked until
// the server is shut down by a system signal
func (s *Server) ListenAndServe() error {
	ctx, cancel := context.WithCancel(context.Background())

	s.build(ctx)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	srv := &http.Server{Addr: fmt.Sprintf(":%d", utask.FPort), Handler: s.httpHandler}

	go func() {
		<-stop
		logrus.Info("Shutting down...")
		cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logrus.Fatal(err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Handler returns the underlying http.Handler of a Server
func (s *Server) Handler(ctx context.Context) http.Handler {
	s.build(ctx)
	return s.httpHandler
}

// build registers all routes and their corresponding handlers for the Server's API
func (s *Server) build(ctx context.Context) {
	if s.httpHandler == nil {
		ginEngine := gin.Default()
		ginEngine.
			Group("/ui", s.authMiddleware).
			StaticFS("/dashboard", http.Dir("./static/dashboard"))
		ginEngine.
			StaticFS("/ui/editor", http.Dir("./static/editor"))

		collectMetrics(ctx)
		ginEngine.GET("/metrics", gin.WrapH(promhttp.Handler()))

		router := fizz.NewFromEngine(ginEngine)

		router.Use(ajaxHeadersMiddleware, errorLogMiddleware)

		tonic.SetErrorHook(jujerr.ErrHook)
		tonic.SetBindHook(yamlBindHook)

		authRoutes := router.Group("/", "Authenticated routes", "Utask CRUD: authentication and authorization is required", s.authMiddleware)
		{
			// public template listing
			authRoutes.GET("/template",
				[]fizz.OperationOption{
					fizz.Summary("List task templates"),
				},
				tonic.Handler(handler.ListTemplates, 200))
			authRoutes.GET("/template/:name",
				[]fizz.OperationOption{
					fizz.Summary("Get task template details"),
				},
				tonic.Handler(handler.GetTemplate, 200))

			// task creation in batches
			authRoutes.POST("/batch",
				[]fizz.OperationOption{
					fizz.Summary("Create a batch of tasks"),
				},
				maintenanceMode,
				tonic.Handler(handler.CreateBatch, 201))

			// task
			authRoutes.POST("/task",
				[]fizz.OperationOption{
					fizz.Summary("Create new task"),
				},
				maintenanceMode,
				tonic.Handler(handler.CreateTask, 201))
			authRoutes.GET("/task",
				[]fizz.OperationOption{
					fizz.Summary("List tasks"),
				},
				tonic.Handler(handler.ListTasks, 200))
			authRoutes.GET("/task/:id",
				[]fizz.OperationOption{
					fizz.Summary("Get task details"),
				},
				tonic.Handler(handler.GetTask, 200))
			authRoutes.PUT("/task/:id",
				[]fizz.OperationOption{
					fizz.Summary("Edit task"),
				},
				maintenanceMode,
				tonic.Handler(handler.UpdateTask, 200))
			authRoutes.POST("/task/:id/wontfix",
				[]fizz.OperationOption{
					fizz.Summary("Cancel task"),
				},
				maintenanceMode,
				tonic.Handler(handler.WontfixTask, 204))
			authRoutes.DELETE("/task/:id",
				[]fizz.OperationOption{
					fizz.Summary("Delete task"),
					fizz.Description("Admin rights required"),
				},
				requireAdmin,
				maintenanceMode,
				tonic.Handler(handler.DeleteTask, 204))

			// comments
			authRoutes.POST("/task/:id/comment",
				[]fizz.OperationOption{
					fizz.Summary("Post new comment on task"),
				},
				maintenanceMode,
				tonic.Handler(handler.CreateComment, 201))
			authRoutes.GET("/task/:id/comment",
				[]fizz.OperationOption{
					fizz.Summary("List task comments"),
				},
				tonic.Handler(handler.ListComments, 200))
			authRoutes.GET("/task/:id/comment/:commentid",
				[]fizz.OperationOption{
					fizz.Summary("Get single task comment"),
				},
				tonic.Handler(handler.GetComment, 200))
			authRoutes.PUT("/task/:id/comment/:commentid",
				[]fizz.OperationOption{
					fizz.Summary("Edit task comment"),
				},
				maintenanceMode,
				tonic.Handler(handler.UpdateComment, 200))
			authRoutes.DELETE("/task/:id/comment/:commentid",
				[]fizz.OperationOption{
					fizz.Summary("Delete task comment"),
				},
				maintenanceMode,
				tonic.Handler(handler.DeleteComment, 204))

			// resolution
			authRoutes.POST("/resolution",
				[]fizz.OperationOption{
					fizz.Summary("Create task resolution"),
					fizz.Summary("This action instantiates a holder for the task's execution state. Only an approved resolver or admin user can perform this action."),
				},
				maintenanceMode,
				tonic.Handler(handler.CreateResolution, 201))
			authRoutes.GET("/resolution",
				[]fizz.OperationOption{
					fizz.Summary("List task resolutions"),
					fizz.Description("By default, only resolution for which the user is responsible will be displayed. Admin users can list every task resolution."),
				},
				tonic.Handler(handler.ListResolutions, 200))
			authRoutes.GET("/resolution/:id",
				[]fizz.OperationOption{
					fizz.Summary("Get the details of a task resolution"),
					fizz.Description("Details include the intermediate results of every step. Admin users can view any resolution's details."),
				},
				tonic.Handler(handler.GetResolution, 200))
			authRoutes.PUT("/resolution/:id",
				[]fizz.OperationOption{
					fizz.Summary("Edit a task's resolution during execution."),
					fizz.Description("Action of last resort if a task needs fixing. Admin users only."),
				},
				requireAdmin,
				maintenanceMode,
				tonic.Handler(handler.UpdateResolution, 204))
			authRoutes.POST("/resolution/:id/run",
				[]fizz.OperationOption{
					fizz.Summary("Execute a task"),
				},
				tonic.Handler(handler.RunResolution, 204))
			authRoutes.POST("/resolution/:id/pause",
				[]fizz.OperationOption{
					fizz.Summary("Pause a task's execution"),
					fizz.Description("This action takes a task out of the execution pipeline, it will not be considered for automatic retry until it is re-run manually."),
				},
				maintenanceMode,
				tonic.Handler(handler.PauseResolution, 204))
			authRoutes.POST("/resolution/:id/extend",
				[]fizz.OperationOption{
					fizz.Summary("Extend max retry limit for a task's execution"),
				},
				maintenanceMode,
				tonic.Handler(handler.ExtendResolution, 204))
			authRoutes.POST("/resolution/:id/cancel",
				[]fizz.OperationOption{
					fizz.Summary("Cancel a task's execution"),
				},
				maintenanceMode,
				tonic.Handler(handler.CancelResolution, 204))

			//	authRoutes.POST("/resolution/:id/rollback",
			//		[]fizz.OperationOption{
			// 			fizz.Summary(""),
			//		},
			//		tonic.Handler(handler.ResolutionRollback, 200))

			authRoutes.GET("/",
				[]fizz.OperationOption{
					fizz.Summary("Redirect to /meta"),
				},
				func(c *gin.Context) {
					c.Redirect(http.StatusMovedPermanently, "/meta")
				})

			authRoutes.GET("/meta",
				[]fizz.OperationOption{
					fizz.Summary("Display service name and user's status"),
				},
				tonic.Handler(rootHandler, 200))

			// admin
			authRoutes.POST("/key-rotate",
				[]fizz.OperationOption{
					fizz.Summary("Re-encrypt all data with latest storage key"),
				},
				requireAdmin,
				tonic.Handler(keyRotate, 200))

			// plugin
			authRoutes.GET("/plugin/script",
				[]fizz.OperationOption{
					fizz.Summary("List of available scripts for script plugin"),
				},
				listScripts)
		}

		router.GET("/unsecured/mon/ping",
			[]fizz.OperationOption{
				fizz.Summary("Assert that the service is running and can talk to it's data backend"),
			},
			pingHandler)
		router.GET("/unsecured/spec.json", nil, router.OpenAPI(&openapi.Info{
			Title:   utask.AppName(),
			Version: utask.Version,
		}, "json"))
		router.GET("/unsecured/stats",
			[]fizz.OperationOption{
				fizz.Summary("Fetch statistics about existing tasks"),
			},
			tonic.Handler(Stats, 200))

		s.httpHandler = router
	}
}

func pingHandler(c *gin.Context) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		c.String(http.StatusInternalServerError, "")
		c.Error(err)
		return
	}
	i, err := dbp.DB().SelectInt(`SELECT 1`)
	if err != nil {
		c.String(http.StatusInternalServerError, "")
		c.Error(err)
		return
	}
	if i != 1 {
		c.String(http.StatusInternalServerError, "")
		c.Error(fmt.Errorf("Unexpected value %d", i))
		return
	}
	c.String(http.StatusOK, "pong")
}

type scriptInfo struct {
	Name             string `json:"name"`
	Size             int64  `json:"size"`
	ModificationTime string `json:"modification_time"`
}

func listScripts(c *gin.Context) {
	sFiles, err := ioutil.ReadDir(utask.FScriptsFolder)
	if err != nil {
		c.String(http.StatusInternalServerError, "")
		c.Error(err)
		return
	}

	payload := []scriptInfo{}

	for _, f := range sFiles {
		if f.Name()[0] != '.' {
			si := scriptInfo{
				Name:             f.Name(),
				Size:             f.Size(),
				ModificationTime: f.ModTime().String(),
			}
			payload = append(payload, si)
		}
	}

	c.JSON(http.StatusOK, payload)
}

type rootOut struct {
	ApplicationName string `json:"application_name"`
	UserIsAdmin     bool   `json:"user_is_admin"`
	Username        string `json:"username"`
	Version         string `json:"version"`
	Commit          string `json:"commit"`
}

func rootHandler(c *gin.Context) (*rootOut, error) {
	return &rootOut{
		ApplicationName: utask.AppName(),
		UserIsAdmin:     auth.IsAdmin(c) == nil,
		Username:        auth.GetIdentity(c),
		Version:         utask.Version,
		Commit:          utask.Commit,
	}, nil
}

func requireAdmin(c *gin.Context) {
	if err := auth.IsAdmin(c); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Next()
}

func maintenanceMode(c *gin.Context) {
	if utask.FMaintenanceMode {
		c.JSON(http.StatusMethodNotAllowed, map[string]string{
			"error": "Maintenance mode activated",
		})
		return
	}
	c.Next()
}

func keyRotate(c *gin.Context) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}
	if err := task.RotateTasks(dbp); err != nil {
		return err
	}
	return resolution.RotateResolutions(dbp)
}
