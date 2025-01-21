package userinput

import (
	"flag"
	"log"
	"os"

	"github.com/BuildAndDestroy/backdoorBoi/pkg/filetransfer"
	"github.com/BuildAndDestroy/backdoorBoi/pkg/netcat"
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
		log.Printf("We made it to %s", Http)
	case Netcat:
		var nni netcat.NetcatInput
		nni.SetNetcatInput(fs)
		fs.Parse(os.Args[2:])
		netcat.NetcatArgLogic(&nni)
	case Proxy:
		log.Printf("We made it to %s", Proxy)
	case Scanner:
		log.Printf("We made it to %s", Scanner)
	case FileTransfer:
		ft := &filetransfer.FileTransfer{}
		fs := flag.NewFlagSet("filetransfer", flag.ExitOnError)
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
