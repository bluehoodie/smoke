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

var parseOtt = []struct {
	runner        *Runner
	contract      *Contract
	body          []byte
	description   string
	err           bool
	expectedKey   string
	expectedValue string
}{
	{
		contract: &Contract{
			Outputs: make(map[string]string),
		},
		err:         false,
		description: "It should not return an error if there are no outputs",
	},
	{
		contract: &Contract{
			Outputs: map[string]string{"value": "NOT_MAPPED.whatever"},
		},
		err:         true,
		description: "It should return an error if there are outputs that does not begin by what is expected (JSON)",
	},
	{
		contract: &Contract{
			Outputs: map[string]string{"value": "JSON.A"},
		},
		runner:        &Runner{test: Test{Globals: make(map[string]string)}},
		body:          []byte(`{"A": 1 }`),
		err:           false,
		expectedKey:   "value",
		expectedValue: "1",
		description:   "Runner should have an value as output",
	},
	{
		contract: &Contract{
			Outputs: map[string]string{"value": "JSON.A"},
		},
		runner:      &Runner{test: Test{Globals: make(map[string]string)}},
		body:        []byte(`OBVIOUSLY NOT A JSON`),
		err:         true,
		description: "It should return an error if the body does not match with what is exected",
	},
}

func TestParseOutputs(t *testing.T) {
	for _, tt := range parseOtt {
		//act
		err := parseOutputs(tt.runner, tt.contract, tt.body)

		//assert
		assert.True(t, (err != nil) == tt.err, tt.description)

		if tt.expectedKey != "" && err == nil {
			assert.Contains(t, tt.runner.test.Globals, tt.expectedKey, tt.description)
			assert.Equal(t, tt.expectedValue, tt.runner.test.Globals[tt.expectedKey], tt.description)
		}

	}

}
