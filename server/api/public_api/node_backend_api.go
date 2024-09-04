package public_api

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AirGo-Official/AirGo/api"
	"github.com/AirGo-Official/AirGo/constant"
	"github.com/AirGo-Official/AirGo/global"
	"github.com/AirGo-Official/AirGo/model"
	"github.com/AirGo-Official/AirGo/service"

	"github.com/gin-gonic/gin"
)

// AGGetNodeInfo
// @Tags [public api] node
// @Summary 获取节点配置信息
// @Produce json
// @Param id query int64 true "节点ID"
// @Param key query string true "节点密钥"
// @Success 200 {object} model.Node "成功"
// @Failure 400  "请求错误"
// @Failure 304  "数据和上次一致"
// @Router /api/public/airgo/node/getNodeInfo [get]
func AGGetNodeInfo(ctx *gin.Context) {
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		ctx.AbortWithStatus(400)
		return
	}
	id := ctx.Query("id")
	nodeIDInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.AbortWithStatus(400)
		return
	}
	var node model.Node
	err = global.DB.Model(&model.Node{}).Where(&model.Node{ID: nodeIDInt}).Preload("Access").First(&node).Error
	if err != nil {
		global.Logrus.Error("AGGetNodeInfo error,id=" + id + ": " + err.Error())
		ctx.AbortWithStatus(400)
		return
	}
	node.TrafficRate = 1
	//处理ss节点加密
	if node.Protocol == "shadowsocks" {
		node.ServerKey = service.AdminNodeSvc.GetShadowsocksServerKey(node)
	}
	//etag
	api.EtagHandler(node, ctx)
}

// AGReportNodeStatus
// @Tags [public api] node
// @Summary 上报节点状态
// @Produce json
// @Param key query string true "节点密钥"
// @Param data body model.AGNodeStatus true "参数"
// @Success 200 {object} string "成功"
// @Failure 400  "请求错误"
// @Failure 304  "数据和上次一致"
// @Router /api/public/airgo/node/AGReportNodeStatus [post]
func AGReportNodeStatus(ctx *gin.Context) {
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		ctx.AbortWithStatus(400)
		return
	}
	var AGNodeStatus model.AGNodeStatus
	err := ctx.ShouldBind(&AGNodeStatus)
	if err != nil {
		global.Logrus.Error("AGReportNodeStatus error: " + err.Error())
		ctx.AbortWithStatus(400)
		return
	}

	var cacheLock sync.Mutex
	cacheLock.Lock()
	defer cacheLock.Unlock()

	cacheStatus, ok := global.LocalCache.Get(fmt.Sprintf("%s%d", constant.CACHE_NODE_STATUS_BY_NODEID, AGNodeStatus.ID))

	if ok {
		oldStatus := cacheStatus.(model.NodeStatus)
		oldStatus.Status = true
		oldStatus.CPU = AGNodeStatus.CPU
		oldStatus.Mem = AGNodeStatus.Mem
		oldStatus.Disk = AGNodeStatus.Disk
		global.LocalCache.Set(fmt.Sprintf("%s%d",
			constant.CACHE_NODE_STATUS_BY_NODEID, AGNodeStatus.ID),
			oldStatus,
			constant.CAHCE_NODE_STATUS_TIMEOUT*time.Minute)
	} else {
		var status model.NodeStatus
		status.Status = true
		status.ID = AGNodeStatus.ID
		status.CPU = AGNodeStatus.CPU
		status.Mem = AGNodeStatus.Mem
		status.Disk = AGNodeStatus.Disk
		global.LocalCache.Set(fmt.Sprintf("%s%d",
			constant.CACHE_NODE_STATUS_BY_NODEID, AGNodeStatus.ID),
			status,
			constant.CAHCE_NODE_STATUS_TIMEOUT*time.Minute)
	}
	ctx.String(200, "success")
}

// AGGetUserlist
// @Tags [public api] node
// @Summary 获取用户列表
// @Produce json
// @Param id query int64 true "节点ID"
// @Param key query string true "节点密钥"
// @Success 200 {object} string "成功"
// @Failure 400  "请求错误"
// @Failure 304  "数据和上次一致"
// @Router /api/public/airgo/user/AGGetUserlist [get]
func AGGetUserlist(ctx *gin.Context) {
	//验证key
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		ctx.AbortWithStatus(400)
		return
	}
	id := ctx.Query("id")
	nodeIDInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.AbortWithStatus(400)
		return
	}
	var node model.Node
	err = global.DB.Model(&model.Node{}).Where(&model.Node{ID: nodeIDInt}).First(&node).Error
	if err != nil {
		ctx.AbortWithStatus(400)
		return
	}
	//节点属于哪些goods
	goods, err := service.AdminShopSvc.FindGoodsByNodeID(nodeIDInt)
	if err != nil {
		ctx.AbortWithStatus(400)
		return
	}
	//goods属于哪些用户
	var goodsArr []int64
	for _, v := range goods {
		goodsArr = append(goodsArr, v.ID)
	}
	var users []model.AGUserInfo //返回给节点服务器的数据，其中的 customer_server id 对应 Xrayr 或 v2bx 中的 uid; 处理上报流量时也要注意对应关系
	err = global.DB.
		Model(&model.CustomerService{}).
		Where("goods_id in (?) and sub_status = ?", goodsArr, true).
		Select("id, sub_uuid AS uuid, user_name, node_connector, node_speed_limit").
		Find(&users).Error
	if err != nil {
		global.Logrus.Error("AGGetUserlist error, id=" + id + ": " + err.Error())
		ctx.AbortWithStatus(400)
		return
	}
	//处理ss加密
	switch node.Protocol {
	case constant.NODE_PROTOCOL_SHADOWSOCKS:
		switch strings.HasPrefix(node.Scy, "2022") {
		case true:
			for k := range users {
				p := users[k].UUID.String()
				if node.Scy == "2022-blake3-aes-128-gcm" {
					p = p[:16]
				}
				p = base64.StdEncoding.EncodeToString([]byte(p))
				users[k].Passwd = p
			}
		default:
			for k := range users {
				users[k].Passwd = users[k].UUID.String()
			}
		}
	default:
	}
	//fmt.Println("users:", users)
	api.EtagHandler(users, ctx)
}

// AGReportUserTraffic
// @Tags [public api] node
// @Summary 上报用户流量
// @Produce json
// @Param key query string true "节点密钥"
// @Param data body model.AGUserTraffic true "参数"
// @Success 200 {object} string "成功"
// @Failure 400  "请求错误"
// @Failure 304  "数据和上次一致"
// @Router /api/public/airgo/user/AGReportUserTraffic [post]
func AGReportUserTraffic(ctx *gin.Context) {
	// 校验节点密钥
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		ctx.AbortWithStatus(400)
		return
	}

	// 绑定并校验请求体数据
	var AGUserTraffic model.AGUserTraffic
	if err := ctx.ShouldBind(&AGUserTraffic); err != nil {
		global.Logrus.Error("AGReportUserTraffic error: " + err.Error())
		ctx.AbortWithStatus(400)
		return
	}

	// 获取节点信息
	node, err := service.AdminNodeSvc.FirstNode(&model.Node{ID: AGUserTraffic.ID})
	if err != nil {
		global.Logrus.Error("AGReportUserTraffic error: " + err.Error())
		ctx.AbortWithStatus(400)
		return
	}

	// 准备记录数据
	var trafficLog model.NodeTrafficLog
	userTrafficLogMap := make(map[int64]model.UserTrafficLog)
	var totalUpload, totalDownload int64

	for _, userTraffic := range AGUserTraffic.UserTraffic {
		upload := int64(float64(userTraffic.Upload) * node.TrafficRate)
		download := int64(float64(userTraffic.Download) * node.TrafficRate)

		userTrafficLogMap[userTraffic.UID] = model.UserTrafficLog{
			SubUserID: userTraffic.UID,
			UserName:  userTraffic.Email,
			U:         upload,
			D:         download,
		}

		totalUpload += upload
		totalDownload += download
	}

	// 记录节点总流量
	trafficLog.NodeID = node.ID
	trafficLog.U = totalUpload
	trafficLog.D = totalDownload

	// 发布任务到队列处理
	_ = global.Queue.Publish(constant.NODE_BACKEND_TASK, &service.NodeBackendServiceMessage{
		Title: constant.NODE_BACKEND_TASK_TITLE_NODE_TRAFFIC,
		Data: &service.NodeTrafficMessage{
			NodeTrafficLog: &trafficLog,
			AGUserTraffic:  &AGUserTraffic,
		},
	})

	_ = global.Queue.Publish(constant.NODE_BACKEND_TASK, &service.NodeBackendServiceMessage{
		Title: constant.NODE_BACKEND_TASK_TITLE_USER_TRAFFIC,
		Data:  userTrafficLogMap,
	})

	// 记录日志
	global.Logrus.Info(fmt.Sprintf("Node[id=%d] user traffic log success", AGUserTraffic.ID))

	// 返回成功响应
	ctx.String(200, "success")
}

// AGReportNodeOnlineUsers
// @Tags [public api] node
// @Summary 上报在线用户
// @Produce json
// @Param key query string true "节点密钥"
// @Param data body model.AGOnlineUser true "参数"
// @Success 200 {object} string "成功"
// @Failure 400  "请求错误"
// @Failure 304  "数据和上次一致"
// @Router /api/public/airgo/user/AGReportNodeOnlineUsers [post]
func AGReportNodeOnlineUsers(ctx *gin.Context) {
	//验证key
	if global.Server.Subscribe.TEK != ctx.Query("key") {
		return
	}
	var AGOnlineUser model.AGOnlineUser
	err := ctx.ShouldBind(&AGOnlineUser)
	if err != nil {
		global.Logrus.Error("error", err.Error())
		ctx.AbortWithStatus(400)
		return
	}
	ctx.String(200, "success")
	//TODO 未用到
}
