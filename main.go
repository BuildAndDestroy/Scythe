package main

import "backdoorBoi/internal/userinput"

// "github.com/BuildAndDestroy/backdoorBoi/internal/userinput"

func main() {
	userinput.UserInputCheck()
	// userinput.ArgLengthCheck()
	// userinput.CommandCheck(os.Args[1])
	userinput.UserCommands()
}
