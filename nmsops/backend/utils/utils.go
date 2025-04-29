package utils

import (
	"errors"
	"math"
	"net"
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

func GetIpListFromCIDRNetworkIp(CIDRNetworkIp string) []string {

	ipSplit := strings.Split(CIDRNetworkIp, "/")

	ip := ConvertIpToNumeric(ipSplit[0])

	networkBits, _ := strconv.Atoi(ipSplit[1])

	var subnet uint32

	for i := 0; i < networkBits; i++ {

		subnet = subnet | 1<<(32-i-1)

	}

	networkIp := ip & subnet

	var IpList []string

	for i := uint32(1); i < 1<<(32-networkBits)-1; i++ {
		IpList = append(IpList, ConvertNumericToIp(networkIp+i))
	}

	return IpList

}

var ErrInvalidCIDRIp = errors.New("invalid CIDR IP")

func ValidateCIDRIp(ip string) bool {

	_, _, err := net.ParseCIDR(ip)

	if err != nil {

		return false

	} else {

		return true

	}

}

func ValidateIpAddress(IpAddress string) bool {

	ip := net.ParseIP(IpAddress)

	if ip == nil {

		return false

	} else {

		return true

	}

}
