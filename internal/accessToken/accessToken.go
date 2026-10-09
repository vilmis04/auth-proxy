package accessToken

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const ACCESS_TOKEN = "access_token"

// MinKeyLength is the minimum HS256 secret size in bytes.
const MinKeyLength = 32

// clockLeeway tolerates small clock differences when validating exp/iat.
const clockLeeway = 30 * time.Second

// Signer creates and validates access tokens. Build it once at startup.
type Signer struct {
	key []byte
	ttl time.Duration
	now func() time.Time
}

// NewSigner fails if the key is shorter than MinKeyLength bytes or ttl is not positive.
func NewSigner(key string, ttl time.Duration) (*Signer, error) {
	if len(key) < MinKeyLength {
		return nil, fmt.Errorf("JWT_KEY must be at least %d bytes, got %d", MinKeyLength, len(key))
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("token lifetime must be positive, got %v", ttl)
	}

	return &Signer{key: []byte(key), ttl: ttl, now: time.Now}, nil
}

// TTL is the lifetime of issued tokens; the cookie max age should match it.
func (s *Signer) TTL() time.Duration {
	return s.ttl
}

func (s *Signer) Create(username string) (*string, error) {
	now := s.now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": username,
			"iat": jwt.NewNumericDate(now),
			"exp": jwt.NewNumericDate(now.Add(s.ttl)),
		})
	signedToken, err := token.SignedString(s.key)
	if err != nil {
		return nil, err
	}

	return &signedToken, nil
}

func (s *Signer) Validate(tokenCookie string) (*string, error) {
	token, err := jwt.Parse(tokenCookie,
		func(t *jwt.Token) (interface{}, error) { return s.key, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(clockLeeway),
		jwt.WithTimeFunc(s.now),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("jwt token is not valid")
	}
	user, err := token.Claims.GetSubject()
	if err != nil {
		return nil, err
	}
	if user == "" {
		return nil, fmt.Errorf("jwt token has no subject")
	}

	return &user, nil
}
