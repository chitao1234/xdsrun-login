package core

import (
	"net/http"
	"time"
)

var client *http.Client

func init() {
	client = &http.Client{
		Timeout: 5 * time.Second,
	}
}

func PerformLoginWithDefaultClient(username, password, domain string) {
	PerformLogin(client, username, password, domain)
}

func CheckStatusWithDefaultClient() {
	CheckStatus(client)
}