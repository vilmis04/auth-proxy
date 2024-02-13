package auth

type signUpRequest struct {
	Username       string `json:"username"`
	RepeatUsername string `json:"repeatUsername"`
	Password       string `json:"password"`
}
