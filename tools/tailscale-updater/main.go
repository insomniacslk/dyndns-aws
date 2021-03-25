package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"tailscale.com/ipn/ipnstate"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <domain name>", os.Args[0])
	}
	flag.Parse()
	domain := flag.Arg(0)
	if domain == "" {
		log.Fatalf("no domain specified")
	}

	// HORRORS AHEAD. I am shelling out to `tailscale` because I am confused by
	// the tailscale.com API. It seems that tailscale.com/client/tailscale is
	// out of sync from github.com/tailscale/tailscale/client/tailscale . The
	// latter contains a Status method, the former doesn't. So for now shellout.
	out, err := exec.Command("tailscale", "status", "-json").Output()
	if err != nil {
		log.Fatalf("Failed to run tailscale executable: %v", err)
	}
	var status ipnstate.Status
	if err := json.Unmarshal(out, &status); err != nil {
		log.Fatalf("Failed to parse tailscale status output: %v", err)
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
