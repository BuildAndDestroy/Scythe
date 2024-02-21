package userinput

import (
	"flag"
	"log"
	"os"
)

// First layer flags that must be called
const (
	Server             string = "Server"
	Client             string = "Client"
	ExceptionStatement string = "Expected 'Client' or 'Server'"
)

// Second layer flags that check true/false to help build a functional beacon
const (
	Netcat     string = "Netcat"
	Scanner    string = "Scanner"
	Proxy      string = "Proxy"
	HttpClient string = "HttpClient"
	HttpServer string = "HttpServer"
)

func UserInputCheck() {
	if len(os.Args) <= 1 {
		log.Fatalln("No arguments provided.")
	}
}

func ArgLengthCheck() {
	// Check for no arguments. If none, exit with help message
	// log.Println(len(os.Args))
	if len(os.Args) <= 2 {
		log.Fatalln(ExceptionStatement)
	}
}

func CommandCheck(command string) {
	// Check for user input matches our const, otherwise throw "exception" and exit
	if len(os.Args) <= 1 {
		log.Fatalln(ExceptionStatement)
	}

	if command == Client || command == Server {
		return
	} else {
		log.Fatalln(ExceptionStatement)
	}
}

func UserCommands() {

	var command string = os.Args[1]

	ArgLengthCheck()
	CommandCheck(command)
	fs := flag.NewFlagSet(command, flag.ExitOnError)
	switch command {
	case Client:
		uic := &UserInputClient{}
		uic.SetFlagClient(fs)
		fs.Parse(os.Args[2:])
		log.Println("Netcat Enabled:", uic.IsNetcatEnabled())
		log.Println("Scanner Enabled:", uic.IsScannerEnabled())
		log.Println("HTTP Client Enabled:", uic.IsHttpClientEnabled())
	case Server:
		log.Println("We made it to server")
	default:
		log.Fatalln("Subcommand does not exist")
		os.Exit(1)
	}
}
