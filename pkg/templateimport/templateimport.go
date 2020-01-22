package templateimport

import (
	"sync"
)

var (
	mu                  sync.Mutex
	discoveredTemplates []string = []string{}
)

// AddTemplate registers a template name currently being imported
func AddTemplate(templateName string) {
	mu.Lock()
	defer mu.Unlock()
	discoveredTemplates = append(discoveredTemplates, templateName)
}

// GetTemplates returns template names currently being imported.
// Used when template needs to validate dependency with others templates (check existance, ...)
func GetTemplates() []string {
	mu.Lock()
	defer mu.Unlock()
	return discoveredTemplates
}

// CleanTemplates flushes the template names list being imported.
func CleanTemplates() {
	mu.Lock()
	defer mu.Unlock()
	discoveredTemplates = []string{}
}
