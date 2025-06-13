package model

type USER_CREDENTIAL struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	SecondPassword string `json:"second_password"`
	YFlag          string `json:"y_flag"`
}
