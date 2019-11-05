package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
)

const maxBodyBytes = 256 * 1024

var yamlBind = yamlBinding{}

type yamlBinding struct{}

func (yamlBinding) Name() string { return "yamlBinding" }
func (yamlBinding) Bind(req *http.Request, obj interface{}) error {
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	return yaml.Unmarshal(bodyBytes, obj, jsonNumberOpt)
}

func yamlBindHook(c *gin.Context, i interface{}) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	if c.Request.ContentLength == 0 || c.Request.Method == http.MethodGet {
		return nil
	}
	if err := c.ShouldBindWith(i, yamlBind); err != nil && err != io.EOF {
		return fmt.Errorf("error parsing request body: %s", err.Error())
	}
	return nil
}

func jsonNumberOpt(dec *json.Decoder) *json.Decoder {
	dec.UseNumber()
	return dec
}
