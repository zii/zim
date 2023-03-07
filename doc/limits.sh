#!/usr/bin/env bash
# 修改最大连接数脚本, 支持百万连接
# wget -q -O - https://qpools1.oss-accelerate.aliyuncs.com/sh/install_limits.sh | bash

set -e

# must CentOS
if [ ! -f /etc/centos-release ]; then
  echo "must CentOS!"
  exit 1
fi

# must root
if [ "$(whoami)" != "root" ]; then
  echo "user must root!"
  exit 1
fi

if ! cat /etc/security/limits.conf | grep -q '^\*\s*hard nofile'; then
  echo "* hard nofile 1048576" >> /etc/security/limits.conf
  echo "* soft nofile 1048576" >> /etc/security/limits.conf
  systemctl daemon-reexec
fi

if ! cat /etc/profile | grep -q '^ulimit -n'; then
  echo "ulimit -n 1048576" >> /etc/profile
fi

ulimit -n 1048576

/sbin/modprobe tcp_hybla
# sysctl net.ipv4.tcp_available_congestion_control

# net.core.somaxconn: 低版本linux最大65535
cat << EOF > /etc/sysctl.conf
net.core.rmem_default = 6291456
net.core.wmem_default = 6291456
net.core.rmem_max = 12582912
net.core.wmem_max = 12582912
net.core.netdev_max_backlog = 1048576
net.core.somaxconn = 1048576
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_tw_recycle = 0
net.ipv4.tcp_keepalive_time = 600
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_max_syn_backlog = 16384
net.ipv4.tcp_max_tw_buckets = 6000
net.ipv4.route.gc_timeout = 100
net.ipv4.tcp_fastopen = 3
net.ipv4.tcp_mem = 262144 524288 786432
net.ipv4.tcp_rmem = 4096 87380 67108864
net.ipv4.tcp_wmem = 4096 16384 4194304
net.ipv4.tcp_mtu_probing = 1
net.ipv4.tcp_max_orphans = 262114
net.ipv4.tcp_congestion_control = hybla
net.ipv4.tcp_synack_retries = 3
net.ipv4.tcp_syn_retries = 3
net.ipv4.ip_local_port_range = 3000 65000
fs.file-max = 6385998
vm.overcommit_memory = 1
EOF

sysctl -p

install_ntp() {
  if ! systemctl status ntpd | grep -q Active; then
    yum install -y ntp
    systemctl enable ntpd
    systemctl start ntpd
  fi
}

install_ntp
