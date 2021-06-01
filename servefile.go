package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/mdp/qrterminal/v3"
)

const COLOR = "\033[33m"
const RESET = "\033[0m"

const BLACK_WHITE = COLOR + "▄" + RESET 
const BLACK_BLACK = COLOR + " " + RESET 
const WHITE_BLACK = COLOR + "▀" + RESET 
const WHITE_WHITE = COLOR + "█" + RESET 

var port int
func init() {
	flag.IntVar(&port, "port", 6666, "port to serve the file on default: 6666")
}

func main() {

	ips, err := externalIP()
	if err != nil {
		log.Fatal(err)
	}

	var ip_string strings.Builder

	for _, ip := range ips {
		ip_string.WriteString(fmt.Sprintf("http://%s:%d\n", ip, port))
	}

	config := qrterminal.Config{
		Level: qrterminal.M,
		Writer: os.Stdout,
		HalfBlocks:     true,
		BlackChar:      BLACK_BLACK,
		WhiteBlackChar: WHITE_BLACK,
		WhiteChar:      WHITE_WHITE,
		BlackWhiteChar: BLACK_WHITE,
		QuietZone: 1,
	}
	qrterminal.GenerateWithConfig(ip_string.String(), config)
}

func externalIP() ([]string, error) {
	var ip_addrs []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ip_addrs, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return ip_addrs, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ip_addrs = append(ip_addrs, ip.String())
		}
	}
	if len(ip_addrs) > 0 {
		return ip_addrs, nil
	}
	return ip_addrs, errors.New("are you connected to the network?")
}
