package dto

import (
	"gin_scaffold/public"
	"time"

	"github.com/gin-gonic/gin"
)

type AdminInfoOutput struct {
	Id           int       `json:"Id"`
	AdminName    string    `json:"name" `
	LoginTime    time.Time `json:"Logintime"`
	Avatar       string    `json:"Avatar"`
	Introduction string    `json:"Introduction"`
	Roles        []string  `json:"Roles"`
}

type ChangePwdInput struct {
	Psw string `json:"psw" form:"psw" comment:"密码" example:"3122351092" validate:"required"`
}

// 绑定并校验参数
func (param *ChangePwdInput) BindingValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
