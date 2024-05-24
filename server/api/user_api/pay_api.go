package user_api

import (
	"github.com/shuangyu233/AirGo_Modify/global"
	"github.com/shuangyu233/AirGo_Modify/model"
	"github.com/shuangyu233/AirGo_Modify/service"
	"github.com/shuangyu233/AirGo_Modify/utils/response"
	"github.com/gin-gonic/gin"
)

// GetEnabledPayList
// @Tags [customer api] pay
// @Summary 获取已激活支付列表
// @Produce json
// @Param Authorization header string false "Bearer 用户token"
// @Success 200 {object} response.ResponseStruct "请求成功；正常：业务代码 code=0；错误：业务代码code=1"
// @Router /api/customer/pay/getEnabledPayList [get]
func GetEnabledPayList(ctx *gin.Context) {
	list, _, err := service.CommonSqlFind[model.Pay, string, []model.Pay]("status = true")
	if err != nil {
		global.Logrus.Error(err.Error())
		response.Fail(err.Error(), nil, ctx)
		return
	}
	response.OK("GetEnabledPayList success", list, ctx)
}
