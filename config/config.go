package config

import "flag"

var ServerWs = "ws://127.0.0.1:3000/node/ws"
var Auth = ""

func Init() {
	wsStr := flag.String("ws", "", "the server ws addr")
	authStr := flag.String("a", "", "the authorization header")
	flag.Parse()
	if *wsStr != "" {
		ServerWs = *wsStr
	}
	if *authStr != "" {
		Auth = *authStr
	}
}
