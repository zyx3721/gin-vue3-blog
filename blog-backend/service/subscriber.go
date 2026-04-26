/*
 * 项目名称：blog-backend
 * 文件名称：subscriber.go
 * 创建时间：2026-04-17 23:40:47
 *
 * 系统用户：Administrator
 * 作　　者：無以菱
 * 联系邮箱：huangjing510@126.com
 * 功能描述：邮件订阅者业务逻辑层，提供订阅、退订、文章推送等业务处理
 */
package service

import (
	"blog-backend/config"
	"blog-backend/logger"
	"blog-backend/model"
	"blog-backend/pkg/email"
	"blog-backend/repository"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SubscriberService 订阅者业务逻辑层结构体
type SubscriberService struct {
	repo        *repository.SubscriberRepository
	emailClient *email.Client
	config      *config.Config
	settingRepo *repository.SettingRepository
}

// NewSubscriberService 创建订阅者业务逻辑层实例
func NewSubscriberService(cfg *config.Config) *SubscriberService {
	return &SubscriberService{
		repo:        repository.NewSubscriberRepository(),
		emailClient: email.Initialize(cfg),
		config:      cfg,
		settingRepo: repository.NewSettingRepository(),
	}
}

// Subscribe 订阅
func (s *SubscriberService) Subscribe(ctx context.Context, emailAddr string) error {
	if s.emailClient == nil {
		return errors.New("邮件服务未配置")
	}

	sub, err := s.repo.GetByEmail(ctx, emailAddr)
	if err == nil {
		if sub.IsActive {
			return errors.New("该邮箱已订阅")
		}
		// 重新激活订阅
		now := time.Now()
		sub.IsActive = true
		sub.SubscribedAt = &now
		sub.UnsubscribedAt = nil
		if err := s.repo.Update(ctx, sub); err != nil {
			return fmt.Errorf("重新激活订阅失败: %w", err)
		}
		return s.sendWelcomeEmail(sub)
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询订阅者失败: %w", err)
	}

	// 创建新订阅者
	now := time.Now()
	newSub := &model.Subscriber{
		Email:        emailAddr,
		IsActive:     true,
		Token:        email.GenerateToken(),
		SubscribedAt: &now,
	}

	if err := s.repo.Create(ctx, newSub); err != nil {
		return fmt.Errorf("创建订阅者失败: %w", err)
	}

	return s.sendWelcomeEmail(newSub)
}

// Unsubscribe 退订
func (s *SubscriberService) Unsubscribe(ctx context.Context, token string) error {
	sub, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("无效的退订链接")
		}
		return fmt.Errorf("查询订阅者失败: %w", err)
	}

	if !sub.IsActive {
		return errors.New("该邮箱已退订")
	}

	now := time.Now()
	sub.IsActive = false
	sub.UnsubscribedAt = &now
	if err := s.repo.Update(ctx, sub); err != nil {
		return fmt.Errorf("退订失败: %w", err)
	}

	return nil
}

// SendArticleNotification 发送文章推送通知（并发发送）
func (s *SubscriberService) SendArticleNotification(ctx context.Context, article *model.Post) error {
	if s.emailClient == nil {
		logger.Warn("邮件服务未配置，跳过文章推送")
		return nil
	}

	subscribers, err := s.repo.GetActiveSubscribers(ctx)
	if err != nil {
		return fmt.Errorf("获取订阅者列表失败: %w", err)
	}

	if len(subscribers) == 0 {
		return nil
	}

	// 并发控制：最多10个并发
	semaphore := make(chan struct{}, 10)
	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0

	// 并发发送邮件
	for _, sub := range subscribers {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(subscriber *model.Subscriber) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := s.sendArticleEmail(subscriber, article); err != nil {
				logger.Warn(fmt.Sprintf("发送文章推送失败 (邮箱: %s): %v", subscriber.Email, err))
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(sub)
	}

	wg.Wait()

	logger.Info(fmt.Sprintf("文章推送完成: 成功 %d/%d", successCount, len(subscribers)))
	return nil
}

// List 获取订阅者列表
func (s *SubscriberService) List(ctx context.Context, page, pageSize int) ([]model.Subscriber, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// Delete 删除订阅者
func (s *SubscriberService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// sendWelcomeEmail 发送欢迎邮件
func (s *SubscriberService) sendWelcomeEmail(sub *model.Subscriber) error {
	unsubscribeURL := fmt.Sprintf("%s/subscribe?action=unsubscribe&token=%s", s.getBlogURL(), sub.Token)
	siteName := s.getSiteName()
	return s.emailClient.SendWelcomeEmail(sub.Email, siteName, unsubscribeURL)
}

// sendArticleEmail 发送文章推送邮件
func (s *SubscriberService) sendArticleEmail(sub *model.Subscriber, article *model.Post) error {
	articleURL := fmt.Sprintf("%s/post/%s", s.getBlogURL(), article.Slug)
	unsubscribeURL := fmt.Sprintf("%s/subscribe?action=unsubscribe&token=%s", s.getBlogURL(), sub.Token)

	summary := article.Summary
	if summary == "" && len(article.Content) > 200 {
		summary = article.Content[:200] + "..."
	}

	siteName := s.getSiteName()
	return s.emailClient.SendArticleNotification(sub.Email, siteName, article.Title, summary, articleURL, unsubscribeURL)
}

// getBlogURL 获取博客前台地址
func (s *SubscriberService) getBlogURL() string {
	return s.config.App.BlogURL
}

// getSiteName 获取网站名称
func (s *SubscriberService) getSiteName() string {
	settings, err := s.settingRepo.GetByGroup("site")
	if err != nil {
		return "菱风叙"
	}

	for _, setting := range settings {
		if setting.Key == "site_name" {
			return setting.Value
		}
	}

	return "菱风叙"
}

// GetActiveCount 获取活跃订阅者数量（公开接口）
func (s *SubscriberService) GetActiveCount(ctx context.Context) (int64, error) {
	return s.repo.CountActive(ctx)
}

// GetTotalCount 获取累积订阅者总数（公开接口，包括已退订的用户）
func (s *SubscriberService) GetTotalCount(ctx context.Context) (int64, error) {
	return s.repo.CountTotal(ctx)
}
