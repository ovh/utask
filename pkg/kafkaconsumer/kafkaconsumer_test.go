package kafkaconsumer

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ghodss/yaml"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/configstore"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/plugins/builtin/echo"
	"github.com/ovh/utask/pkg/plugins/builtin/script"
)

func TestMain(m *testing.M) {
	store := configstore.DefaultStore
	store.InitFromEnvironment()

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.ErrorLevel)

	step.RegisterRunner(echo.Plugin.PluginName(), echo.Plugin)
	step.RegisterRunner(script.Plugin.PluginName(), script.Plugin)

	if err := db.Init(store); err != nil {
		panic(err)
	}

	if err := now.Init(); err != nil {
		panic(err)
	}

	if err := auth.Init(store); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func loadTemplates(t *testing.T, dbp zesty.DBProvider) error {
	templateList := map[string][]byte{}
	files, err := ioutil.ReadDir("./templates_tests")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.Mode().IsRegular() {
			bytesValue, err := os.ReadFile(filepath.Join("./templates_tests", file.Name()))
			if err != nil {
				panic(err)
			}
			templateList[file.Name()] = bytesValue

			var tmpl tasktemplate.TaskTemplate

			if err := yaml.Unmarshal(bytesValue, &tmpl); err != nil {
				return err
			}
			if err := tmpl.Valid(); err != nil {
				return err
			}
			tmpl.Normalize()
			if err := dbp.DB().Insert(&tmpl); err != nil {
				intErr := pgjuju.Interpret(err)
				if !errors.IsAlreadyExists(intErr) {
					return intErr
				}
				existing, err := tasktemplate.LoadFromName(dbp, tmpl.Name)
				if err != nil {
					return err
				}
				tmpl.ID = existing.ID
				if _, err := dbp.DB().Update(&tmpl); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func TestKafkaConsumer(t *testing.T) {
	assert, require := td.AssertRequire(t)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	require.CmpNoError(err)
	err = loadTemplates(t, dbp)
	require.CmpNoError(err)

	ctx, cancel := context.WithCancel(context.Background())

	c, err := StartNewTaskConsumer(ctx, KafkaConfig{
		Brokers:      []string{"localhost:123"},
		KafkaVersion: "2.8.1",
		Group:        "my-consumer-group",
		Topics:       []string{"my-topic"},

		TaskTemplate:      "kafka-task-template",
		RequesterUsername: "rb",
		Input: map[string]interface{}{
			"quantity": json.Number("12"),
			"foo":      "hello you",
		},
	})
	require.CmpNoError(err)

	consumer := MockKafka{}
	c.SetCustomConsumer(&consumer)
	c.StartConsumer(ctx)

	template := "kafka-task-template"
	filter := task.ListFilter{
		Template: &template,
		PageSize: 1,
	}

	time.Sleep(time.Second)
	cancel()

	time.Sleep(time.Second)
	tasks, err := task.ListTasks(dbp, filter)
	require.CmpNoError(err)
	assert.Len(tasks, 1)

	task, err := task.LoadFromID(dbp, tasks[0].ID)
	require.CmpNoError(err)
	assert.Cmp(task.Input["quantity"], json.Number("12"))
	assert.Cmp(task.Input["foo"], "hello you")
}

type MockKafka struct{}

func (mk *MockKafka) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	mgcs := MockConsumerGroupSession{
		ctx: ctx,
	}
	mcgc := MockConsumerGroupClaim{
		msgs: make(chan *sarama.ConsumerMessage),
	}
	handler.Setup(&mgcs)
	go handler.ConsumeClaim(&mgcs, &mcgc)
	mcgc.msgs <- &sarama.ConsumerMessage{
		Topic:     "mytopic",
		Partition: 1,
		Offset:    2,
		Value:     []byte("foobar coco"),
	}
	handler.Cleanup(&mgcs)
	time.Sleep(20 * time.Second)
	return nil
}

// Errors returns a read channel of errors that occurred during the consumer life-cycle.
// By default, errors are logged and not returned over this channel.
// If you want to implement any custom error handling, set your config's
// Consumer.Return.Errors setting to true, and read from this channel.
func (mk *MockKafka) Errors() <-chan error {
	return make(chan error)
}

// Close stops the ConsumerGroup and detaches any running sessions. It is required to call
// this function before the object passes out of scope, as it will otherwise leak memory.
func (mk *MockKafka) Close() error {
	return nil
}
func (mk *MockKafka) Pause(partitions map[string][]int32)  {}
func (mk *MockKafka) Resume(partitions map[string][]int32) {}
func (mk *MockKafka) PauseAll()                            {}
func (mk *MockKafka) ResumeAll()                           {}

// ConsumerGroupSession represents a consumer group member session.
type MockConsumerGroupSession struct {
	ctx context.Context
}

// Claims returns information about the claimed partitions by topic.
func (mcgs *MockConsumerGroupSession) Claims() map[string][]int32 {
	return map[string][]int32{}
}

// MemberID returns the cluster member ID.
func (mcgs *MockConsumerGroupSession) MemberID() string {
	return "foobar"
}

// GenerationID returns the current generation ID.
func (mcgs *MockConsumerGroupSession) GenerationID() int32 {
	return 42
}

// MarkOffset marks the provided offset, alongside a metadata string
// that represents the state of the partition consumer at that point in time. The
// metadata string can be used by another consumer to restore that state, so it
// can resume consumption.
//
// To follow upstream conventions, you are expected to mark the offset of the
// next message to read, not the last message read. Thus, when calling `MarkOffset`
// you should typically add one to the offset of the last consumed message.
//
// Note: calling MarkOffset does not necessarily commit the offset to the backend
// store immediately for efficiency reasons, and it may never be committed if
// your application crashes. This means that you may end up processing the same
// message twice, and your processing should ideally be idempotent.
func (mcgs *MockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
}

// Commit the offset to the backend
//
// Note: calling Commit performs a blocking synchronous operation.
func (mcgs *MockConsumerGroupSession) Commit() {}

// ResetOffset resets to the provided offset, alongside a metadata string that
// represents the state of the partition consumer at that point in time. Reset
// acts as a counterpart to MarkOffset, the difference being that it allows to
// reset an offset to an earlier or smaller value, where MarkOffset only
// allows incrementing the offset. cf MarkOffset for more details.
func (mcgs *MockConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
}

// MarkMessage marks a message as consumed.
func (mcgs *MockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {}

// Context returns the session context.
func (mcgs *MockConsumerGroupSession) Context() context.Context {
	return mcgs.ctx
}

// ConsumerGroupClaim processes Kafka messages from a given topic and partition within a consumer group.
type MockConsumerGroupClaim struct {
	msgs chan *sarama.ConsumerMessage
}

func (mcgc *MockConsumerGroupClaim) Topic() string {
	return "foobar-topic"
}

func (mcgc *MockConsumerGroupClaim) Partition() int32 {
	return 41
}

func (mcgc *MockConsumerGroupClaim) InitialOffset() int64 {
	return 43
}

// HighWaterMarkOffset returns the high water mark offset of the partition,
// i.e. the offset that will be used for the next message that will be produced.
// You can use this to determine how far behind the processing is.
func (mcgc *MockConsumerGroupClaim) HighWaterMarkOffset() int64 {
	return 23
}

// Messages returns the read channel for the messages that are returned by
// the broker. The messages channel will be closed when a new rebalance cycle
// is due. You must finish processing and mark offsets within
// Config.Consumer.Group.Session.Timeout before the topic/partition is eventually
// re-assigned to another group member.
func (mcgc *MockConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	return mcgc.msgs
}
