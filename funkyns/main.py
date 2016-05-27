#!/usr/bin/env python
# coding: utf8

from __future__ import absolute_import, division, print_function, \
    with_statement

import os
import struct
import socket


class TempoDNS(object):
    sock = None

    def run(self):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.sock.bind(('0.0.0.0', 53))
        while True:
            data, addr = yield self.sock.recvfrom(4096)
            self.sock.sendto(data, addr)


class DNSUtil(object):
    """
    rfc1035
    record
                                   1  1  1  1  1  1
     0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                                               |
    /                                               /
    /                      NAME                     /
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                      TYPE                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     CLASS                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                      TTL                      |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                   RDLENGTH                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--|
    /                     RDATA                     /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    """
    QTYPE_ANY = 255
    QTYPE_A = 1
    QTYPE_AAAA = 28
    QTYPE_CNAME = 5
    QTYPE_NS = 2
    QCLASS_IN = 1

    QTYPE_NAMES = {
        1: 'A',
        2: 'NS',
        5: 'CNAME',
        6: 'SOA',
        12: 'PTR',
        15: 'MX',
        16: 'TXT',
        28: 'AAAA',
        33: 'SRV',
        44: 'SSHFP',
        255: 'ANY',  # for request only
        256: 'URI',
    }

    @staticmethod
    def build_address(address):
        # ipv4 only. convert ip to ns record format
        results = []
        for label in address.strip(b'.').split(b'.'):
            l = len(label)
            if l > 63:
                return None
            results.append(chr(l))
            results.append(label)
        results.append(b'\0')
        return b''.join(results)

    @classmethod
    def parse_name(cls, data, offset):
        p = offset
        labels = []
        l = ord(data[p])
        while l > 0:
            if (l & (128 + 64)) == (128 + 64):
                # pointer
                pointer = struct.unpack('!H', data[p:p + 2])[0]
                pointer &= 0x3FFF
                r = cls.parse_name(data, pointer)
                labels.append(r[1])
                p += 2
                # pointer is the end
                return p - offset, b'.'.join(labels)
            else:
                labels.append(data[p + 1:p + 1 + l])
                p += 1 + l
            l = ord(data[p])
        return p - offset + 1, b'.'.join(labels)

    @classmethod
    def parse_ip(cls, addrtype, data, length, offset):
        if addrtype == cls.QTYPE_A:
            return socket.inet_ntop(socket.AF_INET, data[offset:offset + length])
        elif addrtype == cls.QTYPE_AAAA:
            return socket.inet_ntop(socket.AF_INET6, data[offset:offset + length])
        elif addrtype in [cls.QTYPE_CNAME, cls.QTYPE_NS]:
            return cls.parse_name(data, offset)[1]
        else:
            return data[offset:offset + length]

    @classmethod
    def parse_record(cls, data, offset, question=False):
        nlen, name = cls.parse_name(data, offset)
        if not question:
            record_type, record_class, record_ttl, record_rdlength = struct.unpack(
                '!HHiH', data[offset + nlen:offset + nlen + 10]
            )
            ip = cls.parse_ip(record_type, data, record_rdlength, offset + nlen + 10)
            return nlen + 10 + record_rdlength, \
                (name, ip, cls.QTYPE_NAMES[record_type], record_ttl)
        else:
            record_type, record_class = struct.unpack(
                '!HH', data[offset + nlen:offset + nlen + 4]
            )
            return nlen + 4, (name, None, cls.QTYPE_NAMES[record_type], None)


class DNSRequest(object):
    def __init__(self, name, qtype=DNSUtil.QTYPE_A):
        self.req_id = os.urandom(2)
        # flag: Standard Query
        self.flag = 0x0100
        # 1 question, 0 answer RR, 0 authority RR, 0 additional RR
        self.req_nums = (1, 0, 0, 0)

        self.name = name
        self.qtype = qtype

    def bytes(self):
        return '%s%s%s' % (
            struct.pack('!2sH4H', self.req_id, self.flag, *self.req_nums),
            DNSUtil.build_address(self.name),
            struct.pack('!HH', self.qtype, DNSUtil.QCLASS_IN))

    def __repr__(self):
        return '<DNSRequest: %s>' % self.name


class DNSResponse(object):
    def __init__(self, data):
        self.data = data

    def parse(self):
        offset = 12  # lets hope
        if len(self.data) < offset:
            raise ValueError('Bad Response: %r' % self.data)
        fmt = '!2sH4H'
        r = struct.unpack(fmt, self.data[:offset])
        self.req_id, self.flag, self.res_nums = r[0], r[1], r[2:]
        l = self.data[offset]

        questions, answers = [], []
        for i in range(0, self.res_nums[0]):
            l, r = DNSUtil.parse_record(self.data, offset, True)
            offset += l
            if r:
                questions.append(r)
        for i in range(0, self.res_nums[1]):
            l, r = DNSUtil.parse_record(self.data, offset)
            offset += l
            if r:
                answers.append(r)
        for i in range(0, self.res_nums[2]):
            l, r = DNSUtil.parse_record(self.data, offset)
            offset += l
        for i in range(0, self.res_nums[3]):
            l, r = DNSUtil.parse_record(self.data, offset)
            offset += l

        return answers


def test():
    rsp_data = """
    c12d81800001000300000000066d656971696103
    636f6d0000010001c00c000500010000001e0020
    1037343334613237363137373637653937036364
    6e086a69617368756c65c013c028000100010000
    001e0004db9949d2c028000100010000001e0004
    db9949f5""".replace('\n', '').replace(' ', '').decode('hex')
    print(DNSResponse(rsp_data).parse())


if '__main__' == __name__:
    test()













