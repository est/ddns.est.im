package main

import (
    "fmt"
    "net"
    "os"
    "strings"
    "encoding/binary"
)

// http://technet.microsoft.com/en-us/library/dd197470(v=ws.10).aspx
// https://github.com/tonnerre/golang-dns/blob/master/types.go
var RRTYPE = map[uint16]string{
    0x01: "A",
    0x02: "NS",
    0x05: "CNAME",
    0x0C: "PTR",
    0x0F: "MX",
    0x10: "TXT",
    0x1C: "AAAA",
    0x33: "SRV",
    0xFF: "ANY",
}
var Uint16 = binary.BigEndian.Uint16


func showQuery(data[]byte) {
    di := 12 // cursor index, hard coded C0 0C (offset=12)
    dl := data[di] // Cursor data length
    ds := make([]string, 127) // data string array for output. Max 127 subdomains because 255 bytes total
    li := 0 // loop index
    for dl>0 {
        diNext := di+int(dl)+1 // 1 is the dl itself
        s:= string(data[di+1:diNext])
        ds[li] = s
        li++
        di = diNext
        dl = data[di]
    }
    di += 1 // skip \0 termin for domain names
    dt := RRTYPE[Uint16(data[di:di+2])]
    fs := strings.Join(ds[:li], ".") // full string
    fmt.Print("Q: ", fs, " IN ", dt, " ")
    // @ToDO: check all DNS has exactly 1 query?

    // answer count
    ac := Uint16(data[6:8]) + Uint16(data[8:10]) + Uint16(data[10:12])
    if ac > 0 {
        di += 4 // skip question TYPE, CLASS
        // as := make([]string, 31) // 255 bytes hard coded again!
        // qRef := []byte {0xC0, 0x0C}
        // http://www.ccs.neu.edu/home/amislove/teaching/cs4700/fall09/handouts/project1-primer.pdf
        fmt.Println("RR:", ac, " ")
        for ac > 0 {
            if Uint16(data[di:di+2]) & 0xC000 == 0xC000 {
                dt = RRTYPE[Uint16(data[di+2:di+4])]
                dttl := binary.BigEndian.Uint32(data[di+6:di+10])
                di += 10
                adl := int(Uint16(data[di:di+2])) // answer data length
                di += 2
                ad := data[di:di+adl] // answer  data
                di += adl
                fmt.Println("  ", dt, "\t", dttl, "\t", ad)
            } else {
                fmt.Println("  Can't parse RDATA yet: ", data[di:di+2])
                break
            }
            ac--
        }
    }
    fmt.Println(" ")
}

func main() {
    var buf [512]byte
    var up_buf [512]byte
    localServerAddr, _ := net.ResolveUDPAddr("udp", "0:53")
    localServer, err := net.ListenUDP("udp", localServerAddr)
    if err != nil {
        fmt.Println("Can not start server.", err)
        os.Exit(1)
    }
    for{
            
        c, addr, _ := localServer.ReadFrom(buf[:512])
        dataReq := buf[:c]
        // @ToDo: check flag is 0X0100
        showQuery(dataReq)
        go func(){
            upConn, err := net.Dial("udp", "8.8.8.8:53")
            if err != nil {
                fmt.Println("Connect to upstream failed", err)
                // @ToDo: how to handle this? use stale cache?
            }

            upConn.Write(dataReq)
            c, _ = upConn.Read(up_buf[:512])
            dataRsp := up_buf[:c]
            localServer.WriteTo(dataRsp, addr)
            // @ToDo: check response flag is 0x8180. 0x8183 = NXDOMAIN
            showQuery(dataRsp)
        }()
    }
}

