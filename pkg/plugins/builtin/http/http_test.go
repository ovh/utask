package pluginhttp

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_validConfig(t *testing.T) {
	cfg := HTTPConfig{
		URL:            "http://lolcat.host/stuff",
		Method:         "GET",
		TimeoutSeconds: "10",
		DenyRedirects:  "true",
		Auth: Auth{
			Bearer: "my_token",
			Basic: AuthBasic{
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

	// Wrong timeout_seconds
	cfg.TimeoutSeconds = "-2"
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "timeout_seconds is wrong -2")

	// Wrong deny_redirects
	cfg.DenyRedirects = "foo"
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "deny_redirects is wrong foo")
}

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func Test_exec(t *testing.T) {

	NewHTTPClient = func(cfg HTTPClientConfig) HTTPClient {
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
				httpResponse.Body = ioutil.NopCloser(bytes.NewBuffer(bodyResponse))
				httpResponse.StatusCode = 200
				return httpResponse, nil
			},
		}
	}

	cfg := HTTPConfig{
		URL:    "http://lolcat.host/stuff",
		Method: "GET",
		Parameters: []Parameter{
			Parameter{
				Key:   "foo",
				Value: "bar",
			},
		},
		TimeoutSeconds: "10",
		DenyRedirects:  "true",
		Auth: Auth{
			Bearer: "my_token",
		},
	}

	cfgJSON, err := json.Marshal(cfg)
	assert.NoError(t, err)

	output, metadata, err := Plugin.Exec("test", json.RawMessage(""), json.RawMessage(cfgJSON), nil)
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

func TestNewHTTPClient(t *testing.T) {
	NewHTTPClient = defaultHTTPClientFactory
	c := NewHTTPClient(HTTPClientConfig{Timeout: time.Hour, DenyRedirects: true})
	assert.NotNil(t, c)
}
