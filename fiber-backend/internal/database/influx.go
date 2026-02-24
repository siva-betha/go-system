package database

import (
	"fmt"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func ConnectInflux(url, token, username, password string) (influxdb2.Client, string, string) {
	url = strings.TrimSpace(url)
	token = strings.TrimSpace(token)
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	method := "token"
	if token == "" && username != "" && password != "" {
		token = fmt.Sprintf("%s:%s", username, password)
		method = "fallback_credentials"
	} else if token == "plc-monitoring-token-123456789" {
		method = "token_likely_placeholder"
	}

	masked := "empty"
	tLen := len(token)
	if tLen > 4 {
		masked = token[:2] + "..." + token[tLen-2:]
	} else if tLen > 0 {
		masked = "too_short"
	}

	client := influxdb2.NewClient(url, token)
	return client, method, fmt.Sprintf("%s (len:%d)", masked, tLen)
}
