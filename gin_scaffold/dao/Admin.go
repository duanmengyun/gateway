package dao

import (
	"errors"
	"gin_scaffold/dto"
	"gin_scaffold/public"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Admin struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	UserName  string    `json:"user_name" gorm:"column:user_name" description:"管理员用户名"`
	Salt      string    `json:"salt" gorm:"column:salt" description:"盐"`
	Password  string    `json:"password" gorm:"column:password" description:"密码"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

func (a *Admin) TableName() string {
	return "gateway_admin"
}

// select
func (a *Admin) FindAdmin(c *gin.Context, tx *gorm.DB, search *Admin) (*Admin, error) {
	admin := &Admin{}
	err := tx.WithContext(c).Where(search).Find(admin).Error
	if err != nil {
		return nil, err
	}
	return admin, nil
}

// 登录校验信息
func (a *Admin) LoginCheck(c *gin.Context, db *gorm.DB, user *dto.AdminLoginInput) (*Admin, error) {
	admin, err := a.FindAdmin(c, db, (&Admin{UserName: user.UserName, IsDelete: 0}))
	if err != nil {
		return nil, errors.New("用户信息不存在")
	}
	if public.GenSaltpsw(user.Password, admin.Salt) != admin.Password {
		//密码不匹配
		return nil, errors.New("密码错误:请重新输入")
	}
	return admin, nil
}

// 修改数据库中的admin信息
func (a *Admin) Save(c *gin.Context, db *gorm.DB) error {
	if err := db.WithContext(c).Save(a).Error; err != nil {
		return err
	}
	return nil
}
