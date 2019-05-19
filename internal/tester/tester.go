package tester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

const (
	good = "\u2713"
	bad  = "\u2717"
)

var (
	red       = color.New(color.FgRed, color.Bold)
	green     = color.New(color.FgGreen)
	boldGreen = color.New(color.FgGreen, color.Bold)
)

// Contract represents the data for a single test case: the definition of the HTTP call
// and the expected result
type Contract struct {
	Name    string            `json:"name" yaml:"name"`
	Path    string            `json:"path" yaml:"path"`
	Method  string            `json:"method" yaml:"method"`
	Body    string            `json:"body" yaml:"body"`
	Headers map[string]string `json:"headers" yaml:"headers"`

	Locals map[string]string `json:"locals" yaml:"locals"`

	Outputs map[string]string `json:"outputs" yaml:"outputs"`

	ExpectedHTTPCode     int               `json:"http_code_is" yaml:"http_code_is"`
	ExpectedResponseBody string            `json:"response_body_contains" yaml:"response_body_contains"`
	ExpectedResponses    []string          `json:"response_contains" yaml:"response_contains"`
	ExpectedHeaders      map[string]string `json:"response_headers_is" yaml:"response_headers_is"`
}

// Test represents the data for a full test suite
type Test struct {
	Globals   map[string]string `json:"globals" yaml:"globals"`
	Contracts []Contract        `json:"contracts" yaml:"contracts"`
}

// NewTest returns an initialized *Test and any error encountered along the way
func NewTest(inputFile string) (*Test, error) {
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read test file %v", inputFile)
	}

	t := Test{}
	if err := unmarshal(inputFile, data, &t); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal test data")
	}

	t.init()

	return &t, nil
}

func (t *Test) init() {
	if t == nil || len(t.Contracts) == 0 {
		return
	}

	for _, c := range t.Contracts {
		if c.ExpectedResponseBody == "" {
			continue
		}

		c.ExpectedResponses = append(c.ExpectedResponses, c.ExpectedResponseBody)
	}
}

// Runner is the primary struct of this package and is responsible for running the test suite
type Runner struct {
	successOutput io.Writer
	failureOutput io.Writer

	client *http.Client

	test *Test
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

// WithHTTPClient returns an Option which overrides the default http client
func WithHTTPClient(client *http.Client) Option {
	return func(r *Runner) {
		r.client = client
	}
}

// NewRunner returns a *Runner for a given url and Test.
func NewRunner(url string, test *Test, opts ...Option) *Runner {
	runner := &Runner{
		url:  url,
		test: test,

		client:        http.DefaultClient,
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
	var failCount int
	for _, contract := range runner.test.Contracts {
		if err := runner.validateContract(contract); err != nil {
			failure(runner.failureOutput, contract.Name, err.Error())
			failCount++
			continue
		}
		success(runner.successOutput, contract.Name)
	}

	if failCount > 0 {
		red.Fprintf(runner.failureOutput, "FAILED (%d of %d tests failed)\n", failCount, len(runner.test.Contracts))
		return false
	}

	boldGreen.Fprint(runner.successOutput, "OK\n")
	return true
}

func (runner *Runner) validateContract(contract Contract) (err error) {
	if err = parseVariables(runner, &contract); err != nil {
		return
	}

	var resp *http.Response
	resp, err = createAndSendRequest(contract, runner.url, runner.client)
	if err != nil {
		return
	}

	if err = validateHTTPCode(contract, resp); err != nil {
		return
	}

	if len(contract.ExpectedHeaders) > 0 {
		if err = validateHeaders(contract, resp); err != nil {
			return
		}
	}

	if len(contract.ExpectedResponses) > 0 || (contract.Outputs != nil && len(contract.Outputs) > 0) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if err = validateResponseBody(contract, body); err != nil {
			return err
		}

		if err = parseOutputs(runner, &contract, body); err != nil {
			return err
		}
	}

	return
}

func success(out io.Writer, name string) {
	green.Fprintf(out, "%v\t%s\n", good, name)
}

func failure(out io.Writer, name, format string, args ...interface{}) {
	red.Fprintf(out, "%v\t%s: %s\n", bad, name, fmt.Sprintf(format, args...))
}

func createAndSendRequest(contract Contract, url string, client *http.Client) (*http.Response, error) {
	// create request
	uri := strings.Join([]string{url, contract.Path}, "")
	req, err := http.NewRequest(strings.ToUpper(contract.Method), uri, strings.NewReader(contract.Body))
	if err != nil {
		return nil, fmt.Errorf("could not create http request: %v", err)
	}

	// set headers
	for key, value := range contract.Headers {
		req.Header.Set(key, value)
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	return resp, nil
}

func validateHTTPCode(contract Contract, resp *http.Response) error {
	if contract.ExpectedHTTPCode != 0 {
		if resp.StatusCode != contract.ExpectedHTTPCode {
			return fmt.Errorf("expected http response code %d got %d", contract.ExpectedHTTPCode, resp.StatusCode)
		}
	}

	return nil
}

func validateResponseBody(contract Contract, body []byte) error {
	if len(contract.ExpectedResponses) == 0 {
		return nil
	}

	for _, r := range contract.ExpectedResponses {
		// check if it is a regexp
		if strings.HasPrefix(r, "r/") {
			expectedRegexp := r[2:]
			re, err := regexp.Compile(expectedRegexp)
			if err == nil {
				if !re.Match(body) {
					return fmt.Errorf("regular expression did not find any matches in the response body")
				}
			}
		} else if !bytes.Contains(body, []byte(r)) {
			return fmt.Errorf("expected response not found in the body")
		}
	}

	return nil
}

func validateHeaders(contract Contract, resp *http.Response) error {
	for k, v := range contract.ExpectedHeaders {
		if val, ok := resp.Header[k]; ok && len(val) > 0 {
			if v != val[0] {
				return fmt.Errorf("expected header %s value %s got %s ", k, v, val[0])
			}
		} else {
			return fmt.Errorf("expected header %s not found in the response", k)
		}

	}

	return nil
}

func unmarshal(filename string, in []byte, out interface{}) error {
	var unmarshalError error

	ext := strings.Trim(path.Ext(filename), ".")
	switch ext {
	case "yaml":
		fallthrough
	case "yml":
		unmarshalError = yaml.Unmarshal(in, out)
	default:
		unmarshalError = json.Unmarshal(in, out)
	}

	return unmarshalError
}
