package tester

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var extractmaptt = []struct {
	m           map[string]interface{}
	key         string
	expected    interface{}
	description string
}{
	{
		description: "should send nil if there are no map",
		m:           nil,
		key:         "",
		expected:    nil,
	},
	{
		description: "should send nil if there are no '[' on the key",
		m:           make(map[string]interface{}),
		key:         "param",
		expected:    nil,
	},
	{
		description: "should send nil if there is no number between [] on the key",
		m:           make(map[string]interface{}),
		key:         "param[]",
		expected:    nil,
	},
	{
		description: "should send nil if there is no number formated string between [] on the key",
		m:           make(map[string]interface{}),
		key:         "param[abc]",
		expected:    nil,
	},
	{
		description: "should send nil if there is no key present on the map",
		m: map[string]interface{}{
			"param1": nil,
		},
		key:      "param[0]",
		expected: nil,
	},
	{
		description: "should send something if all is correct",
		m: map[string]interface{}{
			"param": []interface{}{"test"},
		},
		key:      "param[0]",
		expected: interface{}("test"),
	},
}

func TestExtractValueFromJSONMap(t *testing.T) {
	for _, tt := range extractmaptt {
		//act
		result := extractValueFromJSONMap(tt.key, tt.m)

		//arrange
		assert.Equal(t, tt.expected, result, tt.description)
	}
}

var jsonParsertt = []struct {
	json          string
	conf          string
	keys          []string
	expectedValue string
	expectedError bool
	description   string
}{
	{
		description:   "should have an error when no field parameter",
		conf:          "A",
		keys:          []string{},
		json:          `{ "A":1 }`,
		expectedError: true,
		expectedValue: "",
	},
	{
		description:   "should have an error when field is nil",
		conf:          "A",
		keys:          nil,
		json:          `{ "A":1 }`,
		expectedError: true,
		expectedValue: "",
	},
	{
		description:   "should have an error when no body parameter",
		conf:          "A",
		keys:          []string{"A"},
		json:          "",
		expectedError: true,
		expectedValue: "",
	},
	{
		description:   "should have an error when no json",
		json:          "Hello World",
		conf:          "A",
		keys:          []string{"A"},
		expectedError: true,
		expectedValue: "",
	},
	{
		description:   "should have an error when value is not present in the middle of the json",
		conf:          "C.D",
		keys:          []string{"C", "D"},
		json:          `{ "A": { "B": 1 } }`,
		expectedError: true,
		expectedValue: "",
	},
	{
		description:   "should have an error when the final value is not present in the json",
		conf:          "A.C",
		keys:          []string{"A", "C"},
		json:          `{ "A": { "B": 1 } }`,
		expectedError: true,
		expectedValue: "",
	},
	{
		description:   "should works when the format is A",
		conf:          "A",
		keys:          []string{"A"},
		json:          `{ "A":1 }`,
		expectedError: false,
		expectedValue: "1",
	},
	{
		description:   "should works when the format is A.B",
		conf:          "A.B",
		keys:          []string{"A", "B"},
		json:          `{ "A": { "B": 1 } }`,
		expectedError: false,
		expectedValue: "1",
	},
	{
		description:   "should works when the format is A.B[0]",
		conf:          "A.B[0]",
		keys:          []string{"A", "B[0]"},
		json:          `{ "A": { "B": [1,2,3] } }`,
		expectedError: false,
		expectedValue: "1",
	},
	{
		description:   "should works when the format is A.B[1].C",
		conf:          "A.B[1].C",
		keys:          []string{"A", "B[1]", "C"},
		json:          `{ "A": { "B": [{"C":1},{"C":2}] } }`,
		expectedError: false,
		expectedValue: "2",
	},
	{
		description:   "should works when the format is [2]",
		conf:          "[2]",
		keys:          []string{"[2]"},
		json:          `[1,2,3]`,
		expectedError: false,
		expectedValue: "3",
	},
	{
		description:   "should works when the format is [0].A",
		conf:          "[0].A",
		keys:          []string{"[0]", "A"},
		json:          `[ {"A":1}, {"A":2} ]`,
		expectedError: false,
		expectedValue: "1",
	},
	{
		description:   "should send an error when the format is not a json but begins by [",
		conf:          "[0]",
		keys:          []string{"[0]"},
		json:          `[HelloWorld]`,
		expectedError: true,
		expectedValue: "",
	},
}

func TestJsonParser(t *testing.T) {
	for _, tt := range jsonParsertt {
		//act
		v, err := jsonParser(tt.conf, tt.keys, []byte(tt.json))

		//assert
		assert.True(t, (err != nil) == tt.expectedError, tt.description)
		assert.Equal(t, tt.expectedValue, v, tt.description)
	}
}

var extractarraytt = []struct {
	key         string
	arr         []interface{}
	expected    interface{}
	description string
}{
	{
		description: "should send nil, if no array sent",
		key:         "key",
		arr:         nil,
		expected:    nil,
	},
	{
		description: "should send nil, if no '[' present on the key",
		key:         "key",
		arr:         make([]interface{}, 1),
		expected:    nil,
	},
	{
		description: "should works when all is correct",
		key:         "[0]",
		arr: []interface{}{
			"Hello",
		},
		expected: "Hello",
	},
}

func TestExtractValueFromJSONArray(t *testing.T) {
	for _, tt := range extractarraytt {
		//act
		v := extractValueFromJSONArray(tt.key, tt.arr)

		//assert
		assert.Equal(t, tt.expected, v, tt.description)
	}
}
