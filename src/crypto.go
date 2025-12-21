package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
)

// functions for the encoding / deconding of messages to the TV, which is symmetric and based on a pass code set up on the TV itself
// spec can be seen defined in ./LG_IP.pdf

var salt = []byte{0x63, 0x61, 0xb8, 0x0e, 0x9b, 0xdc, 0xa6, 0x63, 0x8d, 0x07, 0x20, 0xf2, 0xcc, 0x56, 0x8f, 0xb9}
var iter = 16384
var bsze = 16

func get_key(pass string) []byte {
	key, _ := pbkdf2.Key(sha256.New, pass, salt, iter, bsze)
	return key
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
