#!/bin/bash
set -euo pipefail

# chmod +x scripts/collect.sh

# # Via argument
# ./scripts/collect.sh http://fleetlens.example.com:8080

# # Via environment variable
# FLEETLENS_API_URL=http://fleetlens.example.com:8080 ./scripts/collect.sh

API_URL="${1:-${FLEETLENS_API_URL:-}}"

if [ -z "$API_URL" ]; then
  echo "Usage: $0 <FLEETLENS_API_URL>" >&2
  echo "   or: FLEETLENS_API_URL=http://... $0" >&2
  exit 1
fi

# --- System info ---

HOSTNAME=$(hostname)
OS="linux"
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
esac

# --- Serial number ---

SERIAL=""
if [ -r /sys/class/dmi/id/product_serial ]; then
  SERIAL=$(cat /sys/class/dmi/id/product_serial 2>/dev/null | tr -d '\n')
elif command -v dmidecode &>/dev/null; then
  SERIAL=$(dmidecode -s system-serial-number 2>/dev/null | tr -d '\n')
fi
if [ -z "$SERIAL" ]; then
  echo "WARNING: failed to get serial number" >&2
fi

# --- IP addresses (non-loopback, IPv4) ---

IP_ADDRESSES="[]"
if command -v ip &>/dev/null; then
  ips=$(ip -4 -o addr show scope global 2>/dev/null | awk '{split($4,a,"/"); print a[1]}')
  if [ -n "$ips" ]; then
    IP_ADDRESSES=$(echo "$ips" | jq -R . | jq -s .)
  fi
fi

# --- Network interfaces ---

collect_network_interfaces() {
  local result="[]"

  for iface_path in /sys/class/net/*; do
    local iface=$(basename "$iface_path")
    local mac=$(cat "$iface_path/address" 2>/dev/null || echo "")
    local iface_type=""

    # Detect type
    if [ "$iface" = "lo" ]; then
      iface_type="loopback"
    elif [[ "$iface" =~ ^(wl|wlan) ]]; then
      iface_type="wifi"
    elif [[ "$iface" =~ ^(veth|docker|br-|virbr|vbox|vmnet|tun|tap) ]]; then
      iface_type="virtual"
    else
      iface_type="ethernet"
    fi

    # Collect all IPs for this interface
    local ips="[]"
    if command -v ip &>/dev/null; then
      local ip_list=$(ip -o addr show dev "$iface" 2>/dev/null | awk '{split($4,a,"/"); print a[1]}')
      if [ -n "$ip_list" ]; then
        ips=$(echo "$ip_list" | jq -R . | jq -s .)
      fi
    fi

    result=$(echo "$result" | jq \
      --arg name "$iface" \
      --argjson ips "$ips" \
      --arg mac "$mac" \
      --arg type "$iface_type" \
      '. + [{"name": $name, "ip_addresses": $ips, "mac_address": $mac, "type": $type}]')
  done

  echo "$result"
}

NETWORK_INTERFACES=$(collect_network_interfaces)

# --- Installed software ---

collect_software() {
  local result="[]"

  if command -v dpkg-query &>/dev/null; then
    result=$(dpkg-query -W -f='${Package}\t${Version}\t${Maintainer}\n' 2>/dev/null | \
      awk -F'\t' '{if($1!="") print}' | \
      jq -R 'split("\t") | {"name": .[0], "version": .[1], "publisher": .[2], "installed_at": ""}' | \
      jq -s .)
  elif command -v rpm &>/dev/null; then
    result=$(rpm -qa --queryformat '%{NAME}\t%{VERSION}\t%{VENDOR}\t%{INSTALLTIME:date}\n' 2>/dev/null | \
      jq -R 'split("\t") | {"name": .[0], "version": .[1], "publisher": .[2], "installed_at": .[3]}' | \
      jq -s .)
  else
    echo "WARNING: no supported package manager found (tried dpkg, rpm)" >&2
  fi

  echo "$result"
}

SOFTWARE=$(collect_software)

# --- Users ---

collect_users() {
  local result="[]"
  local admin_groups

  # Get admin group members (sudo and wheel)
  admin_groups=$(awk -F: '/^(sudo|wheel):/ {print $4}' /etc/group 2>/dev/null | tr ',' '\n')

  # Get last login times
  declare -A last_logins
  if command -v lastlog &>/dev/null; then
    while IFS= read -r line; do
      local user=$(echo "$line" | awk '{print $1}')
      if echo "$line" | grep -q "Never logged in"; then
        continue
      fi
      local login_date=$(echo "$line" | awk '{for(i=4;i<=NF;i++) printf "%s ", $i; print ""}' | xargs)
      if [ -n "$user" ] && [ -n "$login_date" ]; then
        last_logins[$user]="$login_date"
      fi
    done <<< "$(lastlog 2>/dev/null | tail -n +2)"
  fi

  while IFS=: read -r username _ uid _ gecos _ shell; do
    # Skip system accounts
    if [[ "$shell" == *"nologin"* ]] || [[ "$shell" == *"/false"* ]]; then
      continue
    fi

    # Display name from GECOS
    local display_name="${gecos%%,*}"
    if [ -z "$display_name" ]; then
      display_name="$username"
    fi

    # Check admin
    local is_admin=false
    if [ "$username" = "root" ] || echo "$admin_groups" | grep -qx "$username"; then
      is_admin=true
    fi

    # Last login
    local last_login="${last_logins[$username]:-}"

    result=$(echo "$result" | jq \
      --arg username "$username" \
      --arg display_name "$display_name" \
      --argjson is_admin "$is_admin" \
      --arg last_login "$last_login" \
      '. + [{"username": $username, "display_name": $display_name, "is_admin": $is_admin, "last_login": $last_login}]')
  done < /etc/passwd

  echo "$result"
}

USERS=$(collect_users)

# --- Build and send payload ---

PAYLOAD=$(jq -n \
  --arg hostname "$HOSTNAME" \
  --argjson ip_addresses "$IP_ADDRESSES" \
  --arg serial_number "$SERIAL" \
  --arg operating_system "$OS" \
  --arg architecture "$ARCH" \
  --argjson network_interfaces "$NETWORK_INTERFACES" \
  --argjson software "$SOFTWARE" \
  --argjson users "$USERS" \
  '{
    "hostname": $hostname,
    "ip_addresses": $ip_addresses,
    "serial_number": $serial_number,
    "operating_system": $operating_system,
    "architecture": $architecture,
    "network_interfaces": $network_interfaces,
    "software": $software,
    "users": $users
  }')

echo "$PAYLOAD" | jq .

HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD" \
  "${API_URL}/api/agents")

if [ "$HTTP_STATUS" -ge 200 ] && [ "$HTTP_STATUS" -lt 300 ]; then
  echo "System info sent successfully (HTTP $HTTP_STATUS)"
else
  echo "ERROR: API returned HTTP $HTTP_STATUS" >&2
  exit 1
fi
