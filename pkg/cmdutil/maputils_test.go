package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNestedStringField(t *testing.T) {
	type output struct {
		ok    bool
		value string
	}

	tests := []struct {
		name      string
		inputMap  map[string]interface{}
		selectors []string
		output    struct {
			ok    bool
			value string
		}
	}{
		{
			name: "successfulNestedFetch",
			inputMap: map[string]interface{}{
				"key1": map[string]interface{}{
					"key2": "value",
				},
			},
			selectors: []string{"key1", "key2"},
			output: output{
				ok:    true,
				value: "value",
			},
		},
		{
			name:      "emptyMapFailedFetch",
			inputMap:  map[string]interface{}{},
			selectors: []string{"key1", "key2"},
			output: output{
				ok:    false,
				value: "",
			},
		},
		{
			name: "keyDoesNotExistFailedFetch",
			inputMap: map[string]interface{}{
				"key1": map[string]interface{}{
					"key2": "value",
				},
			},
			selectors: []string{"key1", "notAKey"},
			output: output{
				ok:    false,
				value: "",
			},
		},
		{
			name: "keyExistsButIsNotAString",
			inputMap: map[string]interface{}{
				"key1": map[string]interface{}{
					"key2": []string{"value"},
				},
			},
			selectors: []string{"key1", "key2"},
			output: output{
				ok:    false,
				value: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := GetNestedStringField(tt.inputMap, tt.selectors...)
			assert.Equal(t, value, tt.output.value)
			assert.Equal(t, ok, tt.output.ok)
		})
	}

}
