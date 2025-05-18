package utils

import (
	"math"
	"strconv"
	"strings"
)

func ConvertIpToNumeric(ip string) uint32 {

	var numericIp uint32

	for index, octet := range strings.Split(ip, ".") {

		octetNum, _ := strconv.Atoi(octet)

		numericIp = numericIp | (uint32(octetNum) << ((3 - index) * 8))

	}

	return numericIp

}

func ConvertNumericToIp(ip uint32) string {

	var ipString string

	for i := range 4 {

		ipString += strconv.Itoa(int((ip>>((3-i)*8))&math.MaxUint8)) + "."

	}

	return ipString[:len(ipString)-1]

}
