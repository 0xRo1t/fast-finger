package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
)

// 这里传入的 redirect_url
func match_finger(rawURL string, redirect_url string) scanResult {

	req, err := http.NewRequest("GET", rawURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "close")
	resp, err := httpc.Do(req)
	if err != nil {
		//fmt.Printf("http response is nil, %s", err.Error())
		return scanResult{}
	}

	if err != nil || resp == nil {
		//fmt.Println(err)
		return scanResult{}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
	//println(string(body))
	//body, _ := io.ReadAll(resp.Body)
	status := fmt.Sprintf("%d", resp.StatusCode)
	bodyStr := string(body)
	lcBody := strings.ToLower(bodyStr)
	headers := flattenHeaders(resp.Header)
	title := extractTitleWithEncoding(resp, body)
	matches := []string{}

	for _, f := range Fingerprints {
		if len(f.Keyword) == 0 {
			continue
		}
		switch f.Method {
		case "keyword":
			switch f.Location {
			case "header":
				allMatched := true
				for _, kw := range f.Keyword {
					kw = strings.TrimSpace(kw)
					if kw == "" {
						continue
					}
					if !strings.Contains(headers, kw) {
						allMatched = false
						break
					}
				}
				if allMatched {
					matches = append(matches, f.Cms)
				}
			case "title":
				allMatched := true
				for _, kw := range f.Keyword {
					if !strings.Contains(strings.ToLower(title), strings.ToLower(strings.TrimSpace(kw))) {
						allMatched = false
						break
					}
				}
				if allMatched {
					matches = append(matches, f.Cms)
				}
			default: // body
				allMatched := true
				for _, kw := range f.Keyword {
					if !strings.Contains(lcBody, strings.ToLower(strings.TrimSpace(kw))) {
						allMatched = false
						break
					}
				}
				if allMatched {
					matches = append(matches, f.Cms)
				}
			}
		}
	}
	r := scanResult{
		InputURL: rawURL,
		FinalURL: redirect_url,
		Status:   status,
		Title:    title,
		Matches:  removeDuplicate(matches),
		BodyLen:  len(body),
	}
	//fmt.Printf("%s%s%s [%s%s%s]\t[源url: %s%s%s]\n", Yellow, r.InputURL, Reset,
	//	Green, r.Status, Reset,
	//	Cyan, strings.Join(r.Matches, ","), Reset,
	//	Yellow, r.FinalURL, Reset)
	if len(r.Matches) > 0 {
		fmt.Printf("%s%s%s [%s%s%s]\t[源url: %s%s%s]\n", Yellow, r.InputURL, Reset,

			//Green, r.Status, Reset,
			Cyan, strings.Join(r.Matches, ","), Reset,
			Yellow, r.FinalURL, Reset)
	}
	return r

}
