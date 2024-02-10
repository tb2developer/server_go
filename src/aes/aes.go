package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"os"

	b64 "encoding/base64"
	"fmt"
)

func Encrypt(src string) string {
	KEY := os.Getenv("KEY1")
	InitialVector := os.Getenv("InitialVector")

	block, err := aes.NewCipher([]byte(KEY))
	if err != nil {
		fmt.Println("key error1", err)
	}
	if src == "" {
		fmt.Println("plain content empty")
	}
	ecb := cipher.NewCBCEncrypter(block, []byte(InitialVector))
	content := []byte(src)
	content = PKCS5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)

	base := b64.StdEncoding.EncodeToString(crypted)
	// base := php.Base64Encode(string(crypted))
	return base
}

func Decrypt(crypt string) string {
	KEY := os.Getenv("KEY1")
	InitialVector := os.Getenv("InitialVector")

	block, err := aes.NewCipher([]byte(KEY))
	if err != nil {
		fmt.Println("key error1", err)
	}
	if len(crypt) == 0 {
		fmt.Println("plain content empty")
	}
	unbase, err := b64.StdEncoding.DecodeString(crypt)
	// unbase, err := php.Base64Decode(crypt)
	if err != nil {
		fmt.Println(err)
	}
	ecb := cipher.NewCBCDecrypter(block, []byte(InitialVector))
	decrypted := make([]byte, len(unbase))
	ecb.CryptBlocks(decrypted, []byte(unbase))

	return string(PKCS5Trimming(decrypted))
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
