package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hashedPassword, err := HashPassword("test")
	if err != nil {
		t.Error(err)
	}
	if hashedPassword == nil {
		t.Errorf("failed to hash password")
	}
}

func TestValidatePassword(t *testing.T) {
	password := "test"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Error(err)
	}

	err = ValidatePassword(password, *hashedPassword)
	if err != nil {
		t.Error(err)
	}

	err = ValidatePassword("failing_test", *hashedPassword)
	if err == nil {
		t.Errorf("function return false positive")
	}
}
