package auth

type User struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

type signUpRequest struct {
	Username       string `json:"username"`
	RepeatPassword string `json:"repeatPassword"`
	Password       string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
