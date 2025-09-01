// main.go
package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// scanUrls 接收 URL 列表，返回扫描结果
func scan_finger_file(urls []string) []scanResult {
	if len(urls) == 0 {
		return nil
	}

	start := time.Now()
	var validResultsCount int32
	var wg sync.WaitGroup
	resultsChan := make(chan scanResult, len(urls))
	seen := make(map[string]struct{})

	for _, url := range urls {
		wg.Add(1)
		go func(u string, redir_url string) {
			redirects := re_url(url)

			defer wg.Done()
			furl := ensureHTTP(u)
			for _, u1 := range furl {
				wg.Add(1)
				go func(u string) {
					defer wg.Done()
					qwe := getRedirectURL(u1) // 这里是跳转的url
					match_finger(qwe, u1)
				}(u1)
			}
			// 这里是处理页面隐藏的完整url
			for _, re := range redirects {
				wg.Add(1)
				go func(re string) {
					defer wg.Done()
					tar := getRedirectURL(re)
					result := match_finger(tar, url)
					matchesKey := strings.Join(result.Matches, "|")
					key := result.InputURL + "|" + matchesKey
					if _, exists := seen[key]; !exists {
						seen[key] = struct{}{}
						resultsChan <- result
					}
				}(re)
			}

		}(url, url)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var finalResults []scanResult
	for result := range resultsChan {
		if len(result.Matches) > 0 {
			atomic.AddInt32(&validResultsCount, 1)
		}
		finalResults = append(finalResults, result)
	}

	fmt.Printf("扫描 %d 个URL 有效数量 %d 总耗时: %s\n", len(urls), validResultsCount, time.Since(start))
	return finalResults
}

func main() {
	initHTTPClient()
	load_fingerjson()
	fmt.Printf("成功加载 finger.json %d 条指纹规则\n", len(Fingerprints))

	filePath := flag.String("f", "", "URL 文件路径，一行一个 URL")
	//web := flag.String("", "-m", "web 模式")
	singleURL := flag.String("u", "", "Single URL to scan")
	flag.Parse()
	if *filePath != "" {
		// CLI 模式
		RunCLI(*filePath)
		return
	} else if *singleURL != "" {
		urls := []string{*singleURL}
		scan_finger_file(urls)
		return
	}
	println("-u 单个url")
	println("-f 文件读取。一行一个")

}
