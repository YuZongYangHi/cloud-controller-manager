package routers

import (
	"github.com/YuZongYangHi/cloud-controller-manager/cmd/cloud-provider-manager/controllers/loadbalance"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()
	apiGroup := r.Group("/api/v1/cloudprovider")

	{
		loadBalanceGroup := apiGroup.Group("/loadbalance")
		loadBalanceGroup.GET("/list", loadbalance.List)
		loadBalanceGroup.POST("/unbind", loadbalance.Released)
		loadBalanceGroup.POST("/bind", loadbalance.Bind)
	}

	return r
}
