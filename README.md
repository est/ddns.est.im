ddns.est.im
===========

dynamic DNS you say?


wat
---

  - set any ddns to any domain you want. open webpage -> my.ddns.est.im -> filling ip or hostname -> done.
    - implicit CNAME
      - cache and aggregating service
        - cache with TTL
        - combine results instead of Round-Robin
        - dynamic weight
        - premium http/smtp checking
        - blacklist & cache invalidation policy
    - explicit CNAME
  - loose API in RESTful manner
  - detect recursive DNS API, in case you are behind a DNS forwarder.
  - monitor port 53 spams
    - socket reuse routing to other apps
    - every query is logged
  - experimental R/W Python Dict Notation Format as "Zone File"
    - functions are also values (lambdas)
    - basically dict()
    - deep mergable
    - iterative parsing streams
    - differential transport in incremental data
  - collect ISP DNS from all over the world
    - act as both forwarder and recursive dns
    - history IP database
