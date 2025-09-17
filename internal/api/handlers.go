package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"subscription-service/internal/log"
	"subscription-service/internal/model"
	"subscription-service/internal/service"
)

type Handlers struct {
	svc    *service.Service
	logger *log.Logger
}

func NewHandlers(s *service.Service, l *log.Logger) *Handlers {
	return &Handlers{svc: s, logger: l}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func parseYYYYMM(s string) (time.Time, error) {
	// Принимаем YYYY-MM
	return time.Parse("2006-01", s)
}

// Create godoc
// @Summary      Создать подписку
// @Description  Добавляет новую подписку пользователю
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription body model.SubscriptionCreate true "Subscription"
// @Success      201  {object} model.Subscription
// @Failure      400  {object} map[string]string
// @Router       /api/v1/subscriptions/ [post]
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var req model.SubscriptionCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid json", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	sub, err := h.svc.Create(r.Context(), req)
	if err != nil {
		h.logger.Warn("create failed", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	h.logger.Info("subscription created", "id", sub.ID, "user_id", sub.UserID)
	writeJSON(w, http.StatusCreated, sub)
}

// GetByID godoc
// @Summary      Получить подписку
// @Description  Возвращает подписку по её ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "UUID подписки"
// @Success      200  {object}  model.Subscription
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/subscriptions/{id}/ [get]
func (h *Handlers) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("subscription not found", "id", id, "err", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

// Update godoc
// @Summary      Обновить подписку
// @Description  Обновляет данные подписки по ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "UUID подписки"
// @Param        subscription body model.SubscriptionUpdate true "Subscription update"
// @Success      200  {object}  model.Subscription
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/subscriptions/{id}/ [put]
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req model.SubscriptionUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	sub, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		h.logger.Warn("update failed", "id", id, "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	h.logger.Info("subscription updated", "id", sub.ID)
	writeJSON(w, http.StatusOK, sub)
}

// Delete godoc
// @Summary      Удалить подписку
// @Description  Удаляет подписку по ID
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "UUID подписки"
// @Success      204  {string}  string  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/subscriptions/{id}/ [delete]
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.logger.Warn("delete failed", "id", id, "err", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	h.logger.Info("subscription deleted", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

// List godoc
// @Summary      Список подписок
// @Description  Возвращает список подписок по фильтрам
// @Tags         subscriptions
// @Produce      json
// @Param        user_id      query  string  false  "UUID пользователя"
// @Param        service_name query  string  false  "Название сервиса"
// @Param        limit        query  int     false  "Лимит"  default(50)
// @Param        offset       query  int     false  "Смещение"  default(0)
// @Success      200  {array}  model.Subscription
// @Failure      400  {object}  map[string]string
// @Router       /api/v1/subscriptions/ [get]
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	q := model.ListQuery{
		UserID:      r.URL.Query().Get("user_id"),
		ServiceName: r.URL.Query().Get("service_name"),
		Limit:       model.ParseInt(r.URL.Query().Get("limit"), 50),
		Offset:      model.ParseInt(r.URL.Query().Get("offset"), 0),
	}
	items, err := h.svc.List(r.Context(), q)
	if err != nil {
		h.logger.Error("list failed", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// Total godoc
// @Summary      Общая стоимость подписок
// @Description  Возвращает суммарную стоимость подписок за выбранный период
// @Tags         subscriptions
// @Produce      json
// @Param        user_id      query  string  true  "UUID пользователя"
// @Param        service_name query  string  false "Название сервиса"
// @Param        from         query  string  true  "Начало периода (YYYY-MM)"
// @Param        to           query  string  true  "Конец периода (YYYY-MM)"
// @Success      200  {object}  map[string]int64
// @Failure      400  {object}  map[string]string
// @Router       /api/v1/subscriptions/total [get]
func (h *Handlers) Total(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user_id")
	if user == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id required"})
		return
	}
	fromS := r.URL.Query().Get("from")
	toS := r.URL.Query().Get("to")
	if fromS == "" || toS == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "from and to required (YYYY-MM)"})
		return
	}
	from, err := parseYYYYMM(fromS)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid from (YYYY-MM)"})
		return
	}
	to, err := parseYYYYMM(toS)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid to (YYYY-MM)"})
		return
	}
	total, err := h.svc.Total(r.Context(), model.TotalQuery{
		UserID:      user,
		ServiceName: r.URL.Query().Get("service_name"),
		From:        from,
		To:          to,
	})
	if err != nil {
		h.logger.Error("total calc failed", "user_id", user, "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"total": total})
}
