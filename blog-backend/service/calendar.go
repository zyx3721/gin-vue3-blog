/*
 * 项目名称：blog-backend
 * 文件名称：calendar.go
 * 创建时间：2026-01-31 16:34:35
 *
 * 系统用户：Administrator
 * 作　　者：無以菱
 * 联系邮箱：huangjing510@126.com
 * 功能描述：日历业务逻辑层，提供Gitee贡献热力图数据查询功能，支持Redis缓存，内置Gitee主页爬取
 */
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"blog-backend/db"
	"blog-backend/logger"
	"go.uber.org/zap"
)

// CalendarService 日历业务逻辑层结构体
type CalendarService struct{}

// NewCalendarService 创建日历业务逻辑层实例
func NewCalendarService() *CalendarService {
	return &CalendarService{}
}

// CalendarResponse 贡献热力图响应结构
type CalendarResponse struct {
	Total         int64           `json:"total"`
	Contributions [][]CalendarDay `json:"contributions"`
}

// CalendarDay 单日贡献数据
type CalendarDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// GetGiteeCalendar 获取Gitee贡献热力图数据（带Redis缓存）
func (s *CalendarService) GetGiteeCalendar(username string) (*CalendarResponse, error) {
	if username == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("gitee_calendar:%s", username)

	// 1. 先尝试从Redis缓存获取
	cachedData, err := db.RDB.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		// 缓存命中，解析并返回
		var response CalendarResponse
		if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
			return &response, nil
		}
		// 如果解析失败，继续从API获取
	}

	// 2. 缓存未命中或解析失败，直接爬取Gitee主页
	logger.Debug("Gitee日历缓存未命中，开始爬取", zap.String("username", username))
	response, err := ScrapeGiteeCalendar(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("获取Gitee贡献数据失败: %v", err)
	}

	// 3. 将数据存入Redis缓存，过期时间20分钟
	cacheData, err := json.Marshal(response)
	if err == nil {
		db.RDB.Set(ctx, cacheKey, string(cacheData), 20*time.Minute)
	}

	return response, nil
}
