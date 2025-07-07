package vpnctl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
	"testing"

	"github.com/goo-apps/vpnctl/internal/middleware"
	"github.com/goo-apps/vpnctl/internal/model"
	"github.com/goo-apps/vpnctl/logger"
	"github.com/stretchr/testify/assert"
)

// --- Mocks and helpers ---
var (
	getLastConnectedProfile = middleware.GetLastConnectedProfile
	setLastConnectedProfile = middleware.SetLastConnectedProfile
)

// Mock logger
var (
	logInfoMsgs  []string
	logErrorMsgs []string
	logWarnMsgs  []string
	infof        = logger.Infof
	errorf       = logger.Errorf
	warningf     = logger.Warningf
)

func TestMain(m *testing.M) {
    infof = mockInfof
    errorf = mockErrorf
    warningf = mockWarningf
    os.Exit(m.Run())
}

func resetLogger() {
	logInfoMsgs = nil
	logErrorMsgs = nil
	logWarnMsgs = nil
}

func mockInfof(format string, args ...interface{}) {
	logInfoMsgs = append(logInfoMsgs, fmt.Sprintf(format, args...))
}
func mockErrorf(format string, args ...interface{}) {
	logErrorMsgs = append(logErrorMsgs, fmt.Sprintf(format, args...))
}
func mockWarningf(format string, args ...interface{}) {
	logWarnMsgs = append(logWarnMsgs, fmt.Sprintf(format, args...))
}

// Mock middleware
var (
	mockLastProfile   string
	mockSetProfileErr error
	mockGetProfileErr error
)

func mockGetLastConnectedProfile() (string, error) {
	return mockLastProfile, mockGetProfileErr
}
func mockSetLastConnectedProfile(profile string) error {
	mockLastProfile = profile
	return mockSetProfileErr
}

// Mock exec.Command
type fakeCmd struct {
	stdout    io.Reader
	stderr    io.Reader
	startErr  error
	waitErr   error
	output    []byte
	outputErr error
}

func (c *fakeCmd) StdoutPipe() (io.ReadCloser, error) {
	if c.stdout == nil {
		return io.NopCloser(strings.NewReader("")), nil
	}
	return io.NopCloser(c.stdout), nil
}
func (c *fakeCmd) StderrPipe() (io.ReadCloser, error) {
	if c.stderr == nil {
		return io.NopCloser(strings.NewReader("")), nil
	}
	return io.NopCloser(c.stderr), nil
}
func (c *fakeCmd) Start() error                    { return c.startErr }
func (c *fakeCmd) Wait() error                     { return c.waitErr }
func (c *fakeCmd) CombinedOutput() ([]byte, error) { return c.output, c.outputErr }
func (c *fakeCmd) Run() error                      { return nil }
func (c *fakeCmd) Output() ([]byte, error)         { return c.output, c.outputErr }

var (
	execCommandContextFunc = func(ctx interface{}, name string, arg ...string) *fakeCmd { return &fakeCmd{} }
	execCommandFunc        = func(name string, arg ...string) *fakeCmd { return &fakeCmd{} }
)

// Patch exec.Command and exec.CommandContext
func patchExecCommand() func() {
	origCmd := execCommandFunc
	origCmdCtx := execCommandContextFunc
	execCommandFunc = func(name string, arg ...string) *fakeCmd {
		return &fakeCmd{}
	}
	execCommandContextFunc = func(ctx interface{}, name string, arg ...string) *fakeCmd {
		return &fakeCmd{}
	}
	return func() {
		execCommandFunc = origCmd
		execCommandContextFunc = origCmdCtx
	}
}

// Patch os.ReadFile, os.WriteFile, os.Open, os.Remove
var (
	readFileFunc  = os.ReadFile
	writeFileFunc = os.WriteFile
	openFileFunc  = os.Open
	removeFunc    = os.Remove
)

func patchFileOps() func() {
	origRead := readFileFunc
	origWrite := writeFileFunc
	origOpen := openFileFunc
	origRemove := removeFunc
	return func() {
		readFileFunc = origRead
		writeFileFunc = origWrite
		openFileFunc = origOpen
		removeFunc = origRemove
	}
}

// Patch user.Current
var userCurrentFunc = user.Current

func patchUserCurrent() func() {
	orig := userCurrentFunc
	return func() { userCurrentFunc = orig }
}

// Patch KillCiscoProcesses, Disconnect, LaunchGUI
var (
	killCiscoProcessesFunc = KillCiscoProcesses
	disconnectFunc         = Disconnect
	launchGUIFunc          = LaunchGUI
)

func patchHelpers() func() {
	origKill := killCiscoProcessesFunc
	origDisc := disconnectFunc
	origLaunch := launchGUIFunc
	return func() {
		killCiscoProcessesFunc = origKill
		disconnectFunc = origDisc
		launchGUIFunc = origLaunch
	}
}

// --- Test cases ---

func TestGetProfilePath(t *testing.T) {
	restore := patchUserCurrent()
	defer restore()
	userCurrentFunc = func() (*user.User, error) {
		return &user.User{HomeDir: "/home/test"}, nil
	}
	assert.Contains(t, getProfilePath("intra"), ".credential_intra")
	assert.Contains(t, getProfilePath("dev"), ".credential_dev")
	assert.Equal(t, "", getProfilePath("unknown"))
}

func TestContains(t *testing.T) {
	assert.True(t, contains("hello world", "world"))
	assert.False(t, contains("hello world", "foo"))
}

func TestReadCredentials(t *testing.T) {
	restore := patchFileOps()
	defer restore()
	readFileFunc = func(path string) ([]byte, error) {
		if strings.Contains(path, "intra") {
			return []byte("user\npass\nyflag\n"), nil
		}
		if strings.Contains(path, "dev") {
			return []byte("user\npass\nyflag\npush\n"), nil
		}
		return nil, errors.New("not found")
	}
	u, p, y, s, err := readCredentials("intra")
	assert.NoError(t, err)
	assert.Equal(t, "user", u)
	assert.Equal(t, "pass", p)
	assert.Equal(t, "yflag", y)
	assert.Equal(t, "", s)

	u, p, y, s, err = readCredentials("dev")
	assert.NoError(t, err)
	assert.Equal(t, "user", u)
	assert.Equal(t, "pass", p)
	assert.Equal(t, "yflag", y)
	assert.Equal(t, "push", s)

	_, _, _, _, err = readCredentials("unknown")
	assert.Error(t, err)
}

func TestConnectWithRetries_ProfileNotFound(t *testing.T) {
	resetLogger()
	connectWithRetries(&model.CREDENTIAL_FOR_LOGIN{}, "unknown", 0)
	assert.Contains(t, logInfoMsgs[len(logInfoMsgs)-1], "Unknown VPN profile")
}

func TestConnectWithRetries_AlreadyConnectedSameProfile(t *testing.T) {
	resetLogger()
	restore := patchHelpers()
	defer restore()
	killCiscoProcessesFunc = func() error { return nil }
	disconnectFunc = func() {}
	launchGUIFunc = func() {}

	mockLastProfile = "dev"
	mockGetProfileErr = nil

	// Patch exec.CommandContext to simulate "Connected"
	execCommandContextFunc = func(ctx interface{}, name string, arg ...string) *fakeCmd {
		return &fakeCmd{output: []byte("Connected")}
	}
	// Patch middleware
	getLastConnectedProfile = mockGetLastConnectedProfile

	connectWithRetries(&model.CREDENTIAL_FOR_LOGIN{}, "dev", 0)
	assert.Contains(t, logInfoMsgs[len(logInfoMsgs)-1], "VPN already connected to profile")
}

func TestConnectWithRetries_AlreadyConnectedDifferentProfile(t *testing.T) {
	resetLogger()
	restore := patchHelpers()
	defer restore()
	killCiscoProcessesFunc = func() error { return nil }
	disconnectCalled := false
	disconnectFunc = func() { disconnectCalled = true }
	launchGUIFunc = func() {}

	mockLastProfile = "intra"
	mockGetProfileErr = nil

	execCommandContextFunc = func(ctx interface{}, name string, arg ...string) *fakeCmd {
		return &fakeCmd{output: []byte("Connected")}
	}
	getLastConnectedProfile = mockGetLastConnectedProfile

	connectWithRetries(&model.CREDENTIAL_FOR_LOGIN{}, "dev", 0)
	assert.True(t, disconnectCalled)
}

func TestConnectWithRetries_KillCiscoProcessesFails(t *testing.T) {
	resetLogger()
	restore := patchHelpers()
	defer restore()
	killCiscoProcessesFunc = func() error { return errors.New("fail") }
	disconnectFunc = func() {}
	launchGUIFunc = func() {}

	mockLastProfile = ""
	mockGetProfileErr = nil

	execCommandContextFunc = func(ctx interface{}, name string, arg ...string) *fakeCmd {
		return &fakeCmd{output: []byte("Not Connected")}
	}
	getLastConnectedProfile = mockGetLastConnectedProfile

	// Patch os.ReadFile to fail
	restoreFiles := patchFileOps()
	defer restoreFiles()
	readFileFunc = func(path string) ([]byte, error) { return nil, errors.New("fail") }

	connectWithRetries(&model.CREDENTIAL_FOR_LOGIN{}, "dev", 0)
	assert.Contains(t, logErrorMsgs[len(logErrorMsgs)-1], "reading VPN profile")
}

func TestConnectWithRetries_HappyPath(t *testing.T) {
	resetLogger()
	restore := patchHelpers()
	defer restore()
	killCiscoProcessesFunc = func() error { return nil }
	disconnectFunc = func() {}
	launchCalled := false
	launchGUIFunc = func() { launchCalled = true }

	mockLastProfile = ""
	mockGetProfileErr = nil
	mockSetProfileErr = nil

	execCommandContextFunc = func(ctx interface{}, name string, arg ...string) *fakeCmd {
		return &fakeCmd{output: []byte("Not Connected")}
	}
	getLastConnectedProfile = mockGetLastConnectedProfile
	setLastConnectedProfile = mockSetLastConnectedProfile

	restoreFiles := patchFileOps()
	defer restoreFiles()
	readFileFunc = func(path string) ([]byte, error) {
		return []byte("script with {{USERNAME}} and {{PASSWORD}} and {{Y}}"), nil
	}
	writeFileFunc = func(path string, data []byte, perm os.FileMode) error { return nil }
	openFileFunc = func(path string) (*os.File, error) {
		return os.NewFile(0, os.DevNull), nil
	}
	removeFunc = func(path string) error { return nil }

	// Patch exec.Command for connect
	execCommandFunc = func(name string, arg ...string) *fakeCmd {
		return &fakeCmd{
			stdout:   bytes.NewBufferString("All good\n"),
			stderr:   bytes.NewBufferString(""),
			startErr: nil,
			waitErr:  nil,
		}
	}

	cred := &model.CREDENTIAL_FOR_LOGIN{
		Username: "user",
		Password: "pass",
		YFlag:    "yflag",
		Push:     "push",
	}
	connectWithRetries(cred, "dev", 0)
	assert.True(t, launchCalled)
}

// Add more tests for error branches as needed...
