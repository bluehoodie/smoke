package tester

import (
	"fmt"
	"io"

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

func success(output io.Writer, name string) {
	green.Fprintf(output, "%v\t%s\n", good, name)
}

func failure(output io.Writer, name, format string, args ...interface{}) {
	red.Fprintf(output, "%v\t%s: %s\n", bad, name, fmt.Sprintf(format, args...))
}
