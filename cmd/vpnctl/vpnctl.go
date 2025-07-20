// Author: rohan.das

// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 goo-apps (rohan.das1203@gmail.com)
// Licensed under the MIT License. See LICENSE file for details.
package vpnctl

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

	"github.com/common-nighthawk/go-figure"
	"github.com/goo-apps/vpnctl/config"
	"github.com/goo-apps/vpnctl/internal/middleware"
	"github.com/goo-apps/vpnctl/internal/model"
	"github.com/goo-apps/vpnctl/logger"
)

var (
	shouldRetryAgentLock    bool
	shouldRetryConnectivity bool
)

// Status checks the current VPN connection status using the Cisco Secure Client command line tool.
// // It runs the command with a timeout to avoid hanging indefinitely.
// // If the command times out, it logs an error and returns.
// // If the command fails, it logs the error and prints the output.
// // If the command succeeds, it logs the success and prints the output.
// // This function is useful for checking if the VPN is currently connected or disconnected.
func Status() {
	logger.Infof("Checking VPN status...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, config.VPN_BINARY_PATH, "status", "-s")
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		logger.Errorf("status check: %v", fmt.Errorf("vpn status timed out"))
		return
	}

	if err != nil {
		logger.Errorf("retrieving VPN status: %v", err)
		fmt.Println(string(output))
		return
	}

	logger.Infof("VPN status retrieved successfully")
	fmt.Println(string(output))

}

// func Status() {
// 	cmd := exec.Command(vpnCmd, "status")
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		logger.Errorf("Error retrieving VPN status.")
// 		return
// 	}
// 	logger.Infof(string(output))
// }

// Disconnect terminates the current VPN connection and kills the Cisco Secure Client GUI.
// It runs the `vpn disconnect` command to disconnect the VPN.
// After disconnecting, it attempts to kill the Cisco Secure Client GUI process using `pkill`.
func DisconnectWithKillPid() {
	logger.Infof("Attempting to disconnect VPN...")
	exec.Command(config.VPN_BINARY_PATH, "disconnect").Run()
	logger.Infof("VPN disconnected")

	// Kill Cisco Secure Client UI only, not vpnagentd
	exec.Command("pkill", "-x", "Cisco Secure Client").Run()
	logger.Infof("Cisco Secure Client UI process killed")

	pids, err := getPIDs("vpn")
	if err != nil {
		logger.Errorf("getting VPN PIDs: %v", err)
		return
	}

	for _, pid := range pids {
		// Use `ps -p <pid> -o command=` to get the full process path/command
		out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "command=").Output()
		if err != nil {
			logger.Errorf("could not inspect process %d: %v", pid, err)
			continue
		}

		cmdline := string(out)
		cmdline = strings.TrimSpace(cmdline)

		if strings.Contains(cmdline, "vpnagentd") {
			logger.Infof("Skipping vpnagentd process (PID %d)", pid)
			continue
		}

		logger.Infof("Killing VPN process with PID %d", pid)

		process, err := os.FindProcess(pid)
		if err != nil {
			logger.Errorf(fmt.Sprintf("finding process %d", pid), err)
			continue
		}

		err = process.Signal(os.Interrupt)
		if err != nil {
			logger.Errorf(fmt.Sprintf("killing VPN process %d", pid), err)
		}
	}

	logger.Infof("VPN disconnected and related processes (excluding vpnagentd) killed")
}

func Disconnect() {
	logger.Infof("Attempting to disconnect VPN...")
	exec.Command(config.VPN_BINARY_PATH, "disconnect").Run()
	logger.Infof("VPN disconnected")
	exec.Command("pkill", "-x", "Cisco Secure Client").Run()
}

// KillGUI kills the Cisco Secure Client GUI and optionally the VPN process
// It uses `pkill` to find and terminate the GUI process by name.
// It also retrieves the VPN process IDs using `pgrep` and sends an interrupt signal to them.
// This function is useful for cleaning up the GUI and VPN processes when they are no longer needed.
func KillGUI() {
	logger.Infof("Killing Cisco Secure Client GUI...")
	exec.Command("pkill", "-x", "Cisco Secure Client").Run()
	logger.Infof("Cisco GUI killed")
	// Optionally, you can also kill the VPN process
	pids, err := getPIDs("vpn")
	if err != nil {
		logger.Errorf("getting VPN PIDs: %v", err)
		return
	}
	for _, pid := range pids {
		process, err := os.FindProcess(pid)
		if err != nil {
			logger.Errorf(fmt.Sprintf("finding process %d", pid), err)
			continue
		}
		logger.Infof(fmt.Sprintf("Killing VPN process with PID %d", pid))
		err = process.Signal(os.Interrupt) // Use SIGINT to gracefully stop
		if err != nil {
			logger.Errorf(fmt.Sprintf("killing VPN process %d", pid), err)
		}
	}
	logger.Infof("VPN processes killed")
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
	logger.Infof("Launching Cisco Secure Client GUI...")
	err := exec.Command("open", config.VPN_GUI_PATH).Run()
	if err != nil {
		logger.Errorf("launching Cisco Secure Client GUI: %v", err)
	}
	logger.Infof("Cisco GUI launched")
}

func KillCiscoProcesses() error {
	processes := []string{"Cisco Secure Client"} // Do NOT include vpnagentd
	var killErrors []string

	for _, name := range processes {
		cmd := exec.Command("pgrep", "-f", name)
		output, err := cmd.Output()
		if err != nil {
			logger.Warningf(fmt.Sprintf("failed to find process: %s", name), err)
			continue
		}

		pids := strings.Fields(string(output))
		for _, pid := range pids {
			killCmd := exec.Command("kill", "-9", pid)
			if err := killCmd.Run(); err != nil {
				msg := fmt.Sprintf("failed to kill PID %v for %v: %v", pid, name, err)
				logger.Errorf(msg, err)
				killErrors = append(killErrors, msg)
			} else {
				logger.Infof(fmt.Sprintf("Killed %v (PID %v)", name, pid))
			}
		}
	}

	time.Sleep(2 * time.Second)

	if len(killErrors) > 0 {
		return fmt.Errorf("not all processes could be killed:\n%v", strings.Join(killErrors, "\n"))
	}
	return nil
}

// Connect establishes a VPN connection using the specified profile.
// It reads the VPN profile script from a hidden path, replaces placeholders with credentials,
// and executes the VPN command with the provided script.
// It checks the current VPN connection status before attempting to connect.
// If the VPN is already connected, it aborts the connection operation.
// It also reads the credentials for the specified profile from a hidden file.
func Connect(credential *model.CREDENTIAL_FOR_LOGIN, profile string) {
	connectWithRetries(credential, profile, config.VPN_CONNECTION_RETRY_COUNT)
}

// connectWithRetries attempts to connect to the VPN with retries.
// It checks the current VPN connection status, reads the profile script, and executes the VPN command.
// If the VPN is already connected to a different profile, it disconnects first.
// It retries the connection if it detects a Cisco VPN agent lock.
// It uses a recursive approach to retry the connection up to a maximum number of retries.
// This function is useful for establishing a VPN connection with error handling and retry logic.
// It reads the credentials for the specified profile from a hidden file.
// It also handles the case where the VPN is already connected to a different profile.
func connectWithRetries(credential *model.CREDENTIAL_FOR_LOGIN, profile string, retryCount int) { // Removed 'credential *model.CREDENTIAL_FOR_LOGIN'
	logger.Infof(fmt.Sprintf("Initiating VPN connection using profile: %v", profile))

	profilePath := getProfilePath(profile)
	if profilePath == "" {
		logger.Infof("Unknown VPN profile: %v", profile)
		return
	}

	// implement keychain here and get rid of credential files
	// username, password, y_flag, secondPassword, err := readCredentials(profile)
	// if err != nil {
	// 	logger.Fatalf("reading credentials for profile %v: %v", profile, err)
	// 	return
	// }

	logger.Infof("Checking current VPN connection status...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, config.VPN_BINARY_PATH, "status")
	output, _ := cmd.CombinedOutput()

	if contains(string(output), "Connected") {
		last, err := middleware.GetLastConnectedProfile()
		if err != nil {
			logger.Warningf("retrieve error: %v", err)
		}
		logger.Infof("Last connected VPN profile: %v", last)

		if last == profile {
			logger.Infof(fmt.Sprintf("VPN already connected to profile: %v. Aborting connect operation.", profile))
			return
		}
		logger.Infof(fmt.Sprintf("VPN connected to profile %v, switching to %v...", last, profile))
		Disconnect()
	}

	if err := KillCiscoProcesses(); err != nil {
		logger.Errorf("failed to kill Cisco processes before reconnect: %v", err)
	}

	// Deprecated: Reading profile script from file ---
	// logger.Infof("reading VPN profile script from (hidden path)")
	// scriptContent, err := os.ReadFile(profilePath)
	// if err != nil {
	// logger.Errorf("reading VPN profile: %v", err)
	// return
	// }

	// username
	// password
	// Y/N for second factor prompt
	// second_password (if applicable)
	var scriptBuilder strings.Builder
	scriptBuilder.WriteString(credential.Username + "\n")
	scriptBuilder.WriteString(credential.Password + "\n")
	scriptBuilder.WriteString(credential.YFlag + "\n")
	if profile == "dev" {
		scriptBuilder.WriteString(credential.Push + "\n")
	}
	script := scriptBuilder.String()

	// Create a temporary file for the script input
	tempFile, err := os.CreateTemp("", "vpn_input_*.txt")
	if err != nil {
		logger.Errorf("creating temp VPN input file: %v", err)
		return
	}
	tempScript := tempFile.Name()

	_, err = tempFile.WriteString(script)
	if err != nil {
		tempFile.Close()
		os.Remove(tempScript)
		logger.Errorf("writing to temp VPN input file: %v", err)
		return
	}
	tempFile.Close()

	defer os.Remove(tempScript)

	logger.Infof("Running VPN command with provided script")

	vpnProfile := ""
	switch profile {
	case "dev":
		vpnProfile = "DEV-VPN-REMOTE"
	case "intra":
		vpnProfile = "INTRA"
	}

	cmd = exec.Command(config.VPN_BINARY_PATH, "connect", vpnProfile, "-s")

	stdinFile, err := os.Open(tempScript)
	if err != nil {
		logger.Errorf("opening temp script: %v", err)
		return
	}
	defer stdinFile.Close()
	defer os.Remove(tempScript)

	cmd.Stdin = stdinFile

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Errorf("getting stdout pipe: %s", err)
		return
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Errorf("getting stderr pipe: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		logger.Errorf("starting VPN command: %v", err)
		return
	}

	stdoutLines := make(chan string, 100)
	done := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("[VPN stdout] " + line)
			stdoutLines <- line
		}
		close(done)
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			logger.Errorf("VPN stderr: %v", errors.New(scanner.Text()))
		}
	}()

	if err := cmd.Wait(); err != nil {
		logger.Errorf("VPN command exited with error: %v", err)
	}

	<-done
	close(stdoutLines)

	shouldRetry := false
	for line := range stdoutLines {
		if strings.Contains(line, "Connect capability is unavailable") {
			logger.Infof("Detected Cisco VPN agent lock. Please manually restart Cisco Secure Client (AnyConnect) and try again.")
			shouldRetry = false
			break
		}
	}

	if shouldRetry && retryCount < config.VPN_CONNECTION_RETRY_COUNT {
		time.Sleep(2 * time.Second)
		logger.Infof(fmt.Sprintf("Retrying VPN connection to profile: %v (attempt %d)", profile, retryCount+1))
		connectWithRetries(credential, profile, retryCount+1)
		return
	}

	go func() {
		if err := middleware.SetLastConnectedProfile(profile); err != nil {
			logger.Errorf("store error: %v", err)
		}
	}()

	LaunchGUI()
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
// Deprecated: This function is depricated as migrated the credentialk management to keyring
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

	logger.Infof(fmt.Sprintf("Reading credentials for %v (hidden path)", profile))
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

	return "", "", "", "", fmt.Errorf("invalid credentials format for profile: %v", profile)
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
