package tester

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func parseVariables(runner *Runner, testCase *Case) error {
	//parse path
	parsedPath, err := replaceVariables(runner, testCase, testCase.Path)
	if err != nil {
		return errors.Wrap(err, "could not parse path")
	}
	testCase.Path = parsedPath

	//parse body
	parsedBody, err := replaceVariables(runner, testCase, testCase.Body)
	if err != nil {
		return errors.Wrap(err, "could not parse body")
	}
	testCase.Body = parsedBody

	//parse headers
	for key, value := range testCase.Headers {
		parsedValue, err := replaceVariables(runner, testCase, value)
		if err != nil {
			return errors.Wrap(err, "could not parse header value")
		}
		testCase.Headers[key] = parsedValue
	}

	return nil
}

func replaceVariables(runner *Runner, testCase *Case, s string) (string, error) {
	matched := variableRegex.FindAllString(s, -1)

	if len(matched) == 0 {
		return s, nil
	}

	for _, match := range matched {
		variableName := strings.Trim(match, "::")

		var replacement string
		var found bool

		// check variables at Case-level first
		if val, ok := testCase.Locals[variableName]; ok {
			replacement = val
			found = true
		}

		// if not found at Case-level, look for variables in the Test
		if val, ok := runner.test.Globals[variableName]; !found && ok {
			replacement = val
			found = true
		}

		// check environment values if still nothing is found
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

	for _, match := range matched {
		variableName := strings.Trim(match, "::")
		if val, ok := testCase.Locals[variableName]; ok {
			s = strings.Replace(s, match, val, -1)
		}
	}

	return s, nil
}
