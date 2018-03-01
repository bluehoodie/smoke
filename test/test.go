package test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

const (
	good = "\u2713"
	bad  = "\u2717"
)

var (
	red       = color.New(color.FgRed, color.Bold)
	green     = color.New(color.FgGreen)
	boldGreen = color.New(color.FgGreen, color.Bold)

	variableRegex *regexp.Regexp
)

func init() {
	re, err := regexp.Compile(`::([\w]+)::`)
	if err != nil {
		panic(err)
	}
	variableRegex = re
}

type Case struct {
	Name    string            `json:"name"`
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`

	Variables map[string]string

	ExpectedHttpCode     int    `json:"http_code_is"`
	ExpectedResponseBody string `json:"response_body_contains"`
}

type Test struct {
	Variables map[string]string `json:"variables"`
	Cases     []Case            `json:"cases"`
}

type Runner struct {
	successOutput io.Writer
	failureOutput io.Writer

	client *http.Client

	test Test
	url  string
}

type Option func(*Runner)

func WithVerboseMode(verboseMode bool) Option {
	return func(r *Runner) {
		if !verboseMode {
			r.successOutput = ioutil.Discard
			r.failureOutput = os.Stderr
		}
	}
}

func NewRunner(url string, test Test, opts ...Option) *Runner {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	runner := &Runner{
		url:  url,
		test: test,

		client:        client,
		successOutput: os.Stdout,
		failureOutput: os.Stdout,
	}

	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

func (runner *Runner) Run() bool {
	ok := true

	// keep track of all response bodies to close
	toClose := make([]io.Closer, 0, len(runner.test.Cases))
	defer func() {
		for _, closer := range toClose {
			closer.Close()
		}
	}()

	failCount := 0
	for _, testCase := range runner.test.Cases {
		err := parseVariables(runner, &testCase)
		if err != nil {
			failure(runner.failureOutput, testCase, "could not parse variables: %v", err)
			failCount++
			ok = false
			continue
		}

		// create request
		path := strings.Join([]string{runner.url, testCase.Path}, "")
		req, err := http.NewRequest(strings.ToUpper(testCase.Method), path, strings.NewReader(testCase.Body))
		if err != nil {
			failure(runner.failureOutput, testCase, "could not create http request: %v", err)
			failCount++
			ok = false
			continue
		}

		// set headers
		for key, value := range testCase.Headers {
			req.Header.Set(key, value)
		}

		// send request
		resp, err := runner.client.Do(req)
		if err != nil {
			failure(runner.failureOutput, testCase, "error sending request: %v", err)
			failCount++
			ok = false
			continue
		}

		// validate http status code
		if testCase.ExpectedHttpCode != 0 {
			if resp.StatusCode != testCase.ExpectedHttpCode {
				failure(runner.failureOutput, testCase, "expected http response code %d got %d", testCase.ExpectedHttpCode, resp.StatusCode)
				failCount++
				ok = false
				continue
			}
		}

		// validate http response body
		if testCase.ExpectedResponseBody != "" {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				failure(runner.failureOutput, testCase, "error reading response body: %v", err)
				failCount++
				ok = false
				continue
			}
			toClose = append(toClose, resp.Body)

			if !strings.Contains(string(body), testCase.ExpectedResponseBody) {
				failure(runner.failureOutput, testCase, "expected response not found in the body")
				failCount++
				ok = false
				continue
			}
		}

		success(runner.successOutput, testCase)
	}

	if !ok {
		red.Fprintf(runner.failureOutput, "FAILED (%d of %d tests failed)\n", failCount, len(runner.test.Cases))
	} else {
		boldGreen.Fprint(runner.successOutput, "OK\n")
	}

	return ok
}

func success(output io.Writer, test Case) {
	green.Fprintf(output, "%v\t%s\n", good, test.Name)
}

func failure(output io.Writer, test Case, format string, args ...interface{}) {
	red.Fprintf(output, "%v\t%s: %s\n", bad, test.Name, fmt.Sprintf(format, args...))
}

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
		if val, ok := testCase.Variables[variableName]; ok {
			replacement = val
			found = true
		}

		// if not found at Case-level, look for variables in the Test
		if val, ok := runner.test.Variables[variableName]; !found && ok {
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
		if val, ok := testCase.Variables[variableName]; ok {
			s = strings.Replace(s, match, val, -1)
		}
	}

	return s, nil
}
