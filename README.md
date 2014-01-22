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
    - basically dict() but ordered
    - reference pointers e.g. parent, sibling, etc.
    - deep traversal mergable
    - iterative/partial parsing for streams
    - differential transport in incremental data
  - collect ISP DNS from all over the world
    - act as both forwarder and recursive dns
    - history IP database
  - DNS over HTTP
    - find the quickest record using TCP PING in ajax


depends
-------

  GOPATH=$PWD go get github.com/mattn/go-sqlite3

notes
----

 - How to tell Windows & *nix:
   - Windows will lookup PTR
   - ttl maybe?
 - LMDB
   - http://symas.com/mdb/
   - lmdb.readthedocs.org 
   - gitorious.org/mdb/mdb
   - http://godoc.org/github.com/szferi/gomdb
 - sqlite
   - https://github.com/mattn/go-sqlite3/blob/master/_example/simple/simple.go



UPSERT in SQLite

Not easy. http://stackoverflow.com/a/7511635/41948

    // insert into record (id, name_r, type_id, value) values (1, 'com reddit', 1, X'173E6D57');

    INSERT OR REPLACE INTO record (id, name_r, ttl, type_id, value)
    SELECT old.id, old.name_r, new.ttl, old.type_id, old.value
    FROM ( SELECT
       'com reddit'    AS name_r, 
       1234           AS ttl,
       1  as type_id,
       X'173E6D57' as value
    ) AS new
    LEFT JOIN (
       SELECT id, name_r, type_id, value
       FROM record
    ) AS old ON new.name_r = old.name_r AND new.type_id = old.type_id AND new.value = old.value;

