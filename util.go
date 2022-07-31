package rsh

import (
	"log"
	"strconv"
)

func parseUint16(s string) uint16 {
	u, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		log.Println("Error parsing uint:", err)
		return 0
	}

	return uint16(u)
}
