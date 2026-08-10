package gen

import (
	"fmt"

	"github.com/google/uuid"
)

type IDGenerator interface {
	// Generate generates a new unique ID
	Generate() string
}

type UUIDGenerator struct {
	prefix string
}

func (g *UUIDGenerator) Generate() string {
	if g.prefix == "" {
		return uuid.New().String()
	}
	return fmt.Sprintf("%s%s", g.prefix, uuid.New().String())
}

func NewUUIDGenerator(prefix string) IDGenerator {
	return &UUIDGenerator{prefix: prefix}
}

type RandomTokenGenerator struct {
	length int
}

func (g *RandomTokenGenerator) Generate() string {
	return uuid.New().String()[:g.length]
}

func NewRandomTokenGenerator(length int) IDGenerator {
	if length <= 0 || length > 36 {
		length = 8 // default length
	}
	return &RandomTokenGenerator{length: length}
}

type SequentialIDGenerator struct {
	current int
}

func (g *SequentialIDGenerator) Generate() string {
	g.current++
	return fmt.Sprintf("%d", g.current)
}

func NewSequentialIDGenerator(start int) IDGenerator {
	if start < 0 {
		start = 0
	}
	return &SequentialIDGenerator{current: start - 1}
}
