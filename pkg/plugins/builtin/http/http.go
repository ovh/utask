package pluginhttp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/plugins/builtin/httputil"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
	"github.com/ovh/utask/pkg/utils"
)

// the HTTP plugin performs an HTTP call
var (
	Plugin = taskplugin.New("http", "0.6", exec,
		taskplugin.WithConfig(validConfig, HTTPConfig{}),
	)
)

// HTTPConfig is the configuration needed to perform an HTTP call
type HTTPConfig struct {
	URL            string      `json:"url"`
	Method         string      `json:"method"`
	Body           string      `json:"body,omitempty"`
	Headers        []Header    `json:"headers,omitempty"`
	TimeoutSeconds string      `json:"timeout_seconds,omitempty"`
	HTTPBasicAuth  BasicAuth   `json:"basic_auth,omitempty"`
	DenyRedirects  string      `json:"deny_redirects,omitempty"`
	Parameters     []Parameter `json:"parameters,omitempty"`
	TrimPrefix     string      `json:"trim_prefix,omitempty"`
}

// Header represents an HTTP header
type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Parameter represents HTTP parameters
type Parameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// BasicAuth represents a HTTP basic auth
type BasicAuth struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func validConfig(config interface{}) error {
	cfg := config.(*HTTPConfig)
	switch cfg.Method {
	case "GET", "POST", "PUT", "DELETE":
	default:
		return fmt.Errorf("Unknown method for HTTP runner: %s", cfg.Method)
	}

	if cfg.TimeoutSeconds != "" {
		if _, err := strconv.ParseUint(cfg.TimeoutSeconds, 10, 16); err != nil {
			return fmt.Errorf("timeout_seconds is wrong %s", err.Error())
		}
	}

	if cfg.DenyRedirects != "" {
		if _, err := strconv.ParseBool(cfg.DenyRedirects); err != nil {
			return fmt.Errorf("deny_redirects is wrong %s", err.Error())
		}
	}

	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*HTTPConfig)

	// do it once and avoid re-copies
	body := []byte(cfg.Body)

	if utask.FDebug {
		fmt.Println(string(body))
	}

	req, err := http.NewRequest(cfg.Method, cfg.URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create http request: %s", err.Error())
	}

	q := req.URL.Query()
	for _, p := range cfg.Parameters {
		q.Add(p.Key, p.Value)
	}
	req.URL.RawQuery = q.Encode()

	if cfg.HTTPBasicAuth.User != "" && cfg.HTTPBasicAuth.Password != "" {
		req.SetBasicAuth(cfg.HTTPBasicAuth.User, cfg.HTTPBasicAuth.Password)
	}

	// best-effort match the body's content-type
	var i interface{}
	reader := bytes.NewReader(body)
	if err := utils.JSONnumberUnmarshal(reader, &i); err == nil {
		req.Header.Set("content-type", "application/json")
	} else if err := xml.Unmarshal(body, &i); err == nil {
		req.Header.Set("content-type", "application/xml")
	}

	for _, h := range cfg.Headers {
		req.Header.Set(h.Name, h.Value)
	}

	ts, _ := strconv.ParseUint(cfg.TimeoutSeconds, 10, 16)

	httpClient := &http.Client{
		// 0 by default
		Timeout: time.Duration(ts) * time.Second,
	}

	dr, _ := strconv.ParseBool(cfg.DenyRedirects)

	if dr {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("HTTP request failed: %s", err.Error())
	}

	// remove response magic prefix
	if cfg.TrimPrefix != "" {
		trimPrefixBytes := []byte(cfg.TrimPrefix)
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("HTTP cannot read response: %s", err.Error())
		}
		resp.Body.Close()
		respBody = bytes.TrimPrefix(respBody, trimPrefixBytes)
		resp.Body = ioutil.NopCloser(bytes.NewReader(respBody))
	}

	return httputil.UnmarshalResponse(resp)
}

// ExecutorMetadata generates json schema to validate the metadata
// returned by the http executor
func ExecutorMetadata() string {
	return taskplugin.NewMetadataSchema().
		WithStatusCode().
		String()
}
