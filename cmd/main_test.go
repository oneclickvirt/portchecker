package main

import (
	"context"
	"testing"

	"github.com/oneclickvirt/portchecker/email"
)

func TestStructuredMainReportHasVersionedSchema(t *testing.T) {
	report := email.CheckMail(context.Background(), nil, nil, nil, nil)
	if report.SchemaVersion != "portchecker.mail/v1" || report.GeneratedAt.IsZero() {
		t.Fatalf("unexpected report metadata: %+v", report)
	}
}
