/*
 * 项目名称：blog-backend
 * 文件名称：client.go
 * 创建时间：2026-04-17 23:40:47
 *
 * 系统用户：Administrator
 * 作　　者：無以菱
 * 联系邮箱：huangjing510@126.com
 * 功能描述：邮件客户端，提供邮件发送功能和限流控制
 */
package email

import (
	"blog-backend/config"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"time"

	"gopkg.in/gomail.v2"
)

// Client 邮件客户端
type Client struct {
	config      *config.Config
	rateLimiter *RateLimiter
}

// RateLimiter 限流器（滑动时间窗口）
type RateLimiter struct {
	mu      sync.RWMutex
	records map[string][]time.Time
	limit   int
	window  time.Duration
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		records: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
}

// Allow 检查是否允许发送
func (r *RateLimiter) Allow(email string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	times := r.records[email]

	var validTimes []time.Time
	for _, t := range times {
		if now.Sub(t) < r.window {
			validTimes = append(validTimes, t)
		}
	}

	if len(validTimes) >= r.limit {
		return false
	}

	validTimes = append(validTimes, now)
	r.records[email] = validTimes
	return true
}

// Initialize 创建邮件客户端
func Initialize(conf *config.Config) *Client {
	if conf == nil || conf.Email.Host == "" {
		return nil
	}

	return &Client{
		config:      conf,
		rateLimiter: NewRateLimiter(5, time.Hour), // 5次/小时
	}
}

// SendWelcomeEmail 发送欢迎邮件
func (c *Client) SendWelcomeEmail(to, siteName, unsubscribeURL string) error {
	if !c.rateLimiter.Allow(to) {
		return fmt.Errorf("发送频率过高，请稍后再试")
	}

	if siteName == "" {
		siteName = "菱风叙"
	}

	subject := fmt.Sprintf("【%s】订阅成功", siteName)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background-color: #f5f5f5; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #0891b2 0%%, #06b6d4 100%%); color: white; padding: 40px 30px; text-align: center; border-radius: 12px 12px 0 0; box-shadow: 0 4px 12px rgba(8, 145, 178, 0.2); }
        .header h2 { margin: 0; font-size: 28px; font-weight: 700; }
        .content { background: white; padding: 40px 30px; border-radius: 0 0 12px 12px; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); }
        .content p { margin: 16px 0; font-size: 15px; color: #475569; }
        .content strong { color: #0891b2; font-weight: 600; }
        .highlight-box { background: rgba(8, 145, 178, 0.05); border-left: 4px solid #0891b2; padding: 16px; margin: 24px 0; border-radius: 4px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; color: #94a3b8; font-size: 13px; text-align: center; }
        .footer a { color: #0891b2; text-decoration: none; }
        .footer a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>🎉 订阅成功</h2>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>感谢您订阅 <strong>%s</strong>！</p>
            <div class="highlight-box">
                <p style="margin: 0;">✅ 您将会收到本站的最新文章推送</p>
            </div>
            <p>我们会在有新文章发布时第一时间通知您。</p>
        </div>
        <div class="footer">
            <p>如果这不是您本人的操作，或您想退订，请点击：<a href="%s">退订链接</a></p>
            <p>此邮件由系统自动发送，请勿直接回复</p>
        </div>
    </div>
</body>
</html>
`, siteName, unsubscribeURL)

	textBody := fmt.Sprintf(`订阅成功

感谢您订阅 %s！

您将会收到本站的最新文章推送。

我们会在有新文章发布时第一时间通知您。

如果这不是您本人的操作，或您想退订，请访问：%s

---
此邮件由系统自动发送，请勿直接回复`, siteName, unsubscribeURL)

	return c.sendEmail(to, subject, htmlBody, textBody)
}

// SendArticleNotification 发送文章推送邮件
func (c *Client) SendArticleNotification(to, siteName, title, summary, articleURL, unsubscribeURL string) error {
	if !c.rateLimiter.Allow(to) {
		return fmt.Errorf("发送频率过高，请稍后再试")
	}

	if siteName == "" {
		siteName = "菱风叙"
	}

	subject := fmt.Sprintf("【%s】新文章：%s", siteName, title)

	htmlTitle := template.HTMLEscapeString(title)
	htmlSummary := template.HTMLEscapeString(summary)
	htmlSummary = strings.ReplaceAll(htmlSummary, "\n", "<br>")

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background-color: #f5f5f5; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #0891b2 0%%, #06b6d4 100%%); color: white; padding: 40px 30px; text-align: center; border-radius: 12px 12px 0 0; box-shadow: 0 4px 12px rgba(8, 145, 178, 0.2); }
        .header h2 { margin: 0; font-size: 28px; font-weight: 700; }
        .content { background: white; padding: 40px 30px; border-radius: 0 0 12px 12px; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); }
        .article-title { font-size: 24px; color: #1a202c; margin: 0 0 20px 0; font-weight: 700; }
        .article-summary { color: #475569; line-height: 1.8; margin-bottom: 30px; font-size: 15px; }
        .btn { display: inline-block; background: linear-gradient(135deg, #0891b2 0%%, #06b6d4 100%%); color: white; padding: 14px 32px; text-decoration: none; border-radius: 8px; font-weight: 600; box-shadow: 0 4px 12px rgba(8, 145, 178, 0.3); }
        .btn:hover { box-shadow: 0 6px 16px rgba(8, 145, 178, 0.4); }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; color: #94a3b8; font-size: 13px; text-align: center; }
        .footer a { color: #0891b2; text-decoration: none; }
        .footer a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>📝 新文章发布</h2>
        </div>
        <div class="content">
            <h2 class="article-title">%s</h2>
            <p class="article-summary">%s</p>
            <div style="text-align: center;">
                <a href="%s" class="btn">阅读全文</a>
            </div>
        </div>
        <div class="footer">
            <p>这是来自 <strong>%s</strong> 的文章推送。</p>
            <p>如需退订，请点击：<a href="%s">退订链接</a></p>
        </div>
    </div>
</body>
</html>
`, htmlTitle, htmlSummary, articleURL, siteName, unsubscribeURL)

	textBody := fmt.Sprintf(`新文章：%s

%s

阅读全文：%s

---
这是来自 %s 的文章推送。
如需退订，请访问：%s`, title, summary, articleURL, siteName, unsubscribeURL)

	return c.sendEmail(to, subject, htmlBody, textBody)
}

// sendEmail 发送邮件
func (c *Client) sendEmail(to, subject, htmlBody, textBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(c.config.Email.Username, c.config.Email.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(c.config.Email.Host, c.config.Email.Port, c.config.Email.Username, c.config.Email.Password)

	return d.DialAndSend(m)
}

// GenerateToken 生成退订令牌
func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
