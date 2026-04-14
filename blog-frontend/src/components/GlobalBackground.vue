<!--
  项目名称：blog-frontend
  文件名称：GlobalBackground.vue
  创建时间：2026-04-14 15:40:07

  系统用户：Administrator
  作　　者：無以菱
  联系邮箱：huangjing510@126.com
  功能描述：全局固定背景图组件，从 app store 读取后台配置的背景图 URL，
           渲染为视口固定层，叠加亮色/暗色半透明遮罩，支持图片预加载淡入。
           仅当后台设置了背景图时渲染，未设置时不渲染，由 body CSS 渐变兜底。
-->
<template>
  <div class="global-bg">
    <!-- 背景图片层：优先后台配置，未配置时使用默认背景图 -->
    <img
      :src="bgSrc"
      alt=""
      class="global-bg-img"
      :class="{ loaded: imgLoaded }"
      @load="imgLoaded = true"
      @error="handleImgError"
    />
    <!-- 半透明遮罩层：保证内容可读 -->
    <div class="global-bg-overlay"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useAppStore } from '@/stores'

const DEFAULT_BG = '/default_background.png'

const appStore = useAppStore()
const imgLoaded = ref(false)

// 优先使用后台配置的背景图，未配置时回退到默认背景图
const bgSrc = computed(() => appStore.bgImage || DEFAULT_BG)

// 图片 URL 变化时重置加载状态
watch(bgSrc, () => {
  imgLoaded.value = false
})

// 后台图片加载失败时回退到默认图
function handleImgError() {
  imgLoaded.value = false
  if (appStore.bgImage) {
    appStore.setBgImage('')
  }
}
</script>

<style scoped>
.global-bg {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

.global-bg-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  opacity: 0;
  transition: opacity 0.8s ease;
}

.global-bg-img.loaded {
  opacity: 1;
}

/* 亮色模式遮罩：轻微提亮，保证玻璃态卡片文字可读 */
.global-bg-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.15);
  transition: background 0.3s ease;
}

/* 暗色模式遮罩：压暗背景，适配深色主题 */
html.dark .global-bg-overlay {
  background: rgba(0, 0, 0, 0.45);
}
</style>
