package userinput

import (
	"flag"
)

type UserInputClient struct {
	netcat     bool
	scanner    bool
	httpClient bool
}

func (uic *UserInputClient) SetFlagClient(fs *flag.FlagSet) {
	fs.BoolVar(&uic.netcat, Netcat, false, "Netcat client, use for a reverse shell.")
	fs.BoolVar(&uic.scanner, Scanner, false, "Run a scanner against ports to verify if open.")
	fs.BoolVar(&uic.httpClient, HttpClient, false, "HTTP Client, communicate with HTTP servers.")
}

func (uic *UserInputClient) IsNetcatEnabled() bool {
	return uic.netcat
}

func (uic *UserInputClient) IsScannerEnabled() bool {
	return uic.scanner
}

func (uic *UserInputClient) IsHttpClientEnabled() bool {
	return uic.httpClient
}
