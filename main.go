// Author: rohan.das

// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 goo-apps (rohan.das1203@gmail.com)
// Licensed under the MIT License. See LICENSE file for details.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	goautobuild "github.com/goo-apps/go-auto-build"
	"github.com/goo-apps/vpnctl/cmd/vpnctl"
	"github.com/goo-apps/vpnctl/config"
	"github.com/goo-apps/vpnctl/internal/handler"
	"github.com/goo-apps/vpnctl/internal/middleware"
	"github.com/goo-apps/vpnctl/internal/model"
	"github.com/goo-apps/vpnctl/logger"
	"golang.org/x/term"

	"github.com/common-nighthawk/go-figure"
)

// introduction to cpnctl fro user
func info() {
	banner := figure.NewColorFigure("VPNCTL", "basic", "green", true)

	// fetch version from remote
	latest, verr := vpnctl.FetchLatestPreOrStableRelease()
	if verr != nil {
		logger.Fatalf("error/timeout fetching latest release: %v", verr)
		logger.Warningf("please re-run the application command: %v", verr)
	}

	// set in DB
	go func() {
		serr := middleware.SetLatestVersionToDB(config.APPLICATION_VERSION)
		if serr != nil {
			logger.Warningf("vpnctl went into some error: %s", serr)
		}
	}()

	version, err := middleware.GetLatestVerisionFromDB()
	if err != nil {
		logger.Errorf("vpnctl went into some error: %v", err)
	}

	banner.Print()
	// fmt.Printf("You're using vpnctl CLI[%v]", config.APPLICATION_ENVIRONMENT)
	fmt.Println("vpnctl - VPN Helper CLI for Cisco Secure Client")
	fmt.Println("👤 Author: @Rohan Das")
	fmt.Println("📧 Email: dev.work.rohan@gmail.com")
	fmt.Printf("#️⃣  Version: %s\n", config.APPLICATION_VERSION)
	// Normalize and compare version strings
	if strings.TrimPrefix(version, "v") == strings.TrimPrefix(latest.TagName, "v") {
		fmt.Println("✅ Your version is up to date!")
	} else {
		fmt.Printf("⚠️  A newer version is available: %s\n", latest.TagName)
		fmt.Printf("👉 Download it from: %s\n", latest.HTMLURL)
		fmt.Printf("📥 Installation manual: %s\n", "https://github.com/goo-apps/vpnctl?tab=readme-ov-file#installation")
	}
	fmt.Print("📌 Run 'vpnctl help' for available commands\n")
	fmt.Println()
}

// showHelp displays the help message for vpnctl commands
func showHelp() {
	fmt.Println("🛡️  vpnctl - A Cisco Secure Client Helper CLI")
	fmt.Println("=============================================")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Command\tDescription")
	fmt.Fprintln(w, "-------\t-----------")
	fmt.Fprintln(w, "vpnctl status\tShow VPN status")
	fmt.Fprintln(w, "vpnctl connect intra\tConnect using intra profile")
	fmt.Fprintln(w, "vpnctl connect dev\tConnect using dev profile")
	fmt.Fprintln(w, "vpnctl disconnect\tDisconnect VPN and kill GUI")
	fmt.Fprintln(w, "vpnctl kill\tKill Cisco Secure Client GUI only")
	fmt.Fprintln(w, "vpnctl gui\tLaunch Cisco GUI")
	fmt.Fprintln(w, "vpnctl credential update\tUpdate your credential")
	fmt.Fprintln(w, "vpnctl credential fetch\tFetch your existing credential")
	fmt.Fprintln(w, "vpnctl credential remove\tRemove your existing credential")
	fmt.Fprintln(w, "vpnctl help\tShow this help message")
	w.Flush()
}

func main() {
	// Initialize logger: logToFile=true, verbosity=2, file=~/.vpnctl.log
	logger.InitLogger(true, "")
	defer logger.Shutdown()

	// load configuration
	configPath := os.Getenv("CONFIG_PATH")
	cerr := config.LoadAllConfigAtOnce(configPath) // loading from embedded config
	if cerr != nil {
		logger.Fatalf("Failed to load configuration: %s", cerr)
	}

	// Initialize the database (ensure it's done before API handlers)
	_, dberr := middleware.InitDB()
	if dberr != nil {
		logger.Errorf("Failed to initialize database: %s", dberr)
		os.Exit(1) // Exit if database initialization fails
	}

	if len(os.Args) < 2 {
		// If no command is provided, show the info
		if config.APPLICATION_ENVIRONMENT == "PRODUCTION" {
			info()
			return
		}
	}

	// _, errrr := servicediscovery.DetectCiscoVPNPath()
	// if errrr != nil {
	// 	logger.Warningf("Failed to detect Cisco Secure Client path: %s", errrr)
	// 	fmt.Println("Please ensure Cisco Secure Client is installed and in your PATH.")
	// }

	var err error
	// Process CLI commands after credentials are set
	if len(os.Args) >= 2 {
		var credential *model.CREDENTIAL_FOR_LOGIN
		cmd := os.Args[1]
		switch cmd {
		case "connect":
			if len(os.Args) < 3 {
				fmt.Print("Please specify profile: intra or dev")
				return
			}
			credential, err = handler.GetOrPromptCredential()
			if err != nil {
				logger.Fatalf("Failed to get credentials: %s", err)
				return
			}
			vpnctl.Connect(credential, os.Args[2])
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
		// case "register-credential":
		// handler.StoreCredential(profile, credential.Username, credential.Password)
		case "credential":
			if len(os.Args) < 3 {
				fmt.Print("Please specify an operation: fetch or update or delete")
				return
			}

			switch os.Args[2] {
			case "fetch":
				creds, err := handler.GetCredential()
				if err != nil {
					logger.Fatalf("Failed to fetch credential from keyring: %s", err)
					return
				}
				data, err := json.MarshalIndent(creds, "", "  ")
				if err != nil {
					fmt.Println("Error marshaling to JSON:", err)
				} else {
					fmt.Println(string(data))
				}
			case "update":
				reader := bufio.NewReader(os.Stdin)
				fmt.Printf("Enter username for '%s': ", config.KEYRING_SERVICE_NAME)
				username, _ := reader.ReadString('\n')
				username = strings.TrimSpace(username)

				fmt.Print("Enter password: ")
				bytePassword, _ := term.ReadPassword(int(os.Stdin.Fd()))
				password := string(bytePassword)

				credential := model.CREDENTIAL_FOR_LOGIN{
					Username: username,
					Password: password,
				}
				// call store function
				err := handler.StoreCredential(credential)
				if err != nil {
					logger.Fatalf("Failed to store credential: %s", err)
					return
				}

			case "remove":
				err := handler.RemoveCredential()
				if err != nil {
					logger.Fatalf("Failed to remove credential: %s", err)
					return
				}

			}
		default:
			fmt.Printf("Unknown command: %s", cmd)
		}
	}

	cfg := &goautobuild.Config{
		ConfigPath:    "",
		OutputBinary:  "build/vpnctl",
		InstallPath:   "/usr/local/bin/vpnctl",
		PollInterval:  15,
		WatchExt:      "",
		ProjectRoot:   ".",
		EnableLogging: true,
		PostBuildMove: true,
		BuildCommand:  "build -o build/vpnctl",
	}
	watcher := goautobuild.NewWatcher(cfg)
	if config.APPLICATION_ENVIRONMENT == "DEVELOPMENT" {
		go watcher.Start()
		select {} // for indefinite run
	}
}
