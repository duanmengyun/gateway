package controller

import (
	"encoding/json"
	"gin_scaffold/dao"
	"gin_scaffold/dto"
	"gin_scaffold/middleware"
	"gin_scaffold/public"
	"time"

	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminLoginController struct {
}

// 编写绑定login对应的action，每个方法就是对应的action
func AdminLoginRegister(router *gin.RouterGroup) {
	adminlogin := &AdminLoginController{}
	//登录的action使用post比较安全
	router.POST("/login", adminlogin.AdminLogin)
	router.GET("/logout", adminlogin.Logout)
}

// AdminLogin godoc
// @Summary 管理员登录
// @Description 管理员输入账号密码进入系统
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminlogin *AdminLoginController) AdminLogin(c *gin.Context) {
	//先查询是否在数据库中，然后判断账号密码是否正确,从dto获取输入的信息
	//1.校验清华求参数，后端在请求前端数据的时候以json的格式来请求
	para := &dto.AdminLoginInput{}
	if err := para.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	//2.从MySQL中读取数据
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	admin := &dao.Admin{}
	if admin, err = admin.LoginCheck(c, tx, para); err != nil {
		//要有弹出信息提示
		middleware.ResponseError(c, 2002, err)
		return
	}
	//设置session
	sessInfo := &dto.AdminSessionInfo{
		ID:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}
	sessBts, err := json.Marshal(sessInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	sess := sessions.Default(c)
	sess.Set(public.AdminSessionInfoKey, string(sessBts))
	sess.Save()

	out := &dto.AdminLoginOutput{Token: admin.UserName}
	middleware.ResponseSuccess(c, out)
}

// AdminLogin godoc
// @Summary 管理员退出
// @Description 管理员退出
// @Tags 管理员接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (adminlogin *AdminLoginController) Logout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Delete(public.AdminSessionInfoKey)
	sess.Save()
	middleware.ResponseSuccess(c, "")
}
