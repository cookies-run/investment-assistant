package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/smtp"
	"os"
	"stock-monitor/pkg/logger"
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

	subject := "投资助手 - 登录验证码"
	body := fmt.Sprintf("您的验证码是：%s\n5分钟内有效，请勿泄露给他人。", code)
	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(addr, auth, user, []string{toEmail}, msg)
}
