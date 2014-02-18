package main


import (
    "log"
    "fmt"
    "net"
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
            dataRsp :=  // @ToDo: NXDOMAIN
            server.Writeln(dataRsp, addr)
        }
    }
}