package notify

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

// MessageFields returns a task state change struct
func (tsu TaskStateUpdate) MessageFields() *TaskStateUpdate {
	return &tsu
}
