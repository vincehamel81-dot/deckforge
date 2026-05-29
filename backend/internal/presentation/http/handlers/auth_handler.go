package handlers

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/vincehamel81-dot/deckforge/internal/infrastructure/auth"
	"github.com/vincehamel81-dot/deckforge/internal/domain/user"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]{3,}$`)

type AuthHandler struct {
	users     user.Repository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewAuthHandler(users user.Repository, jwtSecret string, jwtExpiry time.Duration) *AuthHandler {
	return &AuthHandler{users: users, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	if !usernameRegex.MatchString(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must be alphanumeric and at least 3 characters"})
		return
	}
	exists, err := h.users.ExistsByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}
	u := user.New(req.Username)
	if err := h.users.Create(u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}
	token, err := jwtpkg.Issue(u.ID.String(), u.Username, string(u.Role), h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  gin.H{"id": u.ID, "username": u.Username, "role": u.Role},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	u, err := h.users.FindByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "username not found"})
		return
	}
	token, err := jwtpkg.Issue(u.ID.String(), u.Username, string(u.Role), h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  gin.H{"id": u.ID, "username": u.Username, "role": u.Role},
	})
}
