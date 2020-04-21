#!/bin/sh

set -e

if [ $# -gt 1 ]; then # this is an upgrade - shutdown the old agent
   if which systemctl >/dev/null; then
      systemctl stop pdagent
   else
      service pdagent stop
   fi
fi

exit 0
