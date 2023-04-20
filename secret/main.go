package main

import (
	"crypto/rand"
	"fmt"
	"log"
)

func main() {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%x\n", b)
}
