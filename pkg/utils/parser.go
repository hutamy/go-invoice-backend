package utils

import "strconv"

func GetUintOrZero(p *uint) uint {
	if p == nil {
		return 0
	}
	return *p
}

func ParseIntDefault(v string, def int) int {
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
