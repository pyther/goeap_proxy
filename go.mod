module github.com/pyther/goeap_proxy

go 1.13

require (
	github.com/google/gopacket v1.1.17 // indirect
	github.com/mdlayher/raw v0.0.0-20191009151244-50f2db8cc065 // indirect
)

// https://github.com/mdlayher/raw/pull/64
replace github.com/mdlayher/raw => github.com/pyther/raw v0.0.0-20200508193324-eb26248ef18b

// https://github.com/google/gopacket/pull/781
replace github.com/google/gopacket => github.com/pyther/gopacket v1.1.18-0.20200502044149-9afa69325031
