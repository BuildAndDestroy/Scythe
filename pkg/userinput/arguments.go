package userinput

import (
	"backdoorBoi/pkg/netcat"
	"flag"
	"log"
	"os"
)

// First layer flags that must be called
const (
	Http               string = "Http"
	Netcat             string = "Netcat"
	Proxy              string = "Proxy"
	Scanner            string = "Scanner"
	ExceptionStatement string = "Expected 'Http', 'Netcat', 'Proxy, or 'Scanner'"
)

func UserInputCheck() {
	// Check for no arguments.
	if len(os.Args) <= 1 {
		log.Fatalln("No arguments provided.")
	}
}

func ArgLengthCheck() {
	// Check for less than 1 arg.
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

	if command == Http || command == Netcat || command == Proxy || command == Scanner {
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
	case Http:
		log.Printf("We made it to %s", Http)
	case Netcat:
		log.Printf("We made it to %s", Netcat)
		var nni netcat.NetcatInput
		nni.SetNetcatInput(fs)
		fs.Parse(os.Args[2:])
		var sbn netcat.NetcatStruct
		sbn.NetcatBind(&nni)
	case Proxy:
		log.Printf("We made it to %s", Proxy)
	case Scanner:
		log.Printf("We made it to %s", Scanner)
	default:
		log.Fatalln("Subcommand does not exist")
	}
}
