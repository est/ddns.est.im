package main


import (
    "log"
    "fmt"
    "io/ioutil"
    "net"
    "encoding/binary"
    "encoding/json"
    "github.com/mrjones/oauth"
)

func main1() {
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
    bits, err := ioutil.ReadAll(response.Body)
    return string(bits)
}

func main() {
    byt := []byte(`[{"created_at":"Fri Feb 21 09:05:23 +0000 2014","id":436788732155277312,"id_str":"436788732155277312","text":"\u6211\u6ef4\u90a3\u4e2a\u5988\u5440\uff01RT @bbcchinese: \u7f05\u7538\u514b\u94a6\u90a6\u53d1\u73b0\u5de8\u578b\u7fe1\u7fe0\u539f\u77f3: \u7f05\u7538\u514b\u94a6\u90a6\u8c03\u6d3e\u519b\u4eba\u524d\u5f80\u7fe1\u7fe0\u4ea7\u5730\u5e15\u6562\uff0c\u4fdd\u62a4\u5f53\u5730\u65b0\u53d1\u73b0\u7684\u4e00\u5757\u53ef\u80fd\u91cd50\u5428\u7684\u5de8\u578b\u7fe1\u7fe0\u539f\u77f3\u3002 http:\/\/t.co\/G8FPBi2ZlO","source":"\u003ca href=\"http:\/\/itunes.apple.com\/us\/app\/twitter\/id409789998?mt=12\" rel=\"nofollow\"\u003eTwitter for Mac\u003c\/a\u003e","truncated":false,"in_reply_to_status_id":436787297577480192,"in_reply_to_status_id_str":"436787297577480192","in_reply_to_user_id":791197,"in_reply_to_user_id_str":"791197","in_reply_to_screen_name":"bbcchinese","user":{"id":800076,"id_str":"800076","name":"Andre Pan","screen_name":"POPOEVER","location":"Shanghai, China","description":"somewhere over the rainbow...","url":"http:\/\/t.co\/3GkNrfZHy3","entities":{"url":{"urls":[{"url":"http:\/\/t.co\/3GkNrfZHy3","expanded_url":"http:\/\/about.me\/popoever","display_url":"about.me\/popoever","indices":[0,22]}]},"description":{"urls":[]}},"protected":false,"followers_count":10282,"friends_count":874,"listed_count":202,"created_at":"Wed Feb 28 03:48:29 +0000 2007","favourites_count":666,"utc_offset":28800,"time_zone":"Beijing","geo_enabled":true,"verified":false,"statuses_count":38325,"lang":"en","contributors_enabled":false,"is_translator":false,"is_translation_enabled":false,"profile_background_color":"AAC031","profile_background_image_url":"http:\/\/pbs.twimg.com\/profile_background_images\/66392\/twitter_bg.png","profile_background_image_url_https":"https:\/\/pbs.twimg.com\/profile_background_images\/66392\/twitter_bg.png","profile_background_tile":true,"profile_image_url":"http:\/\/pbs.twimg.com\/profile_images\/535947311\/at_haagen_dazs_normal.jpg","profile_image_url_https":"https:\/\/pbs.twimg.com\/profile_images\/535947311\/at_haagen_dazs_normal.jpg","profile_banner_url":"https:\/\/pbs.twimg.com\/profile_banners\/800076\/1379775970","profile_link_color":"CC3300","profile_sidebar_border_color":"212123","profile_sidebar_fill_color":"3B3E43","profile_text_color":"787878","profile_use_background_image":true,"default_profile":false,"default_profile_image":false,"following":true,"follow_request_sent":false,"notifications":true},"geo":null,"coordinates":null,"place":null,"contributors":null,"retweet_count":0,"favorite_count":0,"entities":{"hashtags":[],"symbols":[],"urls":[{"url":"http:\/\/t.co\/G8FPBi2ZlO","expanded_url":"http:\/\/bbc.in\/1jjhfFF","display_url":"bbc.in\/1jjhfFF","indices":[81,103]}],"user_mentions":[{"screen_name":"bbcchinese","name":"BBCChinese.com","id":791197,"id_str":"791197","indices":[10,21]}]},"favorited":false,"retweeted":false,"possibly_sensitive":false,"lang":"zh"}]`)
    var dat []map[string]interface{}
    if err := json.Unmarshal(byt, &dat); err != nil {
        panic(err)
    }
    fmt.Println(dat[0]["text"])
}