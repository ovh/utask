package notify

import "github.com/ovh/utask"

// utask should be able to notify about inner task events through different channels
// relevant information for the outside world is described by the Message struct
// this package allows for the registration of different senders, capable of handling the Message struct

var (
	senders = make(map[string]NotificationSender)
	// actions represents configuration of each notify actions
	actions utask.NotifyActions
)

// NotificationSender is an object capable of sending a Message struct
// over a notification channel, as determined by its implementation
type NotificationSender interface {
	Send(m *Message, name string)
}

// RegisterSender adds a NotificationSender to the pool of available senders
func RegisterSender(s NotificationSender, name string) {
	senders[name] = s
}

// ListSendersNames returns a list of available senders
func ListSendersNames() []string {
	names := []string{}
	for name := range senders {
		names = append(names, name)
	}
	return names
}

// RegisterActions set available actions
func RegisterActions(na utask.NotifyActions) {
	actions = na
}

// ListActions returns a list of available actions to notify
func ListActions() utask.NotifyActions {
	return actions
}

// Send dispatches a Message struct over all registered senders
func Send(m *Message, params utask.NotifyActionsParameters) {
	if params.Disabled {
		return
	}

	// Empty NotifyBackends list means any
	if len(params.NotifyBackends) == 0 {
		for name, s := range senders {
			go s.Send(m, name)
		}
		return
	}

	// Match given config name /w senders
	for _, n := range params.NotifyBackends {
		for nsname, ns := range senders {
			switch n {
			case nsname:
				go ns.Send(m, nsname)
			}
		}
	}
}
