package main

import (
	"fmt"
	"os"
	"strings"
)

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	var urls []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			urls = append(urls, line)
		}
	}
	return urls, nil
}
func RunCLI(filePath string) {

	lines, err := readLines(filePath)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return
	}
	scan_finger_file(lines)
}
