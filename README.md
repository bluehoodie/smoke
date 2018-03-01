# Smoke

[![Build Status](https://travis-ci.org/bluehoodie/smoke.svg?branch=master)](https://travis-ci.org/bluehoodie/smoke)

A simple application to write and run smoke tests for RESTful APIs.

## Getting Started

### Installing

To install, simply run

```go get github.com/bluehoodie/smoke```

or you can download the source code and run 

```make install```

This will put the executable ```smoke``` in your ```$GOPATH/bin``` directory

## Usage

``` 
Usage of smoke:
  -file string
        file containing the test definition (required)
  -url string
        url endpoint to test (required)
  -v    verbose mode will print out full report including successful results (optional. default: false)

```

### Writing a test file

The test file is a JSON file with the following structure:

```
{
    "variables": {"variable_name": "variable_value", ...}
    "cases" : [<CASE_1>, ... <CASE_N>]
}
```

The variables are a map of string values which can be accessed in every test case defined in the current file.  

Each test CASE is of the form:

```
{
    "name": "<test case name>",
    "path": "<uri endpoint to call for this test. (will be appended to the URL defined in the command)>",
    "method": "<http verb, ie: GET, POST, etc>",
    "body": "<http request body. optional">
    "headers": {"header_name": "header_value" ...} // <map of header values to add to the request. optional>,
    "variables": {"variable_name": "variable_value", ...} // <map of variables specific to this test case. will override the global values>
    
    "http_code_is": <integer representing the expected http code in the result>
    "response_body_contains": <string representing an expected value within the resulting response body>" 
}
```

### Variables

Variables can be used in the path, body or header values. The way a variable is called is by wrapping it in `::`, e.g.: `::variable_name::`
 
The order of precedence for looking for variable values is:

1. variables defined in the test-case map
2. variables defined in the outer variables map
3. environment variables

If no value is found, then the test will fail.

##### Example

If our test case path is defined this way:

```"path": "/get?foo=::token::&bar=1",```

Here, ```::token::``` will be replaced with whichever value is found. 

## Result

Running a test will result in the following possible exit codes:

- 0 : if the tests run and all tests passed
- 1 : if the tests ran but there were some failed tests.
- 2 : if the tests could not be run (error reading or parsing the json test file)

If any tests failed, some output will be written to stderr with more detail about the failed tests.

If verbose mode is on, a report on all tests will be written to stdout.

## License

MIT. see LICENSE file.