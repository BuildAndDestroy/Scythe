package userinput

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/BuildAndDestroy/backdoorBoi/pkg/filetransfer"
	"github.com/BuildAndDestroy/backdoorBoi/pkg/httpclient"
	"github.com/BuildAndDestroy/backdoorBoi/pkg/netcat"
	"github.com/BuildAndDestroy/backdoorBoi/pkg/proxies"
)

// First layer flags that must be called
const (
	Http               string = "Http"
	Netcat             string = "Netcat"
	Proxy              string = "Proxy"
	Scanner            string = "Scanner"
	FileTransfer       string = "FileTransfer"
	ExceptionStatement string = "Expected 'Http', 'Netcat', 'Proxy, 'Scanner', or 'FileTransfer'"
)

func UserInputCheck() {
	// Check for no arguments.
	if len(os.Args) <= 1 {
		log.Fatalln("No arguments provided.")
	}
}

func ArgLengthCheck() {
	// Check for less than 1 arg.
	if len(os.Args) <= 2 {
		log.Fatalln(ExceptionStatement)
	}
}

func CommandCheck(command string) {
	// Check for user input matches our const, otherwise throw "exception" and exit
	if len(os.Args) <= 1 {
		log.Fatalln(ExceptionStatement)
	}

	if command == Http || command == Netcat || command == Proxy || command == Scanner || command == FileTransfer {
		return
	} else {
		log.Fatalln(ExceptionStatement)
	}
}

func UserCommands() {
	// Parse user commands to execute program
	var command string = os.Args[1]

	ArgLengthCheck()
	CommandCheck(command)

	fs := flag.NewFlagSet(command, flag.ExitOnError)
	switch command {
	case Http:
		// log.Printf("We made it to %s", Http)
		opts := httpclient.RequestOptions{}
		opts.SetRequestFlag(fs)
		// log.Println(os.Args[2:])
		err := fs.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("[-] Error parsing flags: %s\n", err)
		}
		// Create result and error channels
		resultChan := make(chan *httpclient.ResponseData)
		errorChan := make(chan error)
		stopChan := make(chan os.Signal, 1)

		// Handle CTRL+C
		signal.Notify(stopChan, os.Interrupt)

		go httpclient.RunWithInterval(&opts, resultChan, errorChan, stopChan)
		// Listen for responses without blocking the loop
		for {
			select {
			case result := <-resultChan:
				if result != nil {
					// log.Printf("[+] Response Body: %s\n", result.Body)
					log.Println(result.Body)
				}
			case err := <-errorChan:
				if err != nil {
					log.Printf("[-] Error: %s\n", err)
				}
			case <-stopChan:
				fmt.Println("\n[-] CTRL+C detected. Stopping the program...")
				return
			}
		}

	case Netcat:
		nni := netcat.NetcatInput{}
		nni.SetNetcatInput(fs)
		err := fs.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("[-] Error parsing flags: %s", err)
		}
		netcat.NetcatArgLogic(&nni)
	case Proxy:
		pxyFlags := &proxies.Flags{}
		pxyFlags.ProxyFlagInput(fs)
		err := fs.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Error parsing flags: %s", err)
		}
		// Start SOCKS5 server
		proxies.ProxyLogic(pxyFlags)
	case Scanner:
		log.Printf("We made it to %s", Scanner)
	case FileTransfer:
		ft := &filetransfer.FileTransfer{}
		ft.FileTransferInput(fs)
		err := fs.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Error parsing flags: %s", err)
		}
		filetransfer.FileTransferLogic(ft)
	default:
		log.Fatalln("Subcommand does not exist")
	}
}
