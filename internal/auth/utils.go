package auth

import "golang.org/x/crypto/bcrypt"

// returns nil if password matches, error if not
func ValidatePassword(password string, hashedPassword []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}

func HashPassword(password string) (*[]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &hashedPassword, nil
}
