package dao

import (
	"gin_scaffold/dto"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id	"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称	"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间	"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (app *App) TableName() string {
	return "gateway_apps"
}
func (app *App) Find(c *gin.Context, tx *gorm.DB, search *App) (*App, error) {
	model := &App{}
	err := tx.WithContext(c).Where(search).Find(model).Error
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (app *App) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.WithContext(c).Save(app).Error; err != nil {
		return err
	}
	return nil
}

func (app *App) AppList(c *gin.Context, tx *gorm.DB, search *dto.AppListInput) ([]App, int64, error) {
	total := int64(0)
	pagelist := []App{}
	offset := (search.PageNo - 1) * search.PageSize
	//where是模糊查询
	query := tx.WithContext(c)
	query = query.Table(app.TableName()).Where("is_delete=0", "name LIKE %?% or describ like %?%", search.Info, search.Info)
	if err := query.Limit(search.PageSize).Offset(offset).Order("id desc").Find(&pagelist).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	query.Limit(search.PageSize).Offset(offset).Count(&total)
	return pagelist, total, nil
}
