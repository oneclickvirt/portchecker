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
		http.Get("https://hits.spiritlhl.net/portchecker.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
	}()
	showVersion := false
	flag.BoolVar(&showVersion, "v", false, "show version")
	if showVersion {
		fmt.Println(model.Version)
		return
	}
	fmt.Println("项目地址:", "https://github.com/oneclickvirt/portchecker")
	res := email.EmailCheck()
	fmt.Println(res)
}
