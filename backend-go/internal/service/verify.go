package service

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math/big"
	"net/smtp"
	"os"
	"stock-monitor/pkg/logger"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type verifyEntry struct {
	Code      string
	ExpiresAt time.Time
}

var (
	verifyCache = make(map[string]*verifyEntry)
	verifyMu    sync.RWMutex
)

func GenerateVerifyCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func SaveVerifyCode(email, code string) {
	verifyMu.Lock()
	defer verifyMu.Unlock()
	verifyCache[email] = &verifyEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
}

func CheckVerifyCode(email, code string) bool {
	verifyMu.Lock()
	defer verifyMu.Unlock()
	entry, ok := verifyCache[email]
	if !ok {
		return false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(verifyCache, email)
		return false
	}
	if !strings.EqualFold(entry.Code, code) {
		return false
	}
	delete(verifyCache, email)
	return true
}

func cleanupExpiredCodes() {
	for {
		time.Sleep(10 * time.Minute)
		verifyMu.Lock()
		now := time.Now()
		for email, entry := range verifyCache {
			if now.After(entry.ExpiresAt) {
				delete(verifyCache, email)
			}
		}
		verifyMu.Unlock()
	}
}

func init() {
	go cleanupExpiredCodes()
}

func SendVerifyEmail(toEmail, code string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	if host == "" || port == "" || user == "" || pass == "" {
		logger.Log.Info("SMTP not configured, printing code to console instead",
			zap.String("email", toEmail),
			zap.String("code", code),
		)
		return nil
	}

	portNum, _ := strconv.Atoi(port)

	subject := "投资助手 - 登录验证码"
	body := fmt.Sprintf("您的验证码是：%s\n5分钟内有效，请勿泄露给他人。", code)
	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)

	// 465 uses SSL/TLS; 587 uses STARTTLS; 25 uses plain text
	if portNum == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, InsecureSkipVerify: false})
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return err
		}
		defer client.Close()

		if err := client.Auth(auth); err != nil {
			return err
		}
		if err := client.Mail(user); err != nil {
			return err
		}
		if err := client.Rcpt(toEmail); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(msg)
		if err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, user, []string{toEmail}, msg)
}
