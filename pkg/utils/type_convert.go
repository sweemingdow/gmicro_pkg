package utils

import (
	"math"
	"strconv"
)

func A2i(val string) int {
	i, err := strconv.Atoi(val)

	if err != nil {
		i = math.MinInt
	}

	return i
}

func I2a(val int) string {
	return strconv.Itoa(val)
}

func A2b(val string) bool {
	if b, err := strconv.ParseBool(val); err != nil {
		return false
	} else {
		return b
	}
}
