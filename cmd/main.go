package main

import (
	"log"

	"github.com/BuildAndDestroy/Scythe/pkg/userinput"
)

func main() {
	if err := userinput.ApplyEmbeddedArgs(); err != nil {
		log.Fatal(err)
	}
	userinput.UserInputCheck()
	userinput.UserCommands()
}
