module github.com/pyther/goeap_proxy

go 1.13

require (
	github.com/google/gopacket v1.1.19
	github.com/mdlayher/raw v0.0.0-20210412142147-51b895745faf
	golang.org/x/net v0.0.0-20210903162142-ad29c8ab022f // indirect
	golang.org/x/sys v0.0.0-20210903071746-97244b99971b // indirect
)

// https://github.com/mdlayher/raw/pull/64
replace github.com/mdlayher/raw => github.com/pyther/raw v0.0.0-20210906162253-71a651591f09

// https://github.com/google/gopacket/pull/781
replace github.com/google/gopacket => github.com/pyther/gopacket v1.1.20-0.20210906161201-08fbbeab9a82
