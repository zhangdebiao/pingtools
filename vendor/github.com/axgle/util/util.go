package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	ipRegex       = `\b(?:\d{1,3}\.){3}\d{1,3}\b$`
	ipRegex2      = "^(25[0-5]|2[0-4]\\d|[0-1]\\d{2}|[1-9]?\\d)\\.(25[0-5]|2[0-4]\\d|[0-1]\\d{2}|[1-9]?\\d)\\.(25[0-5]|2[0-4]\\d|[0-1]\\d{2}|[1-9]?\\d)\\.(25[0-5]|2[0-4]\\d|[0-1]\\d{2}|[1-9]?\\d)$"
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// IPparse parse host/hostfile parameters
func IPparse(host, hostfile string) []string {
	var ips = []string{}

	if hostfile != "" {
		buf, err := ioutil.ReadFile(hostfile)
		if err != nil {
			fmt.Println("open file failed, err:", err)
			os.Exit(-1)
		}
		s1 := strings.Replace(string(buf), "\r\n", "\n", -1)
		s1 = strings.Replace(string(s1), "\r", "\n", -1)
		ips = append(strings.Split(string(s1), "\n"))
		fileLine := len(ips)

		if bytes.Contains(buf, []byte("/")) || bytes.Contains(buf, []byte("-")) {
			for i := 0; i < fileLine; i++ {
				if strings.Contains(ips[i], "-") || strings.Contains(ips[i], "/") {
					ips = append(ips, Iplist(ips[i])...)
				}
			}
			ips = ips[fileLine:]
		}
	} else {
		ips = Iplist(host)
	}

	return ips
}

// Iplist is parse ip address
func Iplist(target string) []string {
	var ipSlice []string

	if strings.Contains(target, "-") {
		// IP segment query

		if strings.Contains(target, "/") {
			cidr1 := strings.Split(target, "/")[0]
			cidr2 := strings.Split(target, "/")[1]
			for _, cidrip := range Segment2IPs(cidr1) {
				ipSlice = append(ipSlice, cidr2IPs(cidrip+"/"+cidr2)...)
			}
		} else if strings.Count(target, "-") == 2 {
			return TwoSegment2IPs(target)
		} else {
			return Segment2IPs(target)
		}

	} else if strings.Contains(target, "/") {
		// C-segment query
		return cidr2IPs(target)
	} else {
		// single IP
		if isIP(target) {
			return append(ipSlice, target)
		}
	}

	return ipSlice
}

func isCIDR(cidr string) bool {
	match, _ := regexp.MatchString(ipRegex, cidr)
	return match
}

func isIP(ip string) (b bool) {
	if m, _ := regexp.MatchString(ipRegex2, ip); !m {
		return false
	}
	return true
}

func increment(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func cidr2IPs(cidr string) []string {
	// CIDR To IP
	var ips []string

	if isCIDR(cidr) {
		fmt.Println("IP Address Format Error!")
		os.Exit(1)
	}

	ipAddr, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println("CIDR IP Format Error!")
		os.Exit(1)
	}

	// CIDR too small eg. 10.0.0.1/7
	cidrlast, err := strconv.Atoi(strings.Split(cidr, "/")[1])
	if cidrlast < 8 {
		fmt.Println("CIDR too Big!")
		os.Exit(1)
	}

	for ip := ipAddr.Mask(ipNet.Mask); ipNet.Contains(ip); increment(ip) {
		if ip[3] != 255 {
			ips = append(ips, ip.String())
		}
	}

	return ips
}

// Segment2IPs ip segment to ip
func Segment2IPs(segment string) []string {
	var ips []string

	ipRange := strings.Split(segment, "-")
	spot0 := strings.Count(ipRange[0], ".")
	spot1 := strings.Count(ipRange[1], ".")

	if spot0 == 3 && spot1 == 0 {
		// c类地址
		ipSegment := strings.Split(ipRange[0], ".")
		startNum, _ := strconv.Atoi(ipSegment[3])
		endNum, _ := strconv.Atoi(ipRange[1])
		realIP := ipRange[0]

		// if a IP address, display the IP address and return
		if !isIP(realIP) {
			fmt.Println("IP Address Format Error!!")
			os.Exit(1)
		}

		if endNum > 256 {
			fmt.Println("IP Range Error!")
			os.Exit(1)
		}

		for n := startNum; n <= endNum; n++ {
			ipTemp := append(ipSegment[0:3], strconv.Itoa(n))
			ip := strings.Join(ipTemp, ".")
			ips = append(ips, ip)
		}
	} else if spot0 == 2 && spot1 == 1 {
		// b类地址
		ipSegment := strings.Split(ipRange[0], ".")
		ipSegment1 := strings.Split(ipRange[1], ".")
		startNum, _ := strconv.Atoi(ipSegment[2])
		endNum, _ := strconv.Atoi(ipSegment1[0])
		realIP := strings.Join(append(ipSegment, ipSegment1[1]), ".")

		if !isIP(realIP) {
			fmt.Println("IP Address Format Error!")
			os.Exit(1)
		}

		if endNum > 256 {
			fmt.Println("IP Range Error!")
			os.Exit(1)
		}

		for n := startNum; n <= endNum; n++ {
			ipTemp := append(ipSegment[0:2], strconv.Itoa(n), ipSegment1[1])
			ip := strings.Join(ipTemp, ".")
			ips = append(ips, ip)
		}
	} else if spot0 == 1 && spot1 == 2 {
		// a类地址
		ipSegment := strings.Split(ipRange[0], ".")
		ipSegment1 := strings.Split(ipRange[1], ".")
		startNum, _ := strconv.Atoi(ipSegment[1])
		endNum, _ := strconv.Atoi(ipSegment1[0])
		realIP := strings.Join(append(ipSegment, ipSegment1[1], ipSegment1[2]), ".")

		if !isIP(realIP) {
			fmt.Println("IP Address Format Error!")
			os.Exit(1)
		}

		if endNum > 256 {
			fmt.Println("IP Range Error!")
			os.Exit(1)
		}

		for n := startNum; n <= endNum; n++ {
			ipTemp := append(ipSegment[0:1], strconv.Itoa(n), ipSegment1[1], ipSegment1[2])
			ip := strings.Join(ipTemp, ".")
			ips = append(ips, ip)
		}
	} else {
		fmt.Println("IP Segment Format Error!")
		os.Exit(1)
	}

	return ips
}

// TwoSegment2IPs parse ip
func TwoSegment2IPs(target string) []string {
	var ips, ipsTemp []string
	var segment1, segment2, segment3 string

	segment1 = strings.Split(target, "-")[0]
	segment2 = strings.Split(target, "-")[1]
	segment3 = strings.Split(target, "-")[2]

	if strings.Count(segment1+segment2, ".") == 3 {
		ipsTemp = append(ipsTemp, Segment2IPs(segment1+"-"+segment2)...)
		for _, ip := range ipsTemp {
			ips = append(ips, Segment2IPs(ip+"-"+segment3)...)
		}
	} else if strings.Count(segment1, ".") == 1 && strings.Count(segment2, ".") == 1 && strings.Count(segment3, ".") == 1 {
		ip1 := segment1 + "-" + segment2
		ip2 := strings.Split(segment3, ".")[1]
		ipsTemp = append(ipsTemp, Segment2IPs(ip1+"."+ip2)...)
		for _, ip := range ipsTemp {

			ip3 := strings.Split(ip, ".")[0] + "." + strings.Split(ip, ".")[1] + "." + strings.Split(ip, ".")[2] + "-" + strings.Split(segment3, ".")[0] + "." + strings.Split(ip, ".")[3]
			ips = append(ips, Segment2IPs(ip3)...)
		}
	}

	return ips
}
