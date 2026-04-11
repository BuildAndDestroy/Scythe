package httpclient

import (
	"strings"
	"testing"
)

func TestRunBuiltinHTTPCommand_RecognizesBuiltins(t *testing.T) {
	t.Parallel()
	cases := []string{
		"whoami", "Whoami", " WHOAMI ",
		"groups", "Groups",
		"environment", "Environment", "ENVIRONMENT",
		"env", "Env",
	}
	for _, c := range cases {
		c := c
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			_, _, handled := runBuiltinHTTPCommand(c)
			if !handled {
				t.Fatalf("expected builtin handled for %q", c)
			}
		})
	}
}

func TestRunBuiltinHTTPCommand_NotExecForEnv(t *testing.T) {
	t.Parallel()
	out, err, handled := runBuiltinHTTPCommand("env")
	if !handled || err != nil {
		t.Fatalf("handled=%v err=%v", handled, err)
	}
	if !strings.Contains(out, "=") && out != "\n" {
		t.Fatalf("expected KEY=value lines, got %q", out)
	}
}

func TestRunBuiltinHTTPCommand_UnknownPassesThrough(t *testing.T) {
	t.Parallel()
	_, _, handled := runBuiltinHTTPCommand("not-a-builtin")
	if handled {
		t.Fatal("expected not handled")
	}
}

func TestRunBuiltinHTTPCommand_WithArgsNotBuiltin(t *testing.T) {
	t.Parallel()
	_, _, handled := runBuiltinHTTPCommand("env -0")
	if handled {
		t.Fatal("multi-token command should not use builtin env")
	}
}
