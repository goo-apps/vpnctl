// Author: rohan.das

// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 goo-apps (rohan.das1203@gmail.com)
// Licensed under the MIT License. See LICENSE file for details.

package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	// "time"

	"github.com/goo-apps/vpnctl/config"
	"github.com/rs/zerolog"
)

var (
	logFile *os.File
	log     *zerolog.Logger
	// logFileSubPath = "/go_vpn/application.log"
)

// InitLogger sets up zerolog for file and/or console output
func InitLogger(logToFile bool, logFilePath string) {
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000" // initialise timeformat
	var writers []io.Writer

	consoleWriter := zerolog.ConsoleWriter{
		Out:          os.Stderr,
		TimeFormat:   "2006-01-02 15:04:05.000", // Full date + time with milliseconds
		TimeLocation: time.Local,                // local timezone
		NoColor:      false,
	}

	// write to console without level may be?

	writers = append(writers, consoleWriter)

	if logToFile {
		if logFilePath == "" {
			home, _ := os.UserHomeDir()
			logFilePath = filepath.Join(home, ".vpnctl", "application.log") // when production release change the go_vpn dir to .vpnctl

		}
		_ = os.MkdirAll(filepath.Dir(logFilePath), 0755)
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if err == nil {
			writers = append(writers, file)
		}
	}

	multi := zerolog.MultiLevelWriter(writers...)
	logger := zerolog.New(multi).With().Timestamp().Str("module", "vpnctl").Logger()
	log = &logger
}

// Shutdown closes the log file cleanly
func Shutdown() {
	if logFile != nil {
		_ = logFile.Close()
	}
}

// callerInfo retrieves the file name and line number of the caller
// func callerInfo() string {
// 	_, file, line, ok := runtime.Caller(2)
// 	if !ok {
// 		return "unknown:0"
// 	}
// 	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
// }

func callerInfo() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown:0"
	}

	// supported logger level, 1 and 2
	// level 1 - absolute logger and 2
	// level 2 - relative logger
	switch config.LOGGER_LEVEL {
	case 1:
		file = filepath.Base(file) // Get only the base file name
	case 2: // Try to get relative path from current working directory
		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, file); err == nil {
				file = rel
			}
		}
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// logging functions
func Debugf(format string, args ...interface{}) {
	log.Debug().
		Str("caller", callerInfo()).
		Msgf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Info().
		Str("caller", callerInfo()).
		Msgf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Error().
		Str("caller", callerInfo()).
		Msgf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	log.Warn().
		Str("caller", callerInfo()).
		Msgf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatal().
		Str("caller", callerInfo()).
		Msgf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	log.Panic().
		Str("caller", callerInfo()).
		Msgf(format, args...)
}

func Tracef(format string, args ...interface{}) {
	log.Trace().
		Str("caller", callerInfo()).
		Msgf(format, args...)
}

// Klog is not used in this version, but if needed, it can be implemented as follows:

// func InitLogger(logToFile bool, verbosity int, logFilePath string) {
// 	klog.InitFlags(nil)

// 	// Set verbosity
// 	_ = flag.Set("v", fmt.Sprintf("%d", verbosity))

// 	if logToFile {
// 		if logFilePath == "" {
// 			home, err := os.UserHomeDir()
// 			if err != nil {
// 				fmt.Printf("Failed to get home dir: %v\n", err)
// 				os.Exit(1)
// 			}
// 			logFilePath = filepath.Join(home, "go_vpn", "vpnctl.log")
// 			_ = os.MkdirAll(filepath.Dir(logFilePath), 0755)
// 		}

// 		// Proper klog flag to write to file
// 		_ = flag.Set("logtostderr", "false")
// 		_ = flag.Set("alsologtostderr", "false")
// 		_ = flag.Set("log_file", logFilePath)
// 	} else {
// 		// Only log to stderr
// 		_ = flag.Set("logtostderr", "true")
// 		_ = flag.Set("alsologtostderr", "false")
// 	}

// 	// Must be called after setting flags
// 	flag.Parse()
// }

// func Shutdown() {
// 	klog.Flush()
// }

// func Infof(format string, args ...interface{}) {
// 	klog.Infof(format, args...)
// }

// func Warningf(format string, args ...interface{}) {
// 	klog.Warningf(format, args...)
// }

// func Errorf(format string, args ...interface{}) {
// 	klog.Errorf(format, args...)
// }

// func Fatalf(format string, args ...interface{}) {
// 	klog.Fatalf(format, args...)
// }
