package pluginhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"testing"

	httputilutask "github.com/ovh/utask/pkg/plugins/builtin/httputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_validConfig(t *testing.T) {
	bearerToken := "my_token"
	cfg := HTTPConfig{
		URL:            "http://lolcat.host/stuff",
		Method:         "GET",
		Timeout:        "10s",
		FollowRedirect: "false",
		Auth: auth{
			Basic: &authBasic{
				User:     "foo",
				Password: "bar",
			},
		},
	}

	cfgJSON, err := json.Marshal(cfg)
	assert.NoError(t, err)

	assert.NoError(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)))

	// Wrong method
	cfg.Method = "RANDOM"
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "Unknown method for HTTP runner: RANDOM")
	cfg.Method = "GET"

	// wrong headers
	cfg.Headers = []parameter{
		{
			Name:  "",
			Value: "foo",
		},
	}
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "headers has invalid name value")
	cfg.Headers = []parameter{
		{
			Name:  "x-foo-header",
			Value: "foo",
		},
	}

	// wrong query params
	cfg.QueryParameters = []parameter{
		{
			Name:  "",
			Value: "foo",
		},
	}
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "query_parameters has invalid name value")
	cfg.QueryParameters = []parameter{
		{
			Name:  "bar",
			Value: "foo",
		},
	}

	// wrong auth: exclusive auth added
	cfg.Auth.Bearer = &bearerToken
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "basic auth and bearer auth are mutually exclusive")
	cfg.Auth.Bearer = nil

	// wrong auth: invalid basic auth
	cfg.Auth.Basic.Password = ""
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "missing either user or password for basic auth")
	cfg.Auth.Basic.Password = "bar"

	// wrong auth: invalid bearer auth
	cfg.Auth.Basic = nil
	empty := ""
	cfg.Auth.Bearer = &empty
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "missing bearer token value")
	cfg.Auth.Basic = &authBasic{
		User:     "foo",
		Password: "bar",
	}
	cfg.Auth.Bearer = nil

	// wrong auth: invalid mTLS auth
	cfg.Auth.MutualTLS = &mTLS{
		ClientCert: "foo",
		ClientKey:  "",
	}
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "missing either client_cert or client_key for mTLS")
	cfg.Auth.MutualTLS = nil

	// no URL
	cfg.URL = ""
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "URL should not be empty without host/path")

	cfg.URL = "http://foobar.example"
	cfg.Path = "/search"
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "incompatible parameters URL + Path")

	cfg.Host = "http://bla.example"
	cfg.Path = ""
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "incompatible parameters URL + host")

	cfg.URL = ""
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.NoError(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)))

}

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func Test_exec(t *testing.T) {

	httputilutask.NewHTTPClient = func(cfg httputilutask.HTTPClientConfig) httputilutask.HTTPClient {
		return MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				reqDump, _ := httputil.DumpRequest(req, false)
				t.Log(string(reqDump))
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "http://lolcat.host/stuff?foo=bar", req.URL.String())
				assert.Equal(t, "Bearer my_token", req.Header.Get("Authorization"))

				var httpResponse = new(http.Response)
				httpResponse.Header = http.Header{"Set-Cookie": {"Cookie-1=foo"}, "Content-Type": {"application/json"}}
				var bodyResponse = []byte(`{"foo": "bar"}`)
				httpResponse.Body = io.NopCloser(bytes.NewBuffer(bodyResponse))
				httpResponse.StatusCode = 200
				return httpResponse, nil
			},
		}
	}

	bearerToken := "my_token"
	cfg := HTTPConfig{
		URL:    "http://lolcat.host/stuff",
		Method: "GET",
		QueryParameters: []parameter{
			{
				Name:  "foo",
				Value: "bar",
			},
		},
		Timeout:        "10s",
		FollowRedirect: "false",
		Auth: auth{
			Bearer: &bearerToken,
		},
	}

	cfgJSON, err := json.Marshal(cfg)
	assert.NoError(t, err)

	output, metadata, _, err := Plugin.Exec("test", json.RawMessage(""), json.RawMessage(cfgJSON), nil)
	require.NoError(t, err)

	assert.NoError(t, err)

	require.NotNil(t, output)
	t.Logf("> %T %+v", output, output)
	mapOutput, ok := output.(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, mapOutput, 1)
	assert.Equal(t, "bar", mapOutput["foo"])

	require.NotNil(t, metadata)
	t.Logf("> %T %+v", metadata, metadata)

	mapMetadata, ok := metadata.(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, mapMetadata, 3)
	assert.Equal(t, 200, mapMetadata["HTTPStatus"])

	mapCookies, ok := mapMetadata["HTTPCookies"].(map[string]string)
	require.True(t, ok)
	assert.Len(t, mapCookies, 1)
	assert.Equal(t, "foo", mapCookies["Cookie-1"])

	mapHeaders, ok := mapMetadata["HTTPHeaders"].(map[string]string)
	require.True(t, ok)
	assert.Len(t, mapHeaders, 2)
	assert.Equal(t, "application/json", mapHeaders["Content-Type"])
	assert.Equal(t, "Cookie-1=foo", mapHeaders["Set-Cookie"])

}
