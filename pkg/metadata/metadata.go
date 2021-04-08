package metadata

import "github.com/gin-gonic/gin"

const (
	ActionMetadataKey = "action-metadata"

	TaskID       = "task_id"
	TemplateName = "template_name"
	ResolutionID = "resolution_id"
	StepName     = "step_name"
	OldState     = "old_state"
	NewState     = "new_state"
	FunctionName = "function_name"
	CommentID    = "comment_id"
	BatchID      = "batch_id"
)

func AddActionMetadata(c *gin.Context, name string, value interface{}) {
	addMetadata(c, ActionMetadataKey, name, value)
}

func addMetadata(c *gin.Context, metadataKey, name string, value interface{}) {
	i, _ := c.Get(metadataKey)
	m, ok := i.(map[string]interface{})
	if !ok {
		m = map[string]interface{}{}
	}
	m[name] = value
	c.Set(metadataKey, m)
}

func GetActionMetadata(c *gin.Context) map[string]string {
	return getMetadata(c, ActionMetadataKey)
}

func getMetadata(c *gin.Context, metadataKey string) map[string]string {
	i, _ := c.Get(metadataKey)
	m, ok := i.(map[string]string)
	if !ok {
		return nil
	}
	return m
}

func SetSUDO(c *gin.Context) {
	c.Set("sudo", true)
}

func IsSUDO(c *gin.Context) bool {
	bs, _ := c.Get("sudo")
	b, ok := bs.(bool)
	if ok {
		return b
	}
	return false
}
