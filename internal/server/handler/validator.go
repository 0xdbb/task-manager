package handler

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// strongPassword is a custom validation function for strong passwords
var StrongPassword validator.Func = func(fieldLevel validator.FieldLevel) bool {
	password := fieldLevel.Field().String()
	// Define the criteria for a strong password
	var (
		minLength    = 8
		hasUppercase = regexp.MustCompile(`[A-Z]`).MatchString
		hasLowercase = regexp.MustCompile(`[a-z]`).MatchString
		hasNumber    = regexp.MustCompile(`[0-9]`).MatchString
		hasSpecial   = regexp.MustCompile(`[!@#\$%\^&\*]`).MatchString
	)

	if len(password) < minLength {
		return false
	}
	if !hasUppercase(password) || !hasLowercase(password) || !hasNumber(password) || !hasSpecial(password) {
		return false
	}
	return true
}
