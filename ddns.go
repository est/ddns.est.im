package main

import (
    "fmt"
    "net"
    // "bufio"
)

func main() {
    var buf [512]byte
    var up_buf [512]byte
    svr_addr, _ := net.ResolveUDPAddr("udp", "0:53")
    server, err1 := net.ListenUDP("udp", svr_addr)
    // up_addr, _ := net.ResolveUDPAddr("udp", )
    for{
            up_server, err2 := net.Dial("udp", "8.8.8.8:53")
            fmt.Println("server ok == ", err1, err2)
            
            c, addr, _ := server.ReadFrom(buf[:512])
            fmt.Println("L Received", buf[:c])

            up_server.Write(buf[:c])
            c, _ = up_server.Read(up_buf[:512])
            server.WriteTo(up_buf[:c], addr)
            fmt.Println("R Received", up_buf[:c])
    }
}