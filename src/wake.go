package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// specific tools to use IP Control's "Wake on Lan" feature, turning a TV on when a MAC address is provided

var wake_address = "255.255.255.255"
var wake_port = 9
var wake_sync_count = 6
var wake_magic_byte byte = 0xff
var wake_address_count = 16

func send_wake(mac string) error {
	packet, err := wake_packet(mac)
	if err != nil {
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(wake_address, fmt.Sprintf("%d", wake_port)))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	_, err = conn.Write(packet)
	return err
}

func wake_packet(mac string) ([]byte, error) {

	block_parts := strings.Split(strings.ReplaceAll(mac, "-", ";"), ":")
	blocks := make([]byte, len(block_parts))
	for i, block := range block_parts {
		val, err := strconv.ParseUint(block, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid MAC address block: %s", block)
		}
		blocks[i] = byte(val)
	}

	packet := make([]byte, wake_sync_count+len(blocks)*wake_address_count)

	for i := range wake_sync_count {
		packet[i] = wake_magic_byte
	}

	for i := wake_sync_count; i < len(packet); i += len(blocks) {
		copy(packet[i:i+len(blocks)], blocks)
	}

	return packet, nil
}
