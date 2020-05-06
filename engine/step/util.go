package step

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Runner represents a component capable of executing a specific action,
// provided a configuration and a context
type Runner interface {
	Exec(stepName string, baseConfig json.RawMessage, config json.RawMessage, ctx interface{}) (interface{}, interface{}, map[string]string, error)
	ValidConfig(baseConfig json.RawMessage, config json.RawMessage) error
	Context(stepName string) interface{}
	Resources(baseConfig json.RawMessage, config json.RawMessage) []string
	MetadataSchema() json.RawMessage
}

var (
	runners     = map[string]Runner{}
	runnerslock sync.RWMutex
)

// RegisterRunner makes a named runner available for use in a Step's configuration
func RegisterRunner(name string, r Runner) error {
	runnerslock.Lock()
	_, exists := runners[name]
	if exists {
		return fmt.Errorf("Step executor conflict! '%s' executor registered twice", name)
	}
	runners[name] = r
	runnerslock.Unlock()
	return nil
}

func getRunner(t string) (Runner, error) {
	runnerslock.RLock()
	defer runnerslock.RUnlock()
	r, ok := runners[t]
	if !ok {
		return nil, fmt.Errorf("No runner for type '%s'", t)
	}
	return r, nil
}
