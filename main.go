package main

import (
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/mdlayher/raw"
	"net"
	"os"
	"time"
)

var Version string
var BuildStamp string
var EAP_MULTICAST_ADDR string = "01:80:c2:00:00:03"

type eapInterface struct {
	name string
	conn *raw.Conn
}

func newInterface(name string, promiscuous bool) *eapInterface {
	x := eapInterface{name: name}

	intf, err := net.InterfaceByName(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "InterfaceByName(%q) failed: %v\n", name, err)
		os.Exit(1)
	}

	conn, err := raw.ListenPacket(intf, uint16(layers.EthernetTypeEAPOL), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ListenPacket(%q) failed: %v\n", name, err)
		os.Exit(1)
	}

	// Listen to Multicast Address or put interfaces in promiscuous mode
	if promiscuous {
		conn.SetPromiscuous(true)
	} else {
		eapAddr, _ := net.ParseMAC(EAP_MULTICAST_ADDR)
		eapMulticastAddr := &raw.Addr{HardwareAddr: eapAddr}
		conn.SetMulticast(eapMulticastAddr)
	}

	x.conn = conn
	return &x
}

func main() {

	var rtrInt string
	var wanInt string
	var promiscuous bool
	var ignoreLogoff bool
	var version bool

	flag.StringVar(&rtrInt, "if-router", "", "interface of the AT&T Router")
	flag.StringVar(&wanInt, "if-wan", "", "interface of the AT&T ONT/WAN")
	flag.BoolVar(&ignoreLogoff, "ignore-logoff", false, "ignore EAPOL-Logoff packets")
	flag.BoolVar(&promiscuous, "promiscuous", false, "place interfaces into promiscuous mode instead of multicast")
	flag.BoolVar(&version, "version", false, "display version")
	flag.Parse()

	if version {
		fmt.Println("Version: ", Version)
		fmt.Println("Build Time: ", BuildStamp)
		os.Exit(0)
	}

	if rtrInt == "" || wanInt == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	// Allow only single instance of goeap_proxy
	// We could potentially tie the lock file to the wan and rtr interfaces
	// But lets keep things simple for now
	l, err := net.Listen("unix", "@/run/goeap_proxy.lock")
	if err != nil {
		fmt.Fprintln(os.Stderr, "goeap_proxy is already running!")
		os.Exit(1)
	}
	defer l.Close()

	wan := newInterface(wanInt, promiscuous)
	rtr := newInterface(rtrInt, promiscuous)

	// Wait until both subroutines exit
	fmt.Printf("proxy started. router: %s, wan: %s\n", rtrInt, wanInt)
	quit := make(chan int)
	go proxyPackets(rtr, wan, ignoreLogoff)
	go proxyPackets(wan, rtr, ignoreLogoff)
	<-quit
}

func proxyPackets(src *eapInterface, dst *eapInterface, ignoreLogoff bool) {
	// This might break for jumbo frames
	recvBuf := make([]byte, 1500)
	for {
		size, _, err := src.conn.ReadFrom(recvBuf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: unexpected read error: %v\n", src.name, err)
			// maybe not necessary, give the system a minute to recover
			time.Sleep(500 * time.Millisecond)
			continue
		}
		packetData := recvBuf[:size]

		var eth layers.Ethernet
		var eapol layers.EAPOL
		var eap layers.EAP
		parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &eapol, &eap)
		decoded := []gopacket.LayerType{}

		// Raw socket only listes for EAPoL Packet
		// This error should be sufficient error handling
		if err := parser.DecodeLayers(packetData, &decoded); err != nil {
			fmt.Fprintf(os.Stderr, "%s: could not decode layers: %v\n", src.name, err)
			continue
		}

		if ignoreLogoff && eapol.Type == layers.EAPOLTypeLogOff {
			fmt.Printf("%s: ignoring %s\n", src.name, eapol.Type)
			continue
		}

		//DEBUG: Print Decoded Layers
		//fmt.Fprintf(os.Stderr, "Decoded: %v\n", decoded)

		// print a log message with useful information
		printPacketInfo(src.name, dst.name, eth, eapol, eap)

		_, err = dst.conn.WriteTo(packetData, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: unexpected write error: %v\n", dst.name, err)
		}
	}

}

func printPacketInfo(src string, dst string, eth layers.Ethernet, eapol layers.EAPOL, eap layers.EAP) {
	line := fmt.Sprintf("%s: ", src)
	line += fmt.Sprintf("%s > %s, %s v%d, len %d", eth.SrcMAC, eth.DstMAC, eapol.Type, eapol.Version, eapol.Length)

	if eap.Code != 0 {
		codeString := EAPCodeToString(eap.Code)
		line += fmt.Sprintf(", %s (%d) id %d", codeString, eap.Code, eap.Id)
	}

	line += fmt.Sprintf(" > %s", dst)
	fmt.Println(line)
}

func EAPCodeToString(code layers.EAPCode) string {
	switch code {
	case layers.EAPCodeRequest:
		return "Request"
	case layers.EAPCodeResponse:
		return "Response"
	case layers.EAPCodeSuccess:
		return "Success"
	case layers.EAPCodeFailure:
		return "Failure"
	}
	return "Unknown"
}
