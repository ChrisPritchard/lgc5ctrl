package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var default_command = "MUTE_STATE"
var default_port = 9761

var salt = []byte{0x63, 0x61, 0xb8, 0x0e, 0x9b, 0xdc, 0xa6, 0x63, 0x8d, 0x07, 0x20, 0xf2, 0xcc, 0x56, 0x8f, 0xb9}
var iter = 16384
var bsze = 16

func main() {

	var verbose, conn, pass = get_settings()
	defer conn.Close()

	command := flag.Arg(0)
	if command == "" {
		if verbose {
			fmt.Printf("a command was not specified, defaulting to %s\n", default_command)
		}
		command = default_command
	}

	key, _ := pbkdf2.Key(sha256.New, pass, salt, iter, bsze)
	if verbose {
		fmt.Printf("generated key: %x\n", key)
	}

	encoded_command := encode(key, command+"\r")
	if verbose {
		fmt.Printf("encoded command: %x\n", encoded_command)
		fmt.Println("sending...")
	}

	conn.Write(encoded_command)

	buffer := make([]byte, 4096)
	n, _ := conn.Read(buffer)

	if n == 0 {
		if verbose {
			fmt.Println("no response received")
		}
		return
	}

	response := buffer[:n]

	if verbose {
		fmt.Printf("received encoded: %x\n", response)
	}

	fmt.Print(decode(key, response))
}

func get_conn(host string, port int, verbose bool) net.Conn {
	full_host := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	if verbose {
		fmt.Printf("trying to connect to %s...\n", full_host)
	}
	conn, err := net.DialTimeout("tcp", full_host, 1*time.Second)
	if err != nil {
		if verbose {
			fmt.Printf("unable to connect to tv on host %s\n", full_host)
		}
		return nil
	}
	return conn
}

func get_settings() (verbose bool, conn net.Conn, pass string) {
	var (
		help      = flag.Bool("h", false, "Print help")
		host_flag = flag.String("i", "", "TV host address - if not specified autodiscover")
		pass_flag = flag.String("w", "", "IP Control Passphrase")
		port_flag = flag.Int("p", default_port, "IP Control TV port")
		network   = flag.String("n", "", "Autodiscover tv on this CIDR")
		//mac     = flag.String("mac", "", "MAC address of TV, to turn it on if off")
		verbose_flag = flag.Bool("v", false, "Verbose output")
	)

	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	verbose = *verbose_flag

	load_env(".env")

	if *host_flag == "" {
		*host_flag = os.Getenv("TV_HOST")
	}

	if *host_flag != "" {
		conn = get_conn(*host_flag, *port_flag, verbose)
	}

	if conn == nil && *network != "" {
		if verbose {
			fmt.Printf("scanning network %s...\n", *network)
		}
		ips, err := scan_network(*network, *port_flag, 2*time.Second)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if len(ips) == 0 {
			fmt.Printf("no TV host found on CIDR %s\n", *network)
			os.Exit(1)
		}
		fmt.Printf("tv found at ip %s\n", ips[0])

		conn = get_conn(ips[0], *port_flag, verbose)
		if conn == nil {
			fmt.Printf("unable to connect to %s on port %d\n", ips[0], *port_flag)
			os.Exit(1)
		}

		*host_flag = ips[0]
	}

	if conn == nil {
		fmt.Println("a host for the tv must be specified, or a network to scan specified")
		os.Exit(1)
	}

	if *pass_flag == "" {
		*pass_flag = os.Getenv("TV_PASS")
	}
	if *pass_flag == "" {
		fmt.Println("IP control requires a pass configured on the tv and specified")
		os.Exit(1)
	}

	fmt.Println("updating .env cache")
	err := os.WriteFile(".env", fmt.Appendf(nil, "TV_HOST=%s\nTV_PASS=%s\n", *host_flag, *pass_flag), 0644)
	if err != nil {
		panic(err)
	}

	return verbose, conn, *pass_flag
}

func scan_network(cidr string, port int, timeout time.Duration) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %v", err)
	}

	var ips []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// Remove network and broadcast addresses
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	results := make(chan string, len(ips))
	done := make(chan struct{})
	var wg sync.WaitGroup

	maxWorkers := 100
	sem := make(chan struct{}, maxWorkers)

	for _, ip := range ips {
		wg.Add(1)
		go func(targetIP string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			select {
			case <-done:
				return
			default:
			}

			addr := net.JoinHostPort(targetIP, fmt.Sprintf("%d", port))
			conn, err := net.DialTimeout("tcp", addr, timeout)
			if err == nil {
				conn.Close()
				select {
				case results <- targetIP:
				case <-done:
				}
			}
		}(ip)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var foundIPs []string
	for ip := range results {
		foundIPs = append(foundIPs, ip)
	}

	return foundIPs, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func encode(key []byte, message string) []byte {
	iv := make([]byte, bsze)
	rand.Read(iv)

	padded := message
	if len(message)%bsze == 0 {
		padded += " "
	} else {
		padding := len(padded) % bsze
		padded += string(bytes.Repeat([]byte{byte(bsze - padding)}, bsze-padding))
	}

	aes, _ := aes.NewCipher(key)
	ivEnc := make([]byte, bsze)
	aes.Encrypt(ivEnc, iv)

	dataEnc := make([]byte, len(padded))
	cbc := cipher.NewCBCEncrypter(aes, iv)
	cbc.CryptBlocks(dataEnc, []byte(padded))

	return append(ivEnc, dataEnc...)
}

func decode(key []byte, enc []byte) string {

	ivEnc := enc[:bsze]
	dataEnc := enc[bsze:]

	aes, _ := aes.NewCipher(key)
	iv := make([]byte, bsze)
	aes.Decrypt(iv, ivEnc)

	data := make([]byte, len(dataEnc))
	cbc := cipher.NewCBCDecrypter(aes, iv)
	cbc.CryptBlocks(data, dataEnc)

	padByte := int(data[len(data)-1])
	if padByte < bsze {
		data = data[:len(data)-padByte]
	}

	if len(data) > 0 && data[len(data)-1] == ' ' {
		data = data[:len(data)-1]
	}

	return string(data)
}

func load_env(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
