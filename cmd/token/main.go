package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

const tokenBytes = 32

func main() {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("generate token: %v", err)
	}

	fmt.Println(base64.RawURLEncoding.EncodeToString(buf))
}
