#!/usr/bin/env python
# coding: utf8


import socket
import logging
import urllib2

import weather
from dns import DNSResponse
from ipip import IPData

logging.basicConfig(level=logging.DEBUG)


class GeoWeather(object):
    geoip = IPData('17monipdb.dat')
    station_ids = dict((x[2], (x)) for x in weather.CHINA_STATIONS)

    @classmethod
    def get_station_by_ip(cls, ip):
        r = cls.geoip.find(ip).strip().split('\t')
        logging.debug('get_station_id_by_ip: %s', r)
        if len(r) == 3:
            return cls.station_ids.get(r[2])


def handler(data, addr):
    """
    receive: data
    return: cname, rr_value
    """
    station = GeoWeather.get_station_by_ip(addr[0])
    logging.debug('ip is %s, %s', addr[0], station[1])
    return station[0].encode('utf8'), addr[0]


def run_server():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind(('0.0.0.0', 53))
    while True:
        data, addr = sock.recvfrom(4096)
        logging.debug('[Req] %s bytes from %s: ', len(data), addr)
        # cname, rr_value = handler(data, addr)
        cname, rr_value = handler(data, ('118.113.59.116', addr[1]))
        data = DNSResponse.make_rsp(data, cname, rr_value)
        logging.debug('[Rsp] %s bytes to %s', len(data), addr)
        sock.sendto(data, addr)

if '__main__' == __name__:
    run_server()
