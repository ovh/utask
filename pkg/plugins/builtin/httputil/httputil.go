package httputil

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/juju/errors"
	"golang.org/x/net/http2"

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

	bodyBytes, err := io.ReadAll(resp.Body)
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

// NewHTTPClient is a factory of HTTPClient
var NewHTTPClient = defaultHTTPClientFactory

// HTTPClient is an interface for decoupling http.Client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPClientConfig is a set of options used to initialize a HTTPClient
type HTTPClientConfig struct {
	Timeout        time.Duration
	FollowRedirect bool
	Transport      http.RoundTripper
}

func defaultHTTPClientFactory(cfg HTTPClientConfig) HTTPClient {
	c := new(http.Client)
	c.Timeout = cfg.Timeout
	if !cfg.FollowRedirect {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	if cfg.Transport != nil {
		c.Transport = cfg.Transport
	}
	return c
}

func GetTransport(opts ...func(*http.Transport) error) (http.RoundTripper, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	for _, o := range opts {
		if err := o(tr); err != nil {
			return tr, err
		}
	}

	_ = http2.ConfigureTransport(tr)
	return tr, nil
}

func WithTLSInsecureSkipVerify(v bool) func(*http.Transport) error {
	return func(t *http.Transport) error {
		if t.TLSClientConfig == nil {
			t.TLSClientConfig = &tls.Config{}
		}

		t.TLSClientConfig.InsecureSkipVerify = v
		return nil
	}
}

func WithTLSClientAuth(cert tls.Certificate) func(*http.Transport) error {
	return func(t *http.Transport) error {
		if t.TLSClientConfig == nil {
			t.TLSClientConfig = &tls.Config{}
		}

		t.TLSClientConfig.Certificates = append(t.TLSClientConfig.Certificates, cert)
		return nil
	}
}

// WithTLSRootCA should be called only once, with multiple PEM encoded certificates as input if needed.
func WithTLSRootCA(caCert []byte) func(*http.Transport) error {
	return func(t *http.Transport) error {
		if t.TLSClientConfig == nil {
			t.TLSClientConfig = &tls.Config{}
		}
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			fmt.Println("http: tls: failed to load default system cert pool, fallback to an empty cert pool")
			caCertPool = x509.NewCertPool()
		}

		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return errors.New("WithTLSRootCA: failed to add a certificate to the cert pool")
		}

		t.TLSClientConfig.RootCAs = caCertPool
		return nil
	}
}
