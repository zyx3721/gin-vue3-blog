/*
 * 项目名称：blog-backend
 * 文件名称：gitee_scraper.go
 * 创建时间：2026-04-14 13:56:40
 *
 * 系统用户：Administrator
 * 作　　者：無以菱
 * 联系邮箱：huangjing510@126.com
 * 功能描述：Gitee贡献热力图数据爬取模块，从Gitee用户主页HTML解析贡献日历数据，
 *          替代原独立运行的gitee-calendar-api微服务
 */
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"blog-backend/logger"
	"go.uber.org/zap"
)

// giteeCalendarRegex Gitee贡献日历正则表达式（编译一次，复用）
// 匹配格式：data-content='X个贡献：YYYY-MM-DD' date='YYYYMMDD'
var giteeCalendarRegex = regexp.MustCompile(`data-content='(\d+)个贡献[^']*'[^>]*date='(\d{8})'`)

// ScrapeGiteeCalendar 从Gitee用户主页爬取贡献热力图数据
// @param ctx context.Context 上下文，用于超时控制和取消
// @param username string Gitee用户名
// @return *CalendarResponse 贡献热力图数据
// @error error 爬取或解析失败时返回错误
func ScrapeGiteeCalendar(ctx context.Context, username string) (*CalendarResponse, error) {
	// 创建带超时的子上下文（10秒）
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 构建请求
	targetURL := "https://gitee.com/" + username
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建Gitee请求失败: %v", err)
	}

	// 设置浏览器模拟请求头，减少被风控概率
	req.Header.Set("Referer", targetURL)
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="122", "Not(A:Brand";v="24", "Microsoft Edge";v="122"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 Edg/122.0.0.0")

	// 发起HTTP请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Gitee主页失败: %v", err)
	}
	defer resp.Body.Close()

	// 非200状态码返回空数据（与原微服务行为一致）
	if resp.StatusCode != http.StatusOK {
		logger.Warn("Gitee主页返回非200状态码",
			zap.String("username", username),
			zap.Int("status", resp.StatusCode))
		return &CalendarResponse{Total: 0, Contributions: [][]CalendarDay{}}, nil
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取Gitee响应失败: %v", err)
	}

	// 正则匹配贡献数据
	matches := giteeCalendarRegex.FindAllStringSubmatch(string(bodyBytes), -1)
	if len(matches) == 0 {
		logger.Debug("未从Gitee主页匹配到贡献数据", zap.String("username", username))
		return &CalendarResponse{Total: 0, Contributions: [][]CalendarDay{}}, nil
	}

	// 提取日期和贡献次数
	type datePair struct {
		date  string
		count int
	}
	pairs := make([]datePair, 0, len(matches))
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		count, _ := strconv.Atoi(strings.TrimSpace(m[1]))
		dateRaw := m[2]
		if len(dateRaw) != 8 {
			continue
		}
		// 日期格式转换：20241217 → 2024-12-17
		dateStr := dateRaw[:4] + "-" + dateRaw[4:6] + "-" + dateRaw[6:8]
		pairs = append(pairs, datePair{date: dateStr, count: count})
	}

	// 按日期排序
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].date < pairs[j].date
	})

	// 构建结果，计算总数
	var total int64
	list := make([]CalendarDay, 0, len(pairs))
	for _, p := range pairs {
		total += int64(p.count)
		list = append(list, CalendarDay{Date: p.date, Count: p.count})
	}

	return &CalendarResponse{
		Total:         total,
		Contributions: splitCalendarDays(list, 7),
	}, nil
}

// splitCalendarDays 将一维CalendarDay切片按n个一组切分为二维切片
// @param items []CalendarDay 待切分的贡献数据
// @param n int 每组的数量（通常为7，表示一周）
// @return [][]CalendarDay 切分后的二维切片
func splitCalendarDays(items []CalendarDay, n int) [][]CalendarDay {
	result := make([][]CalendarDay, 0, (len(items)+n-1)/n)
	for i := 0; i < len(items); i += n {
		end := i + n
		if end > len(items) {
			end = len(items)
		}
		result = append(result, items[i:end])
	}
	return result
}
