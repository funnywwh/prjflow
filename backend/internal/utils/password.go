package utils

import (
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength 验证密码强度：必须包含大小写字母和数字，至少6位
func ValidatePasswordStrength(password string) error {
	if len(password) < 6 {
		return &PasswordValidationError{Message: "密码长度至少6位"}
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	if !hasUpper {
		return &PasswordValidationError{Message: "密码必须包含至少一个大写字母"}
	}
	if !hasLower {
		return &PasswordValidationError{Message: "密码必须包含至少一个小写字母"}
	}
	if !hasDigit {
		return &PasswordValidationError{Message: "密码必须包含至少一个数字"}
	}

	return nil
}

// PasswordValidationError 密码验证错误
type PasswordValidationError struct {
	Message string
}

func (e *PasswordValidationError) Error() string {
	return e.Message
}

