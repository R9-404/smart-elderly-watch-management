package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"suining-care/internal/auth"
	"time"

	"suining-care/internal/response"
)

func (h *Handler) routeRisk(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/risk/events", h.riskAdminEvents)
	mux.HandleFunc("POST /api/v1/risk/rules", h.riskAdminRuleCreate)
	mux.HandleFunc("DELETE /api/v1/risk/rules/{id}", h.riskAdminRuleDelete)
}

func (h *Handler) riskDeniedLogin(w http.ResponseWriter) {
	response.JSON(w, http.StatusUnauthorized, map[string]any{"code": "INVALID_CREDENTIALS", "message": "手机号或密码不正确"})
}
func (h *Handler) riskDeniedInvite(w http.ResponseWriter) {
	response.Error(w, http.StatusNotFound, "INVITE_NOT_FOUND", "邀请码不存在，请向管理员确认")
}

// Unknown routes are admin-only in accesspolicy. Verify principal again in the handler.
func (h *Handler) riskAdmin(w http.ResponseWriter, r *http.Request) (string, bool) {
	if h.Risk == nil {
		response.Error(w, 503, "RISK_DISABLED", "登录风控尚未启用")
		return "", false
	}
	s := h.Sessions.FromRequest(r)
	if s == nil || s.Principal.Role != auth.RoleAdmin {
		response.Error(w, http.StatusForbidden, "FORBIDDEN", "管理员权限不足")
		return "", false
	}
	return fmt.Sprintf("admin:%d", s.Principal.ID), true
}
func (h *Handler) riskAdminEvents(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.riskAdmin(w, r); !ok {
		return
	}
	events, rules := h.Risk.History(100)
	w.Header().Set("Cache-Control", "no-store")
	response.OK(w, map[string]any{"events": events, "rules": rules})
}
func (h *Handler) riskAdminRuleCreate(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.riskAdmin(w, r)
	if !ok {
		return
	}
	var input struct {
		IP            string `json:"ip"`
		Action        string `json:"action"`
		Reason        string `json:"reason"`
		DurationHours int    `json:"durationHours"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	if err := decoder.Decode(&input); err != nil {
		response.Error(w, 400, "INVALID_RULE", "请求内容不合法")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		response.Error(w, 400, "INVALID_RULE", "请求内容不合法")
		return
	}
	if input.DurationHours < 0 || input.DurationHours > 8760 {
		response.Error(w, 400, "INVALID_DURATION", "封禁期限不合法")
		return
	}
	rule, err := h.Risk.AddRule(input.IP, input.Action, input.Reason, actor, time.Duration(input.DurationHours)*time.Hour)
	if err != nil {
		response.Error(w, 400, "INVALID_RULE", "规则无效或暂时无法保存")
		return
	}
	response.OK(w, map[string]any{"id": rule.ID, "cidr": rule.CIDR, "action": rule.Action,
		"reason": rule.Reason, "actor": rule.Actor, "createdAt": rule.CreatedAt, "expiresAt": rule.ExpiresAt})
}
func (h *Handler) riskAdminRuleDelete(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.riskAdmin(w, r)
	if !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	deleted, err := h.Risk.DeleteRule(id, actor)
	if err != nil {
		response.Error(w, 500, "RISK_STORAGE_ERROR", "无法保存规则")
		return
	}
	if !deleted {
		response.Error(w, 404, "RULE_NOT_FOUND", "规则不存在")
		return
	}
	response.OK(w, map[string]any{"deleted": true})
}

func riskSignals(r *http.Request) []string {
	return []string{r.UserAgent(), r.Header.Get("Sec-CH-UA-Model"), r.Header.Get("X-App-Device-ID")}
}

func (h *Handler) allowRiskInvite(w http.ResponseWriter, r *http.Request) bool {
	if h.Risk == nil || h.Risk.BeforeInvite(h.clientIP(r)).Allowed() {
		return true
	}
	h.Risk.RecordDenied(h.clientIP(r), "", "invite", riskSignals(r)...)
	h.riskDeniedInvite(w)
	return false
}
