package notify

import (
	log "github.com/sirupsen/logrus"

	"github.com/ovh/utask"
)

const (
	errSendCommon string = "Error while sending notification on"
)

// WrappedSendError captures an error from Send Notify
func WrappedSendError(err error, m *Message, backend, name string) {
	newLogger(err, m, backend, name).
		Errorf("%s %s", errSendCommon, backend)
}

// WrappedSendErrorWithBody captures an error with a response body from Send Notify.
func WrappedSendErrorWithBody(err error, m *Message, backend, name, body string) {
	newLogger(err, m, backend, name).
		WithField("response_body", body).
		Errorf("%s %s", errSendCommon, backend)
}

// newLogger creates a logger instance with pre-filled fields.
func newLogger(err error, m *Message, backend, name string) *log.Entry {
	return log.WithFields(log.Fields{
		"notify_backend":    backend,
		"notifier_name":     name,
		"task_id":           m.TaskID(),
		"notification_type": m.NotificationType,
		"instance_id":       utask.InstanceID,
	}).WithError(err)
}
