// Author: rohan.das

// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 goo-apps (rohan.das1203@gmail.com)
// Licensed under the MIT License. See LICENSE file for details.

package main

import (
	"fmt"
	"os"

	"github.com/goo-apps/vpnctl/config"
	"github.com/goo-apps/vpnctl/internal/handler"
	"github.com/goo-apps/vpnctl/internal/middleware"
	"github.com/goo-apps/vpnctl/internal/model"
	"github.com/goo-apps/vpnctl/internal/vpnctl"
	"github.com/goo-apps/vpnctl/logger"

	"github.com/common-nighthawk/go-figure"
)

func info() {
	banner := figure.NewColorFigure("VPNCTL", "basic", "green", true)
	banner.Print()
	fmt.Println()
	fmt.Printf("You're using vpnctl CLI[%v]", config.APPLICATION_ENVIRONMENT)
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

	// load configuration
	// configPath := os.Getenv("CONFIG_PATH")
	// if configPath == "" {
	// 	logger.Fatalf("CONFIG_PATH is not set for resource")
	// }
	config.LoadAllConfigAtOnce("") // loading from embedded config

	// Initialize the database (ensure it's done before API handlers)
	_, dberr := middleware.InitDB()
	if dberr != nil {
		logger.Errorf("Failed to initialize database: %s", dberr)
		os.Exit(1) // Exit if database initialization fails
	}

	if len(os.Args) < 2 {
		// If no command is provided, show the info
		info()
		return
	}

	var profile string
	var err error
	// Process CLI commands after credentials are set
	if len(os.Args) >= 2 {
		var credential model.CREDENTIAL_FOR_LOGIN
		cmd := os.Args[1]
		switch cmd {
		case "connect":
			if len(os.Args) < 3 {
				fmt.Print("Please specify profile: intra or dev")
				return
			}
			profile = os.Args[2]
			credential, err = handler.GetOrPromptCredential(profile)
			if err != nil {
				logger.Fatalf("Failed to get credentials: %s", err)
				return
			}
			vpnctl.Connect(&credential, os.Args[2])
		case "disconnect":
			vpnctl.DisconnectWithKillPid()
		case "status":
			vpnctl.Status()
		case "kill":
			vpnctl.KillGUI()
		case "gui":
			vpnctl.LaunchGUI()
		case "help":
			showHelp()
		case "info":
			info()
		case "register-credential":
			// handler.StoreCredential(profile, credential.Username, credential.Password)

		// case "fetch-credential":
		// 	test()
		// case "update-credential":
		// 	test()
		// case "delete-credential":
		// 	test()
		default:
			fmt.Printf("Unknown command: %s", cmd)
		}
	}
}
