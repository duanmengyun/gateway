package controller

import (
	"errors"
	"gin_scaffold/dao"
	"gin_scaffold/dto"
	"gin_scaffold/middleware"
	"gin_scaffold/public"

	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
)

type AppController struct{}

func AppRegister(router *gin.RouterGroup) {
	app := &AppController{}
	router.GET("/app_list", app.AppList)
	router.GET("/app_detail", app.AppDetail)
	router.GET("/app_delete", app.AppDelete)
	router.POST("/app_add", app.AppAdd)
	router.POST("/app_update", app.AppUpdate)
	router.POST("/app_stat", app.AppStat)
}

// AppList godoc
// @Summary 获取租户列表
// @Description 获取租户列表
// @Tags 租户接口
// @ID /app/app_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query string true "每页多少条"
// @Param page_no query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.AppListOutput} "success"
// @Router /app/app_list [get]
func (app *AppController) AppList(c *gin.Context) {
	params := &dto.AppListInput{}
	if err := params.BindingValidParams(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
	}
	appinfo := &dao.App{}
	list, total, err := appinfo.AppList(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
	}
	outputList := []dto.AppListItemOutput{}
	for _, item := range list {
		appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + item.AppID)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		outputList = append(outputList, dto.AppListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			Qps:      item.Qps,
			RealQpd:  appCounter.TotalCount,
			RealQps:  appCounter.QPS,
		})
	}
	output := dto.AppListOutput{
		List:  outputList,
		Total: total,
	}
	middleware.ResponseSuccess(c, output)
	return
}

// AppDetail godoc
// @Summary 获取租户具体信息
// @Description 获取租户具体信息
// @Tags 租户接口
// @ID /app/app_detail
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Success 200 {object} middleware.Response{data=dto.App} "success"
// @Router /app/app_detail [get]
func (app *AppController) AppDetail(c *gin.Context) {
	param := &dto.AppDeleteInput{}
	if err := param.BindingValidParams(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
	}
	appinfo := &dao.App{ID: param.ID}
	if appinfo, err = appinfo.Find(c, tx, appinfo); err != nil {
		middleware.ResponseError(c, 2002, err)
	}
	middleware.ResponseSuccess(c, appinfo)
}

// AppDelete godoc
// @Summary 删除租户
// @Description 删除租户
// @Tags 租户接口
// @ID /app/app_delete
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_delete [get]
func (app *AppController) AppDelete(c *gin.Context) {
	param := &dto.AppDeleteInput{}
	if err := param.BindingValidParams(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
	}
	appinfo := &dao.App{ID: param.ID}
	if appinfo, err = appinfo.Find(c, tx, appinfo); err != nil {
		middleware.ResponseError(c, 2002, err)
	}
	appinfo.IsDelete = 1
	if err = appinfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
	}
	middleware.ResponseSuccess(c, "success")
}

// AppAdd godoc
// @Summary 添加租户
// @Description 添加租户
// @Tags 租户接口
// @ID /app/app_add
// @Accept  json
// @Produce  json
// @Param page_no query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.APPAddHttpInput} "success"
// @Router /app/app_add [post]
func (app *AppController) AppAdd(c *gin.Context) {
	params := &dto.APPAddHttpInput{}
	if err := params.BindingValidParams(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}

	//验证app_id是否被占用
	search := &dao.App{
		AppID: params.AppID,
	}
	if _, err := search.Find(c, lib.GORMDefaultPool, search); err == nil {
		middleware.ResponseError(c, 2002, errors.New("租户ID被占用，请重新输入"))
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	tx := lib.GORMDefaultPool
	info := &dao.App{
		AppID:    params.AppID,
		Name:     params.Name,
		Secret:   params.Secret,
		WhiteIPS: params.WhiteIPS,
		Qps:      params.Qps,
		Qpd:      params.Qpd,
	}
	if err := info.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// AppUpdate godoc
// @Summary 更新租户信息
// @Description 更新租户信息
// @Tags 租户接口
// @ID /app/app_update
// @Accept  json
// @Produce  json
// @Param page_no query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.AppListOutput} "success"
// @Router /app/app_update [post]
func (app *AppController) AppUpdate(c *gin.Context) {
	params := &dto.APPUpdateHttpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
	}
	info, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	info.Name = params.Name
	info.Secret = params.Secret
	info.WhiteIPS = params.WhiteIPS
	info.Qps = params.Qps
	info.Qpd = params.Qpd
	if err := info.Save(c, lib.GORMDefaultPool); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// AppUpdate godoc
// @Summary 租户流量统计
// @Description 更新租户信息
// @Tags 租户接口
// @ID /app/app_stat
// @Accept  json
// @Produce  json
// @Param page_no query string true "租户id"
// @Success 200 {object} middleware.Response{data=dto.AppStatOutput} "success"
// @Router /app/app_stat [get]
func (app *AppController) AppStat(c *gin.Context) {
	param := &dto.AppDeleteInput{}
	if err := param.BindingValidParams(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
	}
	appinfo := &dao.App{ID: param.ID}
	if appinfo, err = appinfo.Find(c, tx, appinfo); err != nil {
		middleware.ResponseError(c, 2002, err)
	}
	middleware.ResponseSuccess(c, &dto.AppStatOutput{})
}
