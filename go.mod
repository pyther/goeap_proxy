module github.com/pyther/goeap_proxy

go 1.13

require (
	github.com/google/gopacket v1.1.19
	github.com/mdlayher/raw v0.0.0-20191009151244-50f2db8cc065
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
)

// https://github.com/mdlayher/raw/pull/64
replace github.com/mdlayher/raw => github.com/pyther/raw v0.0.0-20200508193324-eb26248ef18b

// https://github.com/google/gopacket/pull/781
replace github.com/google/gopacket => github.com/pyther/gopacket v1.1.20-0.20210124173545-be2359faeaf9
