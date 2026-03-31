package service

import (
	"errors"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type MailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewMailService(host, port, username, password, from string) *MailService {
	return &MailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *MailService) SendVerificationCode(to, code string) error {
	subject := "【西外校园】邮箱验证码"
	body := strings.ReplaceAll(`
		<div style="max-width: 600px; margin: 0 auto; padding: 20px; font-family: 'Microsoft YaHei', sans-serif;">
			<div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; border-radius: 10px 10px 0 0;">
				<h1 style="color: white; margin: 0; text-align: center;">西外校园</h1>
			</div>
			<div style="background: #ffffff; padding: 40px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px;">
				<h2 style="color: #1f2937; margin-bottom: 20px;">邮箱验证</h2>
				<p style="color: #6b7280; font-size: 16px; line-height: 1.6;">
					您好！您正在注册西外校园账号，请使用以下验证码完成邮箱验证：
				</p>
				<div style="background: #f3f4f6; padding: 20px; border-radius: 8px; text-align: center; margin: 30px 0;">
					<span style="font-size: 36px; font-weight: bold; color: #667eea; letter-spacing: 8px;">{{CODE}}</span>
				</div>
				<p style="color: #6b7280; font-size: 14px; line-height: 1.6;">
					验证码有效期为 <strong>10 分钟</strong>，请尽快完成验证。
				</p>
				<p style="color: #9ca3af; font-size: 12px; margin-top: 30px;">
					如果您没有进行此操作，请忽略此邮件。
				</p>
				<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 30px 0;" />
				<p style="color: #9ca3af; font-size: 12px; text-align: center;">
					此邮件由系统自动发送，请勿回复
				</p>
			</div>
		</div>
	`, "{{CODE}}", code)

	return s.sendMail(to, subject, body)
}

func (s *MailService) SendGradeNotification(to, realName, semesterName, gradeTableHTML string, changeCount int) error {
	subject := "【知外助手】成绩变化通知"
	body := strings.ReplaceAll(`
		<div style="max-width: 600px; margin: 0 auto; padding: 20px; font-family: 'Microsoft YaHei', sans-serif;">
			<div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; border-radius: 10px 10px 0 0;">
				<h1 style="color: white; margin: 0; text-align: center;">知外助手</h1>
			</div>
			<div style="background: #ffffff; padding: 40px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px;">
				<h2 style="color: #1f2937; margin-bottom: 20px;">成绩变化通知</h2>
				<p style="color: #6b7280; font-size: 16px; line-height: 1.6;">
					尊敬的 <strong>{{REAL_NAME}}</strong>，您好！
				</p>
				<p style="color: #6b7280; font-size: 16px; line-height: 1.6;">
					系统检测到您在 <strong>{{SEMESTER}}</strong> 的成绩发生了变化，共有 <strong style="color: #ef4444;">{{COUNT}}</strong> 条成绩更新。
				</p>
				<div style="background: #f9fafb; padding: 20px; border-radius: 8px; margin: 20px 0; overflow-x: auto;">
					{{GRADE_TABLE}}
				</div>
				<p style="color: #6b7280; font-size: 14px; line-height: 1.6; margin-top: 20px;">
					如果您不想继续接收此类通知，可以在「应用中心 - 成绩订阅」中关闭订阅。
				</p>
				<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 30px 0;" />
				<p style="color: #9ca3af; font-size: 12px; text-align: center;">
					此邮件由知外助手系统自动发送，请勿直接回复
				</p>
			</div>
		</div>
	`, "{{REAL_NAME}}", realName)
	body = strings.ReplaceAll(body, "{{SEMESTER}}", semesterName)
	body = strings.ReplaceAll(body, "{{COUNT}}", fmt.Sprintf("%d", changeCount))
	body = strings.ReplaceAll(body, "{{GRADE_TABLE}}", gradeTableHTML)

	return s.sendMail(to, subject, body)
}

func (s *MailService) SendPasswordReset(to, code string) error {
	subject := "【西外校园】密码重置验证码"
	body := strings.ReplaceAll(`
		<div style="max-width: 600px; margin: 0 auto; padding: 20px; font-family: 'Microsoft YaHei', sans-serif;">
			<div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; border-radius: 10px 10px 0 0;">
				<h1 style="color: white; margin: 0; text-align: center;">西外校园</h1>
			</div>
			<div style="background: #ffffff; padding: 40px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px;">
				<h2 style="color: #1f2937; margin-bottom: 20px;">密码重置</h2>
				<p style="color: #6b7280; font-size: 16px; line-height: 1.6;">
					您正在重置密码，请使用以下验证码：
				</p>
				<div style="background: #f3f4f6; padding: 20px; border-radius: 8px; text-align: center; margin: 30px 0;">
					<span style="font-size: 36px; font-weight: bold; color: #667eea; letter-spacing: 8px;">{{CODE}}</span>
				</div>
				<p style="color: #6b7280; font-size: 14px; line-height: 1.6;">
					验证码有效期为 <strong>10 分钟</strong>。
				</p>
				<p style="color: #ef4444; font-size: 14px; margin-top: 20px;">
					⚠️ 如果您没有进行此操作，您的账号可能存在安全风险，请及时修改密码。
				</p>
			</div>
		</div>
	`, "{{CODE}}", code)

	return s.sendMail(to, subject, body)
}

// SendCustomMail 发送自定义邮件（支持富文本HTML）
func (s *MailService) SendCustomMail(to, subject, htmlContent string) error {
	return s.sendMail(to, subject, htmlContent)
}

func (s *MailService) sendMail(to, subject, body string) error {
	if strings.TrimSpace(s.host) == "" || strings.TrimSpace(s.port) == "" {
		return errors.New("smtp host or port is empty")
	}
	if strings.TrimSpace(s.username) == "" || strings.TrimSpace(s.password) == "" {
		return errors.New("smtp username or password is empty")
	}
	if strings.TrimSpace(s.from) == "" {
		return errors.New("smtp from address is empty")
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	headers := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body
	msg := []byte(headers)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	// 尝试使用 TLS
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		log.Printf("TLS connection failed: %v", err)
		return fmt.Errorf("SMTP TLS connection failed, refusing to send credentials over plaintext")
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(s.from); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
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

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

