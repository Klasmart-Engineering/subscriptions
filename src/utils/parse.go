package utils

import "strconv"

func MustParseInt(i string) int {
	value, err := strconv.ParseInt(i, 10, 32)
	if err != nil {
		panic(err)
	}

	return int(value)
}
