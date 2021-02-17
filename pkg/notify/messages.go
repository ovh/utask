package notify

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ovh/utask"
)

const (
	// corresponds to github.com/ovh/utask/models/task.StateBlocked
	stateBlocked = "BLOCKED"

	// corresponds to the path of a task in the Dashboard UI
	dashboardUriTaskView = "/ui/dashboard/#/task/"
)

// Message represents a generic message to be sent
type Message struct {
	MainMessage      string
	NotificationType string
	Fields           map[string]string
}

// TaskStateUpdate holds a digest of data representing a task state change
type TaskStateUpdate struct {
	Title              string
	PublicID           string
	ResolutionPublicID string
	State              string
	TemplateName       string
	RequesterUsername  string
	ResolverUsername   *string
	PotentialResolvers []string
	StepsDone          int
	StepsTotal         int
	Tags               map[string]string
}

// WrapTaskStateUpdate returns a Message struct formatted for a task state change
func WrapTaskStateUpdate(tsu *TaskStateUpdate) *Message {
	var m Message

	m.MainMessage = fmt.Sprintf("#task #id:%s\n%s", tsu.PublicID, tsu.Title)
	m.NotificationType = TaskStateUpdateKey

	m.Fields = make(map[string]string)

	m.Fields["task_id"] = tsu.PublicID
	m.Fields["title"] = tsu.Title
	m.Fields["state"] = tsu.State
	m.Fields["template"] = tsu.TemplateName
	if tsu.RequesterUsername != "" {
		m.Fields["requester"] = tsu.RequesterUsername
	}
	if tsu.ResolverUsername != nil && *tsu.ResolverUsername != "" {
		m.Fields["resolver"] = *tsu.ResolverUsername
	}
	m.Fields["steps"] = fmt.Sprintf("%d/%d", tsu.StepsDone, tsu.StepsTotal)
	if tsu.PotentialResolvers != nil && len(tsu.PotentialResolvers) > 0 {
		m.Fields["potential_resolvers"] = strings.Join(tsu.PotentialResolvers, " ")
	}
	if tsu.ResolutionPublicID != "" {
		m.Fields["resolution_id"] = tsu.ResolutionPublicID
	}

	if tsu.Tags != nil {
		tags, err := json.Marshal(tsu.Tags)
		if err == nil {
			m.Fields["tags"] = string(tags)
		} else {
			log.Printf("notify error: failed to marshal tags for task #%s: %s", tsu.PublicID, err)
		}
	}

	if cfg, err := utask.Config(nil); err == nil {
		m.Fields["url"] = cfg.BaseURL + cfg.DashboardPathPrefix + dashboardUriTaskView + tsu.PublicID
	}

	return &m
}

type TaskValidation struct {
	Title              string
	PublicID           string
	State              string
	TemplateName       string
	RequesterUsername  string
	PotentialResolvers []string
	Tags               map[string]string
}

// WrapTaskValidation returns a Message struct formatted for a task requiring validation
func WrapTaskValidation(tv *TaskValidation) *Message {
	var m Message

	m.MainMessage = fmt.Sprintf("#task #id:%s\n%s", tv.PublicID, tv.Title)
	m.NotificationType = TaskValidationKey

	m.Fields = make(map[string]string)

	m.Fields["task_id"] = tv.PublicID
	m.Fields["title"] = tv.Title
	m.Fields["state"] = tv.State
	m.Fields["template"] = tv.TemplateName
	if tv.RequesterUsername != "" {
		m.Fields["requester"] = tv.RequesterUsername
	}
	if tv.PotentialResolvers != nil && len(tv.PotentialResolvers) > 0 {
		m.Fields["potential_resolvers"] = strings.Join(tv.PotentialResolvers, " ")
	}

	if tv.Tags != nil {
		tags, err := json.Marshal(tv.Tags)
		if err == nil {
			m.Fields["tags"] = string(tags)
		} else {
			log.Printf("notify error: failed to marshal tags for task #%s: %s", tv.PublicID, err)
		}
	}

	if cfg, err := utask.Config(nil); err == nil {
		m.Fields["url"] = cfg.BaseURL + cfg.DashboardPathPrefix + dashboardUriTaskView + tv.PublicID
	}

	return &m
}

func checkIfDeliverMessage(m *Message, b *notificationBackend) bool {
	send := checkIfDeliverMessageFromState(m, b.defaultNotificationStrategy[m.NotificationType])

	templateName, ok := m.Fields["template"]
	if !ok {
		return send
	}

	actionStrat, ok := b.templateNotificationStrategies[m.NotificationType]
	if !ok {
		return send
	}

	for _, strat := range actionStrat {
		for _, t := range strat.Templates {
			if t != templateName {
				continue
			}

			return checkIfDeliverMessageFromState(m, strat.NotificationStrategy)
		}
	}

	return send
}

func checkIfDeliverMessageFromState(m *Message, strategy string) bool {
	var send bool
	switch strategy {
	case utask.NotificationStrategyAlways:
		send = true
	case utask.NotificationStrategyFailureOnly:
		if v, ok := m.Fields["state"]; ok && v == stateBlocked {
			send = true
		}
	case utask.NotificationStrategySilent:
	}

	return send
}
