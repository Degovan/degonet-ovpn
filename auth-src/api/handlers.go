package api

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/irsyadulibad/degonet-ovpn/auth-go/cmd"
)

type Handler struct {
	Auth           cmd.AuthOperations
	CCD            cmd.CCDOperations
	Mgmt           *Management
	APIKey         string
	VPNSubnet      string
	DefaultNetmask string

	loginMu sync.Mutex
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
	if msg := validateAddresses(req.IP, req.Netmask); msg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   msg,
		})
		return
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

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	var req struct {
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

	if msg := validateAddresses(req.IP, req.Netmask); msg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   msg,
		})
		return
	}

	existing, _ := h.Auth.FindByUsername(username)
	if existing == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "user not found",
		})
		return
	}

	if req.IP != "" && req.IP != existing.IP {
		existingIP, _ := h.Auth.FindByIP(req.IP)
		if existingIP != nil {
			writeJSON(w, http.StatusConflict, map[string]any{
				"success": false,
				"error":   "ip already in use",
			})
			return
		}
	}

	user, err := h.Auth.UpdateUser(username, req.Password, req.IP, req.Netmask)
	if err != nil || user == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to update user",
		})
		return
	}

	// CCD.Create tidak menimpa file lama, jadi hapus dulu supaya IP/netmask baru terpakai.
	h.CCD.Delete(username)
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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}

	if !keyMatches(req.APIKey, h.APIKey) {
		// Login gagal diproses satu per satu dengan jeda, untuk memperlambat brute force.
		h.loginMu.Lock()
		time.Sleep(time.Second)
		h.loginMu.Unlock()
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"success": false,
			"error":   "invalid API key",
		})
		return
	}

	http.SetCookie(w, newSessionCookie(r, h.APIKey))
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, clearSessionCookie(r))
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}

func (h *Handler) Config(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success":         true,
		"vpn_subnet":      h.VPNSubnet,
		"default_netmask": h.DefaultNetmask,
	})
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	clients, err := h.Mgmt.Clients()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"success": false,
			"error":   "management interface unavailable",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"clients": clients,
	})
}

func (h *Handler) KickUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	err := h.Mgmt.Kill(username)
	if errors.Is(err, ErrClientNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "client not connected",
		})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"success": false,
			"error":   "failed to kick client",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}

// validateAddresses mengecek format IPv4; string kosong dianggap valid (tidak diubah).
func validateAddresses(ip, netmask string) string {
	if ip != "" && net.ParseIP(ip).To4() == nil {
		return "invalid ip address"
	}
	if netmask != "" && net.ParseIP(netmask).To4() == nil {
		return "invalid netmask"
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
