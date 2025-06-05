package utils

import (
	"errors"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validations
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("phone", validatePhone)
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errors[field] = field + " is required"
			case "email":
				errors[field] = "Invalid email format"
			case "min":
				errors[field] = field + " must be at least " + e.Param() + " characters"
			case "max":
				errors[field] = field + " cannot exceed " + e.Param() + " characters"
			case "password":
				errors[field] = "Password must contain at least 8 characters with uppercase, lowercase, number and special character"
			case "username":
				errors[field] = "Username can only contain letters, numbers, and underscores"
			case "phone":
				errors[field] = "Invalid phone number format"
			default:
				errors[field] = "Invalid " + field
			}
		}
	}

	return errors
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return matched
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// Indonesian phone number format
	matched, _ := regexp.MatchString(`^(\+62|62|0)[0-9]{9,12}$`, phone)
	return matched
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	if len(username) > 50 {
		return errors.New("username cannot exceed 50 characters")
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	if !matched {
		return errors.New("username can only contain letters, numbers, and underscores")
	}

	return nil
}

func ValidatePhone(phone string) error {
	// Indonesian phone number validation
	matched, _ := regexp.MatchString(`^(\+62|62|0)[0-9]{9,12}$`, phone)
	if !matched {
		return errors.New("invalid Indonesian phone number format")
	}
	return nil
}
