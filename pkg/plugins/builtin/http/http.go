package pluginhttp

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/plugins/builtin/httputil"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
	"github.com/ovh/utask/pkg/utils"
	dac "github.com/ybriffa/go-http-digest-auth-client"
)

// the HTTP plugin performs an HTTP call
var (
	Plugin = taskplugin.New("http", "1.0", exec,
		taskplugin.WithConfig(validConfig, HTTPConfig{}),
		taskplugin.WithResources(resourceshttp),
	)
)

const (
	// TimeoutDefault represents the default value that will be used for HTTP call, if not defined in configuration
	TimeoutDefault = "30s"
)

// HTTPConfig is the configuration needed to perform an HTTP call
type HTTPConfig struct {
	URL                string      `json:"url"`
	Host               string      `json:"host"`
	Path               string      `json:"path"`
	Method             string      `json:"method"`
	Body               string      `json:"body,omitempty"`
	Headers            []parameter `json:"headers,omitempty"`
	Timeout            string      `json:"timeout,omitempty"`
	Auth               auth        `json:"auth,omitempty"`
	FollowRedirect     string      `json:"follow_redirect,omitempty"`
	QueryParameters    []parameter `json:"query_parameters,omitempty"`
	TrimPrefix         string      `json:"trim_prefix,omitempty"`
	InsecureSkipVerify string      `json:"insecure_skip_verify,omitempty"`
	RootCA             string      `json:"root_ca,omitempty"`
}

// parameter represents either headers, query parameters, ...
type parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// auth represents HTTP authentication
type auth struct {
	Basic     *authBasic  `json:"basic"`
	Digest    *authDigest `json:"digest"`
	Bearer    *string     `json:"bearer"`
	MutualTLS *mTLS       `json:"mutual_tls"`
}

// authBasic represents the embedded basic auth inside Auth struct
type authBasic struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

// authDigest represents the embedded digest auth inside Auth struct
type authDigest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type mTLS struct {
	ClientCert string `json:"client_cert"`
	ClientKey  string `json:"client_key"`
}

func validConfig(config interface{}) error {
	cfg := config.(*HTTPConfig)
	if !strings.HasPrefix(cfg.Method, "{{") && !strings.HasSuffix(cfg.Method, "}}") {
		switch cfg.Method {
		case "GET", "POST", "PUT", "DELETE", "PATCH":
		default:
			return fmt.Errorf("unknown method for HTTP runner: %s", cfg.Method)
		}
	}

	if cfg.URL != "" {
		if cfg.Host != "" || cfg.Path != "" {
			return errors.New("URL field conflicts with Host+Path")
		}
	}

	if cfg.Host == "" && cfg.URL == "" {
		return errors.New("missing either URL or Host")
	}

	// skip validation of Timeout, FollowRedirect to allow runtime templating

	for _, p := range cfg.Headers {
		if p.Name == "" {
			return fmt.Errorf("missing header name (with value '%s')", p.Value)
		}
	}

	for _, p := range cfg.QueryParameters {
		if p.Name == "" {
			return fmt.Errorf("missing query parameter name (with value '%s')", p.Value)
		}
	}

	authentication := 0
	for _, authExist := range []bool{cfg.Auth.Basic != nil, cfg.Auth.Bearer != nil, cfg.Auth.Digest != nil} {
		if authExist {
			authentication++
		}
	}
	if authentication > 1 {
		return errors.New("basic|digest|bearer authentications are mutually exclusive")
	}

	if cfg.Auth.Basic != nil {
		if cfg.Auth.Basic.User == "" || cfg.Auth.Basic.Password == "" {
			return fmt.Errorf("missing either user or password for basic auth")
		}
	}

	if cfg.Auth.Digest != nil {
		if cfg.Auth.Digest.User == "" || cfg.Auth.Digest.Password == "" {
			return fmt.Errorf("missing either user or password for digest auth")
		}
	}

	if cfg.Auth.Bearer != nil && *cfg.Auth.Bearer == "" {
		return fmt.Errorf("missing bearer token value")
	}

	if cfg.Auth.MutualTLS != nil {
		if cfg.Auth.MutualTLS.ClientCert == "" || cfg.Auth.MutualTLS.ClientKey == "" {
			return fmt.Errorf("missing either client_cert or client_key for mTLS")
		}
	}

	return nil
}

func resourceshttp(i interface{}) []string {
	cfg := i.(*HTTPConfig)

	var host string
	if cfg.Host == "" {
		uri, _ := url.Parse(cfg.URL)
		if uri != nil {
			host = uri.Host
		}
	} else {
		uri, _ := url.Parse(cfg.Host)
		if uri != nil {
			host = uri.Host
		}
	}

	if host == "" {
		return []string{"socket"}
	}
	return []string{
		"socket",
		"url:" + host,
	}
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*HTTPConfig)

	// do it once and avoid re-copies
	body := []byte(cfg.Body)

	if utask.FDebug {
		fmt.Println(string(body))
	}

	target := cfg.URL
	if target == "" {
		hostURL, err := url.Parse(cfg.Host)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse host: %s", err)
		}
		pathURL, err := url.Parse(cfg.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse path: %s", err)
		}
		pathURL.Host = hostURL.Host
		pathURL.Scheme = hostURL.Scheme
		target = pathURL.String()
	}

	req, err := http.NewRequest(cfg.Method, target, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create HTTP request: %s", err.Error())
	}

	q := req.URL.Query()
	for _, p := range cfg.QueryParameters {
		q.Add(p.Name, p.Value)
	}
	req.URL.RawQuery = q.Encode()

	if cfg.Auth.Bearer != nil {
		var bearer = "Bearer " + *cfg.Auth.Bearer
		req.Header.Add("Authorization", bearer)
	} else if cfg.Auth.Basic != nil {
		req.SetBasicAuth(cfg.Auth.Basic.User, cfg.Auth.Basic.Password)
	}

	for _, h := range cfg.Headers {
		req.Header.Set(h.Name, h.Value)
	}

	// best-effort match the body's content-type
	if len(body) > 0 && req.Header.Get("Content-Type") == "" {
		var i interface{}
		reader := bytes.NewReader(body)
		if err := utils.JSONnumberUnmarshal(reader, &i); err == nil {
			req.Header.Set("Content-Type", "application/json")
		} else if err := xml.Unmarshal(body, &i); err == nil {
			req.Header.Set("Content-Type", "application/xml")
		}
	}

	if cfg.Timeout == "" {
		cfg.Timeout = TimeoutDefault
	}

	var fr bool

	td, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse timeout: %s", err)
	}
	if cfg.FollowRedirect != "" {
		fr, err = strconv.ParseBool(cfg.FollowRedirect)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse follow_redirect: %s", err)
		}
	}
	var insecureSkipVerify bool
	if cfg.InsecureSkipVerify != "" {
		insecureSkipVerify, err = strconv.ParseBool(cfg.InsecureSkipVerify)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse insecure_skip_verify: %s", err)
		}
	}
	httpClientConfig := httputil.HTTPClientConfig{
		Timeout:        td,
		FollowRedirect: fr,
	}

	opts := []func(*http.Transport) error{}
	if insecureSkipVerify {
		opts = append(opts, httputil.WithTLSInsecureSkipVerify(true))
	}

	if cfg.Auth.MutualTLS != nil {
		cert, err := tls.X509KeyPair([]byte(cfg.Auth.MutualTLS.ClientCert), []byte(cfg.Auth.MutualTLS.ClientKey))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse x509 mTLS certificate or key: %s", err)
		}
		opts = append(opts, httputil.WithTLSClientAuth(cert))
	}

	if cfg.RootCA != "" {
		opts = append(opts, httputil.WithTLSRootCA([]byte(cfg.RootCA)))
	}

	if len(opts) > 0 {
		httpClientConfig.Transport, err = httputil.GetTransport(opts...)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to craft a new http transport: %s", err)
		}
	}

	httpClient := httputil.NewHTTPClient(httpClientConfig)

	if cfg.Auth.Digest != nil {
		transport := dac.NewTransport(cfg.Auth.Digest.User, cfg.Auth.Digest.Password)
		transport.HTTPClient = httpClient
		httpClient = &transport
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("can't do HTTP request: %s", err.Error())
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
