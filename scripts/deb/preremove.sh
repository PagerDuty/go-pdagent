#!/bin/sh

set -e

uninstall_init () {
    service pdagent stop
    update-rc.d -f pdagent remove
}

uninstall_systemd () {
    systemctl stop pdagent
    systemctl disable pdagent
}

if [ "$1" = "remove" ]; then  # remove-only; not upgrade.
    if which systemctl >/dev/null; then
        uninstall_systemd
    else
        uninstall_init
    fi
fi

exit 0
