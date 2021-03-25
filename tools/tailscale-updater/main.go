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

// toDNSName modifies a string to make it suitable for DNS names (e.g. convert
// spaces into dashes).
func toDNSName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}

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

	// DANGER ZONE: the host names are modified to be valid DNS labels (e.g.
	// spaces to dashes) so if two hosts generate the same label, the latter
	// will override the former!
	// TODO detect this, and append a counter, e.g. -1, -2 etc.

	// local config first
	if status.Self == nil {
		log.Fatalf("Self config is nil")
	}
	fmt.Printf("-4 -host '%s' -domain '%s' -ip '%s'\n", toDNSName(status.Self.HostName), domain, status.Self.TailAddr)
	// then the peers
	for _, peer := range status.Peer {
		fmt.Printf("-4 -host '%s' -domain '%s' -ip '%s'\n", toDNSName(peer.HostName), domain, peer.TailAddr)
	}
}
