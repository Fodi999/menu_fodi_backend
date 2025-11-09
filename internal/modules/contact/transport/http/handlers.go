package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/contact/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/contact/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
	"go.uber.org/zap"
)

// ContactHandlers обработчики контактной формы
type ContactHandlers struct {
	service *service.ContactService
	logger  *zap.Logger
}

// NewContactHandlers создает новый обработчик
func NewContactHandlers(srv *service.ContactService, logger *zap.Logger) *ContactHandlers {
	return &ContactHandlers{
		service: srv,
		logger:  logger,
	}
}

// SubmitContactForm обрабатывает отправку контактной формы
// @Summary Submit contact form
// @Description Send a message to administrators through contact form
// @Tags Contact
// @Accept json
// @Produce json
// @Param request body dto.ContactRequest true "Contact form data"
// @Success 200 {object} dto.ContactResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Router /api/contact [post]
func (h *ContactHandlers) SubmitContactForm(w http.ResponseWriter, r *http.Request) {
	var req dto.ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request body")
		return
	}

	// Валидация обязательных полей
	if req.Name == "" || req.Email == "" || req.Message == "" {
		httpx.BadRequest(w, "Name, email and message are required")
		return
	}

	// Устанавливаем дефолтный subject
	if req.Subject == "" {
		req.Subject = "Contact Form Submission"
	}

	// Отправляем email
	err := h.service.SendContactEmail(req.Name, req.Email, req.Subject, req.Message)
	if err != nil {
		// Логируем ошибку, но возвращаем успех пользователю (для UX)
		h.logger.Error("Failed to send contact email",
			zap.Error(err),
			zap.String("email", req.Email),
		)
	}

	// Всегда возвращаем успех (даже если email не отправился)
	httpx.Success(w, dto.ContactResponse{
		Success: true,
		Message: "Your message has been sent successfully. We'll get back to you soon!",
	})
}
