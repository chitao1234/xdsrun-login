package core

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var client *http.Client

func init() {
	client = &http.Client{
		Timeout: 5 * time.Second,
	}
}

// Logger forwards log messages from Go to the Android UI.
type Logger interface {
	Log(message string)
}

var uiLogger Logger

// SetLogger sets the UI logger callback.
func SetLogger(l Logger) {
	uiLogger = l
}

// uiLogf sends formatted logs to the UI (if set) and to Android logcat.
func uiLogf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if uiLogger != nil {
		uiLogger.Log(msg)
	}
	log.Printf("%s", msg)
}

// uiLogln sends a single-line log to the UI (if set) and to Android logcat.
func uiLogln(msg string) {
	if uiLogger != nil {
		uiLogger.Log(msg)
	}
	log.Println(msg)
}

func PerformLoginWithDefaultClient(username, password, domain string) {
	PerformLogin(client, username, password, domain)
}

func CheckStatusWithDefaultClient() bool {
	return CheckStatus(client)
}
