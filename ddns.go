package main

import (
    "fmt"
    "net"
    // "bufio"
)

func main() {
    var buf [512]byte
    svr_addr, _ := net.ResolveUDPAddr("udp", "0:15353")
    // up_addr, _ := net.ResolveUDPAddr("udp", "8.8.8.8:53")
    server, err := net.ListenUDP("udp", svr_addr)
    fmt.Println("server ok == ", err==nil)
    c, addr, err := server.ReadFrom(buf[:512])
    data := buf[:c]
    fmt.Println("done", data, c, addr, buf, len(buf), cap(buf), err)
}