package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"time"
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

func get_conn(host string, port int) net.Conn {
	full_host := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	vprintfln("trying to connect to %s...", full_host)

	conn, err := net.DialTimeout("tcp", full_host, 1*time.Second)
	if err != nil {
		vprintfln("unable to connect to tv on host %s", full_host)
		return nil
	}
	vprintfln("successfully connected to tv on %s", full_host)
	return conn
}

func vprintfln(format string, a ...any) {
	if !verbose {
		return
	}
	fmt.Printf(format+"\n", a...)
}
