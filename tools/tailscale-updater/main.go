package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"tailscale.com/ipn/ipnstate"
)

var (
	flagTailscaledUnixSocket = flag.String("u", "/run/tailscale/tailscaled.sock", "Path to tailscaled's UNIX socket")
	flagTailscaleHTTPStatus  = flag.String("s", "http://localhost/localapi/v0/status", "Tailscale's LocalAPI status URL")
)

func getStatus(unixSocket, statusURL string) (*ipnstate.Status, error) {
	// Using Tailscale's LocalAPI to get status information. This is obviously
	// better than shelling out to `tailscale status`. See
	// https://github.com/tailscale/tailscale/blob/main/ipn/localapi/localapi.go
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", *flagTailscaledUnixSocket)
			},
		},
	}
	resp, err := client.Get(*flagTailscaleHTTPStatus)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET to %s failed: %w", statusURL, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body: %w", err)
	}
	var status ipnstate.Status
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Tailscale LocalAPI Status response: %w", err)
	}
	return &status, nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <domain name>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	domain := flag.Arg(0)
	if domain == "" {
		log.Fatalf("no domain specified")
	}
	status, err := getStatus(*flagTailscaledUnixSocket, *flagTailscaleHTTPStatus)
	if err != nil {
		log.Fatalf("Failed to get Tailscale status: %v", err)
	}

	// local config first
	if status.Self == nil {
		log.Fatalf("Self config is nil")
	}
	suffix := "." + status.MagicDNSSuffix
	if !strings.HasSuffix(suffix, ".") {
		suffix += "."
	}
	fmt.Printf("-4 -host '%s' -domain '%s' -ip '%s'\n", strings.TrimSuffix(status.Self.DNSName, suffix), domain, status.Self.TailAddr)
	// then the peers
	for _, peer := range status.Peer {
		if peer.DNSName == "" {
			log.Printf("Warning: skipping peer '%s' with empty DNS name", peer.HostName)
			continue
		}
		fmt.Printf("-4 -host '%s' -domain '%s' -ip '%s'\n", strings.TrimSuffix(peer.DNSName, suffix), domain, peer.TailAddr)
	}
}
