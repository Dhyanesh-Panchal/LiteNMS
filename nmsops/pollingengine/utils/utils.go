package utils

import (
	"math"
	"strconv"
)

func ConvertNumericToIp(ip uint32) string {

	var ipString string

	for i := range 4 {

		ipString += strconv.Itoa(int((ip>>((3-i)*8))&math.MaxUint8)) + "."

	}

	return ipString[:len(ipString)-1]

}
