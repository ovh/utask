package values_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/pkg/utils"
	"sigs.k8s.io/yaml"
)

func TestTmpl(t *testing.T) {
	input := `"step":
    "first":
        "output":
            "result":
                "my-payload": "{\"common-name\":\"utask.example.org\",\"id\":32,\"foo\":{\"bar\":1}}"`
	obj := map[string]map[string]map[string]map[string]interface{}{}
	err := yaml.Unmarshal([]byte(input), &obj)
	td.CmpNil(t, err)

	v := values.NewValues()
	v.SetOutput("first", obj["step"]["first"]["output"])

	output, err := v.Apply("{{ field `step` `first` `output` `result` `my-payload` }}", nil, "foo")
	td.CmpNil(t, err)
	td.Cmp(t, string(output), "{\"common-name\":\"utask.example.org\",\"id\":32,\"foo\":{\"bar\":1}}")

	output, err = v.Apply("{{ field `step` `first` `output` `result` `my-payload` | fromJson | fieldFrom `common-name` }}", nil, "foo")
	td.CmpNil(t, err)
	td.Cmp(t, string(output), "utask.example.org")

	output, err = v.Apply("{{ field `step` `first` `output` `result` `my-payload` | fromJson | fieldFrom `foo` `bar` }}", nil, "foo")
	td.CmpNil(t, err)
	td.Cmp(t, string(output), "1")

	output, err = v.Apply("{{ `{\"common-name\":\"utask.example.org\",\"id\":32}` | fromJson | fieldFrom `invalid` | default `example.org` }}", nil, "foo")
	td.CmpNil(t, err)
	td.Cmp(t, string(output), "example.org")
}

func TestJsonNumber(t *testing.T) {
	input := `
{
  "step": {
    "first": {
      "output": {
          "my-payload": {
            "foo": 0
          }
      }
    }
  }
}
`

	assert, require := td.AssertRequire(t)

	obj := map[string]map[string]map[string]interface{}{}
	err := utils.JSONnumberUnmarshal(strings.NewReader(input), &obj)
	require.Nil(err)

	v := values.NewValues()
	v.SetOutput("first", obj["step"]["first"]["output"])
	v.SetVariables([]values.Variable{
		{
			Name:  "foobar0",
			Value: "ok",
		},
		{
			Name:  "foobar1",
			Value: "{{ `ok` }}",
		},
		{
			Name: "foobar2",
			Expression: `var foo = {"foo":"bar"};
foo.foo;`,
		},
	})

	output := v.GetOutput("first")
	outputMap, ok := output.(map[string]interface{})
	require.True(ok)
	innerOutputMap, ok := outputMap["my-payload"].(map[string]interface{})
	require.True(ok)

	assert.Cmp(fmt.Sprintf("%T", innerOutputMap["foo"]), "json.Number")

	outputba, err := v.Apply("{{ eval `foobar0` }}", nil, "")
	require.Nil(err)
	assert.Cmp(string(outputba), "ok")

	outputba, err = v.Apply("{{ eval `foobar1` }}", nil, "")
	require.Nil(err)
	assert.Cmp(string(outputba), "ok")

	outputba, err = v.Apply("{{ evalCache `foobar2` }}", nil, "")
	require.Nil(err)
	assert.Cmp(string(outputba), "bar")

	newV, err := v.Clone()
	require.Nil(err, "received unexpected error: %s", err)

	newOutput := newV.GetOutput("first")
	newOutputMap, ok := newOutput.(map[string]interface{})
	require.True(ok)
	newInnerOutputMap, ok := newOutputMap["my-payload"].(map[string]interface{})
	require.True(ok)

	assert.Cmp(fmt.Sprintf("%T", newInnerOutputMap["foo"]), "json.Number")

	outputba, err = newV.Apply("{{ eval `foobar0` }}", nil, "")
	require.Nil(err)
	assert.Cmp(string(outputba), "ok")

	outputba, err = newV.Apply("{{ eval `foobar1` }}", nil, "")
	require.Nil(err)
	assert.Cmp(string(outputba), "ok")

	outputba, err = newV.Apply("{{ evalCache `foobar2` }}", nil, "")
	require.Nil(err)
	assert.Cmp(string(outputba), "bar")

	// with iterator now
	iterator := map[string]map[string][]string{
		"foo": map[string][]string{
			"bar": []string{"foobar", "buzz"},
		},
	}
	v.SetIterator(iterator)

	outputba, err = v.Apply("{{ index .iterator.foo.bar 1 }}", nil, "")
	require.Nil(err)
	assert.Cmp(string(outputba), "buzz")

	newV, err = v.Clone()
	require.Nil(err, "received unexpected error: %s", err)

	newOutput = newV.GetOutput("first")
	newOutputMap, ok = newOutput.(map[string]interface{})
	require.True(ok)
	newInnerOutputMap, ok = newOutputMap["my-payload"].(map[string]interface{})
	require.True(ok)

	assert.Cmp(fmt.Sprintf("%T", newInnerOutputMap["foo"]), "json.Number")

	outputba, err = newV.Apply("{{ index .iterator.foo.bar 1 }}", nil, "")
	require.Nil(err)
	assert.Cmp(string(outputba), "buzz")
}
