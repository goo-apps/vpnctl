package main

// Author: Rohan.das
import (
	"fmt"
	"os"

	"vpnctl/internal/credentialstore"
	"vpnctl/internal/vpnctl"
	"vpnctl/logger"

	"github.com/common-nighthawk/go-figure"
	"github.com/gtank/cryptopasta"
)

func main() {
	// Initialize logger
	logger.InitLogger()

	if len(os.Args) < 2 {
		info()
		return
	}

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

	key := cryptopasta.NewEncryptionKey() // You can load from env/config

	err := credentialstore.StoreCredential("ts-rohan.das", "yourPassword123", key[:])
	if err != nil {
		fmt.Println("Error storing credential:", err)
	}

	pass, err := credentialstore.GetCredential("ts-rohan.das", key[:])
	if err != nil {
		fmt.Println("Error getting credential:", err)
	} else {
		fmt.Println("Retrieved password:", pass)
	}

}

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