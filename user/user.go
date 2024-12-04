package user

type User struct {
	Id				int32		
	Username        string `json:"username"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
	Email           string `json:"email"`
}