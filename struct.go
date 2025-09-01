package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html/charset"
)

// Fingerprint 结构体定义了单个指纹的规则
type Fingerprint struct {
	Cms      string   `json:"cms"`
	Keyword  []string `json:"keyword"`
	Location string   `json:"location"`
	Method   string   `json:"method"`
}

// 结构体用于解析整个指纹文件
type FingerprintData struct {
	Fingerprints []Fingerprint `json:"fingerprint"`
}

type scanResult struct {
	InputURL string       `json:"input_url"` // 原始 URL
	FinalURL string       `json:"final_url"` // 带路径的 URL
	Status   string       `json:"status"`    // HTTP 状态码
	Title    string       `json:"title"`     // 页面标题
	Matches  []string     `json:"matches"`   // 指纹匹配结果
	Path     string       `json:"path"`      // 匹配路径
	Err      string       `json:"err"`       // 错误信息
	Body     string       `json:"-"`
	BodyLen  int          `json:"body_len"`
	Children []scanResult // 子链接扫描结果
}

// 用于存储从文件加载的指纹规则
var Fingerprints []Fingerprint

func load_fingerjson() {
	file, err := os.ReadFile("finger.json")
	if err != nil {
		log.Fatalf("无法读取指纹文件: %v", err)
	}
	var fpData FingerprintData
	if err := json.Unmarshal(file, &fpData); err != nil {
		log.Fatalf("无法解析指纹文件: %v", err)
	}
	Fingerprints = fpData.Fingerprints
}

func extractTitleWithEncoding(resp *http.Response, b []byte) string {
	reader, err := charset.NewReader(bytes.NewReader(b), resp.Header.Get("Content-Type"))
	if err != nil {
		return "" // 转码失败直接返回空
	}
	utf8Bytes, err := io.ReadAll(reader)
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`(?is)<\s*title[^>]*>(.*?)</\s*title\s*>`)
	if m := re.FindSubmatch(utf8Bytes); len(m) == 2 {
		return strings.TrimSpace(string(m[1]))
	}
	return ""
}

func ensureHTTP(s string) []string {
	// 移除可能已经存在的协议前缀，以便我们能够重新添加
	if strings.HasPrefix(s, "http://") {
		s = strings.TrimPrefix(s, "http://")
	} else if strings.HasPrefix(s, "https://") {
		s = strings.TrimPrefix(s, "https://")
	}

	// 创建一个包含两个 URL 的切片
	urls := []string{
		"http://" + s,
		"https://" + s,
	}

	return urls
}

func flattenHeaders(h map[string][]string) string {
	var b strings.Builder
	for k, v := range h {
		b.WriteString(strings.ToLower(k) + ":" + strings.Join(v, ",") + "\n")
	}
	return b.String()
}
func remove_urls(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// Result 存储每个 URL 的请求结果
// removeDuplicate 去掉字符串切片里的重复元素
func removeDuplicate(strs []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(strs))
	for _, s := range strs {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}
