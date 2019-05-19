package tester

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	variableRegex *regexp.Regexp
)

func init() {
	re, err := regexp.Compile(`::([\w]+)::`)
	if err != nil {
		panic(err)
	}
	variableRegex = re
}

func parseVariables(runner *Runner, contract *Contract) error {
	//parse path
	parsedPath, err := replaceVariables(runner, contract, contract.Path)
	if err != nil {
		return errors.Wrap(err, "could not parse path")
	}
	contract.Path = parsedPath

	//parse body
	parsedBody, err := replaceVariables(runner, contract, contract.Body)
	if err != nil {
		return errors.Wrap(err, "could not parse body")
	}
	contract.Body = parsedBody

	//parse headers
	for key, value := range contract.Headers {
		parsedValue, err := replaceVariables(runner, contract, value)
		if err != nil {
			return errors.Wrap(err, "could not parse header value")
		}
		contract.Headers[key] = parsedValue
	}

	return nil
}

func replaceVariables(runner *Runner, contract *Contract, s string) (string, error) {
	matched := variableRegex.FindAllString(s, -1)
	if len(matched) == 0 {
		return s, nil
	}

	for _, match := range matched {
		variableName := strings.Trim(match, "::")

		var replacement string
		var found bool

		if val, ok := contract.Locals[variableName]; ok {
			replacement = val
			found = true
		}

		if val, ok := runner.test.Globals[variableName]; !found && ok {
			replacement = val
			found = true
		}

		if !found {
			if val := os.Getenv(strings.ToUpper(variableName)); val != "" {
				replacement = val
				found = true
			}
		}

		if found {
			s = strings.Replace(s, match, replacement, -1)
		} else {
			return s, fmt.Errorf("value for variable %v not found", variableName)
		}
	}

	return s, nil
}

func parseOutputs(runner *Runner, contract *Contract, body []byte) (err error) {
	for key, value := range contract.Outputs {
		s := strings.Split(value, ".")
		var result string
		if len(s) > 1 {
			switch strings.ToLower(s[0]) {
			case "json":
				result, err = parseJSON(value, s[1:], body)
				if err != nil {
					return
				}
			default:
				return fmt.Errorf("value for variable %v not parsable", key)
			}
		}
		runner.test.Globals[key] = result
	}
	return nil
}

func parseJSON(format string, fields []string, body []byte) (value string, err error) {
	if body == nil || fields == nil || len(fields) == 0 {
		return "", fmt.Errorf("bad parameter")
	}

	begin := 0
	var next interface{}

	// If it begins by a [], we should extract the info from an array
	if string(fields[0][0]) == "[" {
		var arr []interface{}
		err = json.Unmarshal(body, &arr)
		if err != nil {
			return
		}
		v := extractValueFromJSONArray(fields[begin], arr)
		if begin == len(fields)-1 { // if it is the value expected
			value = fmt.Sprint(v)
			return
		}
		next = v
		begin++

	}

	jsonMap := make(map[string]interface{})
	if begin == 0 {
		// if it does not start by an array, we should parse the json
		err = json.Unmarshal(body, &jsonMap)
		if err != nil {
			return
		}
	} else {
		// if after have started by an array it continues with an object [0].A
		jsonMap = next.(map[string]interface{})
	}

	tmp := jsonMap
	for i := begin; i < len(fields); i++ {
		if i == len(fields)-1 {
			if string(fields[i][len(fields[i])-1]) == "]" { // It's an array and fields[i] is in the format "param[number]"
				v := extractValueFromJSONMap(fields[i], tmp)
				if v != nil {
					value = fmt.Sprint(v)
					return
				}
			}
			// value
			if val, ok := tmp[fields[i]]; ok {
				value = fmt.Sprint(val)
				return
			}

			return "", fmt.Errorf("value not present in the json object %s", format)
		}

		var o interface{}

		if string(fields[i][len(fields[i])-1]) == "]" { // It's an array and fields[i] is in the format "param[number]"
			o = extractValueFromJSONMap(fields[i], tmp)
		} else {
			if val, ok := tmp[fields[i]]; ok {
				o = val
			} else {
				return "", fmt.Errorf("value not present in the json object %s", format)
			}
		}

		tmp = o.(map[string]interface{})
	}

	return
}

func extractValueFromJSONMap(key string, m map[string]interface{}) interface{} {
	if m == nil {
		return nil
	}

	if index := strings.Index(key, "["); index > -1 {
		name := key[:index]                                   // we extract the param name
		index, err := strconv.Atoi(key[index+1 : len(key)-1]) // we extract the number
		if err == nil {
			if val, ok := m[name]; ok {
				arr := val.([]interface{})
				return arr[index]
			}
		}
	}
	return nil
}

func extractValueFromJSONArray(key string, m []interface{}) interface{} {
	if m == nil {
		return nil
	}

	if index := strings.Index(key, "["); index > -1 {
		index, err := strconv.Atoi(key[index+1 : len(key)-1]) // we extract the number
		if err == nil {
			return m[index]
		}
	}
	return nil
}
