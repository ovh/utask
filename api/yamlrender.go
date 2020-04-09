package api

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/markusthoemmes/goautoneg"
)

const (
	acceptHeader = "Accept"
	jsonFormat   = "json"
	yamlFormat   = "x-yaml"
)

// yamljsonRenderHook will render output regarding the Accept request header
// in JSON or YAML format.
func yamljsonRenderHook(c *gin.Context, statusCode int, payload interface{}) {
	var status int
	if c.Writer.Written() {
		status = c.Writer.Status()
	} else {
		status = statusCode
	}
	if payload != nil {
		accept := goautoneg.ParseAccept(c.Request.Header.Get(acceptHeader))
		destinationFormat := jsonFormat
		for _, format := range accept {
			if format.Type != "application" {
				continue
			}

			switch format.SubType {
			case jsonFormat, yamlFormat:
			default:
				continue
			}

			destinationFormat = format.SubType
			break
		}
		if destinationFormat == yamlFormat {
			yamlOutput, err := yaml.Marshal(payload)
			if err != nil {
				c.JSON(500, map[string]string{"message": fmt.Sprintf("error while marshalling: %s", err)})
				return
			}
			c.Data(status, "application/"+yamlFormat, yamlOutput)
		} else if gin.IsDebugging() {
			c.IndentedJSON(status, payload)
		} else {
			c.JSON(status, payload)
		}
	} else {
		c.String(status, "")
	}
}
