package userinput

import "flag"

type UserInputServer struct {
	netcat     bool
	proxy      bool
	httpServer bool
}

func (uis *UserInputServer) SetFlagServer(fs *flag.FlagSet) {
	fs.BoolVar(&uis.netcat, Netcat, false, "Netcat server, use as a bind shell.")
	fs.BoolVar(&uis.proxy, Proxy, false, "Create a proxy server.")
	fs.BoolVar(&uis.httpServer, HttpServer, false, "Create an HTTP Server")
}
