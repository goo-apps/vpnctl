# vpnctl 🛡️

[![Actions Status](https://github.com/goo-apps/vpnctl/actions/workflows/workflow-file/badge.svg)](https://github.com/goo-apps/vpnctl/actions)

`vpnctl` is a lightweight CLI wrapper to manage Cisco Secure Client VPN sessions on macOS.

---

## Features

- 🔐 Connect to VPN profiles (`intra`, `dev`)
- ✅ Check current VPN status
- ❌ Disconnect VPN + kill GUI
- 🪓 Kill only the GUI process
- 💻 Launch the GUI manually

---

## Installation

```bash
git clone https://github.com/yourname/vpnctl.git
cd vpnctl
go build -o vpnctl
sudo mv vpnctl /usr/local/bin/
```
