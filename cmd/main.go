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

// 检查本地端口是否开放
func isLocalPortOpen(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// 检测是否可连接服务器的指定端口
func checkConnection(host string, port string) string {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return "✘"
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "✘"
	}
	status := "✘"
	if strings.Split(response, " ")[0] == "220" || strings.Split(response, " ")[0] == "+OK" || strings.Contains(response, "* OK") {
		status = "✔"
	}
	return status
}

func emailCheck() string {
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
	pop3Servers := map[string]string{
		"Gmail":     "pop.gmail.com",
		"163":       "pop.163.com",
		"Yandex":    "pop.yandex.com",
		"Office365": "outlook.office365.com",
		"QQ":        "pop.qq.com",
		"Outlook":   "pop-mail.outlook.com",
		"Yahoo":     "pop.mail.yahoo.com",
		"Apple":     "pop.mail.me.com",
		"MailRU":    "pop.mail.ru",
		"AOL":       "pop.aol.com",
		"GMX":       "pop.gmx.com",
		"MailCOM":   "pop.mail.com",
		"Sohu":      "pop.sohu.com",
		"Sina":      "pop.sina.com",
	}
	imapServers := map[string]string{
		"Gmail":     "imap.gmail.com",
		"163":       "imap.163.com",
		"Yandex":    "imap.yandex.com",
		"Office365": "outlook.office365.com",
		"QQ":        "imap.qq.com",
		"Outlook":   "imap-mail.outlook.com",
		"Yahoo":     "imap.mail.yahoo.com",
		"Apple":     "imap.mail.me.com",
		"MailRU":    "imap.mail.ru",
		"AOL":       "imap.aol.com",
		"GMX":       "imap.gmx.com",
		"MailCOM":   "imap.mail.com",
		"Sohu":      "imap.sohu.com",
		"Sina":      "imap.sina.com",
	}
	var wg sync.WaitGroup
	resultChan := make(chan string, len(smtpServers)*3)
	checkAll := func(name string, smtpHost, pop3Host, imapHost string) {
		defer wg.Done()
		smtpResult := checkConnection(smtpHost, "25")
		pop3Result := checkConnection(pop3Host, "110")
		imapResult := checkConnection(imapHost, "143")
		resultChan <- fmt.Sprintf("%-9s %-4s %-4s %-4s", name, smtpResult, pop3Result, imapResult)
	}
	for name := range smtpServers {
		wg.Add(1)
		go checkAll(name, smtpServers[name], pop3Servers[name], imapServers[name])
	}
	wg.Wait()
	close(resultChan)
	var results []string
	for result := range resultChan {
		results = append(results, result)
	}
	var res, status25, status110, status143 string
	if isLocalPortOpen("25") {
		status25 = "✔"
	} else {
		status25 = "✘"
	}
	if isLocalPortOpen("110") {
		status110 = "✔"
	} else {
		status110 = "✘"
	}
	if isLocalPortOpen("143") {
		status143 = "✔"
	} else {
		status143 = "✘"
	}
	res += fmt.Sprintf("%-8s %-4s %-4s %-4s\n", "Platform", "SMTP", "POP3", "IMAP")
	res += fmt.Sprintf("%-10s%-4s %-4s %-4s\n", "LocalPort", status25, status110, status143)
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
	res := emailCheck()
	fmt.Printf(res)
}
