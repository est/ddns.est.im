#!/usr/bin/env python
# coding: utf-8
# author: frk
# ref: https://github.com/17mon/python/blob/master/ipip.py

import struct
import socket
import mmap


class IPData(object):
    def __init__(self, file_name):
        self.file = open(file_name, 'rb')
        self.mm = mmap.mmap(self.file.fileno(), 0, prot=mmap.PROT_READ)
        self.total_length, = struct.unpack(">L", self.mm[:4])
        self.content = self.mm[4:self.total_length]

    def find(self, ip):
        nip = socket.inet_aton(ip)  # this assumes correctness
        ip_dot = ip.split('.')
        p = int(ip_dot[0]) * 4
        start, = struct.unpack('<L', self.content[p:p + 4])

        index_offset, index_length = 0, 0
        max_comp_len = self.total_length - 1028
        start = start * 8 + 1024
        while start < max_comp_len:
            if self.content[start:start + 4] >= nip:
                index_offset, = struct.unpack(
                    "<L",
                    self.content[start + 4:start + 7] + '\0')
                index_length, = struct.unpack('B', self.content[start + 7])
                break
            start += 8

        if index_offset == 0:
            return "N/A"

        res_offset = self.total_length + index_offset - 1024
        return self.mm[res_offset:res_offset + index_length].decode('utf8')


def test():
    import timeit
    print(timeit.timeit(
        "f.find('118.28.8.8')",
        setup="f=__import__('ipip').IPData('17monipdb.dat')",
        number=1000))
    print(timeit.timeit(
        "f.find('118.28.8.8')",
        setup="f=__import__('ipip').IP; f.load('17monipdb.dat')",
        number=1000))


if '__main__' == __name__:
    test()
