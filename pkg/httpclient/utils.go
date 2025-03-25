package httpclient

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/BuildAndDestroy/Scythe/pkg/httpclient/httpclientenv"
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

// Collect environment info and return as JSON payload
func EnvVariableJsonPayload() string {
	hostname, err := httpclientenv.GetHostname()
	if err != nil {
		log.Println("[-] No hostname to report")
		hostname = ""
	}
	fmt.Println("Hostname:", hostname)

	ip, err := httpclientenv.GetIPAddress()
	if err != nil {
		log.Println("[-] Unable to retrieve IP")
		ip = ""
	}
	fmt.Println("IP Address:", ip)

	user, err := httpclientenv.GetCurrentUser()
	if err != nil {
		log.Println("[-] Unable to get current user")
		user = ""
	}
	fmt.Println("Current User:", user)

	groups, err := httpclientenv.GetUserGroups()
	if err != nil {
		log.Println("[-] Unable to get user groups")
		groups = []string{"Unable", "to", "retrieve"}
	}
	fmt.Println("Groups:", strings.Join(groups, ", "))

	data := map[string]interface{}{
		"hostname":   hostname,
		"ip_address": ip,
		"user":       user,
		"groups":     groups,
	}
	// log.Println(ToJSON(data))
	return ToJSON(data)
}
