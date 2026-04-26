/*
 * 项目名称：blog-frontend
 * 文件名称：subscribe.ts
 * 创建时间：2026-04-26 10:30:00
 *
 * 系统用户：Administrator
 * 作　　者：無以菱
 * 联系邮箱：huangjing510@126.com
 * 功能描述：邮件订阅相关 API 接口定义
 */

import { request } from '@/utils/request'

/**
 * 订阅统计信息接口
 */
export interface SubscribeStats {
  total_count: number   // 累积订阅总数（包括已退订）
  active_count: number  // 当前活跃订阅者数量
}

/**
 * 获取订阅统计信息（公开接口）
 * @returns 返回订阅统计信息
 */
export function getSubscribeStats() {
  return request.get<SubscribeStats>('/subscribe/stats')
}
