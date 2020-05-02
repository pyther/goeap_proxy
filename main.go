package main

import (
    "fmt"
    "flag"
    "github.com/mdlayher/raw"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "log"
    "log/syslog"
    "net"
    "time"
    "os"
)

var EAP_MULTICAST_ADDR string = "01:80:c2:00:00:03"

func main() {

    var rtrInt string
    var wanInt string
    var syslog_enable bool
    var promiscuous bool

    flag.StringVar(&rtrInt, "if-router", "", "interface of the AT&T ONT/WAN")
    flag.StringVar(&wanInt, "if-wan", "", "interface of the AT&T Router")
    flag.BoolVar(&syslog_enable, "syslog", false, "log to syslog")
    flag.BoolVar(&promiscuous, "promiscuous", false, "place interfaces into promiscuous mode instead of multicast")
    flag.Parse()

    if rtrInt == "" || wanInt == "" {
        flag.PrintDefaults()
        os.Exit(1)
    }
    flag.Parse()

    if syslog_enable {
        logwriter, _ := syslog.New(syslog.LOG_INFO, "goeap_proxy")
        log.SetOutput(logwriter)
        log.SetFlags(0) //removes timestamps
    }


    // Allow only single instance of goeap_proxy
    // We could potentially tie the lock file to the wan and rtr interfaces
    // But lets keep things simple for now
    l, err := net.Listen("unix", "@/run/goeap_proxy.lock")
        if err != nil {
        log.Fatal("goeap_proxy is already running!")
    }
    defer l.Close()

    proxyEap(rtrInt, wanInt, promiscuous)
}

func proxyEap(rtrInt string, wanInt string, promiscuous bool) {
    // get interface objects
    wanIf, err := net.InterfaceByName(wanInt)
    if err != nil {
        log.Fatalf("interface by name %s: %v", wanInt, err)
    }

    rtrIf, err := net.InterfaceByName(rtrInt)
    if err != nil {
        log.Fatalf("interface by name %s: %v", rtrInt, err)
    }


    // Listen on Interfaces
    wanConn, err := raw.ListenPacket(wanIf, uint16(layers.EthernetTypeEAPOL), nil)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    defer wanConn.Close()

    rtrConn, err := raw.ListenPacket(rtrIf, uint16(layers.EthernetTypeEAPOL), nil)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    defer rtrConn.Close()

    // Listen to Multicast Address or put interfaces in promiscuous mode
    if promiscuous {
        wanConn.SetPromiscuous(true)
        rtrConn.SetPromiscuous(true)
    } else {
        eapAddr, _ := net.ParseMAC(EAP_MULTICAST_ADDR)
        eapMulticastAddr := &raw.Addr{HardwareAddr: eapAddr}
        wanConn.SetMulticast(eapMulticastAddr)
        rtrConn.SetMulticast(eapMulticastAddr)
    }

    // Wait until both subroutines exit
    quit := make(chan int)
    go proxyPackets(rtrInt, rtrConn, wanInt, wanConn)
    go proxyPackets(wanInt, wanConn, rtrInt, rtrConn)
    <-quit
}

func proxyPackets(srcName string, srcConn *raw.Conn, dstName string, dstConn *raw.Conn) {
    // This might break for jumbo frames
    recvBuf := make([]byte, 1500)
    for {
        size, _, err := srcConn.ReadFrom(recvBuf)
        if err != nil {
            log.Printf("unexpected read error: %v\n", err)
            // maybe not necessary, give the system a minute to recover
            time.Sleep(500 * time.Millisecond)
            continue
		}

        // returns Nil if not an Ethernet AND EAPOL packet
        packet := parsePacket(recvBuf[:size])
        if packet == nil {
            continue
        }

        // print a log message with useful information
        printPacketInfo(srcName, dstName, packet)

        // Get the Source Addr of the Ethernet Frame
        ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
        ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
        srcAddr := &raw.Addr{HardwareAddr: ethernetPacket.SrcMAC}

        // Transmit the Packet to the destination interface
        _, err = dstConn.WriteTo(packet.Data(), srcAddr)
    }

}

func parsePacket(data []byte) gopacket.Packet {
    packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)
    eapolLayer := packet.Layer(layers.LayerTypeEAPOL)

    if eapolLayer == nil {
        log.Println("Not an EAPOL Packet")
        return nil
    }
    return packet
}


func printPacketInfo(src string, dst string, packet gopacket.Packet) {
    ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
    eapLayer := packet.Layer(layers.LayerTypeEAP)
    eapolLayer := packet.Layer(layers.LayerTypeEAPOL)

    // We've verified that we have valid packets in parsePacket
    ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
    eapol, _ := eapolLayer.(*layers.EAPOL)

    line := fmt.Sprintf("%s: ", src)
    line += fmt.Sprintf("%s > %s, %s v%d, len %d", ethernetPacket.SrcMAC, ethernetPacket.DstMAC, eapol.Type, eapol.Version, eapol.Length)

    if eapLayer != nil {
        eap, _ := eapLayer.(*layers.EAP)
        codeString := EAPTypeString(eap.Code)
        line += fmt.Sprintf(", %s (%d) id %d", codeString, eap.Code, eap.Id)
    }

    line += fmt.Sprintf(" > %s", dst)
    log.Println(line)
}

func EAPTypeString(code layers.EAPCode) string {
    switch code {
    case 1:
        return "Request"
    case 2:
        return "Response"
    case 3:
        return "Success"
    case 4:
        return "Failure"
    }
    return "Unknown"
}

