package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
)

const COLOR = "\033[93m" // Bright Green
const RESET = "\033[0m"

const LOWER_BLOCK = COLOR + "▄" + RESET
const EMPTY_BLOCK = COLOR + " " + RESET
const UPPER_BLOCK = COLOR + "▀" + RESET
const DOUBLE_BLOCK = COLOR + "█" + RESET

var port int
var enable_ipv6 bool
var file string

func init() {
	flag.StringVar(&file, "file", "", "file to serve")
	flag.StringVar(&file, "f", "", "(shorthand) file to serve")
	flag.IntVar(&port, "port", 6066, "port to serve the file on, default: 6066")
	flag.BoolVar(&enable_ipv6, "enable_ipv6", false, "Include ipv6 addresses, default: false")
	flag.BoolVar(&enable_ipv6, "e", false, "(shorthand) Include ipv6 addresses, default: false")
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	if file == "" {
		fmt.Println("Need a file name. For help: serverfile -h")
		os.Exit(1)
	}

	fileExists(file)

	ips, err := externalIP()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	random_url_base := randSeq(6)
	var ip_string strings.Builder
	for _, ip := range ips {
		ip_string.WriteString(fmt.Sprintf("http://%s:%d/%s\n\n", ip, port, random_url_base))
	}

	config := qrterminal.Config{
		Level:          qrterminal.M,
		Writer:         os.Stdout,
		HalfBlocks:     true,
		BlackChar:      EMPTY_BLOCK,
		WhiteBlackChar: UPPER_BLOCK,
		WhiteChar:      DOUBLE_BLOCK,
		BlackWhiteChar: LOWER_BLOCK,
		QuietZone:      1,
	}
	qrterminal.GenerateWithConfig(ip_string.String(), config)

	http.HandleFunc(fmt.Sprintf("/%s", random_url_base), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(file)))
		http.ServeFile(w, r, file)
	})

	log.Printf("Start web server on port %d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), logRequest(http.DefaultServeMux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func fileExists(file string) {
	_, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
			if ip.To4() != nil {
				ip = ip.To4()
			} else if enable_ipv6 && ip.To16() != nil {
				ip = ip.To16()
			} else {
				continue
			}
			ip_addrs = append(ip_addrs, ip.String())
		}
	}
	if len(ip_addrs) > 0 {
		return ip_addrs, nil
	}
	return ip_addrs, errors.New("are you connected to the network?")
}
