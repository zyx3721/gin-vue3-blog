/*
 * 项目名称：blog-backend
 * 文件名称：subscriber.go
 * 创建时间：2026-04-17 23:40:47
 *
 * 系统用户：Administrator
 * 作　　者：無以菱
 * 联系邮箱：huangjing510@126.com
 * 功能描述：邮件订阅者处理器，提供订阅、退订、订阅者管理功能
 */
package handler

import (
	"blog-backend/config"
	"blog-backend/service"
	"blog-backend/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SubscriberHandler 订阅者处理器结构体
type SubscriberHandler struct {
	service *service.SubscriberService
}

// NewSubscriberHandler 创建订阅者处理器实例
func NewSubscriberHandler(cfg *config.Config) *SubscriberHandler {
	return &SubscriberHandler{
		service: service.NewSubscriberService(cfg),
	}
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Subscribe 订阅
func (h *SubscriberHandler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, "请输入有效的邮箱地址")
		return
	}

	if err := h.service.Subscribe(c.Request.Context(), req.Email); err != nil {
		util.Error(c, 400, err.Error())
		return
	}

	util.SuccessWithMessage(c, "订阅成功！欢迎邮件已发送到您的邮箱", nil)
}

// Unsubscribe 退订
func (h *SubscriberHandler) Unsubscribe(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		util.BadRequest(c, "缺少退订令牌")
		return
	}

	if err := h.service.Unsubscribe(c.Request.Context(), token); err != nil {
		util.Error(c, 400, err.Error())
		return
	}

	util.SuccessWithMessage(c, "退订成功", nil)
}

// List 获取订阅者列表（管理员用）
func (h *SubscriberHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	subscribers, total, err := h.service.List(c.Request.Context(), page, pageSize)
	if err != nil {
		util.ServerError(c, "获取订阅者列表失败")
		return
	}

	util.PageSuccess(c, subscribers, total, page, pageSize)
}

// Delete 删除订阅者（管理员用）
func (h *SubscriberHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.BadRequest(c, "无效的订阅者ID")
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		util.Error(c, 400, err.Error())
		return
	}

	util.SuccessWithMessage(c, "删除成功", nil)
}
