package main

import (
	_ "secretserver/external/mongodb"
	_ "secretserver/secret"

	"secretserver/server"
)

func main() {
	server.Serve()
}
