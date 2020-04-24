#!/bin/sh

set -e

if [ $1 -gt 0 ]; then # this is an upgrade
    : # no-op
else # this is a remove
    : # no-op
fi

exit 0
