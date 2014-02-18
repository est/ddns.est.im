package main

import (
    // "bytes"
    "log"
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


// zone file RR format 
// *.example.com.   3600 IN  MX 10 host1.example.com.


var db *sql.DB // global var for db connection
var tidMap map[uint16] chan []byte

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
    readerIndex += 4 // to the end
    return readerIndex, nameArray[:nameParts], nameTypeId
}

func showQuery(data []byte) {
    // reader index, domain name array, domain name type
    di, ds, dt := parseQuery(data)
    tid := Uint16BE(data[:2])

    dts, ok := RRTYPE[dt]
    if !ok {
        dts = fmt.Sprintf("(%d)", dt)
    }
    fs := strings.Join(ds, ".") // full string
    fmt.Print("Q: tid=", tid, ". ",  fs, " ", dts, " ")
    // @ToDO: check all DNS has exactly 1 query?
    
    // answer count
    ac := Uint16BE(data[6:8]) // + Uint16BE(data[8:10]) + Uint16BE(data[10:12])
    if ac > 0 {
         // di += 4 don't need to skip question TYPE, CLASS, because fixed.
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

    // upsert hack 
    // for fields (id, name_r, ttl, type_id, value, time_added, time_accessed)
    // http://stackoverflow.com/a/7511635/41948
    sql := `
    INSERT OR REPLACE INTO record 
    SELECT 
        old.id, 
        COALESCE(new.name_r, old.name_r), 
        MAX(COALESCE(old.ttl, 0), new.ttl), 
        COALESCE(new.type_id, old.type_id), 
        COALESCE(new.value, old.value), 
        COALESCE(old.time_added, new.time_added), 
        new.time_accessed
    FROM ( SELECT 
        NULL AS id, 
        ? AS name_r, 
        ? AS ttl, 
        ? AS type_id, 
        ? AS value, 
        datetime('now') AS time_added, 
        strftime('%s','now') AS time_accessed 
    ) AS new
    LEFT JOIN (
        SELECT id, name_r, ttl, type_id, value, time_added, time_accessed 
        FROM record WHERE 
            name_r = ? AND 
            type_id = ? AND 
            value = ? 
    ) AS old ON 
        new.name_r = old.name_r AND 
        new.type_id = old.type_id AND 
        new.value = old.value;
    `
    res, err := db.Exec(sql, nameInv, ttl, typeId, value, nameInv, typeId, value)

    if err != nil {
            log.Fatal("INSERT FAIL ", err)
    }

    ra, err := res.RowsAffected()
    log.Println("SQL", ra)

    // rowId, err := res.LastInsertId()
    // now := time.Now().Unix()
}


func doQuery(tid uint16, data []byte){
    var bufUp [512]byte

    upConn, err := net.Dial("udp", "8.8.8.8:53")
    if err != nil {
        fmt.Println("Connect to upstream failed", err)
        // @ToDo: how to handle this? use stale cache?
    }
    upConn.Write(data)
    c, _ := upConn.Read(bufUp[:512])
    tidMap[tid] <- bufUp[:c]
}

func getRecord(data []byte) []byte {
    // make response
    rsp := make([]byte, 512)
    copy(rsp, data)
    // transaction id
    tid := Uint16BE(data[0:2])

    // query db
    rwIndex, nameArray, nameTypeId := parseQuery(data)
    l := len(nameArray)
    nameInvArray := make([]string, l)
    for i, v := range nameArray {
        nameInvArray[l-i-1] = v
    }
    sql := `
    SELECT 
        ttl, time_accessed, value 
    FROM record 
    WHERE 
        name_r=? AND type_id=? AND ttl>0
    ORDER BY ttl DESC
    `
    rows, err := db.Query(sql, strings.Join(nameInvArray, " "), nameTypeId)
    if err != nil {
            panic(err)
    }
    rrCount := 0
    for rows.Next() {
            var ttl uint32
            var tsAccessed uint64
            var value []byte
            if err := rows.Scan(&ttl, &tsAccessed, &value); err != nil {
                    fmt.Println(err)
            }
            rrCount++
            // construct Resource Record
            rr := make([]byte, 12)
            binary.BigEndian.PutUint16(rr[0:2], 0xC00C) // RR compression
            binary.BigEndian.PutUint16(rr[2:4], nameTypeId)// 2 bytes for type ID
            binary.BigEndian.PutUint16(rr[4:6], 0x0001) // class = IN
            binary.BigEndian.PutUint32(rr[6:10], ttl) // 4 bytes for TTL
            binary.BigEndian.PutUint16(rr[10:12], uint16(len(value))) // 2 bytes for value length
            copy(rsp[rwIndex:], rr)
            rwIndex += len(rr)
            copy(rsp[rwIndex:], value)
            rwIndex += len(value)
            fmt.Println("SQL: ", nameArray, ttl, tsAccessed, value)
    }  
    if err := rows.Err(); err != nil {
            panic(err)
    }
    rows.Close()
    if rrCount > 0 {
        binary.BigEndian.PutUint16(rsp[2:4], 0x8180)
        binary.BigEndian.PutUint16(rsp[6:8], uint16(rrCount))
    } else {
        // @ToDO: better key, like src port.
        delete(tidMap, tid) 
        tidMap[tid] = make(chan []byte, 2)
        go doQuery(tid, data)
        select {
        case rsp = <- tidMap[tid]:
            rwIndex = len(rsp)
            log.Println("Relay packet for ", tid)
        case <-time.After(4 * time.Second):
            log.Println("Relay packet timeout ", tid, " Use NXDOMAIN")
            binary.BigEndian.PutUint16(rsp[2:4], 0x8183)
        }
    }
    return rsp[0:rwIndex]
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


    // setup tid map
    tidMap = make(map[uint16] chan []byte)
    // tidMap[0] = make(chan []byte)



    fmt.Println("Start server...")
    // start server, loop forever
    var bufLocal [512]byte
    localServerAddr, _ := net.ResolveUDPAddr("udp", "0:53")
    localServer, err := net.ListenUDP("udp", localServerAddr)
    if err != nil {
        fmt.Println("Can not start server.", err)
        os.Exit(1)
    }
    for{
        c, addr, _ := localServer.ReadFrom(bufLocal[:512])
        dataReq := bufLocal[:c]
        log.Println("Got query")
        showQuery(dataReq)
        go func(addr net.Addr){
            dataRsp := getRecord(dataReq)
            // @ToDo: handl NXDOMAIN or SERVER FAIL.
            localServer.WriteTo(dataRsp, addr)
            // @ToDo: timeout based on transaction ID instead of sequencial.
            showQuery(dataRsp)
        }(addr)
        // _ = addr
    }
}

