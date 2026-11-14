package interpreter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ── Errors ────────────────────────────────────────────────────────────────────

// Frame is one entry of a RuntimeError's Layer call trace. Consecutive
// identical frames (deep recursion) are collapsed into one with a count.
type Frame struct {
	Function string // "layer.name"
	Count    int
}

// RuntimeError is an evaluation error with the source position where it
// happened and the chain of named Layer functions that were executing,
// innermost first.
type RuntimeError struct {
	Line    int
	Column  int
	Message string
	Trace   []Frame
}

func (e *RuntimeError) addFrame(fn string) {
	if n := len(e.Trace); n > 0 && e.Trace[n-1].Function == fn {
		e.Trace[n-1].Count++
		return
	}
	e.Trace = append(e.Trace, Frame{Function: fn, Count: 1})
}

// TraceString renders the call trace, innermost first, compressing
// recursion: "core.fib (×12)" for direct recursion and
// "(core.fact ← debug.fact) ×5" for a cycle through several layers (the
// usual shape of open recursion with a wrapping layer).
func (e *RuntimeError) TraceString() string {
	frames := make([]string, len(e.Trace))
	for idx, f := range e.Trace {
		if f.Count > 1 {
			frames[idx] = fmt.Sprintf("%s (×%d)", f.Function, f.Count)
		} else {
			frames[idx] = f.Function
		}
	}

	var parts []string
	for i := 0; i < len(frames); {
		bestPeriod, bestRepeats := 1, 1
		for period := 2; period <= 4 && i+2*period <= len(frames); period++ {
			repeats := 1
			for i+(repeats+1)*period <= len(frames) && equalStrings(frames[i:i+period], frames[i+repeats*period:i+(repeats+1)*period]) {
				repeats++
			}
			if repeats > 1 && repeats*period > bestRepeats*bestPeriod {
				bestPeriod, bestRepeats = period, repeats
			}
		}
		if bestRepeats > 1 {
			parts = append(parts, fmt.Sprintf("(%s) ×%d", strings.Join(frames[i:i+bestPeriod], " ← "), bestRepeats))
			i += bestPeriod * bestRepeats
		} else {
			parts = append(parts, frames[i])
			i++
		}
	}
	return strings.Join(parts, " ← ")
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (e *RuntimeError) Error() string {
	msg := fmt.Sprintf("runtime error at %d:%d: %s", e.Line, e.Column, e.Message)
	if len(e.Trace) > 0 {
		msg += " [in " + e.TraceString() + "]"
	}
	return msg
}

// CheckIssue is one problem found by the static checks run before a
// program executes.
type CheckIssue struct {
	Line    int
	Column  int
	Message string
}

// CheckError reports every issue found by the static checks. Its Error()
// renders one "check error at L:C: msg" line per issue.
type CheckError struct {
	Issues []CheckIssue
}

func (e *CheckError) Error() string {
	lines := make([]string, len(e.Issues))
	for idx, issue := range e.Issues {
		lines[idx] = fmt.Sprintf("check error at %d:%d: %s", issue.Line, issue.Column, issue.Message)
	}
	return strings.Join(lines, "\n")
}

// ── Display ───────────────────────────────────────────────────────────────────

// Show renders a value for humans: the result shown by the CLI and the
// editor, log messages, and string interpolation. Top-level strings are
// shown raw; strings nested in lists or records are quoted.
func Show(val interface{}) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ShowNested(val)
}

// ShowNested renders a value as it would appear inside a collection, with
// strings quoted — close to Layer literal syntax.
func ShowNested(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return "nil"
	case string:
		return strconv.Quote(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.Itoa(v)
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	case []interface{}:
		parts := make([]string, len(v))
		for idx, elem := range v {
			parts[idx] = ShowNested(elem)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		if t, ok := v["type"].(string); ok && len(v) <= 2 {
			switch t {
			case "Some":
				if inner, ok := v["value"]; ok {
					return "Some(" + ShowNested(inner) + ")"
				}
			case "None":
				if len(v) == 1 {
					return "None"
				}
			}
		}
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for idx, k := range keys {
			parts[idx] = k + ": " + ShowNested(v[k])
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *Closure:
		return v.String()
	case *UIComponent:
		return v.String()
	case *App:
		return v.String()
	case func(...interface{}) (interface{}, error):
		return "<builtin function>"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// typeName returns the Layer-level name of a value's type, for messages.
func typeName(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return "nil"
	case string:
		return "string"
	case int64, int:
		return "integer"
	case float64:
		return "float"
	case bool:
		return "boolean"
	case []interface{}:
		return "list"
	case map[string]interface{}:
		if t, ok := v["type"].(string); ok && (t == "Some" || t == "None") {
			return "option"
		}
		return "record"
	case *Closure, func(...interface{}) (interface{}, error):
		return "function"
	case *UIComponent:
		return "ui component"
	case *App:
		return "app"
	default:
		return fmt.Sprintf("%T", val)
	}
}
