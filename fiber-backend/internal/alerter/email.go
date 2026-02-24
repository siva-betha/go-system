package alerter

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strings"
	"time"
)

func (m *StorageMonitor) startEmailSender() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case email := <-m.emailChan:
				m.sendWithRetry(email)
			case <-m.stopChan:
				return
			}
		}
	}()
}

func (m *StorageMonitor) sendWithRetry(email EmailMessage) {
	for i := 0; i < 3; i++ {
		err := m.sendEmail(email)
		if err == nil {
			log.Printf("Alert email sent successfully to %v", email.To)
			return
		}
		log.Printf("Failed to send email (attempt %d/3): %v", i+1, err)
		if i < 2 {
			time.Sleep(time.Duration(i+1) * 5 * time.Second)
		}
	}
}

func (m *StorageMonitor) sendEmail(email EmailMessage) error {
	addr := fmt.Sprintf("%s:%d", m.config.Email.SMTPHost, m.config.Email.SMTPPort)

	var auth smtp.Auth
	if m.config.Email.Username != "" {
		auth = smtp.PlainAuth("", m.config.Email.Username, m.config.Email.Password, m.config.Email.SMTPHost)
	}

	header := make(map[string]string)
	header["From"] = m.config.Email.From
	header["To"] = strings.Join(email.To, ",")
	header["Subject"] = email.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""

	var message string
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + email.Body

	return smtp.SendMail(addr, auth, m.config.Email.From, email.To, []byte(message))
}

func (m *StorageMonitor) prepareEmail(alert StorageAlert) {
	temp := m.getEmailTemplate(alert.Level)

	data := struct {
		Alert           StorageAlert
		Hostname        string
		Time            string
		Recommendations []string
	}{
		Alert:           alert,
		Hostname:        m.hostname,
		Time:            alert.Timestamp.Format("2006-01-02 15:04:05"),
		Recommendations: m.getRecommendations(alert),
	}

	var body bytes.Buffer
	if err := temp.Execute(&body, data); err != nil {
		log.Printf("Failed to execute email template: %v", err)
		return
	}

	email := EmailMessage{
		To:       m.config.Email.To,
		Subject:  fmt.Sprintf("[%s] Storage Alert: %s on %s", strings.ToUpper(alert.Level), alert.Component, m.hostname),
		Body:     body.String(),
		Priority: 1,
	}

	select {
	case m.emailChan <- email:
	default:
		log.Printf("Email channel full, dropping alert notification")
	}
}

func (m *StorageMonitor) getEmailTemplate(level string) *template.Template {
	templates := map[string]string{
		"warning": `⚠️ STORAGE WARNING ⚠️
Component: {{.Alert.Component}}
Usage: {{printf "%.1f" .Alert.UsedPercent}}%
Free Space: {{.Alert.FreeBytes}} bytes
Host: {{.Hostname}}
Time: {{.Time}}

Recommendations:
{{range .Recommendations}}- {{.}}
{{end}}`,
		"critical": `🔴 CRITICAL STORAGE ALERT 🔴
Component: {{.Alert.Component}}
Usage: {{printf "%.1f" .Alert.UsedPercent}}%
Free Space: {{.Alert.FreeBytes}} bytes
Host: {{.Hostname}}
Time: {{.Time}}

IMMEDIATE ACTION REQUIRED:
{{range .Recommendations}}- {{.}}
{{end}}`,
		"emergency": `🚨 EMERGENCY STORAGE ALERT 🚨
Component: {{.Alert.Component}}
Usage: {{printf "%.1f" .Alert.UsedPercent}}%
Free Space: {{.Alert.FreeBytes}} bytes
Host: {{.Hostname}}
Time: {{.Time}}

SYSTEM ACTION TAKEN:
{{range .Recommendations}}- {{.}}
{{end}}`,
	}
	return template.Must(template.New("email").Parse(templates[level]))
}

func (m *StorageMonitor) getRecommendations(alert StorageAlert) []string {
	switch alert.Level {
	case "warning":
		return []string{"Review data retention policy", "Plan disk expansion", "Monitor ingestion rates"}
	case "critical":
		return []string{"IMMEDIATE: Clean old logs", "Expand disk partition", "Audit database for high growth"}
	case "emergency":
		return []string{"Automatic cleanup triggered", "MANUAL INTERVENTION REQUIRED", "System at risk of crash"}
	}
	return []string{"Check system health"}
}
