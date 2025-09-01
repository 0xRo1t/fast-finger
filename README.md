## 优势
1. 利用了多线程,扫描速率快
2. 爬取了目标站点的 子url （完整的url会爬取），接口不会爬取
3. 优化了301 302 403等一些页面跳转无法扫到指纹的情况

## 用法
fast_finger_amd64 -u url
fast_finger_amd64 -f 1.txt

$env:GOOS="linux"
$env:GOARCH="amd64"
go build

$env:GOOS="windows"
$env:GOARCH="amd64"
go build

