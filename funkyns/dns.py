#!/usr/bin/env python
# coding: utf8

from __future__ import absolute_import, division, print_function, \
    with_statement

import os
import struct
import socket
import logging


class DNSUtil(object):
    """
    rfc1035
    record
     - NAME
     - TYPE
     - CLASS
     - TTL
     - RDLENGTH
     - RDATA
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

    @classmethod
    def parse(cls, data):
        return cls('')

    def bytes(self):
        return '%s%s%s' % (
            struct.pack('!2sH4H', self.req_id, self.flag, *self.req_nums),
            DNSUtil.build_address(self.name),
            struct.pack('!HH', self.qtype, DNSUtil.QCLASS_IN))

    def __repr__(self):
        return '<DNSRequest: %s>' % self.name


class DNSResponse(object):
    def __init__(self):
        pass

    def parse(self, data):
        offset = 12  # lets hope
        if len(data) < offset:
            raise ValueError('Bad Response: %r' % data)
        fmt = '!2sH4H'
        r = struct.unpack(fmt, data[:offset])
        self.req_id, self.flag, self.res_nums = r[0], r[1], r[2:]
        l = data[offset]

        questions, answers = [], []
        for i in range(0, self.res_nums[0]):
            l, r = DNSUtil.parse_record(data, offset, True)
            offset += l
            if r:
                questions.append(r)
        for i in range(0, self.res_nums[1]):
            l, r = DNSUtil.parse_record(data, offset)
            offset += l
            if r:
                answers.append(r)
        for i in range(0, self.res_nums[2]):
            l, r = DNSUtil.parse_record(data, offset)
            offset += l
        for i in range(0, self.res_nums[3]):
            l, r = DNSUtil.parse_record(data, offset)
            offset += l
        return answers

    @classmethod
    def make_rr(cls, name_or_offset, rdata, qtype=DNSUtil.QTYPE_A, ttl=60):
        if isinstance(name_or_offset, int):
            # 0xC00C, leading 11 means compression, points to offset 12
            name = int('11000000', 2) << 8 | int(name_or_offset)
            fmt = '!HHHIH%ss' % len(rdata)
        elif isinstance(name_or_offset, str):
            name = DNSUtil.build_address(name_or_offset)
            fmt = '!%ssHHIH%ss' % (len(name), len(rdata))
        else:
            return ''
        return struct.pack(
            fmt,
            name,
            qtype,
            DNSUtil.QCLASS_IN,
            ttl,  # ttl
            len(rdata),
            rdata
        )

    @classmethod
    def make_rsp(cls, req_data=None, cname_text='', ip=''):
        rsp = bytearray(req_data)
        rsp[2:8] = struct.pack(
            '!H2sH',
            0x8180,  # flag for standard response
            req_data[4:6],
            2,  # two answers. One for CNAME one for A record.
        )
        return ''.join([
            str(rsp),
            cls.make_rr(12, DNSUtil.build_address(cname_text), DNSUtil.QTYPE_CNAME),
            cls.make_rr(cname_text, socket.inet_aton(ip))])


def test_parse_rsp():
    rsp_data = """
    cd008180000100020000000002733309616d617a
    6f6e61777303636f6d0000010001c00c00050001
    0000044700070473332d31c00fc02e0001000100
    00001e000436e7628b""".replace('\n', '').replace(' ', '').decode('hex')
    print(DNSResponse().parse(rsp_data))


if '__main__' == __name__:
    test_parse_rsp()
