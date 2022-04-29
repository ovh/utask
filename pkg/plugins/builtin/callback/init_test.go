package plugincallback

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ovh/configstore"
	"github.com/stretchr/testify/assert"
)

func Test_loadConfig(t *testing.T) {
	tests := []struct {
		utaskCfg      *string
		callbackCfg   *string
		expectedCfg   *CallbackConfig
		expectedError *string
	}{
		{
			callbackCfg: cfg(map[string]interface{}{
				"base_url":    "http://utask.example.com/callbacks/",
				"path_prefix": "path-prefix",
			}),
			expectedCfg: &CallbackConfig{
				BaseURL:    "http://utask.example.com/callbacks",
				PathPrefix: "path-prefix",
			},
		},
		{
			utaskCfg: cfg(map[string]interface{}{
				"base_url": "http://utask.example.com/",
			}),
			expectedCfg: &CallbackConfig{
				BaseURL:    "http://utask.example.com",
				PathPrefix: "",
			},
		},
		{
			expectedError: str("configstore: get %q: no item found", configAlias),
		},
		{
			callbackCfg:   cfg(map[string]interface{}{}),
			expectedError: str("\"base_url\" key not defined in %q", configAlias),
		},
	}

	for _, test := range tests {
		store := configstore.NewStore()
		store.RegisterProvider("tests", configstoreProvider(test.utaskCfg, test.callbackCfg))

		init := NewCallbackInit()
		err := init.loadConfig(store)

		if test.expectedError == nil {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, *test.expectedError)
		}

		if test.expectedCfg != nil {
			assert.Equal(t, test.expectedCfg, &(init.cfg))
		}
	}
}

func configstoreProvider(utaskCfg *string, callbackCfg *string) func() (configstore.ItemList, error) {
	var items []configstore.Item

	if utaskCfg != nil {
		items = append(items, configstore.NewItem("utask-cfg", *utaskCfg, 1))
	}

	if callbackCfg != nil {
		items = append(items, configstore.NewItem("callback-config", *callbackCfg, 1))
	}

	return func() (configstore.ItemList, error) {
		ret := configstore.ItemList{
			Items: items,
		}
		return ret, nil
	}
}

func cfg(v map[string]interface{}) *string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	jsonStr := string(jsonBytes)
	return &jsonStr
}

func str(format string, a ...interface{}) *string {
	ret := fmt.Sprintf(format, a...)
	return &ret
}
