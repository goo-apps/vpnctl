// Author: rohan.das

// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 goo-apps (rohan.das1203@gmail.com)
// Licensed under the MIT License. See LICENSE file for details.
package servicediscovery

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func findVPNInPath() (string, error) {
	path, err := exec.LookPath("vpn")
	if err == nil {
		return path, nil
	}
	return "", fmt.Errorf("vpn binary not found in PATH")
}

func findVPNOnWindows() (string, error) {
	// Try registry
	regQuery := `reg query "HKLM\SOFTWARE\Cisco\Cisco AnyConnect Secure Mobility Client" /v InstallDir`
	out, err := exec.Command("cmd", "/C", regQuery).Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "InstallDir") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					installDir := parts[len(parts)-1]
					vpnPath := installDir + `\vpn.exe`
					if _, err := os.Stat(vpnPath); err == nil {
						return vpnPath, nil
					}
				}
			}
		}
	}
	// Fallback to PATH
	return findVPNInPath()
}

func findVPNOnMac() (string, error) {
	// Try PATH first
	if path, err := findVPNInPath(); err == nil {
		return path, nil
	}
	// Try mdfind for .app bundles
	out, err := exec.Command("mdfind", "kMDItemCFBundleIdentifier == 'com.cisco.anyconnect.gui'").Output()
	if err == nil && len(out) > 0 {
		appPaths := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, appPath := range appPaths {
			vpnPath := appPath + "/Contents/MacOS/vpn"
			if _, err := os.Stat(vpnPath); err == nil {
				return vpnPath, nil
			}
		}
	}
	return "", fmt.Errorf("vpn binary not found")
}

func findVPNOnLinux() (string, error) {
	return findVPNInPath()
}

func DetectCiscoVPNPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return findVPNOnWindows()
	case "darwin":
		return findVPNOnMac()
	case "linux":
		return findVPNOnLinux()
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
