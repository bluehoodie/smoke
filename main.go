package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bluehoodie/smoke/test"
)

var (
	file        *string
	url         *string
	verboseMode *bool
)

func init() {
	file = flag.String("file", "", "file containing the test definition (required)")
	url = flag.String("url", "", "url endpoint to test (required)")
	verboseMode = flag.Bool("v", false, "verbose mode will print out full report including successful results (optional. default: false)")

	flag.Parse()

	if *file == "" || *url == "" {
		flag.Usage()
		os.Exit(2)
	}
}

func main() {
	data, err := ioutil.ReadFile(*file)
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

	ok := test.NewRunner(*url, *t, test.WithVerboseMode(*verboseMode)).Run()
	if !ok {
		os.Exit(1)
	}
}
