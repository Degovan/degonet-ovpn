package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/irsyadulibad/degonet-ovpn/auth-go/cmd"
)

type Handler struct {
	Auth cmd.AuthOperations
	CCD  cmd.CCDOperations
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"status":  "ok",
	})
}

func (h *Handler) AuthUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}

	user, found, err := h.Auth.Login(req.Username, req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "internal error",
		})
		return
	}
	if user == nil || !found {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"success": false,
			"error":   "invalid credentials",
		})
		return
	}

	h.CCD.Create(user)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"user": map[string]any{
			"username": user.Username,
			"ip":       user.IP,
			"netmask":  user.Netmask,
		},
	})
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Auth.ListUsers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to list users",
		})
		return
	}

	result := make([]map[string]any, 0, len(users))
	for _, u := range users {
		result = append(result, map[string]any{
			"username": u.Username,
			"ip":       u.IP,
			"netmask":  u.Netmask,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"users":   result,
	})
}

func (h *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IP       string `json:"ip"`
		Netmask  string `json:"netmask"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}

	if req.Username == "" || req.IP == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "username and ip are required",
		})
		return
	}

	if req.Password == "" {
		req.Password = req.Username
	}
	if req.Netmask == "" {
		req.Netmask = "255.255.255.0"
	}

	existingUser, _ := h.Auth.FindByUsername(req.Username)
	if existingUser != nil {
		writeJSON(w, http.StatusConflict, map[string]any{
			"success": false,
			"error":   "username already exists",
		})
		return
	}

	existingIP, _ := h.Auth.FindByIP(req.IP)
	if existingIP != nil {
		writeJSON(w, http.StatusConflict, map[string]any{
			"success": false,
			"error":   "ip already in use",
		})
		return
	}

	user, err := h.Auth.AddUser(req.Username, req.Password, req.IP, req.Netmask)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to add user",
		})
		return
	}

	h.CCD.Create(user)

	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"user": map[string]any{
			"username": user.Username,
			"ip":       user.IP,
			"netmask":  user.Netmask,
		},
	})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, _ := h.Auth.FindByUsername(username)
	if user == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "user not found",
		})
		return
	}

	deleted, err := h.Auth.DeleteUser(username)
	if err != nil || !deleted {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to delete user",
		})
		return
	}

	h.CCD.Delete(username)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
