package values

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookOutput(t *testing.T) {
	var v = NewValues()
	var myOutputValue = map[string]string{"foo": "bar"}

	v.SetHookOutput("myStep", "myHook", myOutputValue)

	actual := v.GetHookOutput("myStep", "myHook")
	assert.EqualValues(t, myOutputValue, actual)
}

func TestHookMetadata(t *testing.T) {
	var v = NewValues()
	var myOutputValue = map[string]string{"foo": "bar"}

	v.SetHookMetadata("myStep", "myHook", myOutputValue)
	actual := v.GetHookMetadata("myStep", "myHook")

	assert.EqualValues(t, myOutputValue, actual)
}

func TestHookResults(t *testing.T) {
	var v = NewValues()
	var myResultsValue = map[string]string{"foo": "bar"}

	v.SetHookResults("myStep", "myHook", myResultsValue)
	actual := v.GetHookResults("myStep", "myHook")

	assert.EqualValues(t, myResultsValue, actual)
}

func TestStepHooks(t *testing.T) {
	var v = NewValues()
	var myOutputValue = map[string]string{"foo": "bar"}
	v.SetHookMetadata("myStep", "myHook", myOutputValue)
	expected := map[string]interface{}{"myHook": map[string]interface{}{"metadata": map[string]string{"foo": "bar"}}}

	a := v.getStepHooks("myStep")
	assert.EqualValues(t, expected, a)
	v.setStepHooks("myStep", a)
	a = v.getStepHooks("myStep")
	assert.EqualValues(t, expected, a)
	v.unsetStepHooks("myStep")
	a = v.getStepHooks("myStep")
	assert.Nil(t, a)
}
