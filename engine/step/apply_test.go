package step

import (
	"encoding/json"
	"testing"

	"github.com/maxatome/go-testdeep/td"

	"github.com/ovh/utask/engine/values"
)

func TestApply(t *testing.T) {
	v := values.NewValues()

	v.SetInput(map[string]interface{}{
		"ifoo1": "Hello",
		"ifoo2": true,
		"ifoo3": 12,
		"ifoo4": []string{"Hello", "you!"},
	})

	h := json.RawMessage(`{"foo": "{{ .input.ifoo1}} -> {{.input.ifoo3 }}"}`)
	result, err := resolveObject(v, h, nil, "")
	if !td.CmpNoError(t, err) {
		return
	}
	expected := `{"foo":"Hello -> 12"}`
	td.Cmp(t, string(result), expected)

	h = json.RawMessage(`{"foo": "{{ .input.ifoo1}} -> {{.input.ifoo3.notexisting }}"}`)
	result, err = resolveObject(v, h, nil, "")
	if !td.CmpError(t, err) {
		return
	}
	td.CmpContains(t, err.Error(), "can't evaluate field notexisting in type interface")
}

func TestApplyNested(t *testing.T) {
	assert, require := td.AssertRequire(t)

	v := values.NewValues()

	v.SetInput(map[string]interface{}{
		"ifoo1": "Hello",
		"ifoo2": true,
		"ifoo3": 12,
		"ifoo4": []string{"Hello", "you!"},
	})

	h := json.RawMessage(`{"fooo":{"foo": "{{ .input.ifoo1}} -> {{.input.ifoo3 }}","array":["{{.input.ifoo2}}"]}}`)
	result, err := resolveObject(v, h, nil, "")
	require.CmpNoError(err)

	expected := `{"fooo":{"array":["true"],"foo":"Hello -> 12"}}`
	assert.Cmp(string(result), expected)

	h = json.RawMessage(`{"fooo":{"foo": "{{ .input.ifoo1}} -> {{.input.ifoo3.notexisting }}"}}`)
	result, err = resolveObject(v, h, nil, "")
	require.CmpError(err)
	assert.Contains(err.Error(), "can't evaluate field notexisting in type interface")
}

func TestApplyArray(t *testing.T) {
	assert, require := td.AssertRequire(t)

	v := values.NewValues()

	v.SetInput(map[string]interface{}{
		"ifoo1": "Hello",
		"ifoo2": true,
		"ifoo3": 12,
		"ifoo4": []string{"Hello", "you!"},
	})

	h := json.RawMessage(`["fooo",{"foo": "{{ .input.ifoo1}} -> {{.input.ifoo3 }}","array":["{{.input.ifoo2}}"]}]`)
	result, err := resolveObject(v, h, nil, "")
	require.CmpNoError(err)

	expected := `["fooo",{"array":["true"],"foo":"Hello -> 12"}]`
	assert.Cmp(string(result), expected)

	h = json.RawMessage(`["fooo",{"foo": "{{ .input.ifoo1}} -> {{.input.ifoo3.notexisting }}"}]`)
	result, err = resolveObject(v, h, nil, "")
	require.CmpError(err)
	assert.Contains(err.Error(), "can't evaluate field notexisting in type interface")
}
