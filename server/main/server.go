package main

import (
	"github.com/looking-promising/privfiles/server"
	"github.com/looking-promising/privfiles/server/crypto"
)

func main() {
	options := server.ServerOptions{
		MasterKey: crypto.GenerateKey(32),
		StaticDir: "../public"}

	s := server.New(options)
	s.Start()
}
