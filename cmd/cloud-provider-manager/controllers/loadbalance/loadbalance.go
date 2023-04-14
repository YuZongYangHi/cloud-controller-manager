package loadbalance

import (
	"github.com/YuZongYangHi/cloud-controller-manager/cmd/cloud-provider-manager/base"
	"github.com/YuZongYangHi/cloud-controller-manager/cmd/cloud-provider-manager/models"
	"github.com/gin-gonic/gin"
)

func List(ctx *gin.Context) {
	rawQuery := ctx.Request.URL.Query()
	query := map[string]interface{}{}

	for field, values := range rawQuery {
		query[field] = values[0]
	}

	result := models.LoadBalanceModel.List(query)
	base.SuccessResponse(ctx, result)
}

func Bind(ctx *gin.Context) {
	var m models.LoadBalance

	err := ctx.BindJSON(&m)
	if err != nil {
		base.BadRequestResponse(ctx, err.Error())
		return
	}

	if !Valid(&m) {
		base.BadRequestResponse(ctx, "invalid params")
		return
	}

	response, err := models.LoadBalanceModel.Bind(&m)
	if err != nil {
		base.ServerErrorResponse(ctx, err.Error())
		return
	}

	base.SuccessResponse(ctx, response)
}

func Released(ctx *gin.Context) {
	var m models.LoadBalance

	err := ctx.BindJSON(&m)
	if err != nil {
		base.BadRequestResponse(ctx, err.Error())
		return
	}

	if m.Namespace == "" || m.ServiceName == "" {
		base.BadRequestResponse(ctx, "invalid params")
		return
	}

	err = models.LoadBalanceModel.Released(&m)
	if err != nil {
		base.ServerErrorResponse(ctx, err.Error())
		return
	}
	base.SuccessResponse(ctx, "")
}
