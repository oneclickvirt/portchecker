package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oneclickvirt/portchecker/email"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	specs := email.DefaultPlatformSpecs()
	if len(specs) > 2 {
		specs = specs[:2]
	}
	report := email.CheckMail(ctx, specs, nil, nil, nil)
	counts := make(map[email.MailStatus]int)
	for _, group := range [][]email.EndpointResult{report.Local, report.OutboundSMTP25, report.MX, report.Fixed} {
		for _, result := range group {
			counts[result.Status]++
		}
	}
	encoded, _ := json.Marshal(counts)
	fmt.Println(string(encoded))
}
