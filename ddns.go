package main

import (
    "log"
    "fmt"
    "net"
    "os"
    "strings"
    _ "time"
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


// zone file RR format 
// *.example.com.   3600 IN  MX 10 host1.example.com.


var db *sql.DB // global var for db connection

func parseQuery(data []byte)  (int, []string, uint16) {
    readerIndex := 12 // hard coded offset=12
    readerLength := data[readerIndex] // reader data length
    nameArray := make([]string, 127) // array of domain name for output. Max 127 subdomains because 255 bytes total
    nameParts := 0 // number of subdomains
    for readerLength>0 {
        readerIndexNext := readerIndex+int(readerLength)+1 // 1 is the readerLength itself
        s:= string(data[readerIndex+1:readerIndexNext])
        nameArray[nameParts] = s
        nameParts++
        readerIndex = readerIndexNext
        readerLength = data[readerIndex]
    }
    readerIndex += 1 // skip \0 termin for domain names
    nameTypeId := Uint16BE(data[readerIndex:readerIndex+2])
    return readerIndex, nameArray[:nameParts], nameTypeId
}

func showQuery(data []byte) {
    // reader index, domain name array, domain name type
    di, ds, dt := parseQuery(data)

    dts, ok := RRTYPE[dt]
    if !ok {
        dts = fmt.Sprintf("(%d)", dt)
    }
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


func updateRecord(name []string, ttl uint32, typeId uint16, value []byte) {
    // name_ra is reversed in DB for faster prefix lookup
    l := len(name) 
    nameInvArray := make([]string, l)
    for i, v := range name {
        nameInvArray[l-i-1] = v
    }
    nameInv := strings.Join(nameInvArray, " ")


    sql := `
    INSERT OR REPLACE INTO record (id, name_r, ttl, type_id, value, time_accessed)
    SELECT 
        old.id, 
        COALESCE(new.name_r, old.name_r), 
        MAX(COALESCE(old.ttl, 0), new.ttl), 
        COALESCE(new.type_id, old.type_id), 
        COALESCE(new.value, old.value), 
        new.time_accessed 
    FROM ( SELECT
        ? AS name_r, 
        ? AS ttl, 
        ? AS type_id, 
        ? AS value, 
        datetime('now') as time_accessed
    ) AS new
    LEFT JOIN (
        SELECT id, name_r, ttl, type_id, value
        FROM record
    ) AS old ON 
        new.name_r = old.name_r AND 
        new.type_id = old.type_id AND 
        new.value = old.value;
    `
    res, err := db.Exec(sql, nameInv, ttl, typeId, value)

    if err != nil {
            log.Fatal("INSERT FAIL ", err)
    }

    ra, err := res.RowsAffected()
    log.Println("SQL", ra)

    // if err != nil {
    //         log.Fatal(err)
    // }

    // rowId, err := res.LastInsertId()

    // if err != nil {
    //         log.Fatal(err)
    // }
    // now := time.Now().Unix()
    // if rowId > 0 {
    //     sql := "UPDATE record SET time_accessed=>? where id=?"
    //     db.Exec(sql, now, rowId)
    // }
    // if ra == 0 {
    //     sql := "UPDATE record SET ttl=max(ttl, ?), time_accessed=? where name_r=? AND type_id=? AND value=?"
    //     _, err = db.Exec(sql, ttl, now, nameInv, typeId, value)
    //     fmt.Println("ha", err)
    // }
}


func getRecord(data []byte) []byte {
    readerIndex, nameArray, nameType := parseQuery(data)
    l := len(nameArray)
    nameInvArray := make([]string, l)
    for i, v := range nameArray {
        nameInvArray[l-i-1] = v
    }
    sql := "SELECT ttl, value FROM record WHERE name_r=? and type_id=?"
    rows, err := db.Query(sql, strings.Join(nameInvArray, " "), nameType)
    if err != nil {
            log.Fatal(err)
    }
    for rows.Next() {
            var ttl uint16
            var value []byte
            if err := rows.Scan(&ttl, &value); err != nil {
                    log.Fatal(err)
            }
            fmt.Println("SQL: ", readerIndex, nameArray, ttl, value)
    }
    if err := rows.Err(); err != nil {
            log.Fatal(err)
    }
    rows.Close()
    return data
}

func main() {

    // setup sqlite cache.db
    // no modify time. Append only
    var err error
    db, err = sql.Open("sqlite3", "file:cache.db?cache=shared&mode=rwc")
    if err != nil {
        fmt.Println(err)
    }
    defer db.Close()

    sqls := []string{`
    CREATE TABLE IF NOT EXISTS record (
      id INTEGER not null primary key, 
      name_r TEXT,
      ttl INTEGER,
      type_id INTEGER,
      value BLOB, 
      time_added TEXT DEFAULT CURRENT_TIMESTAMP,
      time_accessed INTEGER
    ); `, `
    CREATE UNIQUE INDEX IF NOT EXISTS name_value ON record (name_r COLLATE NOCASE, type_id, value COLLATE BINARY);
    `}
    for _, sql := range sqls {
        _, err = db.Exec(sql)
        if err != nil {
            fmt.Println(err, sql)
            return
        }
    }



    fmt.Println("Start server...")
    // start server, loop forever
    var bufLocal [512]byte
    var bufUp [512]byte
    localServerAddr, _ := net.ResolveUDPAddr("udp", "0:53")
    localServer, err := net.ListenUDP("udp", localServerAddr)
    if err != nil {
        fmt.Println("Can not start server.", err)
        os.Exit(1)
    }
    for{
            
        c, addr, _ := localServer.ReadFrom(bufLocal[:512])
        dataReq := bufLocal[:c]
        // @ToDo: check flag is 0X0100
        showQuery(dataReq)
        getRecord(dataReq)
        go func(){
            upConn, err := net.Dial("udp", "8.8.8.8:53")
            if err != nil {
                fmt.Println("Connect to upstream failed", err)
                // @ToDo: how to handle this? use stale cache?
            }
            upConn.Write(dataReq)
            c, _ = upConn.Read(bufUp[:512])
            dataRsp := bufUp[:c]
            localServer.WriteTo(dataRsp, addr)
            // @ToDo: check response flag is 0x8180. 0x8183 = NXDOMAIN
            showQuery(dataRsp)
        }()
    }
}

