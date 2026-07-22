package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/oneclickvirt/portchecker/email"
	"github.com/oneclickvirt/portchecker/model"
)

type cliOptions struct {
	help, version, jsonOutput bool
	timeout                   time.Duration
	platforms, mxDomain       string
}

func parseCLI(args []string) (cliOptions, error) {
	opts := cliOptions{}
	fs := newFlagSet(&opts, io.Discard)
	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	if fs.NArg() != 0 {
		return opts, fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}
	if opts.help || opts.version {
		return opts, nil
	}
	if opts.timeout <= 0 || opts.timeout > 2*time.Minute {
		return opts, fmt.Errorf("timeout must be greater than zero and at most 2m")
	}
	return opts, nil
}

func newFlagSet(opts *cliOptions, output io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("portchecker", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.BoolVar(&opts.help, "h", false, "show help")
	fs.BoolVar(&opts.version, "v", false, "show version")
	fs.DurationVar(&opts.timeout, "timeout", 30*time.Second, "set mail endpoint check deadline")
	fs.BoolVar(&opts.jsonOutput, "json", false, "output the versioned JSON report")
	fs.StringVar(&opts.platforms, "platforms", "", "comma-separated platform names to check (empty checks all)")
	fs.StringVar(&opts.mxDomain, "mx-domain", "", "add one extra domain for dynamic MX priority checking")
	return fs
}

func printCLIHelp(program string) {
	fmt.Printf("Usage: %s [options]\n", program)
	newFlagSet(&cliOptions{}, os.Stdout).PrintDefaults()
}

func main() {
	opts, err := parseCLI(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, sanitizeErrorText(err.Error()))
		os.Exit(2)
	}
	if opts.help {
		printCLIHelp(os.Args[0])
		return
	}
	if opts.version {
		fmt.Println(model.Version)
		return
	}
	go func() {
		response, err := http.Get("https://hits.spiritlhl.net/portchecker.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
		if err == nil && response != nil {
			_ = response.Body.Close()
		}
	}()
	specs, err := selectPlatformSpecs(email.DefaultPlatformSpecs(), opts.platforms, opts.mxDomain)
	if err != nil {
		fmt.Fprintln(os.Stderr, sanitizeErrorText(err.Error()))
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, "Repo:", "https://github.com/oneclickvirt/portchecker")
	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
	defer cancel()
	report := email.CheckMail(ctx, specs, nil, nil, nil)
	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	fmt.Println(indentLegacyOutput(email.FormatMailReport(report)))
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
