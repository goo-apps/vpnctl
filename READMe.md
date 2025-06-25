# vpnctl

**vpnctl** is a cross-platform CLI tool designed to help users manage Cisco Secure Client VPN connections efficiently from the terminal.

---

## Features

- **Connect/Disconnect** to VPN profiles (`intra`, `dev`)
- **Show VPN status**
- **Launch/Kill Cisco Secure Client GUI**
- **Manage credentials** (store, fetch, update, remove) securely
- **Auto-detect Cisco Secure Client installation**
- **Developer-friendly auto-build/watch mode**

---

## Author

- **Rohan Das**
- Email: [dev.work.rohan@gmail.com](mailto:dev.work.rohan@gmail.com)

---

## Download

Download the latest release: [download vpnctl](https://github.com/goo-apps/vpnctl/releases)

---

## Usage

After installing `vpnctl`, you can use the following commands:

| Command                              | Description                                 |
|---------------------------------------|---------------------------------------------|
| `vpnctl status`                       | Show VPN status                             |
| `vpnctl connect intra`                | Connect using intra profile                 |
| `vpnctl connect dev`                  | Connect using dev profile                   |
| `vpnctl disconnect`                   | Disconnect VPN and kill GUI                 |
| `vpnctl kill`                         | Kill Cisco Secure Client GUI only           |
| `vpnctl gui`                          | Launch Cisco GUI                            |
| `vpnctl credential update`            | Update your credential                      |
| `vpnctl credential fetch`             | Fetch your existing credential              |
| `vpnctl credential remove`            | Remove your existing credential             |
| `vpnctl help`                         | Show help message                           |
| `vpnctl info`                         | Show version and author info                |

---

## Example

```sh
vpnctl connect intra
vpnctl status
vpnctl credential update
vpnctl disconnect
```

---

## Reporting Bugs & Issues

If you encounter any bugs or issues, **please open an issue in the [Issues section](https://github.com/goo-apps/vpnctl/issues) before submitting a pull request (PR)**. This helps us track and discuss problems before code changes are proposed.

---

## Contribution Guidelines

We welcome contributions! Please follow these steps:

1. **Fork** the repository and create your branch from `main`.
2. **Open an issue** to discuss your proposed change before working on a PR.
3. Make your changes with clear, concise commits.
4. Ensure your code passes all tests and lint checks.
5. Submit a **pull request** referencing the related issue.

By contributing, you agree to follow our [Code of Conduct](CODE_OF_CONDUCT.md) and help us maintain a welcoming community.

---

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## Notes

- Ensure Cisco Secure Client is installed and in your system `PATH`.
- Credentials are stored securely using the system keyring.