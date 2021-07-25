#!/bin/sh
#
#       /etc/rc.d/init.d/public-ip
#
#       public-ip daemon
#
# chkconfig:   2345 95 05
# description: a public-ip script

### BEGIN INIT INFO
# Provides:       public-ip
# Required-Start:
# Required-Stop:
# Should-Start:
# Should-Stop:
# Default-Start: 2 3 4 5
# Default-Stop:  0 1 6
# Short-Description: public-ip
# Description: public-ip
### END INIT INFO

cd "$(dirname "$0")"

test -f .env && . $(pwd -P)/.env

_start() {
    test $(ulimit -n) -lt 100000 && ulimit -n 100000
    (env ENV=${ENV:-development} ./public-ip) <&- >public-ip.error.log 2>&1 &
    local pid=$!
    echo -n "Starting public-ip(${pid}): "
    sleep 1
    if (ps ax 2>/dev/null || ps) | grep "${pid} " >/dev/null 2>&1; then
        echo "OK"
    else
        echo "Failed"
    fi
}

_stop() {
    local pid="$(pidof public-ip)"
    echo -n "Stopping public-ip(${pid}): "
    if kill -HUP ${pid}; then
        echo "OK"
    else
        echo "Failed"
    fi
}

_restart() {
    if ! ./public-ip -validate ${ENV:-development}.toml >/dev/null 2>&1; then
        echo "Cannot restart public-ip, please correct public-ip toml file"
        echo "Run './public-ip -validate' for details"
        exit 1
    fi
    _stop
    sleep 1
    _start
}

_usage() {
    echo "Usage: [sudo] $(basename "$0") {start|stop|restart}" >&2
    exit 1
}

_${1:-usage}
