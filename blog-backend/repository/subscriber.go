/*
 * 项目名称：blog-backend
 * 文件名称：subscriber.go
 * 创建时间：2026-04-17 23:40:47
 *
 * 系统用户：Administrator
 * 作　　者：無以菱
 * 联系邮箱：huangjing510@126.com
 * 功能描述：邮件订阅者数据访问层，提供订阅者的数据库操作功能
 */
package repository

import (
	"blog-backend/db"
	"blog-backend/model"
	"context"
)

// SubscriberRepository 订阅者数据访问层结构体
type SubscriberRepository struct{}

// NewSubscriberRepository 创建订阅者数据访问层实例
func NewSubscriberRepository() *SubscriberRepository {
	return &SubscriberRepository{}
}

// Create 创建订阅者
func (r *SubscriberRepository) Create(ctx context.Context, subscriber *model.Subscriber) error {
	return db.DB.WithContext(ctx).Create(subscriber).Error
}

// GetByEmail 根据邮箱获取订阅者
func (r *SubscriberRepository) GetByEmail(ctx context.Context, email string) (*model.Subscriber, error) {
	var subscriber model.Subscriber
	err := db.DB.WithContext(ctx).Where("email = ?", email).First(&subscriber).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

// GetByToken 根据Token获取订阅者
func (r *SubscriberRepository) GetByToken(ctx context.Context, token string) (*model.Subscriber, error) {
	var subscriber model.Subscriber
	err := db.DB.WithContext(ctx).Where("token = ?", token).First(&subscriber).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

// GetActiveSubscribers 获取所有活跃订阅者
func (r *SubscriberRepository) GetActiveSubscribers(ctx context.Context) ([]*model.Subscriber, error) {
	var subscribers []*model.Subscriber
	err := db.DB.WithContext(ctx).Where("is_active = ?", true).Find(&subscribers).Error
	return subscribers, err
}

// List 获取订阅者列表（分页）
func (r *SubscriberRepository) List(ctx context.Context, page, pageSize int) ([]model.Subscriber, int64, error) {
	var subscribers []model.Subscriber
	var total int64

	offset := (page - 1) * pageSize

	if err := db.DB.WithContext(ctx).Model(&model.Subscriber{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.DB.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&subscribers).Error

	return subscribers, total, err
}

// Update 更新订阅者
func (r *SubscriberRepository) Update(ctx context.Context, subscriber *model.Subscriber) error {
	return db.DB.WithContext(ctx).Save(subscriber).Error
}

// Delete 删除订阅者
func (r *SubscriberRepository) Delete(ctx context.Context, id uint) error {
	return db.DB.WithContext(ctx).Delete(&model.Subscriber{}, id).Error
}

// GetByID 根据ID获取订阅者
func (r *SubscriberRepository) GetByID(ctx context.Context, id uint) (*model.Subscriber, error) {
	var subscriber model.Subscriber
	err := db.DB.WithContext(ctx).First(&subscriber, id).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

// CountActive 统计活跃订阅者数量
func (r *SubscriberRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := db.DB.WithContext(ctx).Model(&model.Subscriber{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// CountTotal 统计累积订阅者总数（包括已退订的用户）
func (r *SubscriberRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	err := db.DB.WithContext(ctx).Model(&model.Subscriber{}).Count(&count).Error
	return count, err
}
