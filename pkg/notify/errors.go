package notify

import "fmt"

const (
	errSendCommon string = "Error while sending notification on"
)

// WrappedSendError print a formatted string from Send Notify in case of issue
func WrappedSendError(etype string, err string) {
	fmt.Printf("%s %s: %s", errSendCommon, etype, err)
}
