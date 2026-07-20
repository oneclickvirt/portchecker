package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/oneclickvirt/portchecker/email"
	"github.com/oneclickvirt/portchecker/model"
)

func main() {
	showVersion := flag.Bool("v", false, "show version")
	timeout := flag.Duration("timeout", 30*time.Second, "set mail endpoint check deadline")
	jsonOutput := flag.Bool("json", false, "output the versioned JSON report")
	platforms := flag.String("platforms", "", "comma-separated platform names to check (empty checks all)")
	mxDomain := flag.String("mx-domain", "", "add one extra domain for dynamic MX priority checking")
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
	specs, err := selectPlatformSpecs(email.DefaultPlatformSpecs(), *platforms, *mxDomain)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, "Repo:", "https://github.com/oneclickvirt/portchecker")
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	report := email.CheckMail(ctx, specs, nil, nil, nil)
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

func selectPlatformSpecs(all []email.PlatformSpec, selection, extraDomain string) ([]email.PlatformSpec, error) {
	selected := strings.TrimSpace(selection)
	result := make([]email.PlatformSpec, 0, len(all)+1)
	if selected == "" {
		result = append(result, all...)
	} else {
		allowed := make(map[string]email.PlatformSpec, len(all))
		for _, spec := range all {
			allowed[strings.ToLower(spec.Name)] = spec
		}
		seen := make(map[string]struct{})
		for _, raw := range strings.Split(selected, ",") {
			name := strings.TrimSpace(raw)
			if name == "" {
				continue
			}
			key := strings.ToLower(name)
			spec, ok := allowed[key]
			if !ok {
				return nil, fmt.Errorf("unknown platform %q; use a name from -platforms or leave it empty", name)
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, spec)
		}
		if len(result) == 0 {
			return nil, fmt.Errorf("-platforms must contain at least one known platform")
		}
	}
	if domain := strings.TrimSpace(extraDomain); domain != "" {
		if strings.ContainsAny(domain, " \t\r\n") {
			return nil, fmt.Errorf("-mx-domain contains whitespace: %q", extraDomain)
		}
		result = append(result, email.PlatformSpec{Name: "Custom", Domain: domain})
	}
	return result, nil
}
