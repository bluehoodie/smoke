package tester

import (
	"fmt"
	"os"
	"regexp"
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
