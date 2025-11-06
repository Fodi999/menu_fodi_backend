package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/ai"
	"github.com/jung-kurt/gofpdf"
)

// CertificateData данные для генерации сертификата
type CertificateData struct {
	StudentName           string
	CourseName            string
	Level                 int
	Stars                 int
	CompletionDate        time.Time
	Language              string
	QuizScore             int
	AIPersonalizedMessage string
}

// CertificateService сервис генерации PDF-сертификатов
type CertificateService struct {
	outputDir string
}

// NewCertificateService создаёт новый сервис
func NewCertificateService() *CertificateService {
	outputDir := "./certificates"
	os.MkdirAll(outputDir, 0755)
	return &CertificateService{
		outputDir: outputDir,
	}
}

// GenerateCertificate создаёт PDF-сертификат
func (cs *CertificateService) GenerateCertificate(data CertificateData) (string, error) {
	// 1. Генерируем AI-персонализированное сообщение
	if data.AIPersonalizedMessage == "" {
		message, err := cs.generateAIMessage(data)
		if err != nil {
			log.Printf("[CERT] AI message generation failed: %v, using fallback", err)
			data.AIPersonalizedMessage = cs.getFallbackMessage(data)
		} else {
			data.AIPersonalizedMessage = message
		}
	}

	// 2. Создаём PDF
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape A4
	pdf.AddPage()

	// 3. Оформление сертификата
	cs.drawCertificate(pdf, data)

	// 4. Сохраняем файл
	filename := fmt.Sprintf("certificate_%s_%d.pdf",
		sanitizeFilename(data.StudentName),
		time.Now().Unix())
	filepath := filepath.Join(cs.outputDir, filename)

	err := pdf.OutputFileAndClose(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	log.Printf("[CERT] ✅ Generated certificate: %s", filename)
	return filepath, nil
}

// drawCertificate рисует содержимое сертификата
func (cs *CertificateService) drawCertificate(pdf *gofpdf.Fpdf, data CertificateData) {
	// Цвета
	goldR, goldG, goldB := 212, 175, 55
	darkBlueR, darkBlueG, darkBlueB := 25, 25, 112

	// Фоновая рамка (золотая)
	pdf.SetLineWidth(2)
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.Rect(10, 10, 277, 190, "D")

	pdf.SetLineWidth(0.5)
	pdf.Rect(15, 15, 267, 180, "D")

	// Заголовок "CERTIFICATE OF ACHIEVEMENT"
	pdf.SetFont("Arial", "B", 32)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)

	titles := map[string]string{
		"pl": "CERTYFIKAT UKOŃCZENIA",
		"ua": "СЕРТИФІКАТ ЗАВЕРШЕННЯ",
		"en": "CERTIFICATE OF ACHIEVEMENT",
	}
	title := titles[data.Language]
	if title == "" {
		title = titles["pl"]
	}

	pdf.SetXY(20, 30)
	pdf.CellFormat(257, 15, title, "", 0, "C", false, 0, "")

	// Логотип Culinary Academy
	pdf.SetFont("Arial", "I", 14)
	pdf.SetTextColor(100, 100, 100)
	pdf.SetXY(20, 50)
	pdf.CellFormat(257, 8, "Culinary Academy - AI-Powered Chef Training", "", 0, "C", false, 0, "")

	// Разделитель
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.SetLineWidth(1)
	pdf.Line(80, 63, 217, 63)

	// "This certifies that"
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(60, 60, 60)
	texts := map[string]string{
		"pl": "Niniejszym potwierdza się, że",
		"ua": "Цим підтверджується, що",
		"en": "This certifies that",
	}
	certText := texts[data.Language]
	if certText == "" {
		certText = texts["pl"]
	}

	pdf.SetXY(20, 75)
	pdf.CellFormat(257, 8, certText, "", 0, "C", false, 0, "")

	// Имя студента (крупный шрифт)
	pdf.SetFont("Arial", "B", 28)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	pdf.SetXY(20, 88)
	pdf.CellFormat(257, 12, data.StudentName, "", 0, "C", false, 0, "")

	// "has successfully completed"
	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(60, 60, 60)
	completedTexts := map[string]string{
		"pl": "ukończył z wyróżnieniem kurs",
		"ua": "успішно завершив курс",
		"en": "has successfully completed the course",
	}
	completedText := completedTexts[data.Language]
	if completedText == "" {
		completedText = completedTexts["pl"]
	}

	pdf.SetXY(20, 105)
	pdf.CellFormat(257, 8, completedText, "", 0, "C", false, 0, "")

	// Название курса
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(darkBlueR, darkBlueG, darkBlueB)
	pdf.SetXY(20, 118)
	pdf.CellFormat(257, 10, data.CourseName, "", 0, "C", false, 0, "")

	// Оценки и достижения
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)

	achievementTexts := map[string]string{
		"pl": fmt.Sprintf("Poziom: %d  |  Wynik testu: %d%%  |  Zdobyte Gwiazdki: %d",
			data.Level, data.QuizScore, data.Stars),
		"ua": fmt.Sprintf("Рівень: %d  |  Результат тесту: %d%%  |  Зірки: %d",
			data.Level, data.QuizScore, data.Stars),
		"en": fmt.Sprintf("Level: %d  |  Quiz Score: %d%%  |  Stars Earned: %d",
			data.Level, data.QuizScore, data.Stars),
	}
	achievementText := achievementTexts[data.Language]
	if achievementText == "" {
		achievementText = achievementTexts["pl"]
	}

	pdf.SetXY(20, 135)
	pdf.CellFormat(257, 6, achievementText, "", 0, "C", false, 0, "")

	// AI-персонализированное сообщение (в рамке)
	pdf.SetFillColor(245, 245, 250)
	pdf.Rect(40, 145, 217, 25, "F")
	pdf.SetDrawColor(goldR, goldG, goldB)
	pdf.Rect(40, 145, 217, 25, "D")

	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetXY(45, 150)
	pdf.MultiCell(207, 5, data.AIPersonalizedMessage, "", "C", false)

	// Дата и подпись
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)

	dateFormat := "02.01.2006"
	dateText := data.CompletionDate.Format(dateFormat)

	pdf.SetXY(50, 180)
	pdf.CellFormat(80, 6, dateText, "T", 0, "C", false, 0, "")

	pdf.SetXY(167, 180)
	pdf.CellFormat(80, 6, "Chef Dima Fomin", "T", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(100, 100, 100)

	dateLabels := map[string]string{
		"pl": "Data wydania",
		"ua": "Дата видачі",
		"en": "Date of Issue",
	}
	dateLabel := dateLabels[data.Language]
	if dateLabel == "" {
		dateLabel = dateLabels["pl"]
	}

	pdf.SetXY(50, 186)
	pdf.CellFormat(80, 4, dateLabel, "", 0, "C", false, 0, "")

	signatureLabels := map[string]string{
		"pl": "Podpis Szefa Kuchni",
		"ua": "Підпис Шеф-кухаря",
		"en": "Chef Signature",
	}
	signatureLabel := signatureLabels[data.Language]
	if signatureLabel == "" {
		signatureLabel = signatureLabels["pl"]
	}

	pdf.SetXY(167, 186)
	pdf.CellFormat(80, 4, signatureLabel, "", 0, "C", false, 0, "")
}

// generateAIMessage генерирует персонализированное сообщение от AI
func (cs *CertificateService) generateAIMessage(data CertificateData) (string, error) {
	client := ai.NewGroqClient()

	systemPrompts := map[string]string{
		"pl": "Jesteś szefem kuchni wystawiającym certyfikat. Napisz krótką (max 2 zdania) personalizowaną gratulację dla ucznia. Bądź ciepły i motywujący.",
		"ua": "Ти шеф-кухар, який видає сертифікат. Напиши коротке (макс 2 речення) персоналізоване привітання учневі. Будь теплим та мотивуючим.",
		"en": "You are a chef issuing a certificate. Write a short (max 2 sentences) personalized congratulation for the student. Be warm and motivating.",
	}

	userMessages := map[string]string{
		"pl": fmt.Sprintf("Uczeń %s ukończył kurs '%s' z wynikiem %d%%, zdobywając %d gwiazdek. Poziom %d.",
			data.StudentName, data.CourseName, data.QuizScore, data.Stars, data.Level),
		"ua": fmt.Sprintf("Учень %s завершив курс '%s' з результатом %d%%, заробивши %d зірок. Рівень %d.",
			data.StudentName, data.CourseName, data.QuizScore, data.Stars, data.Level),
		"en": fmt.Sprintf("Student %s completed course '%s' with %d%% score, earning %d stars. Level %d.",
			data.StudentName, data.CourseName, data.QuizScore, data.Stars, data.Level),
	}

	systemPrompt := systemPrompts[data.Language]
	userMessage := userMessages[data.Language]

	if systemPrompt == "" {
		systemPrompt = systemPrompts["pl"]
		userMessage = userMessages["pl"]
	}

	response, err := client.SimpleChat(systemPrompt, userMessage)
	if err != nil {
		return "", err
	}

	return response, nil
}

// getFallbackMessage возвращает стандартное сообщение если AI недоступен
func (cs *CertificateService) getFallbackMessage(data CertificateData) string {
	messages := map[string]string{
		"pl": fmt.Sprintf("Gratulacje, %s! Twoja pasja i determinacja są inspirujące. Kontynuuj swoją kulinarną podróż!",
			data.StudentName),
		"ua": fmt.Sprintf("Вітаємо, %s! Ваша пристрасть та рішучість надихають. Продовжуйте свою кулінарну подорож!",
			data.StudentName),
		"en": fmt.Sprintf("Congratulations, %s! Your passion and determination are inspiring. Continue your culinary journey!",
			data.StudentName),
	}

	message := messages[data.Language]
	if message == "" {
		message = messages["pl"]
	}

	return message
}

// sanitizeFilename очищает имя файла от опасных символов
func sanitizeFilename(name string) string {
	// Простая замена для кириллицы и спецсимволов
	safe := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			safe += string(r)
		} else if r == ' ' {
			safe += "_"
		}
	}
	if safe == "" {
		safe = "student"
	}
	return safe
}
