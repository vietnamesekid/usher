package ui

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)

type Output struct{}

func NewOutput() *Output { return &Output{} }

func (o *Output) Success(msg string) {
	fmt.Fprintln(os.Stdout, green("✓")+" "+msg)
}

func (o *Output) Error(msg string) {
	fmt.Fprintln(os.Stderr, red("✗")+" "+msg)
}

func (o *Output) Warning(msg string) {
	fmt.Fprintln(os.Stdout, yellow("⚠")+" "+msg)
}

func (o *Output) Info(msg string) {
	fmt.Fprintln(os.Stdout, "  "+msg)
}

func (o *Output) Step(msg string) {
	fmt.Fprintln(os.Stdout, cyan("→")+" "+msg)
}

func (o *Output) Table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	sep := make([]string, len(headers))
	for i, h := range headers {
		sep[i] = strings.Repeat("─", len(h))
	}
	fmt.Fprintln(w, strings.Join(sep, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}
