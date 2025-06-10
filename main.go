// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 Rohan
// Licensed under the MIT License. See LICENSE file for details.

package main

import (
	"fmt"
	"os"

	internal "github.com/goo-apps/vpnctl/internal/vpnctl"
	"github.com/goo-apps/vpnctl/logger"

	"github.com/common-nighthawk/go-figure"
)

func info() {
	banner := figure.NewColorFigure("VPNCTL", "", "green", true)
	banner.Print()
	fmt.Println()
	fmt.Println("vpnctl - VPN Helper CLI for Cisco Secure Client")
	fmt.Println("Author: @Rohan Das")
	fmt.Println("Email: ts-rohan.das@rakuten.com")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Your version is up to date!")
	fmt.Print("Run 'vpnctl help' for available commands\n")
	fmt.Println()
}

func showHelp() {
	fmt.Println(`
🛡️  vpnctl - VPN Helper CLI
-----------------------------
vpnctl status           Show VPN status
vpnctl connect intra    Connect using intra profile
vpnctl connect dev      Connect using dev profile
vpnctl disconnect       Disconnect VPN and kill GUI
vpnctl kill             Kill Cisco Secure Client GUI only
vpnctl gui              Launch Cisco GUI
vpnctl help             Show this help message
	`)
}

func main() {
	// Initialize logger
	logger.InitLogger()

	if len(os.Args) < 2 {
		info()
		return
	}

	// user command input
	cmd := os.Args[1]
	switch cmd {
	case "connect":
		if len(os.Args) < 3 {
			fmt.Println("Please specify profile: intra or dev")
			return
		}
		internal.Connect(os.Args[2])
	case "disconnect":
		internal.DisconnectWithKillPid()
	case "status":
		internal.Status()
	case "kill":
		internal.KillGUI()
	case "gui":
		internal.LaunchGUI()
	case "help":
		showHelp()
	case "info":
		info()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		// internal.ShowHelp()
	}

}
