package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vincehamel81-dot/deckforge/internal/domain/user"
	jwtpkg "github.com/vincehamel81-dot/deckforge/internal/infrastructure/auth"
)

var usernameAlphanumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

type AuthHandler struct {
	users          user.Repository
	jwtSecret      string
	jwtExpiry      time.Duration
	maxUsernameLen int
}

func NewAuthHandler(users user.Repository, jwtSecret string, jwtExpiry time.Duration, maxUsernameLen int) *AuthHandler {
	return &AuthHandler{users: users, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry, maxUsernameLen: maxUsernameLen}
}

type registerRequest struct {
	Username string `json:"username" binding:"required" example:"alice42"`
}

// Register godoc
// @Summary Register a new user
// @Description Creates an account with a username-only credential and returns a signed JWT.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body registerRequest true "username (3–15 alphanumeric chars)"
// @Success 201 {object} map[string]interface{} "token + user"
// @Failure 400 {object} map[string]string "validation error"
// @Failure 409 {object} map[string]string "username already taken"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	if n := len(req.Username); n < 3 || n > h.maxUsernameLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("username must be 3–%d characters", h.maxUsernameLen)})
		return
	}
	if !usernameAlphanumeric.MatchString(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must contain only letters and numbers"})
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

// Login godoc
// @Summary Log in as an existing user
// @Description Returns a fresh JWT for an existing username.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body registerRequest true "username"
// @Success 200 {object} map[string]interface{} "token + user"
// @Failure 400 {object} map[string]string "bad request"
// @Failure 404 {object} map[string]string "username not found"
// @Router /auth/login [post]
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
