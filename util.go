package pydantic

import (
	"strconv"
	"strings"
)

func atoiDefault(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func atofDefault(s string) float64 {
	n, _ := strconv.ParseFloat(s, 64)
	return n
}

func splitWords(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Fields(s)
}

func joinWords(in []string) string {
	return strings.Join(in, " ")
}
