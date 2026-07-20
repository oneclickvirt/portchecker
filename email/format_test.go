package email

import (
	"strings"
	"testing"
)

func TestFormatMailReportKeepsCompactLegacyTable(t *testing.T) {
	report := MailReport{
		Local:          []EndpointResult{{Port: 25, Status: MailUnavailable}},
		OutboundSMTP25: []EndpointResult{{Platform: "Gmail", Status: MailAvailable}},
		MX:             []EndpointResult{{Platform: "Gmail", Status: MailTimeout}},
		Fixed: []EndpointResult{
			{Platform: "Gmail", Protocol: "smtp", Status: MailAvailable},
			{Platform: "Gmail", Protocol: "smtps", Status: MailRefused},
		},
	}
	output := FormatMailReport(report)
	for _, want := range []string{"Platform", "OUT25", "MX25", "Gmail", "OK", "TO", "RF", "LocalPort"} {
		if !strings.Contains(output, want) {
			t.Fatalf("formatted report missing %q: %s", want, output)
		}
	}
	if strings.Contains(output, "{") || strings.Contains(output, "schema_version") {
		t.Fatalf("text report exposed JSON: %s", output)
	}
}
