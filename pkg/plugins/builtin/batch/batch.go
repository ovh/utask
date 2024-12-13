package pluginbatch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	jujuErrors "github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/batch"
	"github.com/ovh/utask/pkg/batchutils"
	"github.com/ovh/utask/pkg/constants"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
	"github.com/ovh/utask/pkg/templateimport"
	"github.com/ovh/utask/pkg/utils"
)

// The batch plugin spawns X new µTask tasks, given a template and inputs, and waits for them to be completed.
// Resolver usernames can be dynamically set for the task
var Plugin = taskplugin.New(
	"batch",
	"0.1",
	exec,
	taskplugin.WithConfig(validateConfigBatch, BatchConfig{}),
	taskplugin.WithContextFunc(ctxBatch),
)

// BatchConfig is the necessary configuration to spawn a new task
type BatchConfig struct {
	TemplateName      string                   `json:"template_name" binding:"required"`
	CommonInputs      map[string]interface{}   `json:"common_inputs"`
	CommonJSONInputs  string                   `json:"common_json_inputs"`
	Inputs            []map[string]interface{} `json:"inputs"`
	JSONInputs        string                   `json:"json_inputs"`
	Comment           string                   `json:"comment"`
	WatcherUsernames  []string                 `json:"watcher_usernames"`
	WatcherGroups     []string                 `json:"watcher_groups"`
	Tags              map[string]string        `json:"tags"`
	ResolverUsernames string                   `json:"resolver_usernames"`
	ResolverGroups    string                   `json:"resolver_groups"`
	// How many tasks will run concurrently. 0 for infinity (default). It's supplied as a string to support templating
	SubBatchSizeStr string `json:"sub_batch_size"`
	SubBatchSize    int64  `json:"-"`
}

// quotedString is a string with doubly escaped quotes, so the string stays simply escaped after being processed
// as the plugin's context (see ctxBatch).
type quotedString string

// BatchContext holds data about the parent task execution as well as the metadata of previous runs, if any.
type BatchContext struct {
	ParentTaskID      string `json:"parent_task_id"`
	RequesterUsername string `json:"requester_username"`
	RequesterGroups   string `json:"requester_groups"`
	// RawMetadata of the previous run. Metadata are used to communicate batch progress between runs. It's returned
	// "as is" in case something goes wrong in a subsequent run, to know what the batch's progress was when the
	// error occured.
	RawMetadata quotedString `json:"metadata"`
	// Unmarshalled version of the metadata
	metadata BatchMetadata
	StepName string `json:"step_name"`
}

// BatchMetadata holds batch-progress data, communicated between each run of the plugin.
type BatchMetadata struct {
	BatchID        string `json:"batch_id"`
	RemainingTasks int64  `json:"remaining_tasks"`
	TasksStarted   int64  `json:"tasks_started"`
}

func ctxBatch(stepName string) interface{} {
	return &BatchContext{
		ParentTaskID:      "{{ .task.task_id }}",
		RequesterUsername: "{{.task.requester_username}}",
		RequesterGroups:   "{{ if .task.requester_groups }}{{ .task.requester_groups }}{{ end }}",
		RawMetadata: quotedString(fmt.Sprintf(
			"{{ if (index .step `%s` ) }}{{ if (index .step `%s` `metadata`) }}{{ index .step `%s` `metadata` }}{{ end }}{{ end }}",
			stepName,
			stepName,
			stepName,
		)),
		StepName: stepName,
	}
}

func validateConfigBatch(config any) error {
	conf := config.(*BatchConfig)

	if err := utils.ValidateTags(conf.Tags); err != nil {
		return err
	}

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return fmt.Errorf("can't retrieve connection to DB: %s", err)
	}

	_, err = tasktemplate.LoadFromName(dbp, conf.TemplateName)
	if err != nil {
		if !jujuErrors.IsNotFound(err) {
			return fmt.Errorf("can't load template from name: %s", err)
		}

		// searching into currently imported templates
		templates := templateimport.GetTemplates()
		for _, template := range templates {
			if template == conf.TemplateName {
				return nil
			}
		}

		return jujuErrors.NotFoundf("batch template %q", conf.TemplateName)
	}

	return nil
}

func exec(stepName string, config any, ictx any) (any, any, error) {
	var metadata BatchMetadata
	var stepError error

	conf := config.(*BatchConfig)
	batchCtx := ictx.(*BatchContext)
	if err := parseInputs(conf, batchCtx); err != nil {
		return nil, batchCtx.RawMetadata.Format(), err
	}

	if len(conf.Inputs) == 0 {
		// Empty input, there's nothing to do
		return nil, BatchMetadata{}, nil
	}

	if conf.Tags == nil {
		conf.Tags = make(map[string]string)
	}
	conf.Tags[constants.SubtaskTagParentTaskID] = batchCtx.ParentTaskID

	ctx := auth.WithIdentity(context.Background(), batchCtx.RequesterUsername)
	requesterGroups := strings.Split(batchCtx.RequesterGroups, utask.GroupsSeparator)
	ctx = auth.WithGroups(ctx, requesterGroups)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, batchCtx.RawMetadata.Format(), err
	}

	if err := dbp.Tx(); err != nil {
		return nil, batchCtx.RawMetadata.Format(), err
	}

	if batchCtx.metadata.BatchID == "" {
		// The batch needs to be started
		metadata, err = startBatch(ctx, dbp, conf, batchCtx)
		if err != nil {
			dbp.Rollback()
			return nil, nil, err
		}

		// A step returning a NotAssigned error is set to WAITING by the engine
		stepError = jujuErrors.NewNotAssigned(fmt.Errorf("tasks from batch %q will start shortly", metadata.BatchID), "")
	} else {
		// Batch already started, we either need to start new tasks or check whether they're all done
		metadata, err = runBatch(ctx, conf, batchCtx, dbp)
		if err != nil {
			dbp.Rollback()
			return nil, batchCtx.RawMetadata.Format(), err
		}

		if metadata.RemainingTasks != 0 {
			// A step returning a NotAssigned error is set to WAITING by the engine
			stepError = jujuErrors.NewNotAssigned(fmt.Errorf("batch %q is currently RUNNING", metadata.BatchID), "")
		} else {
			// The batch is done.
			// We increase the resolution's maximum amount of retries to compensate for the amount of runs consumed
			// by child tasks waking up the parent when they're done.
			err := increaseRunMax(dbp, batchCtx.ParentTaskID, batchCtx.StepName)
			if err != nil {
				return nil, batchCtx.RawMetadata.Format(), err
			}
		}
	}

	formattedMetadata, err := formatOutput(metadata)
	if err != nil {
		dbp.Rollback()
		return nil, batchCtx.RawMetadata.Format(), err
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return nil, batchCtx.RawMetadata.Format(), err
	}
	return nil, formattedMetadata, stepError
}

// startBatch creates a batch of tasks as described in the given batchArgs.
func startBatch(
	ctx context.Context,
	dbp zesty.DBProvider,
	conf *BatchConfig,
	batchCtx *BatchContext,
) (BatchMetadata, error) {
	b, err := task.CreateBatch(dbp)
	if err != nil {
		return BatchMetadata{}, err
	}

	taskIDs, err := populateBatch(ctx, b, dbp, conf, batchCtx)
	if err != nil {
		return BatchMetadata{}, err
	}

	return BatchMetadata{
		BatchID:        b.PublicID,
		RemainingTasks: int64(len(conf.Inputs)),
		TasksStarted:   int64(len(taskIDs)),
	}, nil
}

// populateBatch spawns new tasks in the batch and returns their public identifier.
func populateBatch(
	ctx context.Context,
	b *task.Batch,
	dbp zesty.DBProvider,
	conf *BatchConfig,
	batchCtx *BatchContext,
) ([]string, error) {
	tasksStarted := batchCtx.metadata.TasksStarted
	running, err := batchutils.RunningTasks(dbp, b.ID)
	if err != nil {
		return []string{}, err
	}

	// Computing how many tasks to start
	remaining := int64(len(conf.Inputs)) - tasksStarted
	toStart := conf.SubBatchSize - running // How many tasks can be started
	if remaining < toStart || conf.SubBatchSize == 0 {
		// There's less tasks remaining to start than the amount of available running slots or slots are unlimited
		toStart = remaining
	}

	args := batch.TaskArgs{
		TemplateName:     conf.TemplateName,
		CommonInput:      conf.CommonInputs,
		Inputs:           conf.Inputs[tasksStarted : tasksStarted+toStart],
		Comment:          conf.Comment,
		WatcherGroups:    conf.WatcherGroups,
		WatcherUsernames: conf.WatcherUsernames,
		Tags:             conf.Tags,
	}

	taskIDs, err := batch.Populate(ctx, b, dbp, args)
	if err != nil {
		return []string{}, err
	}

	return taskIDs, nil
}

// runBatch runs a batch, spawning new tasks if needed and checking whether they're all done.
func runBatch(
	ctx context.Context,
	conf *BatchConfig,
	batchCtx *BatchContext,
	dbp zesty.DBProvider,
) (BatchMetadata, error) {
	metadata := batchCtx.metadata

	b, err := task.LoadBatchFromPublicID(dbp, metadata.BatchID)
	if err != nil {
		if !jujuErrors.Is(err, jujuErrors.NotFound) {
			return metadata, err
		}
		// else, the batch has been collected (deleted in DB) because no task referenced it anymore.

		if metadata.TasksStarted == int64(len(conf.Inputs)) {
			// There is no more tasks to create, the work is done
			metadata.RemainingTasks = 0
			return metadata, nil
		}
		// else, the batch was collected but we still have tasks to create. We need to recreate the batch with
		// the same public ID and populate it.
		// It can happen when the garbage collector runs after a sub-batch is done, but before the batch plugin
		// could populate the batch with more tasks.

		b = &task.Batch{BatchDBModel: task.BatchDBModel{PublicID: metadata.BatchID}}
		if err := dbp.DB().Insert(&b.BatchDBModel); err != nil {
			return metadata, pgjuju.Interpret(err)
		}
	}

	if metadata.TasksStarted < int64(len(conf.Inputs)) {
		// New tasks still need to be added to the batch

		taskIDs, err := populateBatch(ctx, b, dbp, conf, batchCtx)
		if err != nil {
			return metadata, err
		}

		started := int64(len(taskIDs))
		metadata.TasksStarted += started
		metadata.RemainingTasks -= started // Starting X tasks means that X tasks became DONE
		return metadata, nil
	}
	// else, all tasks are started, we need to wait for the last ones to become DONE

	running, err := batchutils.RunningTasks(dbp, b.ID)
	if err != nil {
		return metadata, err
	}
	metadata.RemainingTasks = running
	return metadata, nil
}

// increaseRunMax increases the maximum amount of runs of the resolution matching the given parentTaskID by the run
// count of the given batchStepName.
// Since child tasks wake their parent up when they're done, the resolution's RunCount gets incremented everytime. We
// compensate this by increasing the RunMax property once the batch is done.
func increaseRunMax(dbp zesty.DBProvider, parentTaskID string, batchStepName string) error {
	t, err := task.LoadFromPublicID(dbp, parentTaskID)
	if err != nil {
		return err
	}

	if t.Resolution == nil {
		return fmt.Errorf("resolution not found for step '%s' of task '%s'", batchStepName, parentTaskID)
	}

	res, err := resolution.LoadLockedFromPublicID(dbp, *t.Resolution)
	if err != nil {
		return err
	}

	step, ok := res.Steps[batchStepName]
	if !ok {
		return fmt.Errorf("step '%s' not found in resolution", batchStepName)
	}

	res.ExtendRunMax(step.TryCount)
	return res.Update(dbp)
}

// parseInputs parses the step's inputs as well as metadata from the previous run (if it exists).
func parseInputs(conf *BatchConfig, batchCtx *BatchContext) error {
	if batchCtx.RawMetadata != "" {
		// Metadata from a previous run is available
		if err := json.Unmarshal([]byte(batchCtx.RawMetadata), &batchCtx.metadata); err != nil {
			return jujuErrors.NewBadRequest(err, "metadata unmarshalling failure")
		}
	}

	if conf.CommonJSONInputs != "" {
		if err := json.Unmarshal([]byte(conf.CommonJSONInputs), &conf.CommonInputs); err != nil {
			return jujuErrors.NewBadRequest(err, "JSON common input unmarshalling failure")
		}
	}

	if conf.JSONInputs != "" {
		if err := json.Unmarshal([]byte(conf.JSONInputs), &conf.Inputs); err != nil {
			return jujuErrors.NewBadRequest(err, "JSON inputs unmarshalling failure")
		}
	}

	if conf.SubBatchSizeStr == "" {
		conf.SubBatchSize = 0
	} else {
		subBatchSize, err := strconv.ParseInt(conf.SubBatchSizeStr, 10, 64)
		if err != nil {
			return jujuErrors.NewBadRequest(err, "parsing failure of field 'SubBatchSize'")
		}
		conf.SubBatchSize = subBatchSize
	}

	return nil
}

// Format formats the utaskString to make sure it's parsable by subsequent runs of the plugin (i.e.: escaping
// double quotes).
func (rm quotedString) Format() string {
	return strings.ReplaceAll(string(rm), `"`, `\"`)
}

// formatOutput formats an output (plugin output or metadata) as a uTask-friendly output.
func formatOutput(result any) (string, error) {
	marshalled, err := json.Marshal(result)
	if err != nil {
		logrus.WithError(err).Error("Couldn't marshal batch metadata")
		return "", err
	}
	return quotedString(marshalled).Format(), nil
}
