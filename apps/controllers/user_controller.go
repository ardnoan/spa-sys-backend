// apps/controllers/user_controller.go
package controllers

import (
	"strconv"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/services"
	"v01_system_backend/apps/utils"

	"github.com/labstack/echo/v4"
)

// UserController bertanggung jawab untuk menangani HTTP requests terkait user
// Controller ini adalah layer pertama yang menerima request dari client
type UserController struct {
	userService *services.UserService // Dependency injection service layer
}

// NewUserController adalah constructor untuk membuat instance UserController
// Menggunakan dependency injection pattern untuk loose coupling
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

// GetAll menangani GET /users - Mengambil semua data user dengan pagination dan filter
// ALUR:
// 1. Parse pagination parameters dari query string
// 2. Parse filter parameters (status, search, role)
// 3. Panggil service untuk business logic
// 4. Return response dalam format JSON
func (h *UserController) GetAll(c echo.Context) error {
	// Inisialisasi struct untuk pagination
	var pagination models.PaginationRequest

	// Bind query parameters ke struct pagination
	// Otomatis map query params seperti ?page=1&page_size=10
	if err := c.Bind(&pagination); err != nil {
		return utils.BadRequestResponse(c, "Invalid pagination parameters", err.Error())
	}

	// Parse filter parameters secara manual dari query string
	var filters models.UserFilters
	filters.Status = c.QueryParam("status") // ?status=active
	filters.Search = c.QueryParam("search") // ?search=john
	filters.Role = c.QueryParam("role")     // ?role=admin

	// Panggil service layer untuk business logic
	// Service akan handle validasi, pagination logic, dan data retrieval
	response, err := h.userService.GetAll(&pagination, &filters)
	if err != nil {
		// Jika ada error, return 500 Internal Server Error
		return utils.InternalServerErrorResponse(c, "Failed to get users", err.Error())
	}

	// Success response dengan data yang sudah diformat
	return utils.SuccessResponse(c, "Users retrieved successfully", response)
}

// GetByID menangani GET /users/:id - Mengambil user berdasarkan ID
// ALUR:
// 1. Extract ID dari URL parameter
// 2. Validasi ID (harus integer)
// 3. Panggil service untuk mengambil data
// 4. Handle error jika user tidak ditemukan
func (h *UserController) GetByID(c echo.Context) error {
	// Extract ID dari URL parameter dan convert ke integer
	// Contoh: /users/123 -> id = 123
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		// Jika ID bukan integer valid, return 400 Bad Request
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	// Panggil service untuk mengambil user by ID
	user, err := h.userService.GetByID(id)
	if err != nil {
		// Check specific error type untuk response yang tepat
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error()) // 404 Not Found
		}
		return utils.InternalServerErrorResponse(c, "Failed to get user", err.Error()) // 500 Error
	}

	return utils.SuccessResponse(c, "User retrieved successfully", user)
}

// Create menangani POST /users - Membuat user baru
// ALUR:
// 1. Parse request body ke struct
// 2. Validasi data input
// 3. Extract user ID dari context (dari JWT token)
// 4. Panggil service untuk create user
func (h *UserController) Create(c echo.Context) error {
	// Inisialisasi struct untuk menampung data request
	var req models.UserCreateRequest

	// Bind JSON request body ke struct
	// Otomatis parse JSON ke struct fields
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Validasi data menggunakan struct tags (validate:"required,email,etc")
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.BadRequestResponse(c, "Validation failed", err.Error())
	}

	// Extract user ID yang sedang login dari JWT context
	// Middleware auth sudah menyimpan user_id di context
	createdBy := c.Get("user_id").(int)

	// Panggil service untuk create user dengan business logic
	user, err := h.userService.Create(&req, createdBy)
	if err != nil {
		// Service akan return error dengan message yang descriptive
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	// Return 201 Created dengan data user yang baru dibuat
	return utils.CreatedResponse(c, "User created successfully", user)
}

// Update menangani PUT /users/:id - Update user berdasarkan ID
// ALUR:
// 1. Extract ID dari URL
// 2. Parse request body
// 3. Validasi data
// 4. Update via service layer
func (h *UserController) Update(c echo.Context) error {
	// Parse ID dari URL parameter
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	// Parse request body untuk update data
	var req models.UserUpdateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Validasi input data
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.BadRequestResponse(c, "Validation failed", err.Error())
	}

	// Get ID user yang melakukan update
	updatedBy := c.Get("user_id").(int)

	// Panggil service untuk update dengan business logic
	user, err := h.userService.Update(id, &req, updatedBy)
	if err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "User updated successfully", user)
}

// Delete menangani DELETE /users/:id - Soft delete user
// ALUR:
// 1. Extract ID dari URL
// 2. Validasi ID
// 3. Soft delete via service (set is_active = false)
func (h *UserController) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	// Get ID user yang melakukan delete
	deletedBy := c.Get("user_id").(int)

	// Panggil service untuk soft delete
	// Service akan cek apakah user tidak menghapus dirinya sendiri
	if err := h.userService.Delete(id, deletedBy); err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "User deleted successfully", nil)
}

// UpdateStatus menangani PUT /users/:id/status - Update status user
// ALUR:
// 1. Parse ID dan status dari request
// 2. Validasi status value
// 3. Update status via service
func (h *UserController) UpdateStatus(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	// Parse request body untuk status update
	var req models.StatusUpdateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	updatedBy := c.Get("user_id").(int)

	// Service akan validasi status value (active/inactive/suspended)
	if err := h.userService.UpdateStatus(id, req.Status, updatedBy); err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "User status updated successfully", nil)
}

// ResetPassword menangani PUT /users/:id/reset-password - Reset password user
func (h *UserController) ResetPassword(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req models.PasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Validasi password requirements
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.BadRequestResponse(c, "Validation failed", err.Error())
	}

	updatedBy := c.Get("user_id").(int)

	// Service akan hash password baru dan simpan ke database
	if err := h.userService.ResetPassword(id, req.NewPassword, updatedBy); err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Password reset successfully", nil)
}

// === ROLE MANAGEMENT METHODS ===

// GetUserRoles menangani GET /users/:id/roles - Ambil role user
func (h *UserController) GetUserRoles(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	// Service akan query dari table user_roles dengan join ke roles
	roles, err := h.userService.GetUserRoles(id)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get user roles", err.Error())
	}

	return utils.SuccessResponse(c, "User roles retrieved successfully", roles)
}

// AssignRoles menangani POST /users/:id/roles - Assign role ke user
func (h *UserController) AssignRoles(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req models.RoleAssignmentRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	assignedBy := c.Get("user_id").(int)

	// Service akan insert ke table user_roles dengan transaction
	if err := h.userService.AssignRoles(id, req.RoleIDs, assignedBy); err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Roles assigned successfully", nil)
}

// RemoveRoles menangani DELETE /users/:id/roles - Remove role dari user
func (h *UserController) RemoveRoles(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req models.RoleRemovalRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	removedBy := c.Get("user_id").(int)

	// Service akan soft delete (set is_active = false) di user_roles
	if err := h.userService.RemoveRoles(id, req.RoleIDs, removedBy); err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Roles removed successfully", nil)
}

// GetUserPermissions menangani GET /users/:id/permissions - Ambil permission user
func (h *UserController) GetUserPermissions(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	// Service akan query permissions melalui user_roles -> role_permissions -> permissions
	permissions, err := h.userService.GetUserPermissions(id)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get user permissions", err.Error())
	}

	return utils.SuccessResponse(c, "User permissions retrieved successfully", permissions)
}

// GetUserActivities menangani GET /users/:id/activities - Ambil activity log user
func (h *UserController) GetUserActivities(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var pagination models.PaginationRequest
	if err := c.Bind(&pagination); err != nil {
		return utils.BadRequestResponse(c, "Invalid pagination parameters", err.Error())
	}

	// Service akan query dari activity_logs table dengan filter user_id
	response, err := h.userService.GetUserActivities(id, &pagination)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get user activities", err.Error())
	}

	return utils.SuccessResponse(c, "User activities retrieved successfully", response)
}
