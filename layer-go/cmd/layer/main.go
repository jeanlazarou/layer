// Command layer runs Layer programs from the command line.
//
//	layer run [-D name=value]... program.layer   run a program, print its result
//	layer run --send '"inc"' --send '{step: 2}' app.layer   drive an app with messages
//	layer check program.layer                    load + static checks, no main expression
//	layer program.layer                          shorthand for "layer run"
//
// Logs (log, trace, print, ...) go to stderr; the result goes to stdout. For
// an app, the result is "model: ..." followed by the view.
// Errors are reported as file:line:col and exit with status 1.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jeanlazarou/layer/pkg/runner"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// conditionVar matches a @layers condition that referenced an undefined
// variable, so the CLI can suggest -D.
var conditionVar = regexp.MustCompile(`^error evaluating condition for layer '[^']+': undefined variable: (\w+)`)

// listFlag collects a repeatable flag.
type listFlag []string

func (l *listFlag) String() string { return strings.Join(*l, ", ") }

func (l *listFlag) Set(s string) error {
	*l = append(*l, s)
	return nil
}

type varFlags map[string]interface{}

func (v varFlags) String() string { return "" }

func (v varFlags) Set(s string) error {
	name, value, ok := strings.Cut(s, "=")
	if !ok || name == "" {
		return fmt.Errorf("expected name=value, got %q", s)
	}
	v[name] = parseValue(value)
	return nil
}

// parseValue interprets a -D value as an integer, float, or boolean when it
// looks like one, and as a string otherwise.
func parseValue(s string) interface{} {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return s
}

func usage(w io.Writer) {
	fmt.Fprintln(w, `usage:
  layer run [-D name=value]... program.layer   run a program and print its result
  layer run --send MSG... app.layer            send messages to an app, then print
                                               its model and view (MSG is a Layer
                                               expression, e.g. '"inc"' or '{step: 2}')
  layer check [-D name=value]... program.layer load and statically check a program
  layer program.layer                          same as "layer run"

-D seeds a top-level variable, e.g. for conditional layers:
  @layers [core, debug if env == "dev"]   →   layer run -D env=dev app.layer`)
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}

	mode := "run"
	switch args[0] {
	case "run", "check":
		mode = args[0]
		args = args[1:]
	case "-h", "--help", "help":
		usage(stdout)
		return 0
	}

	fs := flag.NewFlagSet("layer", flag.ContinueOnError)
	fs.SetOutput(stderr)
	vars := varFlags{}
	fs.Var(vars, "D", "seed a variable: -D name=value (repeatable)")
	var sends listFlag
	fs.Var(&sends, "send", "send a message to an app (a Layer expression; repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		usage(stderr)
		return 2
	}

	path := fs.Arg(0)
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "layer: %v\n", err)
		return 1
	}

	res := runner.Run(string(source), runner.Options{
		Vars:      vars,
		LogSink:   func(line string) { fmt.Fprintln(stderr, line) },
		CheckOnly: mode == "check",
		Send:      sends,
	})

	if !res.Success {
		for _, issue := range res.Errors {
			fmt.Fprintf(stderr, "%s:%s\n", path, issue)
			if m := conditionVar.FindStringSubmatch(issue.Message); m != nil {
				fmt.Fprintf(stderr, "    hint: define it on the command line, e.g. layer run -D %s=value %s\n", m[1], path)
			}
		}
		return 1
	}

	if mode == "check" {
		fmt.Fprintf(stdout, "%s: ok\n", path)
	} else if res.Kind == "app" {
		fmt.Fprintf(stdout, "model: %s\n%s\n", res.ModelText, res.Display)
	} else if res.Kind != "none" {
		fmt.Fprintln(stdout, res.Display)
	}
	return 0
}
