// Author: rohan.das

// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 Rohan (dev.work.rohan@gmail.com)
// Licensed under the MIT License. See LICENSE file for details.

package main

import (
	"fmt"
	"os"

	"github.com/goo-apps/vpnctl/internal/middleware"
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
	fmt.Println("Email: dev.work.rohan@gmail.com")
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
	// Initialize logger: logToFile=true, verbosity=2, file=~/.vpnctl.log
	logger.InitLogger(true, "")
	defer logger.Shutdown()

	// write the logic to set credential by cli command

	// Initialize the database (ensure it's done before API handlers)
	_, err := middleware.InitDB()
	if err != nil {
		logger.Errorf("Failed to initialize database: %s", err)
		os.Exit(1) // Exit if database initialization fails
	}

	if len(os.Args) < 2 {
		// If no command is provided, show the info
		info()
		return
	}

	// Process CLI commands after credentials are set
	if len(os.Args) >= 2 {
		cmd := os.Args[1]
		switch cmd {
		case "connect":
			if len(os.Args) < 3 {
				logger.Warningf("Please specify profile: intra or dev")
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
		// case "register-credential":
		// 	test()
		// case "fetch-credential":
		// 	test()
		// case "update-credential":
		// 	test()
		// case "delete-credential":
		// 	test()
		default:
			logger.Warningf("Unknown command: " + cmd)
		}
	}
}
