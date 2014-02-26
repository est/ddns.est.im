import socket


s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect(('g.cn', 80))
# s.send('GET / HTTP/1.1\r\n\r\n')
# print s.recv(10)

print s.fileno()

r = socket.fromfd(s.fileno(), socket.AF_INET, socket.SOCK_RAW, socket.IPPROTO_TCP)
r.send('wtf\r\n\r\n')
print r.recv(512)
print r.getsockopt(socket.IPPROTO_IP, socket.IP_TTL)
print s.getsockopt(socket.IPPROTO_IP, socket.IP_TTL)



# https://groups.google.com/forum/#!topic/golang-nuts/--8LqUiUfJ0
# http://manpages.ubuntu.com/manpages/trusty/en/man4/ip.4.html
# http://code.google.com/p/go/source/browse/ipv4/sockopt_unix.go?repo=net
# http://stackoverflow.com/a/9237137/41948