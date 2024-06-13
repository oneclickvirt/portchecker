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

type Result struct {
	Platform string
	Status   string
}

func isLocalPortOpen(port string) string {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return "✘"
	}
	ln.Close()
	return "✔"
}

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
	localServers := []string{"25", "465", "110", "995", "143", "993"}
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
	localChan := make(chan Result, len(localServers))
	smtpChan := make(chan Result, len(smtpServers))
	pop3Chan := make(chan Result, len(pop3Servers))
	imapChan := make(chan Result, len(imapServers))
	smtpsChan := make(chan Result, len(smtpServers))
	pop3sChan := make(chan Result, len(pop3Servers))
	imapsChan := make(chan Result, len(imapServers))
	checkLocal := func(port string) {
		defer wg.Done()
		localResult := isLocalPortOpen(port)
		localChan <- Result{port, localResult}
	}
	checkSMTP := func(name, host string) {
		defer wg.Done()
		smtpResult := checkConnection(host, "25")
		smtpChan <- Result{name, smtpResult}
	}
	checkSMTPS := func(name, host string) {
		defer wg.Done()
		smtpSSLResult := checkConnection(host, "465")
		smtpsChan <- Result{name, smtpSSLResult}
	}
	checkPOP3 := func(name, host string) {
		defer wg.Done()
		pop3Result := checkConnection(host, "110")
		pop3Chan <- Result{name, pop3Result}
	}
	checkPOP3S := func(name, host string) {
		defer wg.Done()
		pop3SSLResult := checkConnection(host, "995")
		pop3sChan <- Result{name, pop3SSLResult}
	}
	checkIMAP := func(name, host string) {
		defer wg.Done()
		imapResult := checkConnection(host, "143")
		imapChan <- Result{name, imapResult}
	}
	checkIMAPS := func(name, host string) {
		defer wg.Done()
		imapSSLResult := checkConnection(host, "993")
		imapsChan <- Result{name, imapSSLResult}
	}
	for _, port := range localServers {
		wg.Add(1)
		go checkLocal(port)
	}
	for name, smtpHost := range smtpServers {
		wg.Add(1)
		go checkSMTP(name, smtpHost)
	}
	for name, smtpHost := range smtpServers {
		wg.Add(1)
		go checkSMTPS(name, smtpHost)
	}
	for name, pop3Host := range pop3Servers {
		wg.Add(1)
		go checkPOP3(name, pop3Host)
	}
	for name, pop3Host := range pop3Servers {
		wg.Add(1)
		go checkPOP3S(name, pop3Host)
	}
	for name, imapHost := range imapServers {
		wg.Add(1)
		go checkIMAP(name, imapHost)
	}
	for name, imapHost := range imapServers {
		wg.Add(1)
		go checkIMAPS(name, imapHost)
	}
	wg.Wait()
	close(localChan)
	close(smtpChan)
	close(pop3Chan)
	close(imapChan)
	close(smtpsChan)
	close(pop3sChan)
	close(imapsChan)
	//转换通道提取数据
	temp := []string{}
	smtpChanMap, smtpsChanMap, pop3ChanMap, pop3sChanMap, imapChanMap, imapsChanMap := map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}
	wg.Add(7)
	// 使用goroutine并发处理每个通道
	go func() {
		defer wg.Done()
		for m := range localChan {
			temp = append(temp, m.Status)
		}
	}()
	go func() {
		defer wg.Done()
		for m := range smtpChan {
			smtpChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range smtpsChan {
			smtpsChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range pop3Chan {
			pop3ChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range pop3sChan {
			pop3sChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range imapChan {
			imapChanMap[m.Platform] = m.Status
		}
	}()
	go func() {
		defer wg.Done()
		for m := range imapsChan {
			imapsChanMap[m.Platform] = m.Status
		}
	}()
	wg.Wait()
	var results []string
	results = append(results, fmt.Sprintf("%-9s %-4s %-4s %-4s %-4s %-4s %-4s", "Platform", "SMTP", "SMTPS", "POP3", "POP3S", "IMAP", "IMAPS"))
	results = append(results, fmt.Sprintf("%-10s%-4s %-4s %-4s %-4s %-4s %-4s", "LocalPort", temp[0], temp[1], temp[2], temp[3], temp[4], temp[5]))
	for name, _ := range smtpServers {
		results = append(results, fmt.Sprintf("%-10s%-4s %-4s %-4s %-4s %-4s %-4s", name,
			smtpChanMap[name], smtpsChanMap[name], pop3ChanMap[name], pop3sChanMap[name], imapChanMap[name], imapsChanMap[name]))
	}
	return strings.Join(results, "\n")
}

func main() {
	flag.Parse()
	go func() {
		http.Get("https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fportchecker&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false")
	}()
	fmt.Println("项目地址:", "https://github.com/oneclickvirt/portchecker")
	res := emailCheck()
	fmt.Printf("%s\n", res)
}
