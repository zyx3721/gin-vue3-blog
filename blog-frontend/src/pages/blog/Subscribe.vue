<!--
 * @ProjectName: go-vue3-blog
 * @FileName: Subscribe.vue
 * @CreateTime: 2026-04-18 14:42:07
 * @SystemUser: Administrator
 * @Author: 無以菱
 * @Contact: huangjing510@126.com
 * @Description: 邮件订阅页面组件，提供订阅和退订功能
 -->
<template>
  <div class="subscribe-page">
    <div class="subscribe-layout">
      <div class="subscribe-main">
        <n-spin :show="loading">
          <!-- 页面头部 -->
          <div class="page-header">
            <h1 class="page-title">📧 邮件订阅</h1>
            <p class="page-description">订阅本站，第一时间获取最新文章推送</p>
          </div>

          <!-- 退订成功状态 -->
          <n-card v-if="unsubscribeSuccess" class="result-card">
            <n-result status="success" title="退订成功" description="您已成功取消订阅，不会再收到新文章推送邮件">
              <template #footer>
                <n-button type="primary" @click="resetForm">返回</n-button>
              </template>
            </n-result>
          </n-card>

          <!-- 订阅内容区 -->
          <template v-else>
            <!-- 订阅介绍卡片 -->
            <n-card class="intro-card">
              <div class="intro-content">
                <div class="intro-icon">✉️</div>
                <h2 class="intro-title">为什么要订阅？</h2>
                <div class="intro-features">
                  <div class="feature-item">
                    <div class="feature-icon">🚀</div>
                    <div class="feature-text">
                      <h3>第一时间获取更新</h3>
                      <p>新文章发布时自动推送到您的邮箱</p>
                    </div>
                  </div>
                  <div class="feature-item">
                    <div class="feature-icon">🔒</div>
                    <div class="feature-text">
                      <h3>隐私安全保护</h3>
                      <p>您的邮箱信息将被安全保密，绝不泄露</p>
                    </div>
                  </div>
                  <div class="feature-item">
                    <div class="feature-icon">🎯</div>
                    <div class="feature-text">
                      <h3>随时取消订阅</h3>
                      <p>可以通过邮件中的链接随时取消订阅</p>
                    </div>
                  </div>
                </div>
              </div>
            </n-card>

            <!-- 订阅表单卡片 -->
            <n-card class="subscribe-form-card">
              <!-- 订阅成功 -->
              <n-result
                v-if="subscribeSuccess"
                status="success"
                title="订阅成功"
                description="欢迎邮件已发送到您的邮箱，请查收确认"
              >
                <template #footer>
                  <n-button type="primary" @click="resetForm">继续浏览</n-button>
                </template>
              </n-result>

              <!-- 订阅表单 -->
              <template v-else>
                <h2 class="form-title">立即订阅</h2>
                <n-form ref="formRef" :model="formData" :rules="rules" label-placement="top">
                  <n-form-item label="邮箱地址" path="email">
                    <n-input
                      v-model:value="formData.email"
                      placeholder="请输入您的邮箱地址"
                      size="large"
                      :disabled="loading"
                      @keyup.enter="handleSubscribe"
                    >
                      <template #prefix>
                        <n-icon :component="MailOutline" />
                      </template>
                    </n-input>
                  </n-form-item>

                  <n-button
                    type="primary"
                    size="large"
                    block
                    :loading="loading"
                    @click="handleSubscribe"
                    class="subscribe-button"
                  >
                    <template #icon>
                      <n-icon :component="SendOutline" />
                    </template>
                    立即订阅
                  </n-button>
                </n-form>

                <div class="tips-section">
                  <n-alert type="info" :bordered="false">
                    <template #icon>
                      <n-icon :component="InformationCircleOutline" />
                    </template>
                    订阅后，您将在新文章发布时收到邮件通知。如需取消订阅，可点击邮件中的退订链接。
                  </n-alert>
                </div>
              </template>
            </n-card>

            <!-- 统计信息卡片 -->
            <n-card class="stats-card">
              <div class="stats-content">
                <div class="stat-item">
                  <div class="stat-icon">👥</div>
                  <div class="stat-info">
                    <div class="stat-value">{{ subscriberCount }}</div>
                    <div class="stat-label">订阅用户</div>
                  </div>
                </div>
                <div class="stat-divider"></div>
                <div class="stat-item">
                  <div class="stat-icon">📝</div>
                  <div class="stat-info">
                    <div class="stat-value">{{ postCount }}</div>
                    <div class="stat-label">文章总数</div>
                  </div>
                </div>
                <div class="stat-divider"></div>
                <div class="stat-item">
                  <div class="stat-icon">🔥</div>
                  <div class="stat-info">
                    <div class="stat-value">定期更新</div>
                    <div class="stat-label">持续输出</div>
                  </div>
                </div>
              </div>
            </n-card>
          </template>
        </n-spin>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useMessage, NIcon } from 'naive-ui'
import { MailOutline, SendOutline, InformationCircleOutline } from '@vicons/ionicons5'
import request from '@/utils/request'
import { getAuthorProfile } from '@/api/blog'

const route = useRoute()
const message = useMessage()

const formRef = ref(null)
const loading = ref(false)
const subscribeSuccess = ref(false)
const unsubscribeSuccess = ref(false)
const subscriberCount = ref('0')
const postCount = ref('0')

const formData = reactive({
  email: ''
})

const rules = {
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: ['blur', 'change'] }
  ]
}

// 订阅邮件
const subscribeEmail = async (email) => {
  return request({
    url: '/subscribe',
    method: 'post',
    data: { email }
  })
}

// 退订邮件
const unsubscribeEmail = async (token) => {
  return request({
    url: '/subscribe/unsubscribe',
    method: 'get',
    params: { token }
  })
}

// 获取统计信息
const fetchStats = async () => {
  try {
    // 获取博主资料和统计数据
    const res = await getAuthorProfile()
    if (res.data) {
      postCount.value = res.data.stats.posts.toString()
    }

    // 获取订阅者数量（从后端API获取）
    const subRes = await request({
      url: '/admin/subscribers',
      method: 'get',
      params: { page: 1, page_size: 1 }
    })
    if (subRes.data && subRes.data.total !== undefined) {
      subscriberCount.value = subRes.data.total.toString()
    }
  } catch (error) {
    console.error('获取统计信息失败:', error)
    // 失败时使用默认值
    subscriberCount.value = '0'
    postCount.value = '0'
  }
}

// 处理订阅
const handleSubscribe = async () => {
  try {
    await formRef.value?.validate()
    loading.value = true

    await subscribeEmail(formData.email)
    subscribeSuccess.value = true
    message.success('订阅成功！欢迎邮件已发送到您的邮箱')
  } catch (error) {
    if (error?.message) {
      message.error(error.message)
    }
  } finally {
    loading.value = false
  }
}

// 处理退订
const handleUnsubscribe = async (token) => {
  try {
    loading.value = true
    await unsubscribeEmail(token)
    unsubscribeSuccess.value = true
    message.success('退订成功')
  } catch (error) {
    message.error(error?.message || '退订失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  subscribeSuccess.value = false
  unsubscribeSuccess.value = false
  formData.email = ''
  formRef.value?.restoreValidation()
}

// 页面加载时检查是否是退订操作
onMounted(() => {
  const action = route.query.action
  const token = route.query.token

  if (action === 'unsubscribe' && token) {
    handleUnsubscribe(token)
  }

  fetchStats()
})
</script>

<style scoped>
.subscribe-page {
  min-height: calc(100vh - 180px);
  padding: 32px 0;
}

.subscribe-layout {
  max-width: 900px;
  margin: 0 auto;
  padding: 0 20px;
}

.subscribe-main {
  min-width: 0;
}

.page-header {
  text-align: center;
  margin-bottom: 32px;
}

.page-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 12px 0;
  color: #fff;
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.5);
}

html.dark .page-title {
  color: #fff;
}

.page-description {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.8);
  text-shadow: 0 1px 8px rgba(0, 0, 0, 0.4);
  margin: 0;
}

html.dark .page-description {
  color: rgba(255, 255, 255, 0.7);
}

.result-card {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(8, 145, 178, 0.1);
  padding: 40px 0;
}

html.dark .result-card {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(56, 189, 248, 0.1);
}

/* 介绍卡片 */
.intro-card {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(8, 145, 178, 0.1);
  margin-bottom: 24px;
}

html.dark .intro-card {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(56, 189, 248, 0.1);
}

.intro-content {
  text-align: center;
}

.intro-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.intro-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 32px 0;
  color: #1a202c;
}

html.dark .intro-title {
  color: #e5e5e5;
}

.intro-features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 24px;
  margin-top: 24px;
}

.feature-item {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  text-align: left;
  padding: 20px;
  background: rgba(8, 145, 178, 0.05);
  border-radius: 8px;
  transition: all 0.3s ease;
}

.feature-item:hover {
  background: rgba(8, 145, 178, 0.1);
  transform: translateY(-2px);
}

html.dark .feature-item {
  background: rgba(56, 189, 248, 0.1);
}

html.dark .feature-item:hover {
  background: rgba(56, 189, 248, 0.15);
}

.feature-icon {
  font-size: 32px;
  flex-shrink: 0;
}

.feature-text h3 {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: #1a202c;
}

html.dark .feature-text h3 {
  color: #e5e5e5;
}

.feature-text p {
  font-size: 14px;
  color: #64748b;
  margin: 0;
  line-height: 1.6;
}

html.dark .feature-text p {
  color: #94a3b8;
}

/* 订阅表单卡片 */
.subscribe-form-card {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(8, 145, 178, 0.1);
  margin-bottom: 24px;
}

html.dark .subscribe-form-card {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(56, 189, 248, 0.1);
}

.form-title {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 24px 0;
  color: #1a202c;
  text-align: center;
}

html.dark .form-title {
  color: #e5e5e5;
}

.subscribe-button {
  margin-top: 8px;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
}

.tips-section {
  margin-top: 24px;
}

/* 统计信息卡片 */
.stats-card {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(8, 145, 178, 0.1);
}

html.dark .stats-card {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(56, 189, 248, 0.1);
}

.stats-content {
  display: flex;
  justify-content: space-around;
  align-items: center;
  gap: 24px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  justify-content: center;
}

.stat-icon {
  font-size: 40px;
}

.stat-info {
  text-align: left;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #0891b2;
  line-height: 1.2;
}

html.dark .stat-value {
  color: #38bdf8;
}

.stat-label {
  font-size: 14px;
  color: #64748b;
  margin-top: 4px;
}

html.dark .stat-label {
  color: #94a3b8;
}

.stat-divider {
  width: 1px;
  height: 50px;
  background: rgba(8, 145, 178, 0.2);
}

html.dark .stat-divider {
  background: rgba(56, 189, 248, 0.2);
}

/* 响应式 */
@media (max-width: 768px) {
  .subscribe-page {
    padding: 24px 0;
  }

  .subscribe-layout {
    padding: 0 16px;
  }

  .page-title {
    font-size: 28px;
  }

  .page-description {
    font-size: 14px;
  }

  .intro-features {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .feature-item {
    padding: 16px;
  }

  .stats-content {
    flex-direction: column;
    gap: 20px;
  }

  .stat-divider {
    width: 80%;
    height: 1px;
  }

  .stat-item {
    width: 100%;
  }
}
</style>
