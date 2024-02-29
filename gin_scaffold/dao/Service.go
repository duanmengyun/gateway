package dao

import (
	"errors"
	"gin_scaffold/dto"
	"gin_scaffold/public"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
)

type ServiceDetial struct {
	Info          *Serviceinfo     `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule        `json:"http_rule" description:"http_rule"`
	TCPRule       *TcpRule         `json:"tcp_rule" description:"tcp_rule"`
	GRPCRule      *GrpcRule        `json:"grpc_rule" description:"grpc_rule"`
	LoadBalance   *LoadBalance     `json:"load_balance" description:"load_balance"`
	AccessControl *AcccessControll `json:"access_control" description:"access_control"`
}

var ServiceManagerHandler *ServiceManager

func init() {
	ServiceManagerHandler = NewServiceManager()
}

type ServiceManager struct {
	ServiceMap   map[string]*ServiceDetial
	ServiceSlice []*ServiceDetial
	Locker       sync.Mutex
	init         sync.Once
	err          error
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*ServiceDetial{},
		ServiceSlice: []*ServiceDetial{},
		Locker:       sync.Mutex{},
		init:         sync.Once{},
	}
}

func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*ServiceDetial, error) {
	//1、前缀匹配 /abc ==> serviceSlice.rule
	//2、域名匹配 www.test.com ==> serviceSlice.rule
	//host c.Request.Host
	//path c.Request.URL.Path
	host := c.Request.Host
	host = host[0:strings.Index(host, ":")]
	path := c.Request.URL.Path
	for _, serviceItem := range s.ServiceSlice {
		if serviceItem.Info.LoadType != public.LoadTypeHTTP {
			continue
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			if serviceItem.HTTPRule.Rule == host {
				return serviceItem, nil
			}
		}
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypefixURL {
			if strings.HasPrefix(path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("not matched service")
}

func (s *ServiceManager) LoadOnce() error {
	s.init.Do(func() {
		serviceInfo := &Serviceinfo{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &dto.ServiceListInput{PageNumber: 1, PageSize: 99999}
		list, _, err := serviceInfo.PageList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			serviceDetail, err := tmpItem.ServiceDetial(c, tx, &tmpItem)
			//fmt.Println("serviceDetail")
			//fmt.Println(public.Obj2Json(serviceDetail))
			if err != nil {
				s.err = err
				return
			}
			s.ServiceMap[listItem.ServiceName] = serviceDetail
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}
	})
	return s.err
}

func (s *ServiceManager) GetTcpServiceList() []*ServiceDetial {
	list := []*ServiceDetial{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == public.LoadTypeTCP {
			list = append(list, tempItem)
		}
	}
	return list
}

func (s *ServiceManager) GetGrpcServiceList() []*ServiceDetial {
	list := []*ServiceDetial{}
	for _, serverItem := range s.ServiceSlice {
		tempItem := serverItem
		if tempItem.Info.LoadType == public.LoadTypeGrpc {
			list = append(list, tempItem)
		}
	}
	return list
}
