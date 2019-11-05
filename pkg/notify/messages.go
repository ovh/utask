package notify

import (
	"fmt"
	"strings"
)

// TaskStateUpdate holds a digest of data representing a task state change
type TaskStateUpdate struct {
	Title              string
	PublicID           string
	State              string
	TemplateName       string
	RequesterUsername  string
	ResolverUsername   *string
	PotentialResolvers []string
	StepsDone          int
	StepsTotal         int
}

// Message renders a string representation of a TaskStateUpdate
func (tsu TaskStateUpdate) Message() string {
	return fmt.Sprintf(
		"#task #id:%s %s",
		tsu.PublicID,
		tsu.Title,
	)
}

// Fields returns a list of formatted fields to qualify a message
func (tsu TaskStateUpdate) Fields() []string {
	l := []string{
		fmt.Sprintf("state:%s", tsu.State),
		fmt.Sprintf("template:%s", tsu.TemplateName),
		fmt.Sprintf("steps:%d/%d", tsu.StepsDone, tsu.StepsTotal),
	}

	if tsu.RequesterUsername != "" {
		l = append(l, fmt.Sprintf("requester:%s", tsu.RequesterUsername))
	}
	if tsu.ResolverUsername != nil && *tsu.ResolverUsername != "" {
		l = append(l, fmt.Sprintf("resolver:%s", *tsu.ResolverUsername))
	}
	if tsu.PotentialResolvers != nil && len(tsu.PotentialResolvers) > 0 {
		l = append(l, fmt.Sprintf("potential_resolvers: %s", strings.Join(tsu.PotentialResolvers, " ")))
	}
	return l
}
