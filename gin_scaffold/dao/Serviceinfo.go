package dao

import (
	"gin_scaffold/dto"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Serviceinfo struct {
	ID          int64     `json:"id" gorm:"primary_key"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	UpdatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"更新时间"`
	CreatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"添加时间"`
	IsDelete    int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (s *Serviceinfo) TableName() string {
	return "gateway_service_info"
}

func (s *Serviceinfo) FindService(c *gin.Context, tx *gorm.DB, search *Serviceinfo) (*Serviceinfo, error) {
	service := &Serviceinfo{}
	err := tx.WithContext(c).Where(search).Find(service).Error
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Serviceinfo) Save(c *gin.Context, db *gorm.DB) error {
	if err := db.WithContext(c).Save(s).Error; err != nil {
		return err
	}
	return nil
}

func (s *Serviceinfo) PageList(c *gin.Context, db *gorm.DB, param *dto.ServiceListInput) ([]Serviceinfo, int64, error) {
	total := int64(0)
	pagelist := []Serviceinfo{}
	offset := (param.PageNumber - 1) * param.PageSize
	//where是模糊查询
	query := db.WithContext(c)
	query = query.Table(s.TableName()).Where("is_delete=0", "name LIKE %?% or describ like %?%", param.Info, param.Info)
	if err := query.Limit(param.PageSize).Offset(offset).Order("id desc").Find(&pagelist).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	query.Limit(param.PageSize).Offset(offset).Count(&total)
	return pagelist, total, nil
}

func (s *Serviceinfo) ServiceDetial(c *gin.Context, tx *gorm.DB, search *Serviceinfo) (*ServiceDetial, error) {
	httprule := &HttpRule{ID: search.ID}
	httprule, err := httprule.Find(c, tx, httprule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	tcprule := &TcpRule{ID: search.ID}
	tcprule, err = tcprule.Find(c, tx, tcprule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	grpcrule := &GrpcRule{ID: search.ID}
	grpcrule, err = grpcrule.Find(c, tx, grpcrule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	loadbalance := &LoadBalance{ID: search.ID}
	loadbalance, err = loadbalance.Find(c, tx, loadbalance)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	accesscontrol := &AcccessControll{ID: search.ID}
	accesscontrol, err = accesscontrol.Find(c, tx, accesscontrol)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	res := &ServiceDetial{
		Info:          search,
		HTTPRule:      httprule,
		TCPRule:       tcprule,
		GRPCRule:      grpcrule,
		AccessControl: accesscontrol,
		LoadBalance:   loadbalance,
	}
	return res, nil
}
