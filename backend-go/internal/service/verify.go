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

type emailSendLimit struct {
	LastSent  time.Time
	DayCount  int
	DayString string
}

var (
	emailLimits  = make(map[string]*emailSendLimit)
	emailLimitMu sync.RWMutex
)

func CheckEmailSendLimit(email string) error {
	emailLimitMu.Lock()
	defer emailLimitMu.Unlock()

	now := time.Now()
	today := now.Format("2006-01-02")

	limit, ok := emailLimits[email]
	if !ok {
		return nil
	}

	if now.Sub(limit.LastSent) < 5*time.Minute {
		remain := int(300 - now.Sub(limit.LastSent).Seconds())
		return fmt.Errorf("发送过于频繁，请%d秒后再试", remain)
	}

	if limit.DayString == today && limit.DayCount >= 2 {
		return fmt.Errorf("今日发送次数已达上限，请明天再试")
	}

	return nil
}

func RecordEmailSent(email string) {
	emailLimitMu.Lock()
	defer emailLimitMu.Unlock()

	now := time.Now()
	today := now.Format("2006-01-02")

	limit, ok := emailLimits[email]
	if !ok || limit.DayString != today {
		emailLimits[email] = &emailSendLimit{
			LastSent:  now,
			DayCount:  1,
			DayString: today,
		}
		return
	}

	limit.LastSent = now
	limit.DayCount++
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
	msg := []byte("From: " + user + "\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)

	logger.Log.Info("sending verify email",
		zap.String("from", user),
		zap.String("to", toEmail),
		zap.String("host", host),
		zap.String("port", port),
	)

	// 465 uses SSL/TLS; 587 uses STARTTLS; 25 uses plain text
	if portNum == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, InsecureSkipVerify: false})
		if err != nil {
			logger.Log.Error("SMTP TLS dial failed", zap.Error(err))
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			logger.Log.Error("SMTP new client failed", zap.Error(err))
			return err
		}
		defer client.Close()

		if err := client.Auth(auth); err != nil {
			logger.Log.Error("SMTP auth failed", zap.Error(err))
			return err
		}
		if err := client.Mail(user); err != nil {
			logger.Log.Error("SMTP mail from failed", zap.Error(err))
			return err
		}
		if err := client.Rcpt(toEmail); err != nil {
			logger.Log.Error("SMTP rcpt to failed", zap.Error(err))
			return err
		}
		w, err := client.Data()
		if err != nil {
			logger.Log.Error("SMTP data failed", zap.Error(err))
			return err
		}
		_, err = w.Write(msg)
		if err != nil {
			logger.Log.Error("SMTP write failed", zap.Error(err))
			return err
		}
		if err := w.Close(); err != nil {
			logger.Log.Error("SMTP close failed", zap.Error(err))
			return err
		}
		logger.Log.Info("SMTP email sent successfully", zap.String("to", toEmail))
		return nil
	}

	if err := smtp.SendMail(addr, auth, user, []string{toEmail}, msg); err != nil {
		logger.Log.Error("SMTP SendMail failed", zap.Error(err))
		return err
	}
	logger.Log.Info("SMTP email sent successfully", zap.String("to", toEmail))
	return nil
}
