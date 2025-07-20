#!/bin/bash

# export CONFIG_PATH=/Users/ts-rohan.das/go_vpn/config/resource.toml
# Update environment to PRODUCTION in the TOML file
# awk '
#   BEGIN { in_app=0 }
#   /^\[application\]/ { print; in_app=1; next }
#   in_app && /^environment *=/ { print "environment = \"PRODUCTION\""; in_app=0; next }
#   in_app && /^[[]/ { print "environment = \"PRODUCTION\""; in_app=0 }
#   { print }
#   END { if (in_app) print "environment = \"PRODUCTION\"" }
# ' "$CONFIG_PATH" > "${CONFIG_PATH}.tmp" && mv "${CONFIG_PATH}.tmp" "$CONFIG_PATH"
go mod tidy
go build -o build/vpnctl
sudo mv build/vpnctl /usr/local/bin/vpnctl
vpnctl