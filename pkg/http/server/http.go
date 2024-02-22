package server

import "flag"

type UserInputHttp struct {
	httpServer bool
	httpClient bool
}

func (uis *UserInputHttp) SetFlagServer(fs *flag.FlagSet) {
	fs.BoolVar(&uis.httpServer, "Server", false, "Create an HTTP Server")
	fs.BoolVar(&uis.httpClient, "Client", false, "Create an HTTP Client")
}
