#!/bin/bash
export CONFIG_PATH=/Users/ts-rohan.das/go_vpn/resource.toml
go build -o build/vpnctl
sudo mv build/vpnctl /usr/local/bin/