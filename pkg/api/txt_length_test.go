package api

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/miekg/dns"

	"github.com/joohoi/acme-dns/pkg/acmedns"
	"github.com/joohoi/acme-dns/pkg/nameserver"
)

// TestALabelEqualsHexValueIsAcceptedAndServed drives the whole path a client
// uses: the value is posted to /update with the registration's credentials,
// and is then read back over DNS as a TXT record at the registration's full
// domain. The value has the form label=hex, here 39 characters, which
// upstream refuses twice: for its length and for the '='.
func TestALabelEqualsHexValueIsAcceptedAndServed(t *testing.T) {
	// Built at runtime: a label, '=', and 32 hexadecimal digits.
	value := "cidtls=" + strings.ReplaceAll(uuid.New().String(), "-", "")
	if len(value) != 39 {
		t.Fatalf("test value has %d characters, want 39", len(value))
	}

	router, _, db := setupRouter(false, false)
	server := httptest.NewServer(router)
	defer server.Close()

	reg, err := db.Register(acmedns.Cidrslice{})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	getExpect(t, server).POST("/update").
		WithJSON(map[string]interface{}{"subdomain": reg.Subdomain, "txt": value}).
		WithHeader("X-Api-User", reg.Username.String()).
		WithHeader("X-Api-Key", reg.Password).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		NotContainsKey("error").
		ValueEqual("txt", value)

	// Serve DNS from the same database.
	addr := freeUDPAddr(t)
	config, logger := fakeConfigAndLogger()
	config.General.Domain = "auth.example.test"
	config.General.Nsname = "ns1.auth.example.test"
	config.General.Nsadmin = "admin.example.test"
	config.General.Listen = addr
	ns := nameserver.NewDNSServer(&config, db, logger, "udp")
	ns.ParseRecords()
	started := make(chan struct{})
	ns.SetNotifyStartedFunc(func() { close(started) })
	errChan := make(chan error, 1)
	go ns.Start(errChan)
	select {
	case <-started:
	case err := <-errChan:
		t.Fatalf("DNS server did not start: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("DNS server did not start within 5s")
	}
	if n, ok := ns.(*nameserver.Nameserver); ok {
		defer func() { _ = n.Server.Shutdown() }()
	}

	q := new(dns.Msg)
	q.SetQuestion(dns.Fqdn(reg.Subdomain+".auth.example.test"), dns.TypeTXT)
	in, err := dns.Exchange(q, addr)
	if err != nil {
		t.Fatalf("DNS query: %v", err)
	}
	if in.Rcode != dns.RcodeSuccess {
		t.Fatalf("DNS query: rcode %s", dns.RcodeToString[in.Rcode])
	}
	for _, rr := range in.Answer {
		if txt, ok := rr.(*dns.TXT); ok && len(txt.Txt) == 1 && txt.Txt[0] == value {
			return
		}
	}
	t.Fatalf("the value %q is not served at %s; answer: %v", value, q.Question[0].Name, in.Answer)
}

// TestAValueOutsideTheAlphabetIsStillRefused shows the relaxed check kept a
// character test: a 40-character value containing '+', which is outside the
// URL-safe base64 alphabet plus '=', is refused at /update.
func TestAValueOutsideTheAlphabetIsStillRefused(t *testing.T) {
	value := strings.Repeat("a", 39) + "+"

	router, _, db := setupRouter(false, false)
	server := httptest.NewServer(router)
	defer server.Close()

	reg, err := db.Register(acmedns.Cidrslice{})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	getExpect(t, server).POST("/update").
		WithJSON(map[string]interface{}{"subdomain": reg.Subdomain, "txt": value}).
		WithHeader("X-Api-User", reg.Username.String()).
		WithHeader("X-Api-Key", reg.Password).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().
		ValueEqual("error", "bad_txt")
}

func freeUDPAddr(t *testing.T) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("finding a free UDP port: %v", err)
	}
	addr := pc.LocalAddr().String()
	_ = pc.Close()
	return addr
}
