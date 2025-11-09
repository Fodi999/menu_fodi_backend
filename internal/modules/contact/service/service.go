package service

import (
	"fmt"
	"net/smtp"
	"os"

	"go.uber.org/zap"
)

// EmailConfig - конфигурация SMTP
type EmailConfig struct {
	SMTPHost   string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	AdminEmail string
}

// ContactService - сервис контактной формы
type ContactService struct {
	config EmailConfig
	logger *zap.Logger
}

// NewContactService создает новый сервис
func NewContactService(logger *zap.Logger) *ContactService {
	return &ContactService{
		config: EmailConfig{
			SMTPHost:   os.Getenv("SMTP_HOST"),
			SMTPPort:   os.Getenv("SMTP_PORT"),
			SMTPUser:   os.Getenv("SMTP_USER"),
			SMTPPass:   os.Getenv("SMTP_PASS"),
			AdminEmail: getAdminEmail(),
		},
		logger: logger,
	}
}

// SendContactEmail отправляет email администратору
func (s *ContactService) SendContactEmail(name, email, subject, message string) error {
	// Проверяем конфигурацию
	if s.config.SMTPHost == "" || s.config.SMTPPort == "" {
		s.logger.Warn("SMTP не сконфигурирован, email не отправлен",
			zap.String("from", email),
			zap.String("name", name),
		)
		return nil // Возвращаем nil чтобы не показывать ошибку пользователю
	}

	// Подготовка сообщения
	emailBody := fmt.Sprintf(
		"From: %s <%s>\nSubject: [Contact Form] %s\n\nName: %s\nEmail: %s\n\nMessage:\n%s",
		name, email, subject, name, email, message,
	)

	// Настройка аутентификации
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPass, s.config.SMTPHost)

	// Отправка письма
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	to := []string{s.config.AdminEmail}

	err := smtp.SendMail(addr, auth, s.config.SMTPUser, to, []byte(emailBody))
	if err != nil {
		s.logger.Error("Ошибка отправки email",
			zap.Error(err),
			zap.String("to", s.config.AdminEmail),
			zap.String("from", email),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("📧 Contact form submitted",
		zap.String("name", name),
		zap.String("email", email),
		zap.String("subject", subject),
	)

	return nil
}

// getAdminEmail возвращает email администратора из env или дефолтный
func getAdminEmail() string {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		return "admin@chefacademy.com"
	}
	return adminEmail
}
