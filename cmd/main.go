package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/oneclickvirt/portchecker/email"
	"github.com/oneclickvirt/portchecker/model"
)

func main() {
	showVersion := flag.Bool("v", false, "show version")
	timeout := flag.Duration("timeout", 30*time.Second, "set mail endpoint check deadline")
	jsonOutput := flag.Bool("json", false, "output the versioned JSON report")
	flag.Parse()
	if *showVersion {
		fmt.Println(model.Version)
		return
	}
	go func() {
		response, err := http.Get("https://hits.spiritlhl.net/portchecker.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
		if err == nil && response != nil {
			_ = response.Body.Close()
		}
	}()
	if *timeout <= 0 || *timeout > 2*time.Minute {
		*timeout = 30 * time.Second
	}
	fmt.Fprintln(os.Stderr, "Repo:", "https://github.com/oneclickvirt/portchecker")
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	report := email.CheckMail(ctx, email.DefaultPlatformSpecs(), nil, nil, nil)
	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	fmt.Println(email.FormatMailReport(report))
}
