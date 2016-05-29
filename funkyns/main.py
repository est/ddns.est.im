#!/usr/bin/env python
# coding: utf8


import socket
import logging
import urllib2
import json

import weather as weather_svr
from dns import DNSResponse
from ipip import IPData

logging.basicConfig(level=logging.DEBUG)


def safe_str(s):
    if isinstance(s, unicode):
        return s.encode('utf8')
    return str(s)


class GeoWeather(object):
    geoip = IPData('17monipdb.dat')
    station_ids = dict((x[2], (x)) for x in weather_svr.CHINA_STATIONS)

    @classmethod
    def get_station_by_ip(cls, ip):
        r = cls.geoip.find(ip).strip().split('\t')
        logging.debug('get_station_id_by_ip: %s, %s, %s', *r)
        if len(r) == 3:
            ret = cls.station_ids.get(r[2])
            logging.debug('get_station_id_by_ip: %s, %s, %s', *ret)
            return ret[0], int(ret[1]), ret[2]
        return (None, None, None)


def handler(data, addr):
    """
    receive: data
    return: cname, rr_value
    """
    pinyin, station_id, name = GeoWeather.get_station_by_ip(addr[0])
    logging.debug('ip: %s, station: %s', addr[0], station_id)
    svc = weather_svr.WeatherService(station_id)
    url = svc.build_nmc_cn_city_id_url()
    logging.debug('get city_id %s', url)
    r = json.load(urllib2.urlopen(url))
    city_id = r[3]
    url = svc.build_nmc_cn_forecast_url(city_id)
    logging.debug('get forecast %s', url)
    r = json.load(urllib2.urlopen(url))
    status = svc.parse_nmc_cn(r)
    return status, addr[0]


def run_server():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind(('0.0.0.0', 53))
    while True:
        data, addr = sock.recvfrom(4096)
        logging.debug('[Req] %s bytes from %s: ', len(data), addr)
        # cname, rr_value = handler(data, addr)
        cname, rr_value = handler(data, ('118.113.59.116', addr[1]))
        data = DNSResponse.make_rsp(data, safe_str(cname), safe_str(rr_value))
        logging.debug('[Rsp] %s bytes to %s', len(data), addr)
        sock.sendto(data, addr)

if '__main__' == __name__:
    run_server()
