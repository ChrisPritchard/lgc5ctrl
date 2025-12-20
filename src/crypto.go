package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

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
