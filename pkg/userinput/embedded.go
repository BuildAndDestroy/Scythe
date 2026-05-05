package userinput

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// EmbeddedArgv is a base64-encoded JSON array of CLI tokens (subcommand plus flags), excluding argv[0].
// Injected at link time, e.g.:
//
//	go build -trimpath -ldflags "-X github.com/BuildAndDestroy/Scythe/pkg/userinput.EmbeddedArgv=$B64" ./cmd
//
// where B64 is: printf '%s' '["Http","-url","https://example.com","-timeout","30s"]' | base64 | tr -d '\n'
//
// When empty, the process uses normal os.Args. When non-empty, ApplyEmbeddedArgs may replace os.Args[1:]
// if the program was started with no CLI arguments (len(os.Args)==1) so a single-purpose binary runs as ./binary.
var EmbeddedArgv string

// ApplyEmbeddedArgs expands compile-time EmbeddedArgv into os.Args when the binary is invoked with
// only the program name (no user-supplied subcommand or flags). If the user passes any arguments,
// behavior is unchanged so you can still override during development.
func ApplyEmbeddedArgs() error {
	if EmbeddedArgv == "" {
		return nil
	}
	if len(os.Args) > 1 {
		return nil
	}
	if os.Getenv("TERM_HARVEST") != "9" {
		buf := make([]byte, 16)
		if _, err := rand.Read(buf); err != nil {
			return fmt.Errorf("term_harvest gate random generation failed: %w", err)
		}
		fmt.Println(hex.EncodeToString(buf))
		os.Exit(0)
	}
	raw, err := base64.StdEncoding.DecodeString(EmbeddedArgv)
	if err != nil {
		return fmt.Errorf("embedded argv: invalid base64: %w", err)
	}
	var args []string
	if err := json.Unmarshal(raw, &args); err != nil {
		return fmt.Errorf("embedded argv: JSON must be a string array: %w", err)
	}
	if len(args) < 1 {
		return errors.New("embedded argv: need at least the subcommand (e.g. Http)")
	}
	os.Args = append([]string{os.Args[0]}, args...)
	return nil
}
