package main


import (
    "log"
    "fmt"
    "net"
    "encoding/binary"
)


func main(){
    var bufLocal [512]byte
    serverAddr, _ := net.ResolveUDPAddr("udp", 0:53)
    server, _ = net.ListenUDP("udp", serverAddr)
    for {
        c, addr := server.ReadFrom(bufLocal[:512])
        dataReq := bufLocal[:c]
        log.Println("Got Query", addr)
        go func(){
            var dataRsp [c]byte
            copy(dataRsp, dataReq)
            // NXDOMAIN
            binar.BigEndian.PutUint16(dataRsp[2:4], 0x8183)
            server.Writeln(dataRsp, addr)
        }
    }
}