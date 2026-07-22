package main

import (
	"context"
	"testing"
	"time"

	"github.com/oneclickvirt/portchecker/email"
)

func TestSelectPlatformSpecsFiltersAndAddsMXDomain(t *testing.T) {
	all := []email.PlatformSpec{{Name: "Gmail", Domain: "gmail.com"}, {Name: "QQ", Domain: "qq.com"}}
	got, err := selectPlatformSpecs(all, "qq,QQ", "example.test")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "QQ" || got[1].Name != "Custom" || got[1].Domain != "example.test" {
		t.Fatalf("unexpected selected specs: %+v", got)
	}
}

func TestParseCLIRejectsInvalidTimeoutAndPositionalArguments(t *testing.T) {
	for _, args := range [][]string{{"-timeout", "0s"}, {"-timeout", "3m"}, {"extra"}} {
		if _, err := parseCLI(args); err == nil {
			t.Fatalf("expected arguments %v to be rejected", args)
		}
	}
	opts, err := parseCLI([]string{"-timeout", "45s", "-json", "-platforms", "Gmail"})
	if err != nil || opts.timeout != 45*time.Second || !opts.jsonOutput || opts.platforms != "Gmail" {
		t.Fatalf("valid options were not retained: opts=%#v err=%v", opts, err)
	}
}

func TestParseCLIHelpAndVersionIgnoreInvalidTimeout(t *testing.T) {
	for _, args := range [][]string{{"-h", "-timeout", "0s"}, {"-v", "-timeout", "3m"}} {
		if _, err := parseCLI(args); err != nil {
			t.Fatalf("help/version arguments %v returned %v", args, err)
		}
	}
}

func TestSelectPlatformSpecsRejectsUnknownOrWhitespaceDomain(t *testing.T) {
	all := []email.PlatformSpec{{Name: "Gmail"}}
	if _, err := selectPlatformSpecs(all, "missing", ""); err == nil {
		t.Fatal("expected unknown platform error")
	}
	if _, err := selectPlatformSpecs(all, "", "bad domain"); err == nil {
		t.Fatal("expected whitespace domain error")
	}
}

func TestStructuredMainReportHasVersionedSchema(t *testing.T) {
	report := email.CheckMail(context.Background(), nil, nil, nil, nil)
	if report.SchemaVersion != "portchecker.mail/v1" || report.GeneratedAt.IsZero() {
		t.Fatalf("unexpected report metadata: %+v", report)
	}
}
