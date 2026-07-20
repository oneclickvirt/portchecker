package email

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/oneclickvirt/portchecker/model"
)

type fixtureMXResolver struct {
	records map[string][]*net.MX
	errors  map[string]error
}

type mxResolverFunc func(context.Context, string) ([]*net.MX, error)

func (f mxResolverFunc) LookupMX(ctx context.Context, domain string) ([]*net.MX, error) {
	return f(ctx, domain)
}

func (r fixtureMXResolver) LookupMX(_ context.Context, domain string) ([]*net.MX, error) {
	if err := r.errors[domain]; err != nil {
		return nil, err
	}
	return r.records[domain], nil
}

type greetingDialer struct{}

func (greetingDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	for _, port := range []string{":465", ":995", ":993"} {
		if strings.HasSuffix(address, port) {
			return nil, errors.New("connection refused")
		}
	}
	return greetingConnection(), nil
}

type listeningDialer struct{}

func (listeningDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	return greetingConnection(), nil
}

func greetingConnection() net.Conn {
	client, server := net.Pipe()
	go func() {
		defer server.Close()
		_, _ = server.Write([]byte("220 fixture.example ESMTP ready\r\n"))
	}()
	return client
}

type refusingDialer struct{}

func (refusingDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	return nil, errors.New("connection refused")
}

type closedListener struct{}

func (closedListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (closedListener) Close() error              { return nil }
func (closedListener) Addr() net.Addr            { return fixtureAddr("fixture") }

type fixtureAddr string

func (a fixtureAddr) Network() string { return string(a) }
func (a fixtureAddr) String() string  { return string(a) }

func TestCheckMailSeparatesCapabilitiesAndSortsMX(t *testing.T) {
	report := CheckMail(context.Background(), []PlatformSpec{{
		Name: "Example", Domain: "example.com", SMTPHost: "smtp.example.com",
	}}, fixtureMXResolver{records: map[string][]*net.MX{
		"example.com": {{Host: "mx20.example.com.", Pref: 20}, {Host: "mx10.example.com.", Pref: 10}},
	}}, greetingDialer{}, nil)

	if len(report.Local) != 6 {
		t.Fatalf("expected six local listen probes, got %d", len(report.Local))
	}
	if len(report.OutboundSMTP25) != 1 || report.OutboundSMTP25[0].Status != MailAvailable {
		t.Fatalf("unexpected outbound SMTP result: %+v", report.OutboundSMTP25)
	}
	if len(report.MX) != 2 || report.MX[0].Preference != 10 || report.MX[1].Preference != 20 {
		t.Fatalf("MX results are not preference ordered: %+v", report.MX)
	}
	if report.MX[0].Kind != KindMXSMTP25 || report.MX[0].Status != MailAvailable {
		t.Fatalf("unexpected MX result: %+v", report.MX[0])
	}
	if len(report.Fixed) != 2 || report.Fixed[0].Kind != KindFixedProtocol {
		t.Fatalf("unexpected fixed results: %+v", report.Fixed)
	}
}

func TestCheckMailRecordsMXLookupFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	report := CheckMail(ctx, []PlatformSpec{{Name: "Example", Domain: "example.com"}}, fixtureMXResolver{
		errors: map[string]error{"example.com": &net.DNSError{Err: "i/o timeout", IsTimeout: true}},
	}, greetingDialer{}, nil)
	if len(report.MX) != 1 || report.MX[0].Status != MailTimeout {
		t.Fatalf("unexpected MX timeout result: %+v", report.MX)
	}
}

func TestCheckMailUsesLoopbackConnectionAsListenerEvidence(t *testing.T) {
	report := CheckMail(context.Background(), nil, fixtureMXResolver{}, listeningDialer{}, nil)
	if len(report.Local) != 6 {
		t.Fatalf("expected six local listen probes, got %d", len(report.Local))
	}
	for _, endpoint := range report.Local {
		if endpoint.Status != MailAvailable || endpoint.Evidence != "loopback_connect_succeeded" || endpoint.Error != "" {
			t.Fatalf("port %d did not preserve local-listener evidence: %+v", endpoint.Port, endpoint)
		}
	}
}

func TestCheckMailDoesNotTreatFreeLocalPortAsListening(t *testing.T) {
	report := CheckMail(context.Background(), nil, fixtureMXResolver{}, refusingDialer{}, nil)
	for _, endpoint := range report.Local {
		if endpoint.Status != MailUnavailable || endpoint.Evidence != "no_loopback_listener" {
			t.Fatalf("port %d was falsely reported listening: %+v", endpoint.Port, endpoint)
		}
	}
}

func TestDefaultPlatformSpecsPreserveExistingPlatforms(t *testing.T) {
	specs := DefaultPlatformSpecs()
	if len(specs) != len(model.Platforms) || specs[0].Name == "" {
		t.Fatalf("expected existing model platforms, got %+v", specs)
	}
	for _, spec := range specs {
		if spec.Domain == "" {
			t.Fatalf("platform %q has no explicit MX domain", spec.Name)
		}
		if spec.Name == "Yahoo" && spec.Domain != "yahoo.com" {
			t.Fatalf("Yahoo domain = %q, want yahoo.com", spec.Domain)
		}
	}
}

func TestDefaultYahooDomainDrivesPreferenceOrderedMXLookup(t *testing.T) {
	var yahoo PlatformSpec
	for _, spec := range DefaultPlatformSpecs() {
		if spec.Name == "Yahoo" {
			yahoo = spec
			break
		}
	}
	queried := make(chan string, 1)
	resolver := mxResolverFunc(func(_ context.Context, domain string) ([]*net.MX, error) {
		queried <- domain
		return []*net.MX{{Host: "mx20.yahoo.test.", Pref: 20}, {Host: "mx10.yahoo.test.", Pref: 10}}, nil
	})
	report := CheckMail(context.Background(), []PlatformSpec{yahoo}, resolver, greetingDialer{}, nil)
	if got := <-queried; got != "yahoo.com" {
		t.Fatalf("default Yahoo MX query domain = %q, want yahoo.com", got)
	}
	if len(report.MX) != 2 || report.MX[0].Preference != 10 || report.MX[1].Preference != 20 {
		t.Fatalf("default Yahoo MX results not preference ordered: %+v", report.MX)
	}
}
