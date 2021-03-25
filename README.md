# dyndns-aws

This tool updates the specified AWS Route53 record with the external IP address
of the host running it, similarly to a dyndns updater.

The external IP address is obtained using the WhatIsMyIP.com API.

The Route53 credentials are read from `~/.aws/credentials` as documented at
https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html .

## Usage

```
$ ./dyndns-aws -h
Usage of ./dyndns-aws:
  -4	Force IPv4
  -6	Force IPv6
  -domain string
    	Domain name to update DNS record for
  -dryrun
    	Do not actually update the DNS record
  -host string
    	Host name to update DNS record for
  -iface string
    	If specified, use the primary IP address of this network interface, even if not routable
  -ip string
    	Force a custom IP address instead of a NIC's IP address
```

For example:
```
$ dyndns-aws -host www -domain slackware.it
```

## Installation

```
go get github.com/insomniacslk/dyndns-aws
```

