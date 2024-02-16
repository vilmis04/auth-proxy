package auth

type signUpRequest struct {
	Username       string `json:"username"`
	RepeatPassword string `json:"repeatPassword"`
	Password       string `json:"password"`
}
