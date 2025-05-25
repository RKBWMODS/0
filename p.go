package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var scraper = []string{
	"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
	"https://raw.githubusercontent.com/mmpx12/proxy-list/master/socks5.txt",
	"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-socks5.txt",
	"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/socks5.txt",
	"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/socks5.txt",
	"https://raw.githubusercontent.com/roosterkid/openproxylist/main/SOCKS5_RAW.txt",
	"https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt",
	"https://raw.githubusercontent.com/HyperBeats/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/rdavydov/proxy-list/main/proxies/socks5.txt",
	"https://raw.githubusercontent.com/zevtyardt/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/sunny9577/proxy-scraper/master/proxies.txt",
	"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt",
	"https://raw.githubusercontent.com/saschazesiger/Free-Proxies/master/proxies/socks5.txt",
	"https://raw.githubusercontent.com/MuRongPIG/Proxy-Master/main/socks5.txt",
	"https://raw.githubusercontent.com/ALIILAPRO/Proxy/main/socks5.txt",
	"https://raw.githubusercontent.com/almroot/proxylist/master/list.txt",
	"https://raw.githubusercontent.com/opsxcq/proxy-list/master/list/socks5.txt",
	"https://raw.githubusercontent.com/UserR3X/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/Zaeem20/FREE_PROXIES_LIST/master/socks5.txt",
	"https://raw.githubusercontent.com/rdavydov/proxy-list/main/proxies_anonymous/socks5.txt",
	"https://raw.githubusercontent.com/hendrikbgr/Free-Proxy-Repo/main/proxy_list/socks5.txt",
	"https://raw.githubusercontent.com/officialputuid/KangProxy/KangProxy/socks5.txt",
	"https://raw.githubusercontent.com/Androkon/Free-Proxies/main/proxies/socks5.txt",
	"https://raw.githubusercontent.com/saisuiu/Lionkings-Http-Proxys-Proxies/main/socks5.txt",
	"https://raw.githubusercontent.com/B4RC0DE-TM/proxy-list/main/socks5.txt",
	"https://raw.githubusercontent.com/ALIILAPRO/Proxy/main/socks5_all.txt",
	"https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/protocols/socks5/data.txt",
	"https://raw.githubusercontent.com/handeveloper1/Proxy/main/Proxies-Ercin/socks5.txt",
	"https://raw.githubusercontent.com/Vann-Dev/proxy-list/main/proxies/https.txt",
	"https://raw.githubusercontent.com/handeveloper1/Proxy/main/Proxies-Ercin/https.txt",
	"https://raw.githubusercontent.com/handeveloper1/Proxy/main/Proxy-Zaeem20/https.txt",
	"https://raw.githubusercontent.com/handeveloper1/Proxy/main/Proxy-hendrikbgr/proxy_list.txt",
}

var (
	goodProxies  []string
	totalProxies int
	checkedCount int
	badCount     int
	wg           sync.WaitGroup
	mu           sync.Mutex
	startTime    time.Time
)

func color(s string, code string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", code, s)
}

func printLogo() {
	logo := []string{
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣀⣠⣤⣤⣀⡠",
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣤⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣧",
		"⠀⠀⠀⠀⠀⠀⠈⠀⠄⠀⣀⣤⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿",
		"⠀⠀⠀⠀⠀⠀⠀⢀⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠈",
		"⠀⠀⠀⠀⢀⣁⢾⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⢋⣭⡍⣿⣿⣿⣿⣿⣿⠐",
		"⠀⢀⣴⣶⣶⣝⢷⡝⢿⣿⣿⣿⠿⠛⠉⠀⠂⣰⣿⣿⢣⣿⣿⣿⣿⣿⣿⡇",
		"⢀⣾⣿⣿⣿⣿⣧⠻⡌⠿⠋⠡⠁⠈⠀⠀⢰⣿⣿⡏⣸⣿⣿⣿⣿⣿⣿⣿",
		"⣼⣿⣿⣿⣿⣿⣿⡇⠁⠀⠀⠐⠀⠀⠀⠀⠈⠻⢿⠇⢻⣿⣿⣿⣿⣿⣿⡟",
		"⠙⢹⣿⣿⣿⠿⠋⠀⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠻⢿⣿⣿⡿⠟⠁",
		"⠀⠀⠉⠁															 ",
	}
	for _, line := range logo {
		fmt.Println(color(line, "1;36"))
		time.Sleep(80 * time.Millisecond)
	}
}

func printInfoPanel(total int) {
	fmt.Println(color("╭────────────[ INFO PROXY ]────────────╮", "1;33"))
	fmt.Printf("│ TOTAL SCRAP  : %s\n", color(fmt.Sprintf("%d", total), "1;32"))
	fmt.Printf("│ STATUS       : %s\n", color("READY", "1;36"))
	fmt.Printf("│ SUPPORT      : %s\n", color("SOKS5", "1;36"))
	fmt.Println(color("╰──────────────────────────────────────╯", "1;33"))
}

func main() {
	clearScreen()
	fmt.Println()
	proxies := fetchAllProxies()
	totalProxies = len(proxies)

	printLogo()
	printInfoPanel(totalProxies)

	fmt.Print(color("\n$ [ Diz~Stresser ] Ketik 'mulai' \n>> ", "1;32"))
	var cmd string
	fmt.Scanln(&cmd)
	if strings.ToLower(cmd) != "mulai" {
		fmt.Println(color("[?] Tidak ada command!", "1;31"))
		return
	}

	startTime = time.Now()
	go showProgress()

	checkProxies(proxies)
	saveToFile("proxy.txt")
	duration := time.Since(startTime)

	clearScreen()
	fmt.Println(color("\n╭─────────────[ SELESAI ]─────────────╮", "1;32"))
	fmt.Printf("│ PROXY AKTIF  : %s\n", color(fmt.Sprintf("%d", len(goodProxies)), "1;32"))
	fmt.Printf("│ PROXY MATI   : %s\n", color(fmt.Sprintf("%d", badCount), "1;31"))
	fmt.Printf("│ DURASI       : %s\n", duration.Truncate(time.Second))
	fmt.Printf("│ FILE DISIMPAN: %s\n", "proxy.txt")
	fmt.Println(color("╰──────────────────────────────────────╯", "1;32"))
	execTermuxAPI()
}

func fetchAllProxies() []string {
	var proxies []string
	client := &http.Client{Timeout: 15 * time.Second}
	for _, url := range scraper {
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && strings.Contains(line, ":") {
				proxies = append(proxies, line)
			}
		}
		resp.Body.Close()
	}
	return proxies
}

func checkProxies(proxies []string) {
	sem := make(chan struct{}, 600)
	for _, proxy := range proxies {
		wg.Add(1)
		sem <- struct{}{}
		go func(p string) {
			defer wg.Done()
			if isProxyAlive(p) {
				mu.Lock()
				goodProxies = append(goodProxies, p)
				mu.Unlock()
			} else {
				mu.Lock()
				badCount++
				mu.Unlock()
			}
			mu.Lock()
			checkedCount++
			mu.Unlock()
			<-sem
		}(proxy)
	}
	wg.Wait()
}

func isProxyAlive(proxy string) bool {
	conn, err := net.DialTimeout("tcp", proxy, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func saveToFile(path string) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println(color("[!] Gagal menyimpan file.", "1;31"))
		return
	}
	defer file.Close()
	for _, proxy := range goodProxies {
		file.WriteString(proxy + "\n")
	}
}

func showProgress() {
	for checkedCount < totalProxies {
		mu.Lock()
		percent := float64(checkedCount) / float64(totalProxies) * 100
		mu.Unlock()
		fmt.Printf("\r[%3.0f%% ] CHECKING %s", percent, spinner())
		time.Sleep(150 * time.Millisecond)
	}
	fmt.Printf("\r[%3.0f%%] SELESAI!\n", 100.0)
}

func spinner() string {
	spinners := []string{"|", "/", "-", "\\"}
	return spinners[time.Now().UnixNano()/1e6%4]
}

func clearScreen() {
	exec.Command("clear").Run()
}

func execTermuxAPI() {
	exec.Command("termux-toast", "Proxy checker selesai!").Run()
	exec.Command("termux-vibrate", "-d", "300").Run()
}
