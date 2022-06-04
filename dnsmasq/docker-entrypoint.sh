#!/bin/sh

# Start dnsmasq in foreground and with no-daemon, but backgrounded
dnsmasq -kd -C /etc/dnsmasq/dnsmasq.conf &

# Start watcher foregrounded
/bin/kubernetes-ingress-dns