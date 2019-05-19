package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"net/http"
	"os"
	"time"

	"github.com/bluehoodie/smoke/internal/tester"
)

var opts struct {
	Verbose bool   `short:"v" long:"verbose" description:"print out full report including successful results"`
	File    string `short:"f" long:"file" default:"./smoke_test.yaml" description:"file containing the test definition"`
	URL     string `short:"u" long:"url" default:"https://httpbin.org" description:"url endpoint to test"`
	Port    int    `short:"p" long:"port" description:"port the service is running on"`
	Timeout int    `short:"t" long:"timeout" default:"1" description:"timeout in seconds for each http request made"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(2)
	}

	t, err := tester.NewTest(opts.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(2)
	}

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

	runner := tester.NewRunner(url, t,
		tester.WithVerboseModeOn(opts.Verbose),
		tester.WithHTTPClient(client),
	)

	ok := runner.Run()
	if !ok {
		os.Exit(1)
	}
}
