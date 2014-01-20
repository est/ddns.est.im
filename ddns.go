package main

import (
    "fmt"
    "net"
    "os"
    "strings"
)


func showQuery(data[]byte) {
    di := 12 // cursor index, hard coded C0 0C (offset=12)
    dl := data[di] // Cursor data length
    ds := make([]string, 0) // output data string
    for dl>0 {
        di_next := di+int(dl)
        s:= string(data[di+1:di_next+1])
        ds = append(ds, s)
        di = di_next+1
        dl = data[di]
    }
    fs := strings.Join(ds, ".")
    fmt.Println(fs)
}

func main() {
    var buf [512]byte
    var up_buf [512]byte
    svr_addr, _ := net.ResolveUDPAddr("udp", "0:53")
    server, err := net.ListenUDP("udp", svr_addr)
    if err != nil {
        fmt.Println("Can not start server.", err)
        os.Exit(1)
    }
    for{
            
        c, addr, _ := server.ReadFrom(buf[:512])
        data_req := buf[:c]
        showQuery(data_req)
        go func(){
            up_server, err := net.Dial("udp", "8.8.8.8:53")
            if err != nil {
                fmt.Println("Connect to upstream failed", err)
                // @ToDo: how to handle this? use stale cache?
            }

            up_server.Write(data_req)
            c, _ = up_server.Read(up_buf[:512])
            data_rsp := up_buf[:c]
            server.WriteTo(data_rsp, addr)
            showQuery(data_req)
        }()
    }
}

