package controller

import (
	"errors"
	"fmt"
	"gin_scaffold/dao"
	"gin_scaffold/dto"
	"gin_scaffold/middleware"
	"gin_scaffold/public"
	"strings"
	"time"

	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
)

type ServiceController struct{}

func ServiceRegister(router *gin.RouterGroup) {
	service := &ServiceController{}
	router.GET("/service_list", service.ServiceList)
	router.GET("/delete_service", service.DeleteService)
	router.GET("/service_detial", service.ServiceDetial)
	router.POST("/create_http_service", service.CreateHTTPService)
	router.POST("/create_tcp_service", service.CreateTcpService)
	router.POST("/create_grpc_service", service.CreateGrpcService)
	router.POST("/update_http_service", service.UpdateHTTPService)
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_no query int true "当前页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/service_list [get]
func (service *ServiceController) ServiceList(c *gin.Context) {
	params := &dto.ServiceListInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//从db中分页读取基本信息
	serviceInfo := &dao.Serviceinfo{}
	list, total, err := serviceInfo.PageList(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}

	//格式化输出信息
	outList := []dto.ServiceListItemOutput{}
	for _, listItem := range list {
		serviceDetail, err := listItem.ServiceDetial(c, tx, &listItem)
		if err != nil {
			middleware.ResponseError(c, 2007, err)
			return
		}
		//1、http后缀接入 clusterIP+clusterPort+path
		//2、http域名接入 domain
		//3、tcp、grpc接入 clusterIP+servicePort
		serviceAddr := "unknown"
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}
		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeGrpc {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		}
		ipList := serviceDetail.LoadBalance.GetIPlistByModel()
		outItem := dto.ServiceListItemOutput{
			ID:          listItem.ID,
			LoadType:    listItem.LoadType,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         0,
			Qpd:         0,
			TotalNode:   len(ipList),
		}
		outList = append(outList, outItem)
	}
	out := &dto.ServiceListOutput{
		Total: total,
		List:  outList,
	}
	middleware.ResponseSuccess(c, out)
}

// ServiceList godoc
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /service/service_detial
// @Accept  json
// @Produce  json
// @Param info query string false "id"
// @Success 200 {object} middleware.Response{data=dto.ServiceDetial} "success"
// @Router /service/service_list [get]
func (service *ServiceController) ServiceDetial(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.Serviceinfo{}
	serviceInfo.ID = params.ID
	serviceInfo, err = serviceInfo.FindService(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}
	servicedetial, err := serviceInfo.ServiceDetial(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2006, err)
		return
	}
	middleware.ResponseSuccess(c, servicedetial)
}

// DeleteService godoc
// @Summary 删除服务
// @Description 删除服务
// @Tags 服务管理
// @ID /service/delete_service
// @Accept  json
// @Produce  json
// @Param id query string true "服务id"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/delete_service [get]
func (service *ServiceController) DeleteService(c *gin.Context) {
	param := &dto.ServiceDeleteInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceinfo := &dao.Serviceinfo{ID: param.ID}
	serviceinfo, err = serviceinfo.FindService(c, tx, serviceinfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	serviceinfo.IsDelete = 1
	if err = serviceinfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "success")
}

// CreateHTTPService godoc
// @Summary 创建http服务
// @Description 创建http服务
// @Tags 服务管理
// @ID /service/create_http_service
// @Accept  json
// @Produce  json
// @Param body body dto.CreateHTTPServiceInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/create_http_service [post]
func (service *ServiceController) CreateHTTPService(c *gin.Context) {
	param := &dto.CreateHTTPServiceInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	if len(strings.Split(param.IpList, ",")) != len(strings.Split(param.WeightList, ",")) {
		middleware.ResponseError(c, 2004, errors.New("ip list wrong"))
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	tx = tx.Begin()
	//先创建serviceinfo还是先创建httprule，因为主键是确定的 ,用事务开始!!!
	serviceinfo := &dao.Serviceinfo{ServiceName: param.ServiceName}
	if _, err = serviceinfo.FindService(c, tx, serviceinfo); err == nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, errors.New("服务已存在"))
		return
	}
	httpUrl := &dao.HttpRule{RuleType: param.RuleType, Rule: param.Rule}
	if _, err = httpUrl.Find(c, tx, httpUrl); err == nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, errors.New("httpurl已存在"))
		return
	}
	servicenew := &dao.Serviceinfo{
		ServiceName: param.ServiceName,
		ServiceDesc: param.ServiceDesc}
	if err := servicenew.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, errors.New("创建服务信息失败"))
		return
	}
	httprule := &dao.HttpRule{
		ServiceID:      servicenew.ID,
		RuleType:       public.LoadTypeHTTP,
		Rule:           param.Rule,
		NeedHttps:      param.NeedHttps,
		NeedWebsocket:  param.NeedWebsocket,
		NeedStripUri:   param.NeedStripUri,
		UrlRewrite:     param.UrlRewrite,
		HeaderTransfor: param.HeaderTransfor,
	}
	if err := httprule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, errors.New("创建httprule信息失败"))
		return
	}
	acesscontrol := &dao.AcccessControll{
		ServiceID:         servicenew.ID,
		OpenAuth:          param.OpenAuth,
		BlackList:         param.BlackList,
		WhiteList:         param.WhiteList,
		ClientIPFlowLimit: param.ClientipFlowLimit,
		ServiceFlowLimit:  param.ServiceFlowLimit,
	}
	if err := acesscontrol.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, errors.New("创建访问控制信息失败"))
		return
	}
	loadbalance := &dao.LoadBalance{
		ServiceID:              servicenew.ID,
		RoundType:              param.RoundType,
		IpList:                 param.IpList,
		WeightList:             param.WeightList,
		UpstreamConnectTimeout: param.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  param.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    param.UpstreamIdleTimeout,
		UpstreamMaxIdle:        param.UpstreamMaxIdle,
	}
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, errors.New("创建负载均衡信息失败"))
		return
	}
	middleware.ResponseSuccess(c, "success")
}

// CreateHTTPService godoc
// @Summary 更新http服务
// @Description 更新http服务
// @Tags 服务管理
// @ID /service/update_http_service
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/update_http_service [post]
func (service *ServiceController) UpdateHTTPService(c *gin.Context) {
	param := &dto.ServiceUpdateHTTPInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	if len(strings.Split(param.IpList, ",")) != len(strings.Split(param.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("IP列表和权重列表不匹配"))
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	tx.Begin()
	serviceinfo := &dao.Serviceinfo{
		ServiceName: param.ServiceName,
	}
	serviceinfo, err = serviceinfo.FindService(c, tx, serviceinfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, err)
		return
	}
	detail, err := serviceinfo.ServiceDetial(c, tx, serviceinfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}
	info := detail.Info
	info.ServiceDesc = param.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}
	httprule := detail.HTTPRule
	httprule.NeedHttps = param.NeedHttps
	httprule.NeedStripUri = param.NeedStripUri
	httprule.NeedWebsocket = param.NeedWebsocket
	httprule.UrlRewrite = param.UrlRewrite
	httprule.HeaderTransfor = param.HeaderTransfor
	if err := httprule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	Loadbalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		Loadbalance = detail.LoadBalance
	}
	Loadbalance.ID = info.ID
	Loadbalance.RoundType = param.RoundType
	Loadbalance.IpList = param.IpList
	Loadbalance.WeightList = param.WeightList
	Loadbalance.UpstreamConnectTimeout = param.UpstreamConnectTimeout
	Loadbalance.UpstreamHeaderTimeout = param.UpstreamHeaderTimeout
	Loadbalance.UpstreamIdleTimeout = param.UpstreamIdleTimeout
	Loadbalance.UpstreamMaxIdle = param.UpstreamMaxIdle
	if err := Loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}
	accesscontrol := &dao.AcccessControll{}
	if detail.AccessControl != nil {
		accesscontrol = detail.AccessControl
	}
	accesscontrol.ServiceID = info.ID
	accesscontrol.OpenAuth = param.OpenAuth
	accesscontrol.BlackList = param.BlackList
	accesscontrol.WhiteList = param.WhiteList
	accesscontrol.ServiceFlowLimit = param.ServiceFlowLimit
	accesscontrol.ClientIPFlowLimit = param.ClientipFlowLimit
	if err := accesscontrol.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "success")
}

// CreateGrpcService godoc
// @Summary 创建grpc服务
// @Description 创建grpc服务
// @Tags 服务管理
// @ID /service/create_grpc_service
// @Accept  json
// @Produce  json
// @Param body body dto.CreateGrpcServiceInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/create_grpc_service [post]
func (service *ServiceController) CreateGrpcService(c *gin.Context) {
	param := &dto.CreateGrpcServiceInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceinfo := &dao.Serviceinfo{
		ServiceName: param.ServiceName,
		IsDelete:    0,
	}
	if _, err := serviceinfo.FindService(c, tx, serviceinfo); err == nil {
		middleware.ResponseError(c, 2002, errors.New("该服务已存在"))
		return
	}
	tcprule := &dao.TcpRule{
		Port: param.Port,
	}
	if _, err := tcprule.Find(c, tx, tcprule); err == nil {
		middleware.ResponseError(c, 2003, errors.New("改端口已被tcp服务占用"))
	}
	grpcrule := &dao.GrpcRule{
		Port: param.Port,
	}
	if _, err := grpcrule.Find(c, tx, grpcrule); err == nil {
		middleware.ResponseError(c, 2004, errors.New("改端口已被grpc服务占用"))
		return
	}
	if len(strings.Split(param.IpList, ",")) != len(strings.Split(param.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("IP列表和权重列表不匹配"))
		return
	}
	tx.Begin()
	info := &dao.Serviceinfo{
		ServiceName: param.ServiceName,
		ServiceDesc: param.ServiceDesc,
		LoadType:    public.LoadTypeGrpc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	grpcnew := &dao.GrpcRule{
		ServiceID:      info.ID,
		Port:           param.Port,
		HeaderTransfor: param.HeaderTransfor,
	}
	if err := grpcnew.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, errors.New("创建grpc信息失败"))
		return
	}
	acesscontrol := &dao.AcccessControll{
		ServiceID:         info.ID,
		OpenAuth:          param.OpenAuth,
		BlackList:         param.BlackList,
		WhiteList:         param.WhiteList,
		ClientIPFlowLimit: param.ClientIPFlowLimit,
		ServiceFlowLimit:  param.ServiceFlowLimit,
	}
	if err := acesscontrol.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, errors.New("创建访问控制信息失败"))
		return
	}
	loadbalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  param.RoundType,
		IpList:     param.IpList,
		WeightList: param.WeightList,
		ForbidList: param.ForbidList,
	}
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2009, errors.New("创建负载均衡信息失败"))
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "success")
}

// UpdateGrpcService godoc
// @Summary 更新grpc服务
// @Description 更新grpc服务
// @Tags 服务管理
// @ID /service/update_grpc_service
// @Accept  json
// @Produce  json
// @Param body body dto.UpdateGrpcServiceInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/update_grpc_service [post]
func (service *ServiceController) UpdateGrpcService(c *gin.Context) {
	param := &dto.UpdateGrpcServiceInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	if len(strings.Split(param.IpList, ",")) != len(strings.Split(param.WeightList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}
	tx.Begin()
	info := &dao.Serviceinfo{
		ID: param.ID,
	}
	detail, err := info.ServiceDetial(c, tx, info)
	if err != nil {
		//说明该服务不存在
		middleware.ResponseError(c, 2003, err)
		return
	}
	info = detail.Info
	info.ServiceDesc = param.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}
	Loadbalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		Loadbalance = detail.LoadBalance
	}
	Loadbalance.ServiceID = info.ID
	Loadbalance.RoundType = param.RoundType
	Loadbalance.IpList = param.IpList
	Loadbalance.WeightList = param.WeightList
	Loadbalance.ForbidList = param.ForbidList
	if err := Loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}
	grpcrule := &dao.GrpcRule{}
	if detail.TCPRule != nil {
		grpcrule = detail.GRPCRule
	}
	grpcrule.ServiceID = info.ID
	grpcrule.HeaderTransfor = param.HeaderTransfor
	if err := grpcrule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	accesscontrol := &dao.AcccessControll{}
	if detail.AccessControl != nil {
		accesscontrol = detail.AccessControl
	}
	accesscontrol.ServiceID = info.ID
	accesscontrol.OpenAuth = param.OpenAuth
	accesscontrol.BlackList = param.BlackList
	accesscontrol.WhiteHostName = param.WhiteHostName
	accesscontrol.WhiteList = param.WhiteList
	accesscontrol.ClientIPFlowLimit = param.ClientIPFlowLimit
	accesscontrol.ServiceFlowLimit = param.ServiceFlowLimit
	if err := accesscontrol.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "success")
}

// CreateTcpService godoc
// @Summary 创建tcp服务
// @Description 创建tcp服务
// @Tags 服务管理
// @ID /service/create_tcp_service
// @Accept  json
// @Produce  json
// @Param body body dto.CreateTcpServiceInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/create_tcp_service [post]
func (service *ServiceController) CreateTcpService(c *gin.Context) {
	param := &dto.CreateTcpServiceInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceinfo := &dao.Serviceinfo{
		ServiceName: param.ServiceName,
		IsDelete:    0,
	}
	if _, err := serviceinfo.FindService(c, tx, serviceinfo); err != nil {
		middleware.ResponseError(c, 2002, errors.New("该服务已存在"))
	}
	grpcrule := &dao.GrpcRule{
		Port: param.Port,
	}
	if _, err := grpcrule.Find(c, tx, grpcrule); err == nil {
		middleware.ResponseError(c, 2003, errors.New("改端口已被grpc服务占用"))
		return
	}
	tcprule := &dao.TcpRule{
		Port: param.Port,
	}
	if _, err := tcprule.Find(c, tx, tcprule); err == nil {
		middleware.ResponseError(c, 2004, errors.New("改端口已被tcp服务占用"))
	}
	if len(strings.Split(param.IpList, ",")) != len(strings.Split(param.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("IP列表和权重列表不匹配"))
		return
	}
	tx.Begin()
	info := &dao.Serviceinfo{
		ServiceName: param.ServiceName,
		ServiceDesc: param.ServiceDesc,
		LoadType:    public.LoadTypeTCP,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	grpcnew := &dao.TcpRule{
		ServiceID: info.ID,
		Port:      param.Port,
	}
	if err := grpcnew.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, errors.New("创建tcp信息失败"))
		return
	}
	acesscontrol := &dao.AcccessControll{
		ServiceID:         info.ID,
		OpenAuth:          param.OpenAuth,
		BlackList:         param.BlackList,
		WhiteList:         param.WhiteList,
		ClientIPFlowLimit: param.ClientIPFlowLimit,
		ServiceFlowLimit:  param.ServiceFlowLimit,
	}
	if err := acesscontrol.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, errors.New("创建访问控制信息失败"))
		return
	}
	loadbalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  param.RoundType,
		IpList:     param.IpList,
		WeightList: param.WeightList,
		ForbidList: param.ForbidList,
	}
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2009, errors.New("创建负载均衡信息失败"))
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "success")
}

// UpdateTcpService godoc
// @Summary 更新tcp服务
// @Description 更新tcp服务
// @Tags 服务管理
// @ID /service/update_tcp_service
// @Accept  json
// @Produce  json
// @Param body body dto.UpdateTcpServiceInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/update_tcp_service [post]
func (service *ServiceController) UpdateTcpService(c *gin.Context) {
	param := &dto.UpdateTcpServiceInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	if len(strings.Split(param.IpList, ",")) != len(strings.Split(param.WeightList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}
	tx.Begin()
	info := &dao.Serviceinfo{
		ID: param.ID,
	}
	detail, err := info.ServiceDetial(c, tx, info)
	if err != nil {
		//说明该服务不存在
		middleware.ResponseError(c, 2003, err)
		return
	}
	info = detail.Info
	info.ServiceDesc = param.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}
	Loadbalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		Loadbalance = detail.LoadBalance
	}
	Loadbalance.ServiceID = info.ID
	Loadbalance.RoundType = param.RoundType
	Loadbalance.IpList = param.IpList
	Loadbalance.WeightList = param.WeightList
	Loadbalance.ForbidList = param.ForbidList
	if err := Loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}
	tcprule := &dao.TcpRule{}
	if detail.TCPRule != nil {
		tcprule = detail.TCPRule
	}
	tcprule.ServiceID = info.ID
	tcprule.Port = param.Port
	if err := tcprule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	accesscontrol := &dao.AcccessControll{}
	if detail.AccessControl != nil {
		accesscontrol = detail.AccessControl
	}
	accesscontrol.ServiceID = info.ID
	accesscontrol.OpenAuth = param.OpenAuth
	accesscontrol.BlackList = param.BlackList
	accesscontrol.WhiteHostName = param.WhiteHostName
	accesscontrol.WhiteList = param.WhiteList
	accesscontrol.ClientIPFlowLimit = param.ClientIPFlowLimit
	accesscontrol.ServiceFlowLimit = param.ServiceFlowLimit
	if err := accesscontrol.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "success")
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /service/service_stat
// @Accept  json
// @Produce  json
// @Param id query string true "服务id"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/service_stat [post]
func (service *ServiceController) ServiceStat(c *gin.Context) {
	param := &dto.ServiceDeleteInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}
	//怎么计算流量
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
	}
	serviceinfo := &dao.Serviceinfo{ID: param.ID}
	if serviceinfo, err := serviceinfo.FindService(c, tx, serviceinfo); err != nil {
		middleware.ResponseError(c, 2002, err)
	}
	servicedetial, err := serviceinfo.ServiceDetial(c, tx, serviceinfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
	}
	todayList := []int64{}
	yesterdayList := []int64{}
	for i := 0; i <= time.Now().Hour(); i++ {
		todayList = append(todayList, 0)
	}
	for i := 0; i <= 23; i++ {
		yesterdayList = append(yesterdayList, 0)
	}
	out := &dto.ServiceStatOutput{Today: todayList, Yesterday: yesterdayList}

	middleware.ResponseSuccess(c, out)
}
