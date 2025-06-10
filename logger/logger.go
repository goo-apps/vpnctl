// Author: Rohan.das
package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	logger          *log.Logger
	suppressConsole = false
)

func InitLogger() {
	logFile := filepath.Join(os.Getenv("HOME"), "go_vpn", "vpn_connect.log")
	_ = os.MkdirAll(filepath.Dir(logFile), 0755)
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	logger = log.New(file, "[vpnctl] ", log.LstdFlags|log.Lshortfile)
}

func LogInfo(msg string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, file, line, _ := runtime.Caller(1)
	logMsg := fmt.Sprintf("INFO [%s] %s:%d %s", timestamp, filepath.Base(file), line, msg)
	logger.Println(logMsg)
	fmt.Println(logMsg)
}

func LogError(err error, context string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, file, line, _ := runtime.Caller(1)
	logMsg := fmt.Sprintf("ERROR [%s] %s:%d %s: %v", timestamp, filepath.Base(file), line, context, err)
	logger.Println(logMsg)
	fmt.Println(logMsg)
}

func LogWarn(err error, context string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, file, line, _ := runtime.Caller(1)
	logMsg := fmt.Sprintf("WARN [%s] %s:%d %s: %v", timestamp, filepath.Base(file), line, context, err)
	logger.Println(logMsg)
	fmt.Println(logMsg)
}

func SetSuppressConsole(suppress bool) {
	suppressConsole = suppress
}
