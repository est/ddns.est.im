#!/bin/bash
echo -e '\xdf\x12\x00\x10\x00\x01\x00\x00\x00\x00\x00\x01\x07chengdu\x05tempo\x03est\x02im\x00\x00\x01\x00\x01\x00\x00)\x05\xc8\x00\x00\x80\x00\x00\x00' | nc -uw 2 localhost 5354