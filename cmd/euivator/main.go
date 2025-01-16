package main

import (
	"fmt"
	"log"

	"github.com/ttl256/euivator/internal/hwaddr"
)

func main() {
	addr, err := hwaddr.EUI48FromBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(addr) //nolint: forbidigo // fine
}
