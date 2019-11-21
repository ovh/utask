package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/auth"
)

func errorLogMiddleware(c *gin.Context) {
	c.Next()

	for _, err := range c.Errors.Errors() {
		logrus.WithFields(logrus.Fields{
			"status":          c.Writer.Status(),
			"method":          c.Request.Method,
			"path":            c.Request.URL.Path,
			"runner_instance": utask.InstanceID,
		}).Error(err)
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
