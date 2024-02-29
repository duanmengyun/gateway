package controller

import (
	"encoding/json"
	"fmt"
	"gin_scaffold/dao"
	"gin_scaffold/dto"
	"gin_scaffold/middleware"
	"gin_scaffold/public"

	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminController struct{}

func AdminRegister(g *gin.RouterGroup) {
	param := AdminController{}
	//展示用户信息的方法
	g.GET("/admin_info", param.GetAdminInfo)
	g.GET("/changepsw", param.ChangeAdminPsw)
}

// ListPage godoc
// @Summary 获取管理员信息
// @Description 获取管理员信息
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (info *AdminController) GetAdminInfo(c *gin.Context) {
	//获取用户信息,需要获取当前用户的信息，怎么知道当前用户的id呢？根据session来获取
	sess := sessions.Default(c)
	sessionget := sess.Get(public.AdminSessionInfoKey)
	//超时或者退出会删除session，超时由redis的过期时间自动管理，退出时由系统设置删除
	sessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessionget)), sessionInfo); err != nil {
		middleware.ResponseError(c, 2004, err)
	}
	//sessionInfo是session数据
	//admin := &dao.Admin{Id: sessionInfo.Id, Name: sessionInfo.AdminName}
	out := &dto.AdminInfoOutput{Id: sessionInfo.ID, AdminName: sessionInfo.UserName, LoginTime: sessionInfo.LoginTime, Avatar: "i dont know where to find it", Introduction: "I am duanmengyun", Roles: []string{"admin"}}
	middleware.ResponseSuccess(c, out)
}

// ListPage godoc
// @Summary 修改用户密码
// @Description 修改用户密码
// @Tags 管理员接口
// @ID /admin/changepsw
// @Accept  json
// @Produce  json
// @Param body body dto.ChangePwdInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/changepsw [post]
func (info *AdminController) ChangeAdminPsw(c *gin.Context) {
	para := &dto.ChangePwdInput{}
	if err := para.BindingValidParams(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	//获取用户信息,需要获取当前用户的信息，怎么知道当前用户的id呢？根据session来获取
	sess := sessions.Default(c)
	sessionget := sess.Get(public.AdminSessionInfoKey)
	//超时或者退出会删除session，超时由redis的过期时间自动管理，退出时由系统设置删除
	sessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessionget)), sessionInfo); err != nil {
		middleware.ResponseError(c, 2004, err)
	}
	//获取db连接
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//从db中读取信息并修改
	admin := &dao.Admin{}
	admin, err = admin.FindAdmin(c, db, &dao.Admin{UserName: sessionInfo.UserName})
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}
	//生成新密码
	admin.Password = public.GenSaltpsw(para.Psw, admin.Salt)
	//写回到数据库
	err = admin.Save(c, db)
	if err != nil {
		middleware.ResponseError(c, 2006, err)
		return
	}

	middleware.ResponseSuccess(c, "")
}
