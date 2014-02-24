package main


import (
    "log"
    "fmt"
    "io/ioutil"
    _"bytes"
    "strings"
    "net"
    "encoding/binary"
    "encoding/json"
    "github.com/mrjones/oauth"
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

func main() {
    var bufLocal [512]byte
    serverAddr, _ := net.ResolveUDPAddr("udp", "0:53")
    server, _ := net.ListenUDP("udp", serverAddr)
    for {
        c, addr, _ := server.ReadFrom(bufLocal[:512])
        dataReq := bufLocal[:c]
        log.Println("Got Query", addr, "tid=", binary.BigEndian.Uint16(dataReq[0:2]))
        go func(){
            dataRsp := bufLocal[:c]
            if false {
                binary.BigEndian.PutUint16(dataRsp[2:4], 0x8183)
                server.WriteTo(dataRsp, addr)
            } else {
                binary.BigEndian.PutUint16(dataRsp[2:4], 0x8180)
                binary.BigEndian.PutUint16(dataRsp[6:8], 1)

                twitterName := parseName(dataReq)
                fmt.Println(twitterName)

                dataAns := []byte{
                    0xC0, 0x0C, // original name
                    0x00, 0x10, // TXT
                    0x00, 0x01, // class
                    0x00, 0x00, 0x0e, 0x10, // ttl == 3600
                    0x00, 0x02, // length
                    0x01,
                    'h',
                }
                server.WriteTo(append(dataRsp, dataAns...), addr)
            }
        }()
    }
}

func getTwitter(userName string) string {
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
        "https://api.twitter.com/1.1/statuses/home_timeline.json", // API URL 
        map[string]string{"count": "1"},
        &at)
    defer response.Body.Close()
    if err != nil {
        log.Fatal(err)
    }

    byt, err := ioutil.ReadAll(response.Body)

    var dat []map[string]interface{}
    if err := json.Unmarshal(byt, &dat); err != nil {
        panic(err)
    }
    s, ok := dat[0]["text"].(string)
    if ok == false{
        panic(s)
    }
    return(s)
}
