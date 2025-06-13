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

const maxRetries = 1

// Status checks the current VPN connection status using the Cisco Secure Client command line tool.
// It runs the command with a timeout to avoid hanging indefinitely.
// If the command times out, it logs an error and returns.
// If the command fails, it logs the error and prints the output.
// If the command succeeds, it logs the success and prints the output.
// This function is useful for checking if the VPN is currently connected or disconnected.
func Status() {
	logger.Infof("Checking VPN status...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, vpnCmd, "status", "-s")
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		logger.Errorf("status check", fmt.Errorf("vpn status timed out"))
		return
	}

	if err != nil {
		logger.Errorf("retrieving VPN status", err)
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
	exec.Command(vpnCmd, "disconnect").Run()
	logger.Infof("VPN disconnected")
	exec.Command("pkill", "-x", "Cisco Secure Client").Run()
	logger.Infof("Cisco Secure Client process killed")

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
	logger.Infof("VPN disconnected and related processes killed")
	// killVPNProcesses() // <-- updated
}

func Disconnect() {
	logger.Infof("Attempting to disconnect VPN...")
	exec.Command(vpnCmd, "disconnect").Run()
	logger.Infof("VPN disconnected")
	exec.Command("pkill", "-x", "Cisco Secure Client").Run()
	logger.Infof("Cisco Secure Client process killed")

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
		logger.Errorf("getting VPN PIDs", err)
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
	err := exec.Command("open", guiApp).Run()
	if err != nil {
		logger.Errorf("launching Cisco Secure Client GUI", err)
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
			logger.Errorf(fmt.Sprintf("failed to find process: %v", name), err)
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

// WORKING VERSION
// func Connect(profile string) {
// 	logger.Infof(fmt.Sprintf("Initiating VPN connection using profile: %v", profile))

// 	profilePath := getProfilePath(profile)
// 	if profilePath == "" {
// 		logger.Infof("Unknown VPN profile")
// 		return
// 	}

// 	username, password, y, y2, err := readCredentials(profile)
// 	if err != nil {
// 		logger.Errorf(err, "reading credentials")
// 		return
// 	}

// 	logger.Infof("Checking current VPN connection status...")
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	cmd := exec.CommandContext(ctx, vpnCmd, "status")
// 	output, _ := cmd.CombinedOutput()

// 	// check the current vpn profile
// 	if contains(string(output), "Connected") {
// 		last, err := middleware.GetLastConnectedProfile()
// 		if err != nil {
// 			logger.Errorf(err, "retrieve error:")
// 		}
// 		logger.Infof("Last connected VPN profile: " + last)

// 		if last == profile {
// 			logger.Infof(fmt.Sprintf("VPN already connected to profile: %v. Aborting connect operation.", profile))
// 			return
// 		}
// 		logger.Infof(fmt.Sprintf("VPN connected to profile %v, switching to %v...", last, profile))
// 		Disconnect()
// 		KillCiscoProcesses()
// 	}

// 	logger.Infof(fmt.Sprintf("Reading VPN profile script from (hidden path) %v", ""))
// 	scriptContent, err := os.ReadFile(profilePath)
// 	if err != nil {
// 		logger.Errorf(err, "reading VPN profile")
// 		return
// 	}

// 	script := strings.ReplaceAll(string(scriptContent), "{{USERNAME}}", username)
// 	script = strings.ReplaceAll(script, "{{PASSWORD}}", password)
// 	script = strings.ReplaceAll(script, "{{Y}}", y)
// 	if profile == "dev" {
// 		script = strings.ReplaceAll(script, "{{Y2}}", y2)
// 	}

// 	tempScript := filepath.Join(os.TempDir(), "vpn_input.txt")
// 	err = os.WriteFile(tempScript, []byte(script), 0600)
// 	if err != nil {
// 		logger.Errorf(err, "writing temp VPN input file")
// 		return
// 	}

// 	logger.Infof("Running VPN command with provided script")

// 	vpnProfile := ""
// 	switch profile {
// 	case "dev":
// 		vpnProfile = "DEV-VPN-REMOTE"
// 	case "intra":
// 		vpnProfile = "INTRA"
// 	}

// 	cmd = exec.Command(vpnCmd, "connect", vpnProfile, "-s")

// 	stdinFile, err := os.Open(tempScript)
// 	if err != nil {
// 		logger.Errorf(err, "opening temp script")
// 		return
// 	}
// 	defer stdinFile.Close()
// 	defer os.Remove(tempScript)

// 	cmd.Stdin = stdinFile

// 	// Declare the stdout channels outside so they're accessible later
// 	stdoutLines := make(chan string, 100)
// 	done := make(chan struct{})

// 	stdoutPipe, err := cmd.StdoutPipe()
// 	if err != nil {
// 		logger.Errorf(err, "getting stdout pipe")
// 		return
// 	}
// 	stderrPipe, err := cmd.StderrPipe()
// 	if err != nil {
// 		logger.Errorf(err, "getting stderr pipe")
// 		return
// 	}

// 	if err := cmd.Start(); err != nil {
// 		logger.Errorf(err, "starting VPN command")
// 		return
// 	}

// 	// Stream stdout
// 	go func() {
// 		scanner := bufio.NewScanner(stdoutPipe)
// 		for scanner.Scan() {
// 			line := scanner.Text()
// 			fmt.Println("[VPN stdout] " + line)
// 			stdoutLines <- line
// 		}
// 		close(done)
// 	}()

// 	// Stream stderr
// 	go func() {
// 		scanner := bufio.NewScanner(stderrPipe)
// 		for scanner.Scan() {
// 			logger.Errorf(errors.New(scanner.Text()), "VPN stderr")
// 		}
// 	}()

// 	if err := cmd.Wait(); err != nil {
// 		logger.Errorf(err, "VPN command exited with error")
// 	}

// 	// Scan the output for Cisco lock issue
// 	<-done
// 	close(stdoutLines)

// 	shouldRetry := false
// 	for line := range stdoutLines {
// 		if strings.Contains(line, "Connect capability is unavailable") {
// 			logger.Infof("Detected Cisco VPN agent lock. Restarting Cisco services...")
// 			KillCiscoProcesses()
// 			shouldRetry = true
// 			break
// 		}
// 	}

// 	if shouldRetry {
// 		time.Sleep(2 * time.Second)
// 		logger.Infof("Retrying VPN connection to profile: " + profile)
// 		Connect(profile)
// 		return
// 	}

// 	if err := middleware.SetLastConnectedProfile(profile); err != nil {
// 		logger.Errorf(err, "store error:")
// 	}

// 	LaunchGUI()
// }

// Connect establishes a VPN connection using the specified profile.
// It reads the VPN profile script from a hidden path, replaces placeholders with credentials,
// and executes the VPN command with the provided script.
func Connect(profile string) {
	connectWithRetries(profile, 0)
}

// connectWithRetries attempts to connect to the VPN with retries.
// It checks the current VPN connection status, reads the profile script, and executes the VPN command.
// If the VPN is already connected to a different profile, it disconnects first.
// It retries the connection if it detects a Cisco VPN agent lock.
// It uses a recursive approach to retry the connection up to a maximum number of retries.
// This function is useful for establishing a VPN connection with error handling and retry logic.
// It reads the credentials for the specified profile from a hidden file.
// It also handles the case where the VPN is already connected to a different profile.
func connectWithRetries(profile string, retryCount int) {
	logger.Infof(fmt.Sprintf("Initiating VPN connection using profile: %v", profile))

	profilePath := getProfilePath(profile)
	if profilePath == "" {
		logger.Infof("Unknown VPN profile")
		return
	}

	// get username from env
	username_env := middleware.GetUsernameFromEnv()


	result, err := middleware.GetCredential(username_env)
	if err != nil {
		logger.Errorf("Error during fetching credential %v", err)
		logger.Warningf("If you haven't set your credential yet, please use the /set-credential to register your credential; Check help for more information!")
	}

	username := result.Username
	password := result.Password
	secondPassword := result.SecondPassword
	y_flag := result.YFlag

	// username, password, secondPassword, y_flag, err := readCredentials(profile)
	if err != nil {
		logger.Errorf("reading credentials", err)
		return
	}

	logger.Infof("Checking current VPN connection status...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, vpnCmd, "status")
	output, _ := cmd.CombinedOutput()

	if contains(string(output), "Connected") {
		last, err := middleware.GetLastConnectedProfile()
		if err != nil {
			logger.Errorf("retrieve error: %v", err)
		}
		logger.Infof("Last connected VPN profile: " + last)

		if last == profile {
			logger.Infof(fmt.Sprintf("VPN already connected to profile: %v. Aborting connect operation.", profile))
			return
		}
		logger.Infof(fmt.Sprintf("VPN connected to profile %v, switching to %v...", last, profile))
		Disconnect()
	}

	if err := KillCiscoProcesses(); err != nil {
		logger.Errorf("failed to kill Cisco processes before reconnect", err)
	}

	logger.Infof("Reading VPN profile script from (hidden path)")
	scriptContent, err := os.ReadFile(profilePath)
	if err != nil {
		logger.Errorf("reading VPN profile", err)
		return
	}

	script := strings.ReplaceAll(string(scriptContent), "{{USERNAME}}", username)
	script = strings.ReplaceAll(script, "{{PASSWORD}}", password)
	script = strings.ReplaceAll(script, "{{Y}}", y_flag)
	if profile == "dev" {
		script = strings.ReplaceAll(script, "{{SECOND_PASSWORD}}", secondPassword)
	}

	tempScript := filepath.Join(os.TempDir(), "vpn_input.txt")
	err = os.WriteFile(tempScript, []byte(script), 0600)
	if err != nil {
		logger.Errorf("writing temp VPN input file", err)
		return
	}
	defer os.Remove(tempScript)

	logger.Infof("Running VPN command with provided script")

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
		logger.Errorf("opening temp script", err)
		return
	}
	defer stdinFile.Close()
	defer os.Remove(tempScript)

	cmd.Stdin = stdinFile

	// Capture stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Errorf("getting stdout pipe", err)
		return
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Errorf("getting stderr pipe", err)
		return
	}

	// Start command before scanning
	if err := cmd.Start(); err != nil {
		logger.Errorf("starting VPN command", err)
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
			logger.Errorf("VPN stderr", errors.New(scanner.Text()))
		}
	}()

	if err := cmd.Wait(); err != nil {
		logger.Errorf("VPN command exited with error", err)
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

	if shouldRetry && retryCount < maxRetries {
		time.Sleep(2 * time.Second)
		logger.Infof(fmt.Sprintf("Retrying VPN connection to profile: %v (attempt %d)", profile, retryCount+1))
		connectWithRetries(profile, retryCount+1)
		return
	}

	go func ()  {
		if err := middleware.SetLastConnectedProfile(profile); err != nil {
			logger.Errorf("store error:", err)
		}
	}() 

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
