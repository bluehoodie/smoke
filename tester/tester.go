package tester

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
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

// Case represents the data for a single test case
type Case struct {
	Name    string            `json:"name"`
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`

	Locals map[string]string `json:"locals"`

	ExpectedHTTPCode     int    `json:"http_code_is"`
	ExpectedResponseBody string `json:"response_body_contains"`
}

// Test represents the data for a full test suite
type Test struct {
	Globals map[string]string `json:"globals"`

	Cases []Case `json:"cases"`
}

// Runner is the primary struct of this package and is responsible for running the test suite
type Runner struct {
	successOutput io.Writer
	failureOutput io.Writer

	client *http.Client

	test Test
	url  string
}

// Option is a function which can change some properties of the Runner
type Option func(*Runner)

// WithVerboseModeOn returns an Option which sets the verbosity of the runner.  Default is false.
func WithVerboseModeOn(verboseMode bool) Option {
	return func(r *Runner) {
		if verboseMode {
			r.successOutput = os.Stdout
			r.failureOutput = os.Stdout
		}
	}
}

// NewRunner returns a *Runner for a given url and Test.
func NewRunner(url string, test Test, opts ...Option) *Runner {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	runner := &Runner{
		url:  url,
		test: test,

		client:        client,
		successOutput: ioutil.Discard,
		failureOutput: os.Stderr,
	}

	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

// Run is the method which runs the Test associated with this Runner.
// Returns a bool representing the result of the test.
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
			failure(runner.failureOutput, testCase.Name, "could not parse variables: %v", err)
			failCount++
			ok = false
			continue
		}

		// create request
		path := strings.Join([]string{runner.url, testCase.Path}, "")
		req, err := http.NewRequest(strings.ToUpper(testCase.Method), path, strings.NewReader(testCase.Body))
		if err != nil {
			failure(runner.failureOutput, testCase.Name, "could not create http request: %v", err)
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
			failure(runner.failureOutput, testCase.Name, "error sending request: %v", err)
			failCount++
			ok = false
			continue
		}

		// validate http status code
		if testCase.ExpectedHTTPCode != 0 {
			if resp.StatusCode != testCase.ExpectedHTTPCode {
				failure(runner.failureOutput, testCase.Name, "expected http response code %d got %d", testCase.ExpectedHTTPCode, resp.StatusCode)
				failCount++
				ok = false
				continue
			}
		}

		// validate http response body
		if testCase.ExpectedResponseBody != "" {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				failure(runner.failureOutput, testCase.Name, "error reading response body: %v", err)
				failCount++
				ok = false
				continue
			}
			toClose = append(toClose, resp.Body)

			if !strings.Contains(string(body), testCase.ExpectedResponseBody) {
				failure(runner.failureOutput, testCase.Name, "expected response not found in the body")
				failCount++
				ok = false
				continue
			}
		}

		success(runner.successOutput, testCase.Name)
	}

	if !ok {
		red.Fprintf(runner.failureOutput, "FAILED (%d of %d tests failed)\n", failCount, len(runner.test.Cases))
	} else {
		boldGreen.Fprint(runner.successOutput, "OK\n")
	}

	return ok
}
