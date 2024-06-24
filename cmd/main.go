package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/oneclickvirt/portchecker/email"
	"github.com/oneclickvirt/portchecker/model"
)

func main() {
	flag.Parse()
	go func() {
		http.Get("https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fportchecker&count_bg=%2323E01C&title_bg=%23555555&icon=sonarcloud.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false")
	}()
	showVersion := false
	flag.BoolVar(&showVersion, "v", false, "show version")
	if showVersion {
		fmt.Println(model.Version)
		return
	}
	fmt.Println("项目地址:", "https://github.com/oneclickvirt/portchecker")
	res := email.EmailCheck()
	fmt.Printf("%s\n", res)
}
