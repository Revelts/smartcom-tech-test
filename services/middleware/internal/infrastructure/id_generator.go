package infrastructure

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type UUIDGenerator struct{}

func NewUUIDGenerator() (gen *UUIDGenerator) {
	gen = &UUIDGenerator{}
	return
}

func (g *UUIDGenerator) Generate() (id string, err error) {
	bytes := make([]byte, 16)
	var n int
	n, err = rand.Read(bytes)
	if err != nil {
		err = fmt.Errorf("failed to generate random bytes: %w", err)
		return
	}
	if n != 16 {
		err = fmt.Errorf("insufficient random bytes generated")
		return
	}

	id = hex.EncodeToString(bytes)
	return
}
