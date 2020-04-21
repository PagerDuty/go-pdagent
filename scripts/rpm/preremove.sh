#!/bin/sh

set -e

uninstall_init () {
    service pdagent stop
    chkconfig --del pdagent
}

uninstall_systemd () {
    systemctl stop pdagent
    systemctl disable pdagent
}

if [ $1 -gt 0 ]; then # this is an upgrade
    : # no-op
else # this is a remove
    if which systemctl >/dev/null; then
        uninstall_systemd
    else
        uninstall_init
    fi
fi

exit 0
