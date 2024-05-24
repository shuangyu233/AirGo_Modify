package public_api

import (
	"github.com/shuangyu233/AirGo_Modify/global"
	"github.com/shuangyu233/AirGo_Modify/service"
	"github.com/shuangyu233/AirGo_Modify/utils/response"
	"github.com/gin-gonic/gin"
)

// GetDefaultArticleList
// @Tags [public api] article
// @Summary 获取默认的首页弹窗和自定义内容
// @Produce json
// @Success 200 {object} response.ResponseStruct "请求成功；正常：业务代码 code=0；错误：业务代码code=1"
// @Router /api/public/article/getDefaultArticleList [get]
func GetDefaultArticleList(ctx *gin.Context) {
	data, err := service.ArticleSvc.GetDefaultArticle()
	if err != nil {
		global.Logrus.Error(err)
		response.Fail("GetDefaultArticleList error:"+err.Error(), nil, ctx)
		return
	}
	response.OK("GetDefaultArticleList success", data, ctx)
}
