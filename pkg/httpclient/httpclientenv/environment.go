package httpclientenv

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

// Give me the hostname
func GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}

// Get the IP we are using to connect back to the server
func GetIPAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue // Skip this for now. We should really collect this shit
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil {
					return v.IP.String(), nil
				}
			}
		}
	}
	return "", fmt.Errorf("[-] No active network interface found")
}

// Get our current user
func GetCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Username, nil
}

// GetUserGroups retrieves the groups the current user belongs to.
func GetUserGroups() ([]string, error) {
	if runtime.GOOS == "windows" {
		return GetWindowsGroups()
	}
	return GetUnixGroups()
}

// getUnixGroups retrieves user groups on Unix-like systems.
func GetUnixGroups() ([]string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	groups, err := currentUser.GroupIds()
	if err != nil {
		return nil, err
	}

	groupNames := []string{}
	for _, gid := range groups {
		group, err := user.LookupGroupId(gid)
		if err == nil {
			groupNames = append(groupNames, group.Name)
		}
	}
	return groupNames, nil
}

// Gather windows details. Relies on cmd.exe
// Need to find a better way to handle this or we will get caught
func GetWindowsGroups() ([]string, error) {
	cmd := exec.Command("whoami", "/groups")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var groups []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "]") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				groups = append(groups, fields[0])
			}
		}
	}
	return groups, nil
}
