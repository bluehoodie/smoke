package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/bluehoodie/smoke/tester"
	"github.com/jessevdk/go-flags"
	yaml "gopkg.in/yaml.v2"
)

var opts struct {
	Verbose bool   `short:"v" long:"verbose" description:"print out full report including successful results"`
	File    string `short:"f" long:"file" default:"./smoke_test.json" description:"file containing the test definition"`
	URL     string `short:"u" long:"url" default:"http://localhost" description:"url endpoint to test"`
	Port    int    `short:"p" long:"port" description:"port the service is running on"`
	Timeout int    `short:"t" long:"timeout" default:"1" description:"timeout in seconds for each http request made"`
}

func init() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(2)
	}
}

func main() {
	data, err := ioutil.ReadFile(opts.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read test file: %v\n", err)
		os.Exit(2)
	}

	t := tester.Test{}
	if err := unmarshal(opts.File, data, &t); err != nil {
		fmt.Fprintf(os.Stderr, "could not unmarshal test data: %v\n", err)
		os.Exit(2)
	}

	prepareResponseBody(&t)

	url := opts.URL
	if opts.Port != 0 {
		url = fmt.Sprintf("%s:%d", url, opts.Port)
	}

	client := &http.Client{
		Timeout: time.Duration(opts.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	testRunner := tester.NewRunner(url, t,
		tester.WithVerboseModeOn(opts.Verbose),
		tester.WithHTTPClient(client),
	)

	ok := testRunner.Run()
	if !ok {
		os.Exit(1)
	}
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

// In order to keep retro compatibility with ExpectedResponseBody, and at the same time introducing multiple response check with ExpectedResponse
func prepareResponseBody(t *tester.Test) {
	if t != nil && t.Contracts != nil && len(t.Contracts) > 0 {
		for _, c := range t.Contracts {
			if c.ExpectedResponseBody != "" {
				c.ExpectedResponse = append(c.ExpectedResponse, c.ExpectedResponseBody)
			}
		}
	}
}
