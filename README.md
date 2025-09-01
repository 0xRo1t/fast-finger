## 
1. 只用于扫描根url， 也就是  x.x.x.x/
2. 如果扫描不是根url，不会有结果，因为注释掉了 根url的扫描
main.go里面的
//fmt.Printf("扫描原本的url %s\n", furl)
//match_finger(furl)

## 因为默认不处理 302 或者 301 跳转了
所以指纹只需要写 跳转后的页面即可


$env:GOOS="windows"
$env:GOARCH="amd64"
go build -o echo-pro.exe