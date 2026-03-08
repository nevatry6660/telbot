package util

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidPhone = errors.New("invalid phone number")

func NormalizePhone(input string) (local, full string) {

	clean := strings.NewReplacer("+", "", "-", "", " ", "").Replace(input)

	switch {
	case strings.HasPrefix(clean, "62"):
		local = clean[2:]
	case strings.HasPrefix(clean, "0"):
		local = clean[1:]
	default:
		local = clean
	}

	full = "62" + local
	return
}

func ValidatePhone(input string) (local, full string, err error) {
	local, full = NormalizePhone(input)

	if len(local) < 9 || len(local) > 13 {
		return "", "", ErrInvalidPhone
	}

	for _, c := range local {
		if !unicode.IsDigit(c) {
			return "", "", ErrInvalidPhone
		}
	}

	return local, full, nil
}

func IsPhoneNumber(s string) bool {
	clean := strings.NewReplacer("+", "", "-", "", " ", "").Replace(s)
	if len(clean) < 9 || len(clean) > 15 {
		return false
	}
	for _, c := range clean {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}
