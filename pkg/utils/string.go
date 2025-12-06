package utils

import (
	"strconv"
	"strings"
)

type Hp struct {
	Host string
	Port int
}

func ExtractHp(addresses string) []Hp {
	outer := strings.Split(addresses, ",")
	hpSli := make([]Hp, 0, len(outer))
	for _, out := range outer {
		if len(out) == 0 {
			continue
		}

		hpSli = append(hpSli, ExtractOneHp(out))
	}

	return hpSli
}

func ExtractOneHp(addr string) Hp {
	sli := strings.Split(addr, ":")
	port, _ := strconv.Atoi(sli[1])

	return Hp{sli[0], port}
}
