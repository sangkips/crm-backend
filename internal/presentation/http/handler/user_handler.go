package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List handles listing users with pagination
// @Summary List Users
// @Description Get a paginated list of users with their roles
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(15)
// @Param search query string false "Search query"
// @Success 200 {object} response.APIResponse
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "15"))
	search := c.Query("search")

	output, err := h.userService.ListUsers(c.Request.Context(), &service.ListUsersInput{
		Page:    page,
		PerPage: perPage,
		Search:  search,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	// Transform users to response format (exclude sensitive data)
	users := make([]gin.H, len(output.Users))
	for i, user := range output.Users {
		users[i] = gin.H{
			"id":         user.ID,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"username":   user.Username,
			"photo":      user.Photo,
			"roles":      user.Roles,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		}
	}

	response.OK(c, "Users retrieved successfully", gin.H{
		"items": users,
		"pagination": gin.H{
			"current_page": output.Page,
			"per_page":     output.PerPage,
			"total":        output.Total,
			"total_pages":  output.TotalPages,
			"has_next":     output.Page < output.TotalPages,
			"has_prev":     output.Page > 1,
		},
	})
}

// Get handles getting a single user by ID
// @Summary Get User
// @Description Get a user by ID with roles and permissions
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.APIResponse
// @Router /users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "User retrieved successfully", gin.H{
		"user": gin.H{
			"id":          user.ID,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"email":       user.Email,
			"username":    user.Username,
			"photo":       user.Photo,
			"store_name":  user.StoreName,
			"roles":       user.Roles,
			"permissions": user.GetPermissions(),
			"created_at":  user.CreatedAt,
			"updated_at":  user.UpdatedAt,
		},
	})
}

// UpdateRolesRequest represents the request body for updating user roles
type UpdateRolesRequest struct {
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

// UpdateRoles handles updating user roles
// @Summary Update User Roles
// @Description Update the roles assigned to a user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateRolesRequest true "Role IDs"
// @Success 200 {object} response.APIResponse
// @Router /users/{id}/roles [put]
func (h *UserHandler) UpdateRoles(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req UpdateRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	user, err := h.userService.UpdateUserRoles(c.Request.Context(), &service.UpdateUserRolesInput{
		UserID:  userID,
		RoleIDs: req.RoleIDs,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "User roles updated successfully", gin.H{
		"user": gin.H{
			"id":          user.ID,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"email":       user.Email,
			"username":    user.Username,
			"roles":       user.Roles,
			"permissions": user.GetPermissions(),
		},
	})
}

// Delete handles deleting a user
// @Summary Delete User
// @Description Soft delete a user
// @Tags users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.APIResponse
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	// Prevent self-deletion
	currentUserID := GetUserID(c)
	if currentUserID != nil && *currentUserID == userID {
		response.BadRequest(c, "Cannot delete your own account")
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "User deleted successfully", nil)
}

// ListRoles handles listing all available roles
// @Summary List Roles
// @Description Get all available roles with their permissions
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /roles [get]
func (h *UserHandler) ListRoles(c *gin.Context) {
	roles, err := h.userService.ListRoles(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Roles retrieved successfully", gin.H{
		"roles": roles,
	})
}

// ListPermissions handles listing all available permissions
// @Summary List Permissions
// @Description Get all available permissions
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /permissions [get]
func (h *UserHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.userService.ListPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response.OK(c, "Permissions retrieved successfully", gin.H{
		"permissions": permissions,
	})
}
