#!/bin/sh

set -e

install_init () {
    cp /var/lib/pdagent/scripts/pdagent.init /etc/init.d/pdagent
    chmod +x /etc/init.d/pdagent
    chkconfig --add pdagent
    service pdagent restart
}

install_systemd () {
    cp /var/lib/pdagent/scripts/pdagent.service /lib/systemd/system
    systemctl enable pdagent
    systemctl start pdagent
}

# Create pdagent user & group
/usr/bin/getent passwd pdagent >/dev/null || \
    /usr/sbin/adduser --system --shell /bin/false --no-create-home pdagent

chown -R pdagent:pdagent /var/db/pdagent /var/lib/pdagent /var/log/pdagent /var/run/pdagent

if which systemctl >/dev/null; then
    install_systemd
else
    install_init
fi

exit 0
