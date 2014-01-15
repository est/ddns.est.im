package main

import (
    "fmt"
    "net"
    // "bufio"
)

func main() {
    var buf []byte
    svr_addr, _ := net.ResolveUDPAddr("udp", ":5353")
    server, _ := net.ListenUDP("udp", svr_addr)
    c, addr, err := server.ReadFrom(buf)
    fmt.Println("done", c, addr, err)
}