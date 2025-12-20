package main

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"flag"
	"fmt"
	"strings"
)

var default_command = "MUTE_STATE"
var default_port = 9761

var salt = []byte{0x63, 0x61, 0xb8, 0x0e, 0x9b, 0xdc, 0xa6, 0x63, 0x8d, 0x07, 0x20, 0xf2, 0xcc, 0x56, 0x8f, 0xb9}
var iter = 16384
var bsze = 16

var verbose bool

func main() {

	var conn, pass = get_settings()
	defer conn.Close()

	command := strings.Join(flag.Args(), " ")
	if command == "" {
		vprintfln("a command was not specified, defaulting to %s", default_command)
		command = default_command
	}

	key, _ := pbkdf2.Key(sha256.New, pass, salt, iter, bsze)
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
