package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// init initializes the validator package with a new validator instance
// and registers custom validations.
func init() {
	validate = validator.New()

	// Custom validations can be added here
	validate.RegisterValidation("strong_password", validateStrongPassword)
}

// ValidateStruct validates the given struct using the validator package
// and returns a slice of string errors or an empty slice if the struct is valid.
//
// The struct must have struct tags that match the validator package's
// validation tags. The most common tags are "required", "email", "min", and
// "max". Additionally, the "strong_password" tag can be used to validate
// passwords against the complexity requirements of at least 8 characters,
// one uppercase letter, one lowercase letter, one digit, and one special
// character.
//
// The returned errors are human-readable and can be used to display the
// validation errors to the user.
func ValidateStruct(s interface{}) []string {
	var errors []string

	err := validate.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return []string{"Invalid validation"}
		}

		for _, err := range err.(validator.ValidationErrors) {
			var errorMessage string
			switch err.Tag() {
			case "required":
				errorMessage = fmt.Sprintf("%s is required", err.Field())
			case "email":
				errorMessage = fmt.Sprintf("%s must be a valid email", err.Field())
			case "min":
				errorMessage = fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param())
			case "max":
				errorMessage = fmt.Sprintf("%s must be at most %s characters", err.Field(), err.Param())
			case "strong_password":
				errorMessage = fmt.Sprintf("%s does not meet password complexity requirements", err.Field())
			default:
				errorMessage = fmt.Sprintf("%s is invalid", err.Field())
			}
			errors = append(errors, errorMessage)
		}
	}

	return errors
}

// validateStrongPassword checks if a password is strong enough.
//
// A strong password is at least 8 characters long and contains at least one
// uppercase letter, one lowercase letter, one digit, and one special character.
// The function returns true if the password is valid, false otherwise.
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check minimum length
	if len(password) < 8 {
		return false
	}

	// Check for at least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return false
	}

	// Check for at least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return false
	}

	// Check for at least one digit
	if !regexp.MustCompile(`\d`).MatchString(password) {
		return false
	}

	// Check for at least one special character
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return false
	}

	return true
}

// SanitizeInput removes any leading or trailing whitespace from a given string.
// It is meant to be used when accepting user input to prevent any malicious
// or accidental whitespace from causing issues in the application.
func SanitizeInput(input string) string {
	return strings.TrimSpace(input)
}

// IsValidEmail checks if a given email is valid according to a regular expression
// that matches most common email address formats.
//
// The regular expression used is:
//
//	^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$
//
// It matches email addresses with the following characteristics:
//   - The local part (before the @) can contain alphanumeric characters, dot,
//     underscore, percent, plus, hyphen, and equals.
//   - The local part must be at least 1 character long.
//   - The domain part (after the @) can contain alphanumeric characters, dot,
//     and hyphen.
//   - The domain part must be at least 2 characters long.
//   - The domain part must end with at least one top-level domain (TLD) that is
//     at least 2 characters long.
//
// This regular expression does not validate all possible valid email addresses
// as specified in the RFC 5322 specification. It is meant to be a simple and
// fast way to validate most common email address formats.
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return emailRegex.MatchString(email)
}
