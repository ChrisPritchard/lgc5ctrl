package main

import (
	"flag"
	"fmt"
	"strings"
)

var default_command = "MUTE_STATE"
var verbose bool

func main() {

	var conn, pass = get_settings()
	defer conn.Close()

	command := strings.Join(flag.Args(), " ")
	if command == "" {
		vprintfln("a command was not specified, defaulting to %s", default_command)
		command = default_command
	}

	key := get_key(pass)
	vprintfln("generated key: %x", key)

	encoded_command := encode(key, command+"\r")
	vprintfln("encoded command: %x", encoded_command)
	vprintfln("sending...")

	conn.Write(encoded_command)

	buffer := make([]byte, 4096)
	n, _ := conn.Read(buffer)

	if n == 0 {
		vprintfln("no response received")
		return
	}

	response := buffer[:n]
	vprintfln("received encoded: %x", response)
	fmt.Print(decode(key, response))
}

func vprintfln(format string, a ...any) {
	if !verbose {
		return
	}
	fmt.Printf(format+"\n", a...)
}
