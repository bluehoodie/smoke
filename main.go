package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bluehoodie/smoke/test"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Verbose bool   `short:"v" long:"verbose" description:"print out full report including successful results"`
	File    string `short:"f" long:"file" default:"./smoke_test.json" description:"file containing the test definition"`
	Url     string `short:"u" long:"url" default:"http://localhost" description:"url endpoint to test"`
	Port    int    `short:"p" long:"port" description:"port the service is running on"`
}

func init() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatalln(err)
		os.Exit(2)
	}
}

func main() {
	data, err := ioutil.ReadFile(opts.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not read test file: %v\n", err)
		os.Exit(2)
	}

	t := &test.Test{}
	err = json.Unmarshal(data, t)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not unmarshal test data: %v\n", err)
		os.Exit(2)
	}

	url := opts.Url
	if opts.Port != 0 {
		url = fmt.Sprintf("%s:%d", url, opts.Port)
	}

	ok := test.NewRunner(url, *t, test.WithVerboseModeOn(opts.Verbose)).Run()
	if !ok {
		os.Exit(1)
	}
}
