package main


import (
    "log"
    _ "fmt"
    "io/ioutil"
    "net"
    "encoding/binary"
    "encoding/json"
    "github.com/mrjones/oauth"
)

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
            // copy(dataRsp[:c], dataReq[:c])
            // NXDOMAIN
            binary.BigEndian.PutUint16(dataRsp[2:4], 0x8183)
            server.WriteTo(dataRsp, addr)
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
        "https://api.twitter.com/1.1/statuses/home_timeline.json",
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
