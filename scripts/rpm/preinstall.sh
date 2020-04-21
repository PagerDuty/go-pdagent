#!/bin/sh

set -e

if [ $1 -gt 1 ]; then # this is an upgrade
    if which systemctl >/dev/null && [ -e /etc/init.d/pdagent ]; then
        /etc/init.d/pdagent stop
    fi
else # this is a remove
    : # no-op
fi

exit 0
