package api

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/metadata"
	"github.com/wI2L/fizz"
)

var requestIDHeader = http.CanonicalHeaderKey("X-Request-Id")

func auditLogsMiddleware(c *gin.Context) {
	now := time.Now()
	c.Next()
	requestDuration := time.Since(now)

	// Unescape the querystring for readability.
	q, _ := url.QueryUnescape(c.Request.URL.RawQuery)

	fields := logrus.Fields{
		"status":          c.Writer.Status(),
		"method":          c.Request.Method,
		"path":            c.Request.URL.Path,
		"query":           q,
		"user_agent":      c.Request.UserAgent(),
		"duration":        requestDuration.Seconds(),
		"duration_ms":     requestDuration.Milliseconds(),
		"request_host":    c.Request.Host,
		"remote_ip":       c.ClientIP(),
		"runner_instance": utask.InstanceID,
		"request_id":      c.Request.Header.Get(requestIDHeader),
		"log_type":        "api",
	}
	if op, _ := fizz.OperationFromContext(c); op != nil {
		fields["action"] = op.ID
	}
	if user := c.GetString(auth.IdentityProviderCtxKey); user != "" {
		fields["user"] = user
	}
	for k, v := range metadata.GetActionMetadata(c) {
		fields["action_metadata_"+k] = v
	}
	if metadata.IsSUDO(c) {
		fields["sudo"] = true
	}

	errs := c.Errors.Errors()

	if len(errs) > 0 {
		fields["success"] = false
		logrus.WithFields(fields).WithError(
			errors.New(strings.Join(errs, "\n")),
		).Error("error")
	} else {
		fields["success"] = true
		logrus.WithFields(fields).Info("success")
	}
}

func ajaxHeadersMiddleware(c *gin.Context) {
	//Specifies a URI that may access the resource.
	//For requests without credentials, the server may specify '*' as a wildcard,
	//thereby allowing any origin to access the resource.
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	//Methods Allowed => TODO : Is better that each method expose his allow-method
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")

	//Used in response to a preflight request to indicate which HTTP headers can be used when making the actual request.
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	//Lets a server whitelist headers that browsers are allowed to access.
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	c.Next()
}

func authMiddleware(authProvider func(*http.Request) (string, error)) func(c *gin.Context) {
	if authProvider != nil {
		return func(c *gin.Context) {
			user, err := authProvider(c.Request)
			if err != nil {
				if errors.IsUnauthorized(err) {
					c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
				}
				c.AbortWithError(http.StatusUnauthorized, err)
				return
			}
			c.Set(auth.IdentityProviderCtxKey, user)
			c.Next()
		}
	}
	return func(c *gin.Context) { c.Next() }
}

func groupAuthMiddleware(authProvider func(*http.Request) (string, []string, error)) func(c *gin.Context) {
	if authProvider != nil {
		return func(c *gin.Context) {
			user, groups, err := authProvider(c.Request)
			if err != nil {
				if errors.IsUnauthorized(err) {
					c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
				}
				_ = c.AbortWithError(http.StatusUnauthorized, err)
				return
			}
			c.Set(auth.IdentityProviderCtxKey, user)
			c.Set(auth.GroupProviderCtxKey, groups)
			c.Next()
		}
	}
	return func(c *gin.Context) { c.Next() }
}
