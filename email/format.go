package email

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/oneclickvirt/portchecker/model"
)

// FormatMailReport renders the structured result in the compact table used by
// the standalone command. JSON callers should encode MailReport directly.
func FormatMailReport(report MailReport) string {
	var builder strings.Builder
	builder.WriteString("Platform      OUT25   MX25    SMTP    SMTPS   POP3    POP3S   IMAP    IMAPS\n")
	local := make([]string, 0, len(report.Local))
	for _, endpoint := range report.Local {
		local = append(local, fmt.Sprintf("%d:%s", endpoint.Port, compactMailStatus(endpoint.Status)))
	}
	if len(local) > 0 {
		fmt.Fprintf(&builder, "%-12s  %s\n", "LocalPort", strings.Join(local, " "))
	}

	platforms := orderedMailPlatforms(report)
	for _, platform := range platforms {
		fixed := make(map[string]MailStatus)
		for _, endpoint := range report.Fixed {
			if endpoint.Platform == platform {
				fixed[strings.ToLower(endpoint.Protocol)] = endpoint.Status
			}
		}
		fmt.Fprintf(&builder, "%s  %s  %s  %s  %s  %s  %s  %s  %s\n",
			padMailCell(platform, 12),
			padMailCell(compactMailStatus(platformMailStatus(report.OutboundSMTP25, platform)), 6),
			padMailCell(compactMailStatus(platformMailStatus(report.MX, platform)), 6),
			padMailCell(compactMailStatus(fixed["smtp"]), 6),
			padMailCell(compactMailStatus(fixed["smtps"]), 6),
			padMailCell(compactMailStatus(fixed["pop3"]), 6),
			padMailCell(compactMailStatus(fixed["pop3s"]), 6),
			padMailCell(compactMailStatus(fixed["imap"]), 6),
			padMailCell(compactMailStatus(fixed["imaps"]), 6),
		)
	}
	builder.WriteString("Status: OK=available  --=unavailable  TO=timeout  RF=refused  ER=error  NA=unsupported")
	return builder.String()
}

func orderedMailPlatforms(report MailReport) []string {
	seen := make(map[string]struct{})
	for _, group := range [][]EndpointResult{report.OutboundSMTP25, report.MX, report.Fixed} {
		for _, endpoint := range group {
			if endpoint.Platform != "" {
				seen[endpoint.Platform] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(seen))
	for _, platform := range model.Platforms {
		if _, ok := seen[platform]; ok {
			result = append(result, platform)
			delete(seen, platform)
		}
	}
	var extra []string
	for platform := range seen {
		extra = append(extra, platform)
	}
	sort.Strings(extra)
	return append(result, extra...)
}

func platformMailStatus(results []EndpointResult, platform string) MailStatus {
	status := MailStatus("")
	for _, result := range results {
		if result.Platform != platform {
			continue
		}
		if result.Status == MailAvailable {
			return MailAvailable
		}
		if status == "" || mailStatusPriority(result.Status) > mailStatusPriority(status) {
			status = result.Status
		}
	}
	return status
}

func mailStatusPriority(status MailStatus) int {
	switch status {
	case MailError:
		return 5
	case MailTimeout:
		return 4
	case MailRefused:
		return 3
	case MailUnavailable:
		return 2
	case MailUnsupported:
		return 1
	default:
		return 0
	}
}

func compactMailStatus(status MailStatus) string {
	switch status {
	case MailAvailable:
		return "OK"
	case MailTimeout:
		return "TO"
	case MailRefused:
		return "RF"
	case MailError:
		return "ER"
	case MailUnsupported:
		return "NA"
	case MailUnavailable, "":
		return "--"
	default:
		return "--"
	}
}

func padMailCell(value string, width int) string {
	if utf8.RuneCountInString(value) > width {
		runes := []rune(value)
		value = string(runes[:max(width-3, 0)]) + "..."
	}
	return value + strings.Repeat(" ", max(width-utf8.RuneCountInString(value), 0))
}
