package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/ph34rd/gohtml2md/html2mdutil"
	"os"
)

func parseArgs() (in string, out string, strip bool, err error) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: html2md [options] input.html output.md\nOptions:\n")
		fs.PrintDefaults()
	}

	stripPtr := fs.Bool("s", false, "Strip unknown tags")

	err = fs.Parse(os.Args[1:])

	if err != nil {
		return
	}

	args := fs.Args()
	argsLen := len(args)

	switch {
	case argsLen > 2:
		err = errors.New("too many parameters")
	case argsLen == 0 || len(args[0]) == 0:
		err = errors.New("missing parameter: input")
	case argsLen == 1 || len(args[1]) == 0:
		err = errors.New("missing parameter: output")
	default:
		in = args[0]
		out = args[1]
		strip = *stripPtr
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		fs.Usage()
	}

	return
}

func processInput() int {
	in, out, strip, err := parseArgs()

	if err != nil {
		return 2
	}

	htmlFile, err := os.Open(in)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}

	defer htmlFile.Close()

	mdFile, err := os.Create(out)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}

	defer mdFile.Close()

	err = html2mdutil.Process(htmlFile, mdFile, strip)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(processInput())
}
