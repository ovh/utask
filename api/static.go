package api

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	notWrittenSize = -1
	defaultStatus  = http.StatusOK
)

// StaticFilePatternReplaceMiddleware is a middleware that modifies the response body with a given replace pattern
// Used inside utask to change response body of static files at flight, to customize path prefixes.
func StaticFilePatternReplaceMiddleware(oldnew ...string) func(c *gin.Context) {
	return func(c *gin.Context) {
		buffer := &bytes.Buffer{}
		originalWriter := c.Writer
		wb := &responseWriter{
			ResponseWriter: originalWriter,
			status:         defaultStatus,
			size:           notWrittenSize,
			bufwriter:      buffer,
		}
		wb.replacer = strings.NewReplacer(oldnew...)
		c.Writer = wb
		c.Next()
		str := wb.replacer.Replace(buffer.String())
		wb.Header().Set("Content-Length", strconv.FormatInt(int64(len(str)), 10))
		wb.WriteHeaderNow()
		if _, err := originalWriter.WriteString(str); err != nil {
			logrus.WithError(err).Error("unable to respond to static content")
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	size      int64
	status    int
	replacer  *strings.Replacer
	bufwriter *bytes.Buffer
}

// WriteHeader sends an HTTP response header with the provided
// status code.
func (w *responseWriter) WriteHeader(code int) {
	if code <= 0 || w.status == code || w.Written() {
		return
	}

	w.status = code
}

// WriteHeaderNow forces to write HTTP headers
func (w *responseWriter) WriteHeaderNow() {
	if w.Written() {
		return
	}

	w.size = 0
	w.ResponseWriter.WriteHeader(w.status)
}

// Write writes a given bytes array into the response buffer.
func (w *responseWriter) Write(data []byte) (n int, err error) {
	n, err = w.bufwriter.Write(data)
	if err != nil {
		logrus.WithError(err).Error("responseWriter: unable to write data")
	}
	w.size += int64(n)
	return

}

// WriteString writes a given string into the response buffer.
func (w *responseWriter) WriteString(s string) (n int, err error) {
	n, err = io.WriteString(w.bufwriter, s)
	if err != nil {
		logrus.WithError(err).Error("responseWriter: unable to write string")
	}
	w.size += int64(n)
	return
}

// Status returns the HTTP response status code of the current request.
func (w *responseWriter) Status() int {
	return w.status
}

// Size returns the number of bytes already written into the response http body.
func (w *responseWriter) Size() int {
	return int(w.size)
}

// Writter returns true if the response body was already written.
func (w *responseWriter) Written() bool {
	return w.size != notWrittenSize
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotify interface.
func (w *responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flush interface.
func (w *responseWriter) Flush() {
	w.WriteHeaderNow()
	w.ResponseWriter.(http.Flusher).Flush()
}

// Pusher implements the http.Pusher interface
func (w *responseWriter) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}
