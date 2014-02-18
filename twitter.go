package main


import (
    "log"
    // "fmt"
    "net"
    "encoding/binary"
)


func main() {
    var bufLocal [512]byte
    serverAddr, _ := net.ResolveUDPAddr("udp", "0:53")
    server, _ := net.ListenUDP("udp", serverAddr)
    for {
        c, addr, _ := server.ReadFrom(bufLocal[:512])
        dataReq := bufLocal[:c]
        log.Println("Got Query", addr, dataReq[0:1])
        go func(){
            dataRsp := bufLocal[:c]
            // copy(dataRsp[:c], dataReq[:c])
            // NXDOMAIN
            binary.BigEndian.PutUint16(dataRsp[2:4], 0x8183)
            server.WriteTo(dataRsp, addr)
        }()
    }
}