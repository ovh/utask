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

	var cfg = `
	{
		"url": "http://lolcat.host/stuff",
		"method": "GET",
		"timeout_seconds": "10",
		"deny_redirects": "true",
		"auth": {
			"bearer": "my_token",
			"basic": {
				"user": "foo",
				"password": "bar"
			}
		}
	}`

	assert.NoError(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfg)))
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

	var cfg = `
	{
		"url": "http://lolcat.host/stuff",
		"method": "GET",
		"parameters": [{"key": "foo", "value": "bar"}],
		"timeout_seconds": "10",
		"deny_redirects": "true",
		"auth": {
			"bearer": "my_token"
		}
	}`

	output, metadata, err := Plugin.Exec("test", json.RawMessage(""), json.RawMessage(cfg), nil)
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
