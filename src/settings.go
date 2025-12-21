package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// argument parsing functions plus a parser for local .env files. also includes a small port scanner to find the tv bases on a cidr network

var default_port = 9761

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

func get_settings() (conn net.Conn, pass string) {
	var (
		help         = flag.Bool("h", false, "Print help")
		host_flag    = flag.String("i", "", "TV host address - if not specified autodiscover")
		pass_flag    = flag.String("w", "", "IP Control Passphrase")
		port_flag    = flag.Int("p", default_port, "IP Control TV port")
		network      = flag.String("n", "", "Autodiscover tv on this CIDR")
		mac          = flag.String("mac", "", "MAC address of TV; if present a wake on lan will be sent")
		verbose_flag = flag.Bool("v", false, "Verbose output")
	)

	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	verbose = *verbose_flag

	if *mac != "" {
		vprintfln("mac provided, sending wake on lan")
		err := send_wake(*mac)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		vprintfln("completed")
		os.Exit(0)
	}

	load_env(".env")

	if *host_flag == "" {
		*host_flag = os.Getenv("TV_HOST")
	}

	if *host_flag != "" {
		conn = get_conn(*host_flag, *port_flag)
	}

	if conn == nil && *network != "" {
		vprintfln("scanning network %s...", *network)
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

		conn = get_conn(ips[0], *port_flag)
		if conn == nil {
			fmt.Printf("unable to connect to %s on port %d\n", ips[0], *port_flag)
			os.Exit(1)
		}

		*host_flag = ips[0]
	}

	if conn == nil {
		fmt.Println("a host for the tv must be specified, or a network to scan specified")
		flag.Usage()
		os.Exit(1)
	}

	if *pass_flag == "" {
		*pass_flag = os.Getenv("TV_PASS")
	}
	if *pass_flag == "" {
		fmt.Println("IP control requires a pass configured on the tv and specified")
		flag.Usage()
		os.Exit(1)
	}

	vprintfln("updating .env cache")
	err := os.WriteFile(".env", fmt.Appendf(nil, "TV_HOST=%s\nTV_PASS=%s\n", *host_flag, *pass_flag), 0644)
	if err != nil {
		panic(err)
	}

	return conn, *pass_flag
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
