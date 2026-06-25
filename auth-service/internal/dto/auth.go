package dto

type UserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,gte=6"`
}

type SignUpResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type SignInResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
