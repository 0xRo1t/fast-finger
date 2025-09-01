package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

func initHTTPClient() {
	httpc = &http.Client{
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			DialContext:         (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			MaxIdleConns:        1000,
			DisableCompression:  true, // 👈 添加这一行
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, MinVersion: tls.VersionTLS10,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					// 👇 加上 CBC 的老套件
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA}}, // 跳过证书校验
		},
		Timeout: 30 * time.Second, // 整体请求超时
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 阻止自动跳转，直接返回原始 301/302
			return http.ErrUseLastResponse
		},
	}

}

var httpc *http.Client

func readBodyUtf8(resp *http.Response) (string, error) {
	var reader io.ReadCloser
	var err error

	// 1. 处理压缩
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer reader.Close()
	case "deflate":
		reader = flate.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	rawBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	// 2. 转 UTF-8
	encReader, err := charset.NewReader(bytes.NewReader(rawBytes), resp.Header.Get("Content-Type"))
	if err != nil {
		// 如果没法识别编码，就直接当 UTF-8 用
		return string(rawBytes), nil
	}

	utf8Bytes, err := io.ReadAll(encReader)
	if err != nil {
		return "", err
	}

	return string(utf8Bytes), nil
}
func getRedirectURL(target string) string {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return target
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	resp, err := httpc.Do(req)
	if err != nil || resp == nil {
		return target
	}
	defer resp.Body.Close()

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		loc := resp.Header.Get("Location")
		if loc != "" {
			parsedRaw, err1 := url.Parse(target)
			parsedLoc, err2 := url.Parse(loc)
			if err1 == nil && err2 == nil && !parsedLoc.IsAbs() {
				loc = parsedRaw.ResolveReference(parsedLoc).String()
			}
			target = loc
		}
	}
	finalURL := target
	if resp.StatusCode == 401 || resp.StatusCode == 402 || resp.StatusCode == 200 {

		//body, _ := io.ReadAll(resp.Body) // 读取整个响应体
		bodyStr, err := readBodyUtf8(resp)
		//fmt.Println("==========", bodyStr)
		redirectRegexes := []*regexp.Regexp{
			regexp.MustCompile(`(?i)<meta[^>]+url=['"]?([^'">]+)['"]?`),
			regexp.MustCompile(`(?i)window\.(?:location|top\.location)(?:\.href)?\s*=\s*['"]([^'"]+)['"]`),
			regexp.MustCompile(`(?i)<(?:frameset|frame)[^>]+src=['"]?([^'"]+)['"]?`),
			//regexp.MustCompile(`(?i)\.location\s*=\s*['"]([^'"]+)['"]`),
		}
		// 在 for 循环之前添加这行
		for _, re := range redirectRegexes {
			matches := re.FindStringSubmatch(bodyStr)

			if len(matches) == 2 {
				finalURL = strings.TrimSpace(matches[1])
				break
			}
		}

		if finalURL == target {
			return target
		}

		parsedTarget, err := url.Parse(target)
		if err != nil {
			return target
		}
		parsedFinal, err := url.Parse(finalURL)
		if err != nil {
			return target
		}

		if !parsedFinal.IsAbs() {
			finalURL = parsedTarget.ResolveReference(parsedFinal).String()
		}
	}
	//fmt.Println("跳转", finalURL)
	return finalURL
}

func re_url(rawurl string) []string {
	//fmt.Printf("============re_url rawurl============\n", rawurl)
	urls := []string{}
	//urls := []string{rawurl}  // 这种可以让这个函数返回带原来url的
	req, err := http.NewRequest("GET", rawurl, nil)
	if err != nil {
		return urls
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.5845.111 Safari/537.36")
	resp, err := httpc.Do(req)
	if err != nil || resp == nil {
		//fmt.Println("http request error:", err)
		return urls
	}
	defer resp.Body.Close()
	visited := make(map[string]bool)
	// 如果状态码是 301/302，就解析 Location 头
	//fmt.Printf(" %s ", resp.StatusCode)
	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		loc := resp.Header.Get("Location")
		if loc != "" {
			parsedRaw, err1 := url.Parse(rawurl)
			parsedLoc, err2 := url.Parse(loc)
			if err1 == nil && err2 == nil && !parsedLoc.IsAbs() {
				loc = parsedRaw.ResolveReference(parsedLoc).String()
			}
			if !visited[loc] {
				visited[loc] = true
				urls = append(urls, loc)
			}
		}
	}
	//========================================================

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
	bodyStr := string(body)
	// 只匹配完整 URL
	re := regexp.MustCompile(`(?i)<a[^>]+href=['"]?(https?://[^'">]+)['"]?`)
	matches := re.FindAllStringSubmatch(bodyStr, -1)

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		link := strings.TrimSpace(m[1])
		if link == "" {
			continue
		}
		blacklist := []string{
			// CMS、博客、论坛官网
			"wordpress.org",
			"joomla.org",
			"drupal.org",
			"phpbb.com",
			"discourse.org",
			"nginx.com",
			"beian.miit.gov.cn",
			".gov.cn",
			// Web 服务器官网
			"nginx.org",
			"apache.org",
			"lighttpd.net",
			"caddyserver.com",

			// CDN / 公共托管
			"cloudflare.com",
			"fastly.com",
			"akamai.com",
			"github.com",
			"gitlab.com",
			"bitbucket.org",

			// 公共 JS / CSS 库
			"cdnjs.com",
			"bootstrapcdn.com",
			"jquery.com",

			// 常见示例域名
			"example.com",
			"example.org",
			"example.net",

			// 操作系统、软件官网
			"microsoft.com",
			"apple.com",
			"google.com",
			"mozilla.org",
			"ubuntu.com",
			"debian.org",
			"centos.org",
		}

		// 检查黑名单
		skip := false
		for _, b := range blacklist {
			if strings.Contains(link, b) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		if !visited[link] {
			visited[link] = true
			urls = append(urls, link)
		}
	}
	//fmt.Printf("re_函数拿到的 urls %s\n", urls)
	return urls
}
