https://gist.github.com/gleicon/1074227
http://www.isnowfy.com/introduction-to-gevent/


GOPATH=$PWD go get github.com/mrjones/oauth
GOPATH=$PWD go get github.com/miekg/dns


 
tweet() { printf "$(eval echo $(dig +short @127.0.0.1 NAME_HERE | sed 's/\\\([0-9]*\)/\\\\`printf x%x \1`/g'))\n"; }


https://www.opensource.apple.com/source/bind9/bind9-45.1/bind9/lib/dns/rdata.c
static isc_result_t txt_totext()

https://www.opensource.apple.com/source/bind9/bind9-45.1/bind9/bin/dig/nslookup.c
static void printrdata()

GOPATH=$PWD go build -ldflags '-s' twitter.go

see msg_control/cmsghdr from `man 2  recvmsg`

replacement: http://golang.org/src/pkg/net/fd_unix.go?h=Recvmsg

http://godoc.org/code.google.com/p/go.net/ipv4