package main

import (
    "fmt"
    "net"
    "os"
    "strings"
    "time"
    "encoding/binary"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

// http://technet.microsoft.com/en-us/library/dd197470(v=ws.10).aspx
// https://github.com/tonnerre/golang-dns/blob/master/types.go
var RRTYPE = map[uint16]string{
    0x01: "A",
    0x02: "NS",
    0x05: "CNAME",
    0x06: "SOA",
    0x0C: "PTR",
    0x0F: "MX",
    0x10: "TXT",
    0x1C: "AAAA",
    0x33: "SRV",
    0xFF: "ANY",
}
var Uint16BE = binary.BigEndian.Uint16
var Uint16LE = binary.LittleEndian.Uint16

var db *sql.DB // global var for db connection

func showQuery(data []byte) {
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
    dt := Uint16BE(data[di:di+2])
    dts, ok := RRTYPE[dt]
    if !ok {
        dts = fmt.Sprintf("(%d)", dt)
    }
    ds = ds[:li]
    fs := strings.Join(ds, ".") // full string
    fmt.Print("Q: ", fs, " ", dts, " ")
    // @ToDO: check all DNS has exactly 1 query?
    
    // answer count
    ac := Uint16BE(data[6:8]) // + Uint16BE(data[8:10]) + Uint16BE(data[10:12])
    if ac > 0 {
        di += 4 // skip question TYPE, CLASS
        // as := make([]string, 31) // 255 bytes hard coded again!
        // qRef := []byte {0xC0, 0x0C}
        // http://www.ccs.neu.edu/home/amislove/teaching/cs4700/fall09/handouts/project1-primer.pdf
        fmt.Println("RR:", ac, " ")
        for ac > 0 {
            if Uint16BE(data[di:di+2]) & 0xC000 == 0xC000 {
                dt := Uint16BE(data[di+2:di+4])
                dts, ok := RRTYPE[dt]
                if !ok {
                    dts = fmt.Sprintf("(%d)", dt)
                }
                dttl := binary.BigEndian.Uint32(data[di+6:di+10])
                di += 10
                adl := int(Uint16BE(data[di:di+2])) // answer data length
                di += 2
                ad := data[di:di+adl] // answer  data
                di += adl
                fmt.Println("  ", dts, "\t", dttl, "\t", ad)
                go updateRecord(ds, dttl, dt, ad)
            } else {
                fmt.Println("  Can't parse RDATA yet: ", data[di:di+2])
                break
            }
            ac--
        }
    }
    fmt.Println(" ")
}


func updateRecord(name []string, ttl uint32, type_id uint16, value []byte) {
    // name_ra is reversed in DB for faster prefix lookup
    l := len(name) 
    name_ra := make([]string, l)
    for i,j := 0,l-1; j>=0; i,j=i+1,j-1 {
        name_ra[i] = name[j]
    }
    name_r := strings.Join(name_ra, " ")
    // fmt.Println("wtf", strings.Join(name_ra, " "), time.Now().Unix()+int64(ttl), type_id, value )
    sql := "INSERT OR IGNORE INTO record (name_r, ts_expires, type_id, value) VALUES (?,?,?,?);"
    expire := time.Now().Unix()+int64(ttl)
    res, _ := db.Exec(sql, name_r, expire, type_id, value)
    ra, _ := res.RowsAffected()
    if ra == 0 {
        sql := "UPDATE record SET ts_expires=?, ts_accessed=>? where name_r=? AND value=?"
        db.Exec(sql, expire, time.Now(), name_r, value)
    }
}

func main() {

    // setup sqlite cache.db
    // no modify time. Append only
    var err error
    db, err = sql.Open("sqlite3", "./cache.db")
    if err != nil {
        fmt.Println(err)
    }
    defer db.Close()

    sql := `
    CREATE TABLE IF NOT EXISTS record (
      id INTEGER not null primary key, 
      name_r TEXT,
      ts_added TEXT DEFAULT CURRENT_TIMESTAMP,
      ts_expires INTEGER,
      ts_accessed TEXT DEFAULT CURRENT_TIMESTAMP,
      type_id INTEGER,
      value BLOB
    ); `
    _, err = db.Exec(sql)
    if err != nil {
        fmt.Println(err, sql)
        return
    }
    sql = `CREATE UNIQUE INDEX IF NOT EXISTS name_value ON record (name_r COLLATE NOCASE, value COLLATE NOCASE);`
    _, err = db.Exec(sql)
    if err != nil {
        fmt.Println(err, sql)
        return
    }



    fmt.Println("Start server...")
    // start server, loop forever
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

