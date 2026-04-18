<!--
 * @ProjectName: go-vue3-blog
 * @FileName: SubscriberManage.vue
 * @CreateTime: 2026-04-15 10:00:00
 * @SystemUser: Administrator
 * @Author: 無以菱
 * @Contact: huangjing510@126.com
 * @Description: 订阅者管理页面组件，提供订阅者的查看、删除等管理功能
 -->
<template>
  <div class="subscriber-manage-page">
    <div class="page-header">
      <h1 class="page-title">订阅者管理</h1>
      <!-- 视图切换按钮（仅桌面端显示） -->
      <n-button-group v-if="!isMobile" size="small" class="view-toggle-group">
        <n-button :type="viewMode === 'table' ? 'primary' : 'default'" @click="viewMode = 'table'">
          <template #icon>
            <n-icon :component="GridOutline" />
          </template>
          表格
        </n-button>
        <n-button :type="viewMode === 'card' ? 'primary' : 'default'" @click="viewMode = 'card'">
          <template #icon>
            <n-icon :component="AppsOutline" />
          </template>
          卡片
        </n-button>
      </n-button-group>
    </div>

    <!-- 内容区域 -->
    <div class="content-area">
      <div v-if="isMobile || viewMode === 'card'" class="card-list">
        <n-card v-for="subscriber in subscribers" :key="subscriber.id" class="list-card" :size="isMobile ? 'small' : 'medium'">
          <template #header>
            <div class="card-header-content">
              <span class="email">{{ subscriber.email }}</span>
              <n-tag :type="subscriber.is_active ? 'success' : 'default'" :size="isMobile ? 'tiny' : 'small'">
                {{ subscriber.is_active ? '活跃' : '已退订' }}
              </n-tag>
            </div>
          </template>
          <div class="card-content">
            <div class="info-item">
              <span class="label">订阅时间：</span>
              <span class="value">{{ formatDate(subscriber.subscribed_at, 'YYYY-MM-DD HH:mm') }}</span>
            </div>
            <div v-if="!subscriber.is_active && subscriber.unsubscribed_at" class="info-item">
              <span class="label">退订时间：</span>
              <span class="value">{{ formatDate(subscriber.unsubscribed_at, 'YYYY-MM-DD HH:mm') }}</span>
            </div>
          </div>
          <template #footer>
            <n-space justify="end" :size="isMobile ? 'small' : 'medium'">
              <n-button :size="isMobile ? 'tiny' : 'small'" type="error" @click="handleDelete(subscriber.id)">
                删除
              </n-button>
            </n-space>
          </template>
        </n-card>
      </div>

      <n-data-table
        v-else-if="viewMode === 'table'"
        :columns="columns"
        :data="subscribers"
        :loading="loading"
        :single-line="false"
      />

      <!-- 分页 - 位于表格右下角 -->
      <div class="pagination-wrapper">
        <n-pagination
          v-if="total > 0"
          v-model:page="currentPage"
          :page-count="totalPages"
          :page-size="pageSize"
          :page-slot="isMobile ? 3 : 7"
          :simple="isMobile"
          @update:page="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, h, watch } from 'vue'
import { useMessage, useDialog, NButton, NButtonGroup, NIcon, NTag, NSpace } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { GridOutline, AppsOutline } from '@vicons/ionicons5'
import request from '@/utils/request'
import { formatDate } from '@/utils/format'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const subscribers = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const viewMode = ref<'table' | 'card'>('table')
const isMobile = ref(false)

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 获取订阅者列表
const getSubscribers = async (page = 1) => {
  try {
    loading.value = true
    const response = await request({
      url: '/admin/subscribers',
      method: 'get',
      params: {
        page,
        page_size: pageSize.value
      }
    })
    subscribers.value = response.data.list || []
    total.value = response.data.total || 0
  } catch (error) {
    message.error(error?.message || '获取订阅者列表失败')
  } finally {
    loading.value = false
  }
}

// 删除订阅者
const deleteSubscriber = async (id: number) => {
  try {
    await request({
      url: `/admin/subscribers/${id}`,
      method: 'delete'
    })
    message.success('删除成功')
    await getSubscribers(currentPage.value)
  } catch (error) {
    message.error(error?.message || '删除失败')
  }
}

// 处理删除
const handleDelete = (id: number) => {
  dialog.warning({
    title: '确认删除',
    content: '确定要删除这个订阅者吗？此操作不可恢复。',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: () => {
      deleteSubscriber(id)
    }
  })
}

// 处理分页变化
const handlePageChange = (page: number) => {
  currentPage.value = page
  getSubscribers(page)
}

// 检测屏幕尺寸
const checkMobile = () => {
  isMobile.value = window.innerWidth < 768
}

// 表格列定义
const columns = computed<DataTableColumns>(() => [
  {
    title: 'ID',
    key: 'id',
    width: 80
  },
  {
    title: '邮箱',
    key: 'email',
    minWidth: 200
  },
  {
    title: '状态',
    key: 'is_active',
    width: 100,
    render: (row: any) => {
      return h(
        NTag,
        {
          type: row.is_active ? 'success' : 'default',
          size: 'small'
        },
        { default: () => (row.is_active ? '活跃' : '已退订') }
      )
    }
  },
  {
    title: '订阅时间',
    key: 'subscribed_at',
    width: 180,
    render: (row: any) => formatDate(row.subscribed_at, 'YYYY-MM-DD HH:mm')
  },
  {
    title: '退订时间',
    key: 'unsubscribed_at',
    width: 180,
    render: (row: any) => row.unsubscribed_at ? formatDate(row.unsubscribed_at, 'YYYY-MM-DD HH:mm') : '-'
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row: any) => {
      return h(
        NSpace,
        { size: 'small' },
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                type: 'error',
                onClick: () => handleDelete(row.id)
              },
              { default: () => '删除' }
            )
          ]
        }
      )
    }
  }
])

// 监听视图模式变化，保存到本地存储
watch(viewMode, (newMode) => {
  localStorage.setItem('subscriber-view-mode', newMode)
})

onMounted(() => {
  getSubscribers()
  checkMobile()
  window.addEventListener('resize', checkMobile)

  // 从本地存储恢复视图模式
  const savedMode = localStorage.getItem('subscriber-view-mode')
  if (savedMode === 'card' || savedMode === 'table') {
    viewMode.value = savedMode
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped lang="scss">
.subscriber-manage-page {
  padding: 20px;

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    .page-title {
      font-size: 24px;
      font-weight: 600;
      margin: 0;
    }

    .view-toggle-group {
      display: flex;
      gap: 0;
    }
  }

  .content-area {
    background: var(--card-color);
    border-radius: 8px;
    padding: 20px;

    .card-list {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
      gap: 16px;

      .list-card {
        .card-header-content {
          display: flex;
          justify-content: space-between;
          align-items: center;

          .email {
            font-weight: 500;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }
        }

        .card-content {
          .info-item {
            margin-bottom: 8px;
            font-size: 14px;

            .label {
              color: var(--text-color-3);
            }

            .value {
              color: var(--text-color-2);
            }
          }
        }
      }
    }

    .pagination-wrapper {
      display: flex;
      justify-content: flex-end;
      margin-top: 20px;
    }
  }
}

@media (max-width: 768px) {
  .subscriber-manage-page {
    padding: 12px;

    .page-header {
      .page-title {
        font-size: 20px;
      }
    }

    .content-area {
      padding: 12px;

      .card-list {
        grid-template-columns: 1fr;
        gap: 12px;
      }
    }
  }
}
</style>
