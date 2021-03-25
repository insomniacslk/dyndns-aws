# tailscale-updater

This tool prints the command line arguments to pass to `dyndns-aws` in order to
update the IP addresses of all your tailscale devices into Route53.

Example:
```
$ ./tailscale-updater | xargs -L1 -r dyndns-aws
```
