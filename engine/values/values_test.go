package values_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/utask/engine/values"
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
