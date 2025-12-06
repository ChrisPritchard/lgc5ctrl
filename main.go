package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"strings"
)

var port = 9761
var salt = []byte{0x63, 0x61, 0xb8, 0x0e, 0x9b, 0xdc, 0xa6, 0x63, 0x8d, 0x07, 0x20, 0xf2, 0xcc, 0x56, 0x8f, 0xb9}
var iter = 16384
var bsze = 16

func main() {
	load_env(".env")
	host := net.JoinHostPort(os.Getenv("TV_HOST"), fmt.Sprintf("%d", port))
	pass := os.Getenv("TV_PASS")

	conn, err := net.Dial("tcp", host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	key, _ := pbkdf2.Key(sha256.New, pass, salt, iter, bsze)

	cmd := strings.Join(os.Args[1:], " ")
	if len(cmd) == 0 {
		cmd = "MUTE_STATE"
	}

	conn.Write(encode(key, cmd+"\r"))

	buffer := make([]byte, 4096)
	n, _ := conn.Read(buffer)

	if n == 0 {
		fmt.Println("no response received")
		return
	}

	response := buffer[:n]
	fmt.Print(decode(key, response))
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
