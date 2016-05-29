#!/usr/bin/env python
# coding: utf8


import socket
import logging
import time
import datetime
import urllib2
import json
import re

import dns
import weather as weather_svr
from ipip import IPData
from caching import ExpiringCache

NMC_TTL = 3600  # nmc.cn has 1 hour http call cache.

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


@ExpiringCache(timeout=NMC_TTL)
def get_weather_condition(station_id):
    logging.debug('Start HTTP calls...')
    station_id = int(station_id)
    url = weather_svr.WeatherService.build_nmc_cn_city_id_url(station_id)
    logging.debug('get city_id %s', url)
    r = json.load(urllib2.urlopen(url))
    city_id = r[3]
    url = weather_svr.WeatherService.build_nmc_cn_forecast_url(city_id)
    logging.debug('get forecast %s', url)
    r = json.load(urllib2.urlopen(url))
    ret = weather_svr.WeatherService.parse_nmc_cn(r)
    logging.debug('return %s', ret)
    return ret


def handler(req, addr):
    """
    receive: data
    return: cname, rr_value
    """

    m = re.search(r'^(\w+)\.(?:tempo|weather|tq|tianqi)\.est\.im\.?$',
                  req.name)
    if m:
        pinyin = m.group(1)
        svc = weather_svr.WeatherService(pinyin)
        if not svc.station_id:
            m = False  # a hack to handle non-exist city
    else:
        return 'tempo.est.im', addr[0]
    if not m:  # elif re.search(r'^tempo\.est\.im\.?$', req.name):
        pinyin, station_id, name = GeoWeather.get_station_by_ip(addr[0])
        logging.debug('ip: %s, station: %s', addr[0], station_id)
        svc = weather_svr.WeatherService(station_id)

    condition = get_weather_condition(svc.station_id)
    status = '%s.%s.tempo.est.im' % (condition, pinyin)

    return status, addr[0]


def run_server():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind(('0.0.0.0', 53))
    while True:
        req_data, addr = sock.recvfrom(4096)
        t0 = time.time()
        # logging.debug('[Req] %s bytes from %s: ', len(data), addr)
        req = dns.DNSRequest.parse(req_data)
        # cname, rr_value = handler(data, addr)
        cname, rr_value = handler(req, addr)
        rsp_data = dns.DNSResponse.make_rsp(req_data, safe_str(cname), safe_str(rr_value))
        sock.sendto(rsp_data, addr)
        logging.info(
            '[Query] %s %s:%s %sB->%sB in %03dms. [Ask]: %s in %s, [Rsp]: %s(%s)',
            datetime.datetime.fromtimestamp(t0).isoformat(' '),
            addr[0], addr[1], len(req_data), len(rsp_data),
            (time.time() - t0) * 1000.0,
            req.name, req.qtype_name, cname, rr_value)

if '__main__' == __name__:
    run_server()
