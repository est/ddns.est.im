
Funky Name Server
=

Use DNS to provide weather (and others) lookup.

Usage
==
    
    ping beijing.tempo.est.im

or 

    nslookup beijing tempo.est.im

_tempo_ means weather in Italian.


License:
==

BSD


Changelog:
==
  
 - 2016-07-24 v2.0 fixed major bug, stripped query to build response, especially for OPT type in `Additional RR`
 - 2016-05-29 v1.5 Caching with better perf.
 - 2016-05-28 v1.0 released for first working prototype


ToDo
==

 
 - [ ] UTF-7 for Linux/macOS. e.g. `echo +2D3eAg- | iconv -f utf-7`
 - [ ] Stock lookup
 - [ ] Phone prefix
 - [ ] Airport code
 - [ ] 48h or 72h format `ping beijing.48h.tempo.est.im`
 - [ ] PM2.5, maybe aqicn.org or embassy stats
 - [ ] le Windowd hack for CJK chars.
 

Ideas
==

 - ver.est.im
  - [ ] Android or iOS version detector lol

 - dl.est.im
  - [ ] lookup by TXT to grab download URL.