package httpclient

import (
	"encoding/json"
	"strings"
)

// ParseHeaders converts a comma-separated string of headers into a map.
func ParseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}
	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

// ToJSON converts a map[string]interface{} to a pretty JSON string.
func ToJSON(data map[string]interface{}) string {
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonData)
}
