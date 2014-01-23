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


INSERT OR REPLACE INTO record 
SELECT 
    old.id, 
    COALESCE(new.name_r, old.name_r), 
    MAX(COALESCE(old.ttl, 0), new.ttl), 
    COALESCE(new.type_id, old.type_id), 
    COALESCE(new.value, old.value), 
    COALESCE(old.time_added, new.time_added), 
    new.time_accessed 
FROM ( SELECT
    'com reddit' AS name_r, 
    1234 AS ttl, 
    1 AS type_id, 
    X'173E6D57' AS value, 
    datetime('now') AS time_added, 
    datetime('now') AS time_accessed 
) AS new
LEFT JOIN (
    SELECT id, name_r, ttl, type_id, value, time_added
    FROM record
) AS old ON 
    new.name_r = old.name_r AND 
    new.type_id = old.type_id AND 
    new.value = old.value;

    
        old.id, 
        COALESCE(new.name_r, old.name_r), 
        MAX(COALESCE(old.ttl, 0), new.ttl), 
        COALESCE(new.type_id, old.type_id), 
        COALESCE(new.value, old.value), 
        COALESCE(old.time_added, new.time_added), 


package main

import "fmt"

func main() {
  cb := make(chan int, 1)
        tidMap := map[int] chan int{
    1234: cb,
  }
  
  tidMap[1234] <- 333
  fmt.Println(<-tidMap[1234])
  tidMap[1234] <- 111
  tidMap[1234] <- 112         // blocks
  fmt.Println(<-tidMap[1234])
  fmt.Println(<-tidMap[1231]) // blocks



  cb := make(chan int, 1)
  cb <- 1111
  tidMap := map[int] chan int{}
  tidMap[1234] = cb
  fmt.Println(<-tidMap[1234])
  tidMap[1234] <- 111
  fmt.Println(<-tidMap[1234])

}






CNAME cache

0000   00 03 81 80 00 01 00 07 00 00 00 00 0a 75 75 75  .............uuu
0010   75 31 32 33 38 37 35 09 77 6f 72 64 70 72 65 73  u123875.wordpres
0020   73 03 63 6f 6d 00 00 01 00 01 c0 0c 00 05 00 01  s.com...........
0030   00 00 38 40 00 05 02 6c 62 c0 17 c0 36 00 01 00  ..8@...lb...6...
0040   01 00 00 01 2c 00 04 4c 4a fe 7b c0 36 00 01 00  ....,..LJ.{.6...
0050   01 00 00 01 2c 00 04 4c 4a fe 78 c0 36 00 01 00  ....,..LJ.x.6...
0060   01 00 00 01 2c 00 04 42 9b 09 ee c0 36 00 01 00  ....,..B....6...
0070   01 00 00 01 2c 00 04 c0 00 50 fa c0 36 00 01 00  ....,....P..6...
0080   01 00 00 01 2c 00 04 42 9b 0b ee c0 36 00 01 00  ....,..B....6...
0090   01 00 00 01 2c 00 04 c0 00 51 fa                 ....,....Q.