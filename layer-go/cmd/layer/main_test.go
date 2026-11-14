package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jeanlazarou/layer/pkg/interpreter"
)

// repoRoot is the repository root, relative to this package directory.
var repoRoot, _ = filepath.Abs("../../..")

func runCLI(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

var (
	resultComment = regexp.MustCompile(`(?m)^#\s*Result:\s*(.+)$`)
	// Apps document their behaviour with messages and the model they lead to:
	//   # Send: "inc"
	//   # Model: {count: 1}
	sendComment  = regexp.MustCompile(`(?m)^#\s*Send:\s*(.+)$`)
	modelComment = regexp.MustCompile(`(?m)^#\s*Model:\s*(.+)$`)
)

// TestExamples runs every program in examples/. When an example documents
// its result with a "# Result: ..." comment, the output must match it. An
// app example's "# Send: ..." messages are sent in order, and the final
// model must match its "# Model: ..." comment.
func TestExamples(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(repoRoot, "examples", "*.layer"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no examples found: %v", err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			t.Chdir(t.TempDir()) // the data layer writes files relative to cwd
			interpreter.ResetStorage()
			src, _ := os.ReadFile(file)

			args := []string{"run"}
			for _, m := range sendComment.FindAllSubmatch(src, -1) {
				args = append(args, "--send", string(m[1]))
			}
			code, stdout, stderr := runCLI(t, append(args, file)...)
			if code != 0 {
				t.Fatalf("exit %d\n%s", code, stderr)
			}
			if m := modelComment.FindSubmatch(src); m != nil {
				want := "model: " + strings.TrimSpace(string(m[1]))
				if got := strings.SplitN(stdout, "\n", 2)[0]; got != want {
					t.Errorf("after the documented messages: %q, documented as %q", got, want)
				}
			}
			if m := resultComment.FindSubmatch(src); m != nil {
				want := strings.TrimSpace(string(m[1]))
				if got := strings.TrimSpace(stdout); got != want {
					t.Errorf("result %q, documented as %q", got, want)
				}
			}
		})
	}
}

var layerBlock = regexp.MustCompile("(?s)```layer\n(.*?)```")

// TestReadmePrograms runs every complete program (a ```layer block that
// starts with @layers) in README.md, so the front page never shows code
// that does not work.
func TestReadmePrograms(t *testing.T) {
	readme, err := os.ReadFile(filepath.Join(repoRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, m := range layerBlock.FindAllSubmatch(readme, -1) {
		block := string(m[1])
		if !strings.HasPrefix(strings.TrimSpace(block), "@layers") {
			continue
		}
		count++
		file := filepath.Join(t.TempDir(), "readme.layer")
		if err := os.WriteFile(file, []byte(block), 0o644); err != nil {
			t.Fatal(err)
		}
		if code, _, stderr := runCLI(t, "run", file); code != 0 {
			t.Errorf("README program %d failed:\n%s\n--- program ---\n%s", count, stderr, block)
		}
	}
	if count == 0 {
		t.Fatal("no complete programs found in README.md")
	}
}

func writeProgram(t *testing.T, src string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "p.layer")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestCLIVariablesDriveConditionalLayers(t *testing.T) {
	file := writeProgram(t, `@layers [core, loud if env == "dev"]
@layer core { fn greet() = "hello" }
@layer loud { fn greet() = "HELLO" }
greet()`)

	// The condition references env, so it must be defined; -D provides it.
	if code, _, stderr := runCLI(t, "run", file); code != 1 || !strings.Contains(stderr, "condition for layer 'loud'") {
		t.Errorf("without -D: exit %d, stderr %q", code, stderr)
	}
	if _, out, _ := runCLI(t, "run", "-D", "env=prod", file); strings.TrimSpace(out) != "hello" {
		t.Errorf("with -D env=prod: got %q", out)
	}
	if _, out, _ := runCLI(t, "run", "-D", "env=dev", file); strings.TrimSpace(out) != "HELLO" {
		t.Errorf("with -D env=dev: got %q", out)
	}
}

func TestCLIReportsLocatedErrors(t *testing.T) {
	file := writeProgram(t, "@layers [core]\n@layer core {\n  fn f(x) = x + 1\n}\nf(\"a\")")
	code, _, stderr := runCLI(t, file)
	if code != 1 {
		t.Errorf("exit code %d, want 1", code)
	}
	if !strings.Contains(stderr, file+":3:15: runtime error: operator +") || !strings.Contains(stderr, "in core.f") {
		t.Errorf("stderr: %s", stderr)
	}
}

func TestCLICheckDoesNotRunMain(t *testing.T) {
	file := writeProgram(t, "@layers [core, debug]\nlog(\"side effect\")")
	code, stdout, stderr := runCLI(t, "check", file)
	if code != 0 || !strings.HasSuffix(strings.TrimSpace(stdout), ": ok") || strings.Contains(stderr, "side effect") {
		t.Errorf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}

	bad := writeProgram(t, "@layers [core]\n@layer core {\n fn f() = nope\n}\n1")
	if code, _, stderr := runCLI(t, "check", bad); code != 1 || !strings.Contains(stderr, "3:11: check error: undefined name 'nope'") {
		t.Errorf("code=%d stderr=%q", code, stderr)
	}
}

func TestCLIUsage(t *testing.T) {
	if code, _, _ := runCLI(t); code != 2 {
		t.Errorf("no args: exit %d, want 2", code)
	}
	if code, out, _ := runCLI(t, "--help"); code != 0 || !strings.Contains(out, "usage") {
		t.Errorf("--help: exit %d, out %q", code, out)
	}
}

func TestCLISendsMessagesToApps(t *testing.T) {
	file := writeProgram(t, `@layers [core, ui]
@layer core {
  fn update(msg, m) = match msg with | "inc" -> {m | n: m.n + 1} | {set: v} -> {m | n: v}
}
@layer ui {
  fn page(m) = text("n = {m.n}")
}
app({init: {n: 0}, update: update, view: page})`)

	code, stdout, stderr := runCLI(t, "run", "--send", `{set: 40}`, "--send", `"inc"`, "--send", `"inc"`, file)
	if code != 0 || !strings.HasPrefix(stdout, "model: {n: 42}\n<text content=\"n = 42\" />") {
		t.Errorf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}

	if code, _, stderr := runCLI(t, "run", "--send", `"inc"`, writeProgram(t, "@layers [core]\n1")); code != 1 || !strings.Contains(stderr, "does not return an app") {
		t.Errorf("--send on a non-app: code=%d stderr=%q", code, stderr)
	}
}
