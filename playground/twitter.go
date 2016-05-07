package main


import (
    "log"
    "fmt"
    "errors"
    "io/ioutil"
    _"bytes"
    "strings"
    "syscall"
    "reflect"
    "net"
    "encoding/binary"
    "encoding/json"
    "github.com/mrjones/oauth"
    "github.com/qiniu/iconv"
)

type DNSQuery struct {
    name []byte
    type_ uint16
    class uint16
}

type DNSAnswer struct {
    name []byte
    type_ uint16
    class uint16
    ttl uint32
    length uint16
    data []byte
}

type DNSPacket struct {
    xid uint16
    flags uint16
    num_questions uint16
    num_answers uint16
    num_authorities uint16
    num_additionals uint16
    question DNSQuery
    answer DNSAnswer
}


// copy paste from http://code.google.com/p/go/source/browse/ipv4/helper_unix.go?repo=net#38
var errInvalidConnType = errors.New("invalid conn type")
func sysfd(c net.Conn) (int, error) {
        cv := reflect.ValueOf(c)
        switch ce := cv.Elem(); ce.Kind() {
        case reflect.Struct:
                netfd := ce.FieldByName("conn").FieldByName("fd")
                switch fe := netfd.Elem(); fe.Kind() {
                case reflect.Struct:
                        fd := fe.FieldByName("sysfd")
                        return int(fd.Int()), nil
                }
        }
        return 0, errInvalidConnType
}


func parseName(input []byte) string {
    var ret []string
    offset := uint16(12)
    l :=  input[offset]
    for l > 0 {
        // fmt.Println("length", l)
        if l > 0xC0 {
            offset = binary.BigEndian.Uint16(input[offset:offset+2]) & uint16(0x3FFF)
            l = input[offset]
        } else {
            offset += uint16(1)
            s := string(input[offset:offset+uint16(l)])
            // fmt.Println("name part", s, l, offset)
            ret = append(ret, s)
            offset = offset + uint16(l)
            l = input[offset]
        }
    }
    return strings.Join(ret, ".")
}

func formatName(input []byte) []byte {
    return nil
}


func setAnswer(dataRsp []byte, data []byte, type_ uint16) []byte{
    binary.BigEndian.PutUint16(dataRsp[2:4], 0x8180)
    binary.BigEndian.PutUint16(dataRsp[6:8], 1)

    dataLength := len(data)

    dataAns := []byte{
        0xC0, 0x0C, // original name
        0x00, 0x00, // empty type. TXT=0x10, CNAME=0x05
        0x00, 0x01, // class
        0x00, 0x00, 0x0e, 0x10, // ttl == 3600
        0x00, 0x00, // length, placeholder
    }
    binary.BigEndian.PutUint16(dataAns[2:4], type_)
    binary.BigEndian.PutUint16(dataAns[10:12], uint16(dataLength+2))
    dataAns = append(dataAns, byte(dataLength))
    dataRsp = append(dataRsp, dataAns...)
    dataRsp = append(dataRsp, data...)
    dataRsp = append(dataRsp, 0x00)
    return  dataRsp
}

func getTwitter(userName string) []byte {
    c := oauth.NewConsumer(
        "ARtsaTvMriWxoq5tu2Zw", // consumerKey,
        "5LsCecr6Jb2G59Tl1qR9xinFAkB4MsnHKK5j0sGnu0", // consumerSecret,
        oauth.ServiceProvider{
            RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
            AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
            AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
    })
    ad := make(map[string]string)
    at :=  oauth.AccessToken{
        "5082501-dGoMoSaYmPspoyQrd9pgWzqVqF0ROZz6D0Fl5uecOw", // access token
        "VevWKlH42x8pwAhiH5GfEoN9V17ZY7i0IcFXEAaECRJVy",   // access secret
        ad,
    }

    response, err := c.Get(
        "https://api.twitter.com/1.1/statuses/user_timeline.json", // API URL 
        map[string]string{
            "count": "1",
            "screen_name": userName,
        },
        &at)
    defer response.Body.Close()
    if err != nil {
        log.Println(err)
        return make([]byte, 0)
    }

    byt, err := ioutil.ReadAll(response.Body)

    var dat []map[string]interface{}
    if err := json.Unmarshal(byt, &dat); err != nil {
        log.Println("json1", err)
        return make([]byte, 0)
    }
    s, ok := dat[0]["text"].(string)
    if ok == false{
        log.Println("json2", err)
        return make([]byte, 0)
    }
    // b, err := json.Marshal(s)
    log.Println("Tweet: ", s)
    return []byte(s)
}



func main() {
    bufLocal := make([]byte, 512)
    serverAddr, _ := net.ResolveUDPAddr("udp", "0:53")
    server, _ := net.ListenUDP("udp", serverAddr)
    for {

        fd, err := sysfd(server)
        if err != nil {
            log.Panic()
        }

        // this only gets the server TTL
        // ttl, _ := syscall.GetsockoptInt(int(f.Fd()), syscall.IPPROTO_IP, syscall.IP_TTL)

        err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_RECVTTL, 1)
        if err != nil {
            log.Panic()
        }

        // see msg_control/cmsghdr from `man 2  recvmsg`
        oob := make([]byte, 32) // actually 16
        n, oobn, _, addr, _ := server.ReadMsgUDP(bufLocal, oob)
        dataReq := bufLocal[:n]
        ttl := binary.LittleEndian.Uint32(oob[12:16])
        log.Printf(
            "Q: %v xid=%v fd=%v ttl=%v\n", 
            addr, binary.BigEndian.Uint16(dataReq[0:2]), fd, ttl)
        fmt.Println(oob[:oobn])
        go func(){
            dataRsp := bufLocal[:n]
            if false {
                binary.BigEndian.PutUint16(dataRsp[2:4], 0x8183)
                server.WriteTo(dataRsp, addr)
            } else {
                twitterName := parseName(dataReq)
                fmt.Println(twitterName)
                tweetData := getTwitter(twitterName) // could be `printf` in bash
                if ttl > 64 {
                    cd, err := iconv.Open("gbk", "utf-8")
                    if err != nil {
                        fmt.Println("iconv.Open failed!")
                        return
                    }
                    defer cd.Close()
                    var buf [512]byte
                    tweetData, _, _ = cd.Conv(tweetData, buf[:])

                }
                dataRsp = setAnswer(dataRsp, tweetData, 0x05)
                server.WriteTo(dataRsp, addr)
            }
        }()
    }
}