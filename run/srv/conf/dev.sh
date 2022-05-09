#!/bin/bash
# 需要先kill掉旧的dlv
DELVE_PID=$(ps aux | grep -E 'dlv|nuwas' | grep -v 'grep' | awk  '{print $2}');
if [ -n "$DELVE_PID" ] > /dev/null 2>&1 ;
  then echo "Located existing delve process running with PID: $DELVE_PID. Killing." ;
  kill -9 $DELVE_PID;
fi;

dlv exec --headless --continue  --accept-multiclient --api-version=2 --listen=:2345 --log /tmp/nuwas -- $BUILD_MAIN_FLAG