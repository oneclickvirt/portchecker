package email

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/oneclickvirt/portchecker/model"
)

// EndpointKind identifies the distinct mail capabilities reported by the
// checker. In particular, local listen and outbound SMTP are separate tests.
type EndpointKind string

const (
	KindLocalListen    EndpointKind = "local_listen"
	KindOutboundSMTP25 EndpointKind = "outbound_smtp25"
	KindMXSMTP25       EndpointKind = "mx_smtp25"
	KindFixedProtocol  EndpointKind = "fixed_protocol"
)

type MailStatus string

const (
	MailAvailable   MailStatus = "available"
	MailUnavailable MailStatus = "unavailable"
	MailTimeout     MailStatus = "timeout"
	MailRefused     MailStatus = "refused"
	MailError       MailStatus = "error"
	MailUnsupported MailStatus = "unsupported"
)

type EndpointResult struct {
	Kind       EndpointKind `json:"kind"`
	Platform   string       `json:"platform,omitempty"`
	Host       string       `json:"host,omitempty"`
	Domain     string       `json:"domain,omitempty"`
	Protocol   string       `json:"protocol,omitempty"`
	Port       int          `json:"port"`
	TLS        bool         `json:"tls"`
	Preference uint16       `json:"mx_preference,omitempty"`
	Status     MailStatus   `json:"status"`
	Evidence   string       `json:"evidence,omitempty"`
	Error      string       `json:"error,omitempty"`
	LatencyMS  int64        `json:"latency_ms"`
}

type MailReport struct {
	SchemaVersion  string           `json:"schema_version"`
	Local          []EndpointResult `json:"local"`
	OutboundSMTP25 []EndpointResult `json:"outbound_smtp25"`
	MX             []EndpointResult `json:"mx"`
	Fixed          []EndpointResult `json:"fixed"`
	GeneratedAt    time.Time        `json:"generated_at"`
}

// PlatformSpec is a provider's fixed endpoints and its domain for MX lookup.
type PlatformSpec struct {
	Name     string
	Domain   string
	SMTPHost string
	POP3Host string
	IMAPHost string
}

// MXResolver is injectable for offline and deterministic tests.
type MXResolver interface {
	LookupMX(ctx context.Context, domain string) ([]*net.MX, error)
}

type systemMXResolver struct{}

func (systemMXResolver) LookupMX(ctx context.Context, domain string) ([]*net.MX, error) {
	return net.DefaultResolver.LookupMX(ctx, domain)
}

// Dialer is the minimal transport dependency needed by CheckMail.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type systemDialer struct{ net.Dialer }

func (d systemDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.Dialer.DialContext(ctx, network, address)
}

// LocalListener is retained for source compatibility with the first
// structured API. CheckMail no longer binds ports to infer listener state;
// local listeners are detected by loopback connections through Dialer.
type LocalListener interface {
	Listen(network, address string) (net.Listener, error)
}

type protocolEndpoint struct {
	protocol string
	port     int
	tls      bool
	host     string
}

type endpointJob struct {
	result EndpointResult
	secure bool
}

// DefaultPlatformSpecs converts the existing model maps into structured specs.
func DefaultPlatformSpecs() []PlatformSpec {
	result := make([]PlatformSpec, 0, len(model.Platforms))
	for _, name := range model.Platforms {
		spec := PlatformSpec{
			Name: name, Domain: model.Domains[name], SMTPHost: model.SmtpServers[name],
			POP3Host: model.Pop3Servers[name], IMAPHost: model.ImapServers[name],
		}
		result = append(result, spec)
	}
	return result
}

// CheckMail runs local, outbound, dynamic-MX, and fixed endpoint checks.
func CheckMail(ctx context.Context, specs []PlatformSpec, mx MXResolver, dialer Dialer, _ LocalListener) MailReport {
	if ctx == nil {
		ctx = context.Background()
	}
	if mx == nil {
		mx = systemMXResolver{}
	}
	if dialer == nil {
		dialer = systemDialer{net.Dialer{Timeout: 5 * time.Second}}
	}
	report := MailReport{SchemaVersion: "portchecker.mail/v1", GeneratedAt: time.Now().UTC()}
	for _, port := range []int{25, 465, 110, 995, 143, 993} {
		report.Local = append(report.Local, probeLocalListener(ctx, dialer, port))
	}
	var outboundJobs, fixedJobs []endpointJob
	for _, spec := range specs {
		if spec.SMTPHost != "" {
			outboundJobs = append(outboundJobs, endpointJob{result: EndpointResult{Kind: KindOutboundSMTP25, Platform: spec.Name, Host: spec.SMTPHost, Protocol: "smtp", Port: 25}})
		}
		fixed := []protocolEndpoint{
			{protocol: "smtp", port: 25, host: spec.SMTPHost}, {protocol: "smtps", port: 465, tls: true, host: spec.SMTPHost},
			{protocol: "pop3", port: 110, host: spec.POP3Host}, {protocol: "pop3s", port: 995, tls: true, host: spec.POP3Host},
			{protocol: "imap", port: 143, host: spec.IMAPHost}, {protocol: "imaps", port: 993, tls: true, host: spec.IMAPHost},
		}
		for _, endpoint := range fixed {
			if endpoint.host == "" {
				continue
			}
			fixedJobs = append(fixedJobs, endpointJob{result: EndpointResult{Kind: KindFixedProtocol, Platform: spec.Name, Host: endpoint.host, Protocol: endpoint.protocol, Port: endpoint.port, TLS: endpoint.tls}, secure: endpoint.tls})
		}
	}
	report.OutboundSMTP25 = runEndpointJobs(ctx, dialer, outboundJobs, 32)
	report.Fixed = runEndpointJobs(ctx, dialer, fixedJobs, 32)

	// Resolve domains concurrently, retaining platform order and MX preference.
	mxBySpec := make([][]EndpointResult, len(specs))
	var mxWG sync.WaitGroup
	mxSem := make(chan struct{}, 8)
	for i, spec := range specs {
		if spec.Domain == "" {
			continue
		}
		mxWG.Add(1)
		go func(i int, spec PlatformSpec) {
			defer mxWG.Done()
			select {
			case mxSem <- struct{}{}:
			case <-ctx.Done():
				mxBySpec[i] = []EndpointResult{{Kind: KindMXSMTP25, Platform: spec.Name, Domain: spec.Domain, Port: 25, Status: MailTimeout, Error: "timeout"}}
				return
			}
			defer func() { <-mxSem }()
			mxRecords, err := mx.LookupMX(ctx, spec.Domain)
			if err != nil {
				status, message := classifyMailError(err)
				mxBySpec[i] = []EndpointResult{{Kind: KindMXSMTP25, Platform: spec.Name, Domain: spec.Domain, Port: 25, Status: status, Error: message}}
				return
			}
			sort.SliceStable(mxRecords, func(i, j int) bool { return mxRecords[i].Pref < mxRecords[j].Pref })
			for _, record := range mxRecords {
				if record == nil || strings.TrimSpace(record.Host) == "" {
					continue
				}
				mxBySpec[i] = append(mxBySpec[i], EndpointResult{Kind: KindMXSMTP25, Platform: spec.Name, Domain: spec.Domain, Host: strings.TrimSuffix(record.Host, "."), Port: 25, Preference: record.Pref})
			}
		}(i, spec)
	}
	mxWG.Wait()
	// Probe each domain's MX hosts sequentially by preference while allowing
	// independent domains to proceed concurrently.
	mxProbeSem := make(chan struct{}, 8)
	for i := range mxBySpec {
		if len(mxBySpec[i]) == 0 || mxBySpec[i][0].Status != "" {
			continue
		}
		mxWG.Add(1)
		go func(i int) {
			defer mxWG.Done()
			select {
			case mxProbeSem <- struct{}{}:
			case <-ctx.Done():
				for j := range mxBySpec[i] {
					mxBySpec[i][j].Status = MailTimeout
					mxBySpec[i][j].Error = "timeout"
				}
				return
			}
			defer func() { <-mxProbeSem }()
			for j := range mxBySpec[i] {
				mxBySpec[i][j] = probeEndpoint(ctx, dialer, mxBySpec[i][j], false)
			}
		}(i)
	}
	mxWG.Wait()
	for _, results := range mxBySpec {
		report.MX = append(report.MX, results...)
	}
	return report
}

func probeLocalListener(ctx context.Context, dialer Dialer, port int) EndpointResult {
	result := EndpointResult{Kind: KindLocalListen, Port: port}
	started := time.Now()
	defer func() { result.LatencyMS = time.Since(started).Milliseconds() }()
	var lastErr error
	refused := 0
	for _, target := range localListenerTargets() {
		probeCtx, cancel := context.WithTimeout(ctx, time.Second)
		conn, err := dialer.DialContext(probeCtx, target.network, net.JoinHostPort(target.host, fmt.Sprintf("%d", port)))
		cancel()
		if err == nil {
			_ = conn.Close()
			result.Status = MailAvailable
			result.Evidence = "loopback_connect_succeeded"
			return result
		}
		lastErr = err
		status, _ := classifyMailError(err)
		if status == MailRefused {
			refused++
		}
		if ctx.Err() != nil {
			result.Status, result.Error = classifyMailError(ctx.Err())
			return result
		}
	}
	if refused > 0 {
		result.Status = MailUnavailable
		result.Evidence = "no_loopback_listener"
		return result
	}
	result.Status, result.Error = classifyMailError(lastErr)
	return result
}

func localListenerTargets() []struct{ network, host string } {
	targets := []struct{ network, host string }{{network: "tcp4", host: "127.0.0.1"}, {network: "tcp6", host: "::1"}}
	seen := map[string]struct{}{"tcp4|127.0.0.1": {}, "tcp6|::1": {}}
	addresses, _ := net.InterfaceAddrs()
	for _, address := range addresses {
		var ip net.IP
		switch value := address.(type) {
		case *net.IPNet:
			ip = value.IP
		case *net.IPAddr:
			ip = value.IP
		}
		if ip == nil || (!ip.IsGlobalUnicast() && !ip.IsLoopback()) {
			continue
		}
		network := "tcp6"
		if ip.To4() != nil {
			network = "tcp4"
			ip = ip.To4()
		}
		host := ip.String()
		key := network + "|" + host
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		targets = append(targets, struct{ network, host string }{network: network, host: host})
	}
	sort.Slice(targets[2:], func(i, j int) bool {
		left, right := targets[i+2], targets[j+2]
		return left.network+"|"+left.host < right.network+"|"+right.host
	})
	return targets
}

func runEndpointJobs(ctx context.Context, dialer Dialer, jobs []endpointJob, concurrency int) []EndpointResult {
	if concurrency <= 0 {
		concurrency = 1
	}
	results := make([]EndpointResult, len(jobs))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, job := range jobs {
		wg.Add(1)
		go func(i int, job endpointJob) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				job.result.Status = MailTimeout
				job.result.Error = "timeout"
				results[i] = job.result
				return
			}
			defer func() { <-sem }()
			results[i] = probeEndpoint(ctx, dialer, job.result, job.secure)
		}(i, job)
	}
	wg.Wait()
	return results
}

func probeEndpoint(ctx context.Context, dialer Dialer, result EndpointResult, secure bool) EndpointResult {
	started := time.Now()
	if strings.TrimSpace(result.Host) == "" || result.Port <= 0 || result.Port > 65535 {
		result.Status = MailUnsupported
		result.Error = "invalid endpoint"
		return result
	}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(result.Host, fmt.Sprintf("%d", result.Port)))
	if err != nil {
		result.Status, result.Error = classifyMailError(err)
		result.LatencyMS = time.Since(started).Milliseconds()
		return result
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if secure {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: result.Host, MinVersion: tls.VersionTLS12})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			result.Status, result.Error = classifyMailError(err)
			result.LatencyMS = time.Since(started).Milliseconds()
			return result
		}
		conn = tlsConn
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		result.Status, result.Error = classifyMailError(err)
	} else if isMailGreeting(result.Protocol, line) {
		result.Status = MailAvailable
	} else {
		result.Status = MailUnavailable
		result.Error = "unexpected greeting"
	}
	result.LatencyMS = time.Since(started).Milliseconds()
	return result
}

func isMailGreeting(protocol, line string) bool {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "220") || strings.HasPrefix(line, "+OK") || strings.Contains(line, "* OK") {
		return true
	}
	return protocol == "smtp" && strings.HasPrefix(line, "2")
}

func classifyMailError(err error) (MailStatus, string) {
	if err == nil {
		return MailAvailable, ""
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return MailTimeout, "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return MailTimeout, "timeout"
	}
	if strings.Contains(strings.ToLower(err.Error()), "connection refused") {
		return MailRefused, "connection_refused"
	}
	return MailError, "connection_error"
}
