#!/bin/bash
go build -o build/vpnctl
sudo mv build/vpnctl /usr/local/bin/