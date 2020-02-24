package httputil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/juju/errors"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
	"github.com/ovh/utask/pkg/utils"
)

// UnmarshalFunc is a type of function capable of taking the body of an http response and
// deserialize it into a target interface{}
type UnmarshalFunc func(action []byte, values interface{}) error

var (
	unmarshalers = map[string]UnmarshalFunc{
		"application/json": func(data []byte, target interface{}) error {
			return utils.JSONnumberUnmarshal(bytes.NewReader(data), target)
		},
	}
)

// UnmarshalResponse takes an http Response and returns:
// - its body, deserialized if content-type appropriate
// - metadata such as headers and status code
func UnmarshalResponse(resp *http.Response) (interface{}, interface{}, error) {

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("can't read body: %s", err.Error())
	}

	metadata := map[string]interface{}{
		taskplugin.HTTPStatus: resp.StatusCode,
	}
	headers := map[string]string{}
	for k, list := range resp.Header {
		if len(list) > 0 {
			headers[k] = list[0]
		}
	}
	metadata[taskplugin.HTTPHeaders] = headers

	cookies := map[string]string{}
	for _, c := range resp.Cookies() {
		if c != nil {
			cookies[c.Name] = c.Value
		}
	}
	metadata[taskplugin.HTTPCookies] = cookies

	var output interface{}
	contentType := strings.SplitN(resp.Header.Get("Content-Type"), ";", 2)
	unmarshaler, ok := unmarshalers[contentType[0]]
	if ok && len(bodyBytes) > 0 {
		var payload interface{}
		err = unmarshaler(bodyBytes, &payload)
		if err != nil {
			return nil, metadata, fmt.Errorf("can't unmarshal body: %s", err.Error())
		}
		output = payload
	} else {
		output = string(bodyBytes)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := fmt.Errorf("failed api request: %d: %s", resp.StatusCode, string(bodyBytes))
		if resp.StatusCode > 399 && resp.StatusCode < 500 {
			return output, metadata, errors.NewBadRequest(err, "Client error")
		}
		return output, metadata, err
	}

	return output, metadata, nil
}
