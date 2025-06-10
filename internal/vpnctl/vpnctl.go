package internal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/goo-apps/vpnctl/internal/middleware"
	"github.com/goo-apps/vpnctl/logger"

	"github.com/common-nighthawk/go-figure"
)

var (
	vpnCmd = "/opt/cisco/secureclient/bin/vpn"
	guiApp = "/Applications/Cisco/Cisco Secure Client.app/"
)

// Status checks the current VPN connection status using the Cisco Secure Client command line tool.
// It runs the command with a timeout to avoid hanging indefinitely.
// If the command times out, it logs an error and returns.
// If the command fails, it logs the error and prints the output.
// If the command succeeds, it logs the success and prints the output.
// This function is useful for checking if the VPN is currently connected or disconnected.
func Status() {
	logger.LogInfo("Checking VPN status...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, vpnCmd, "status", "-s")
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		logger.LogError(fmt.Errorf("vpn status timed out"), "status check")
		return
	}

	if err != nil {
		logger.LogError(err, "retrieving VPN status")
		fmt.Println(string(output))
		return
	}

	logger.LogInfo("VPN status retrieved successfully")
	fmt.Println(string(output))

}

// func Status() {
// 	cmd := exec.Command(vpnCmd, "status")
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		logger.LogError("Error retrieving VPN status.")
// 		return
// 	}
// 	logger.LogInfo(string(output))
// }

// Disconnect terminates the current VPN connection and kills the Cisco Secure Client GUI.
// It runs the `vpn disconnect` command to disconnect the VPN.
// After disconnecting, it attempts to kill the Cisco Secure Client GUI process using `pkill`.
func DisconnectWithKillPid() {
	logger.LogInfo("Attempting to disconnect VPN...")
	exec.Command(vpnCmd, "disconnect").Run()
	logger.LogInfo("VPN disconnected")
	exec.Command("pkill", "-x", "Cisco Secure Client").Run()
	logger.LogInfo("Cisco Secure Client process killed")

	// Optionally, you can also kill the VPN process
	pids, err := getPIDs("vpn")
	if err != nil {
		logger.LogError(err, "getting VPN PIDs")
		return
	}
	for _, pid := range pids {
		process, err := os.FindProcess(pid)
		if err != nil {
			logger.LogError(err, fmt.Sprintf("finding process %d", pid))
			continue
		}
		logger.LogInfo(fmt.Sprintf("Killing VPN process with PID %d", pid))
		err = process.Signal(os.Interrupt) // Use SIGINT to gracefully stop
		if err != nil {
			logger.LogError(err, fmt.Sprintf("killing VPN process %d", pid))
		}
	}
	logger.LogInfo("VPN disconnected and related processes killed")
	// killVPNProcesses() // <-- updated
}

func Disconnect() {
	logger.LogInfo("Attempting to disconnect VPN...")
	exec.Command(vpnCmd, "disconnect").Run()
	logger.LogInfo("VPN disconnected")
	exec.Command("pkill", "-x", "Cisco Secure Client").Run()
	logger.LogInfo("Cisco Secure Client process killed")

}

// KillGUI kills the Cisco Secure Client GUI and optionally the VPN process
// It uses `pkill` to find and terminate the GUI process by name.
// It also retrieves the VPN process IDs using `pgrep` and sends an interrupt signal to them.
// This function is useful for cleaning up the GUI and VPN processes when they are no longer needed.
func KillGUI() {
	logger.LogInfo("Killing Cisco Secure Client GUI...")
	exec.Command("pkill", "-x", "Cisco Secure Client").Run()
	logger.LogInfo("Cisco GUI killed")
	// Optionally, you can also kill the VPN process
	pids, err := getPIDs("vpn")
	if err != nil {
		logger.LogError(err, "getting VPN PIDs")
		return
	}
	for _, pid := range pids {
		process, err := os.FindProcess(pid)
		if err != nil {
			logger.LogError(err, fmt.Sprintf("finding process %d", pid))
			continue
		}
		logger.LogInfo(fmt.Sprintf("Killing VPN process with PID %d", pid))
		err = process.Signal(os.Interrupt) // Use SIGINT to gracefully stop
		if err != nil {
			logger.LogError(err, fmt.Sprintf("killing VPN process %d", pid))
		}
	}
	logger.LogInfo("VPN processes killed")
}

// getPIDs retrieves process IDs for a given process name using pgrep
// It runs the `pgrep -f` command to find all processes matching the name.
// It returns a slice of integers representing the process IDs.
// If an error occurs during execution, it returns an error.
// This function is useful for finding and managing processes by name.
func getPIDs(name string) ([]int, error) {
	out, err := exec.Command("pgrep", "-f", name).Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var pids []int
	for _, line := range lines {
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

// LaunchGUI starts the Cisco Secure Client GUI application.
// It uses the `open` command to launch the GUI application.
// If an error occurs while launching the GUI, it logs the error.
// This function is useful for starting the GUI after a successful VPN connection.
func LaunchGUI() {
	logger.LogInfo("Launching Cisco Secure Client GUI...")
	err := exec.Command("open", guiApp).Run()
	if err != nil {
		logger.LogError(err, "launching Cisco Secure Client GUI")
	}
	logger.LogInfo("Cisco GUI launched")
}

// Connect establishes a VPN connection using the specified profile.
// It reads the VPN profile script from a hidden path, replaces placeholders with credentials,
// and executes the VPN command with the provided script.
// It checks the current VPN connection status before attempting to connect.
// If the VPN is already connected, it aborts the connection operation.
// It also reads the credentials for the specified profile from a hidden file.
func Connect(profile string) {
	logger.LogInfo(fmt.Sprintf("Initiating VPN connection using profile: %s", profile))

	profilePath := getProfilePath(profile)
	if profilePath == "" {
		logger.LogInfo("Unknown VPN profile")
		return
	}

	username, password, y, y2, err := readCredentials(profile)
	if err != nil {
		logger.LogError(err, "reading credentials")
		return
	}

	logger.LogInfo("Checking current VPN connection status...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, vpnCmd, "status")
	output, _ := cmd.CombinedOutput()

	// check the current vpn profile
	if contains(string(output), "Connected") {
		last, err := middleware.GetLastConnectedProfile()
		if err != nil {
			logger.LogError(err, "retrieve error:")
		}
		logger.LogInfo("Last connected VPN profile: " + last)

		if last == profile {
			logger.LogInfo(fmt.Sprintf("VPN already connected to profile: %s. Aborting connect operation.", profile))
			return
		}
		logger.LogInfo(fmt.Sprintf("VPN connected to profile %s, switching to %s...", last, profile))
		Disconnect()
	}

	logger.LogInfo(fmt.Sprintf("Reading VPN profile script from (hidden path) %s", ""))
	scriptContent, err := os.ReadFile(profilePath)
	if err != nil {
		logger.LogError(err, "reading VPN profile")
		return
	}

	script := strings.ReplaceAll(string(scriptContent), "{{USERNAME}}", username)
	script = strings.ReplaceAll(script, "{{PASSWORD}}", password)
	script = strings.ReplaceAll(script, "{{Y}}", y)
	if profile == "dev" {
		script = strings.ReplaceAll(script, "{{Y2}}", y2)
	}

	tempScript := filepath.Join(os.TempDir(), "vpn_input.txt")
	err = os.WriteFile(tempScript, []byte(script), 0600)
	if err != nil {
		logger.LogError(err, "writing temp VPN input file")
		return
	}

	logger.LogInfo("Running VPN command with provided script")

	vpnProfile := ""
	switch profile {
	case "dev":
		vpnProfile = "DEV-VPN-REMOTE"
	case "intra":
		vpnProfile = "INTRA"
	}

	cmd = exec.Command(vpnCmd, "connect", vpnProfile, "-s")

	stdinFile, err := os.Open(tempScript)
	if err != nil {
		logger.LogError(err, "opening temp script")
		return
	}
	defer stdinFile.Close()
	defer os.Remove(tempScript)

	cmd.Stdin = stdinFile

	// Capture stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.LogError(err, "getting stdout pipe")
		return
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.LogError(err, "getting stderr pipe")
		return
	}

	// Start command before scanning
	if err := cmd.Start(); err != nil {
		logger.LogError(err, "starting VPN command")
		return
	}

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			fmt.Println("[VPN stdout] " + scanner.Text())
		}
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			logger.LogError(errors.New(scanner.Text()), "VPN stderr")
		}
	}()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		logger.LogError(err, "VPN command exited with error")
		return
	}

	// set the last connected profile in the database
	if err := middleware.SetLastConnectedProfile(profile); err != nil {
		logger.LogError(err, "store error:")
	}

	LaunchGUI()
}

// ShowHelp prints the help message for the vpnctl CLI tool.
// It lists all available commands and their descriptions.
// This function is useful for users to understand how to use the tool and what commands are available.
func ShowHelp() {
	fmt.Println(`
🛡️  vpnctl - VPN Helper CLI
-----------------------------
vpnctl status           Show VPN status
vpnctl connect intra    Connect using intra profile
vpnctl connect dev      Connect using dev profile
vpnctl disconnect       Disconnect VPN and kill GUI with kill pid
vpnctl kill             Kill Cisco Secure Client GUI only
vpnctl gui              Launch Cisco GUI
vpnctl help             Show this help message
	`)
}

// getProfilePath returns the file path for the specified VPN profile.
// It constructs the path based on the user's home directory and the profile name.
// The function supports "intra" and "dev" profiles, returning the appropriate credential file path.
// If the profile name is not recognized, it returns an empty string.
// This function is useful for locating the VPN profile scripts needed for connection.
func getProfilePath(name string) string {
	u, _ := user.Current()
	switch name {
	case "intra":
		return filepath.Join(u.HomeDir, ".vpnctl", ".credential", ".credential_intra")
	case "dev":
		return filepath.Join(u.HomeDir, ".vpnctl", ".credential", ".credential_dev")
	default:
		return ""
	}
}

// contains checks if the given text contains the specified keyword.
// It uses the strings.Contains function to perform the check.
// This function is useful for determining if a specific keyword is present in a string,
// such as checking the output of a command for a specific status message.
func contains(text, keyword string) bool {
	return strings.Contains(text, keyword)
}

// readCredentials reads the credentials for the specified VPN profile from a hidden file.
// It constructs the file path based on the user's home directory and the profile name.
// It supports "intra" and "dev" profiles, returning the appropriate credentials.
// The function reads the file, splits it into lines, and returns the credentials as strings.
// If the file cannot be read or the format is invalid, it returns an error.
// This function is useful for securely retrieving the VPN credentials needed for connection.
func readCredentials(profile string) (string, string, string, string, error) {
	basePath := filepath.Join(os.Getenv("HOME"), ".vpnctl", ".credential")
	var credentialsPath string

	switch profile {
	case "intra":
		credentialsPath = filepath.Join(basePath, ".credential_intra")
	case "dev":
		credentialsPath = filepath.Join(basePath, ".credential_dev")
	default:
		return "", "", "", "", fmt.Errorf("unsupported profile for credentials")
	}

	logger.LogInfo(fmt.Sprintf("Reading credentials for %s (hidden path)", profile))
	creds, err := os.ReadFile(credentialsPath)
	if err != nil {
		return "", "", "", "", err
	}

	lines := strings.Split(strings.TrimSpace(string(creds)), "\n")
	if profile == "intra" && len(lines) >= 3 {
		return strings.TrimSpace(lines[0]), strings.TrimSpace(lines[1]), strings.TrimSpace(lines[2]), "", nil
	} else if profile == "dev" && len(lines) >= 4 {
		return strings.TrimSpace(lines[0]), strings.TrimSpace(lines[1]), strings.TrimSpace(lines[2]), strings.TrimSpace(lines[3]), nil
	}

	return "", "", "", "", fmt.Errorf("invalid credentials format for profile: %s", profile)
}

// Info prints the information about the vpnctl CLI tool.
// It displays the banner, author information, email, version, and a message indicating that the version is up to date.
// It also provides a hint to run 'vpnctl help' for available commands.
// This function is useful for providing users with basic information about the tool and how to get started.
func Info() {
	banner := figure.NewColorFigure("VPNCTL", "slant", "green", true)
	banner.Print()
	fmt.Println("vpnctl - VPN Helper CLI for Cisco Secure Client")
	fmt.Println("Author: @Rohan Das")
	fmt.Println("Email: ts-rohan.das@rakuten.com")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Your version is up to date!")
	fmt.Print("Run 'vpnctl help' for available commands\n")
}
