package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// 检查本地25端口是否开放
func isLocalPortOpen(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// 检测是否可连接SMTP服务器的25端口
func checkSMTPConnection(name string, host string, port string, resultChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		resultChan <- fmt.Sprintf("%-9s %-19s: ❌", name, host)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		resultChan <- fmt.Sprintf("%-9s %-19s: ❌", name, host)
		return
	}
	status := "❌"
	if strings.Split(response, " ")[0] == "220" {
		status = "✔"
	}
	resultChan <- fmt.Sprintf("%-9s %-19s: %s", name, host, status)
}

func smtpCheck(language string) string {
	smtpServers := map[string]string{
		"Gmail":     "smtp.gmail.com",
		"163":       "smtp.163.com",
		"Yandex":    "smtp.yandex.com",
		"Office365": "smtp.office365.com",
		"QQ":        "smtp.qq.com",
		"Outlook":   "smtp.outlook.com",
		"Yahoo":     "smtp.mail.yahoo.com",
		"Apple":     "smtp.mail.me.com",
		"MailRU":    "smtp.mail.ru",
		"AOL":       "smtp.aol.com",
		"GMX":       "smtp.gmx.com",
		"MailCOM":   "smtp.mail.com",
		"Sohu":      "smtp.sohu.com",
		"Sina":      "smtp.sina.com",
	}

	var wg sync.WaitGroup
	resultChan := make(chan string, len(smtpServers))

	for name, host := range smtpServers {
		wg.Add(1)
		go checkSMTPConnection(name, host, "25", resultChan, &wg)
	}
	wg.Wait()
	close(resultChan)
	var results []string
	for result := range resultChan {
		results = append(results, result)
	}
	var res string
	if language == "zh" {
		if isLocalPortOpen("25") {
			res += "本地25端口                   : ✔\n"
		} else {
			res += "本地25端口                   : ❌\n"
		}
	} else {
		if isLocalPortOpen("25") {
			res += "Local 25 Port               : ✔\n"
		} else {
			res += "Local 25 Port               : ❌\n"
		}
	}
	for _, result := range results {
		res += result + "\n"
	}
	return res
}

func main() {
	flag.Parse()
	go func() {
		http.Get("https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fportchecker&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false")
	}()
	fmt.Println("项目地址:", "https://github.com/oneclickvirt/portchecker")
	languagePtr := flag.String("l", "", "Language parameter (en or zh)")
	var language string
	if *languagePtr == "" {
		language = "zh"
	} else {
		language = strings.ToLower(*languagePtr)
	}
	res := smtpCheck(language)
	fmt.Printf(res)
}