package main

// This tool updates the specified AWS Route53 record with the external IP of
// the device running it.
// The external IP address is obtained using the WhatIsMyIP.com API.
//
// Credentials are read from ~/.aws/credentials as documented in
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

var (
	flagHost   = flag.String("host", "", "Host name to update DNS record for")
	flagDomain = flag.String("domain", "", "Domain name to update DNS record for")
)

func updateAddress(ip net.IP, name, domain string) error {
	log.Printf("Updating DNS record %s.%s with IP address %s", name, domain, ip)
	// normalize domain: lower-case, one trailing dot. This is how zones
	// are returned by the Route53 API.
	// No check for empty zone.
	domain = strings.ToLower(strings.TrimRight(domain, ".")) + "."

	s := session.Must(session.NewSession())
	r := route53.New(s)
	out, err := r.ListHostedZones(&route53.ListHostedZonesInput{})
	if err != nil {
		log.Fatal(err)
	}
	found := false
	var zoneID string
	for idx, hz := range out.HostedZones {
		if hz.Name == nil {
			log.Printf("Warning: got nil zone name at index %d", idx)
			continue
		}
		if strings.ToLower(*hz.Name) == domain {
			found = true
			zoneID = *hz.Id
			break
		}
	}
	if !found {
		return fmt.Errorf("Domain %s not found on this account", domain)
	}
	var Type string
	if to4 := ip.To4(); to4 != nil {
		Type = "A"
	} else if to16 := ip.To16(); to16 != nil {
		Type = "AAAA"
	} else {
		return fmt.Errorf("invalid IP address '%s'", ip)
	}
	ttl := int64(3600)
	result, err := r.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(fmt.Sprintf("%s.%s", name, domain)),
						Type: aws.String(Type),
						TTL:  &ttl,
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ip.String()),
							},
						},
					},
				},
			},
		},
		HostedZoneId: &zoneID,
	})
	if err != nil {
		return fmt.Errorf("failed to update record set %s.%s: %v", name, domain, err)
	}
	log.Printf("DEBUG: %+v", result)
	return nil
}

func getExternalAddress() (net.IP, error) {
	// you can also use ipv4bot and ipv6bot for v4-only and v6-only
	req, err := http.Get("https://bot.whatismyipaddress.com")
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	addr, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(string(addr))
	if ip == nil {
		return nil, fmt.Errorf("Failed to parse ip '%s'", addr)
	}
	return ip, nil
}

func main() {
	flag.Parse()
	if *flagHost == "" {
		log.Fatalf("Host name flag not specified")
	}
	if *flagDomain == "" {
		log.Fatalf("Domain name flag not specified")
	}

	addr, err := getExternalAddress()
	if err != nil {
		log.Fatalf("Failed to get external IP address: %v", err)
	}

	if err := updateAddress(addr, *flagHost, *flagDomain); err != nil {
		log.Fatalf("Failed to update DNS record %s.%s with IP address %s: %v", *flagHost, *flagDomain, addr, err)
	}
}
