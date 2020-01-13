package httputil

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalResponse(t *testing.T) {
	var httpResponse = new(http.Response)
	httpResponse.Header = http.Header{"Set-Cookie": {"Cookie-1=foo"}, "Content-Type": {"application/json"}}
	var bodyResponse = []byte(`{"foo": "bar"}`)
	httpResponse.Body = ioutil.NopCloser(bytes.NewBuffer(bodyResponse))
	httpResponse.StatusCode = 200

	output, metadata, err := UnmarshalResponse(httpResponse)

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
