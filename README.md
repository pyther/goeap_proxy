# goeap_proxy

Proxy EAP packet between network interfaces. Proxy written in golang.

Inspired by [`eap_proxy`](https://github.com/jaysoffian/eap_proxy/), which in turn was inspired by [`1x_prox`](http://www.dslreports.com/forum/r30693618-) posted to the “[AT&T Residential Gateway Bypass - True bridge mode!](https://www.dslreports.com/forum/r29903721-AT-T-Residential-Gateway-Bypass-True-bridge-mode)” discussion in the “AT&T U-verse” DSLReports forum.


## Goals
1. Proxy EAP Packets, that's it.
2. Use a library for packet processing [gopacket/layers](https://github.com/google/gopacket/tree/master/layers)
3. Use a library for raw socket handling away [mdlayher/raw](https://github.com/mdlayher/raw/)
4. Keep It Simple and Easy to Understand

## Usage
```
root@OpenWrt:~# goeap_proxy --help
Usage of goeap_proxy:
  -if-router string
    	interface of the AT&T ONT/WAN
  -if-wan string
    	interface of the AT&T Router
  -ignore-logoff
        ignore EAP-Logoff packets
  -promiscuous
    	place interfaces into promiscuous mode instead of multicast
  -syslog
    	log to syslog
```

### Ignore Logoff
It has been reported that some gateways such as the Pace 5268ac will send a EAPOLLogOff causing a sporadic outage. Use the `-ignore-logoff` flag if you encounter this issue. This is not needed for the BGW210. 

### Example Run
```
root@OpenWrt:~# goeap_proxy -if-router eth3 -if-wan eth2
2020/05/02 19:45:18 eth3: 88:71:b1:a1:b1:c1 > 01:80:c2:00:00:03, EAPOLLogOff v2, len 0 > eth2
2020/05/02 19:45:18 eth3: 88:71:b1:a1:b1:c1 > 01:80:c2:00:00:03, EAPOLStart v2, len 0 > eth2
2020/05/02 19:45:18 eth2: 00:90:d0:63:ff:01 > 01:80:c2:00:00:03, EAP v1, len 4, Failure (4) id 115 > eth3
2020/05/02 19:45:18 eth2: 00:90:d0:63:ff:01 > 01:80:c2:00:00:03, EAP v1, len 15, Request (1) id 116 > eth3
2020/05/02 19:45:18 eth2: 00:90:d0:63:ff:01 > 88:71:b1:a1:c1:b1, EAP v1, len 15, Request (1) id 116 > eth3
...
```

## Build
**Standard Build**
```
$ go build -o goeap_proxy main.go
```

**Against Musl**
```
$ CC=/usr/local/bin/musl-gcc go build -o goeap_proxy main.go
```
Note: This is useful for testing changes on router distributions that use musl without needing to go through the packaging process.

**For OpenWRT**
Openwrt feed: [pyther/openwrt-feed](https://github.com/pyther/openwrt-feed)
Package building instructions in README

## Other Projects
- [jaysoffian/eap_proxy](https://github.com/jaysoffian/eap_proxy): python implementation with a primary focus on EdgeOS
- [nsubtil/eaproxy](https://github.com/nsubtil/eaproxy): C++ client
- [mjonuschat/eap_parrot](https://github.com/mjonuschat/eap_parrot): go client similar to goeap_proxy
