# Web UI 调试界面

Cosy 框架的调试功能提供了一个现代化的 Web UI 界面，让您可以直观地监控和调试应用程序。

## 在线演示

<div style="position: relative; border: 1px solid var(--vp-c-divider); border-radius: 8px; overflow: hidden; margin: 16px 0;">
  <iframe id="debug-demo-iframe" src="/debug-demo/" style="width: 100%; height: 680px; border: none;"></iframe>
  <button
    id="debug-demo-fullscreen"
    title="全屏"
    style="position: absolute; top: 8px; right: 8px; z-index: 10; width: 32px; height: 32px; border: none; border-radius: 6px; background: rgba(0,0,0,0.45); color: #fff; cursor: pointer; display: flex; align-items: center; justify-content: center; opacity: 0.6; transition: opacity 0.2s;"
    onmouseenter="this.style.opacity='1'"
    onmouseleave="this.style.opacity='0.6'"
    onclick="(function(btn){var c=btn.parentElement;if(document.fullscreenElement){document.exitFullscreen()}else{c.requestFullscreen()}})(this)">
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
      <path id="fs-icon" d="M1.5 1a.5.5 0 0 0-.5.5v4a.5.5 0 0 1-1 0v-4A1.5 1.5 0 0 1 1.5 0h4a.5.5 0 0 1 0 1h-4zM10 .5a.5.5 0 0 1 .5-.5h4A1.5 1.5 0 0 1 16 1.5v4a.5.5 0 0 1-1 0v-4a.5.5 0 0 0-.5-.5h-4a.5.5 0 0 1-.5-.5zM.5 10a.5.5 0 0 1 .5.5v4a.5.5 0 0 0 .5.5h4a.5.5 0 0 1 0 1h-4A1.5 1.5 0 0 1 0 14.5v-4a.5.5 0 0 1 .5-.5zm15 0a.5.5 0 0 1 .5.5v4a1.5 1.5 0 0 1-1.5 1.5h-4a.5.5 0 0 1 0-1h4a.5.5 0 0 0 .5-.5v-4a.5.5 0 0 1 .5-.5z"/>
    </svg>
  </button>
</div>

<script setup>
import { onMounted, onUnmounted } from 'vue'

function onFullscreenChange() {
  const icon = document.getElementById('fs-icon')
  const iframe = document.getElementById('debug-demo-iframe')
  if (!icon) return
  if (document.fullscreenElement) {
    icon.setAttribute('d', 'M5.5 0a.5.5 0 0 1 .5.5v4A1.5 1.5 0 0 1 4.5 6h-4a.5.5 0 0 1 0-1h4a.5.5 0 0 0 .5-.5v-4a.5.5 0 0 1 .5-.5zm5 0a.5.5 0 0 1 .5.5v4a.5.5 0 0 0 .5.5h4a.5.5 0 0 1 0 1h-4A1.5 1.5 0 0 1 10 4.5v-4a.5.5 0 0 1 .5-.5zM0 10.5a.5.5 0 0 1 .5-.5h4A1.5 1.5 0 0 1 6 11.5v4a.5.5 0 0 1-1 0v-4a.5.5 0 0 0-.5-.5h-4a.5.5 0 0 1-.5-.5zm10 0a.5.5 0 0 1 .5-.5h4a.5.5 0 0 1 0 1h-4a.5.5 0 0 0-.5.5v4a.5.5 0 0 1-1 0v-4z')
    if (iframe) iframe.style.height = '100vh'
  } else {
    icon.setAttribute('d', 'M1.5 1a.5.5 0 0 0-.5.5v4a.5.5 0 0 1-1 0v-4A1.5 1.5 0 0 1 1.5 0h4a.5.5 0 0 1 0 1h-4zM10 .5a.5.5 0 0 1 .5-.5h4A1.5 1.5 0 0 1 16 1.5v4a.5.5 0 0 1-1 0v-4a.5.5 0 0 0-.5-.5h-4a.5.5 0 0 1-.5-.5zM.5 10a.5.5 0 0 1 .5.5v4a.5.5 0 0 0 .5.5h4a.5.5 0 0 1 0 1h-4A1.5 1.5 0 0 1 0 14.5v-4a.5.5 0 0 1 .5-.5zm15 0a.5.5 0 0 1 .5.5v4a1.5 1.5 0 0 1-1.5 1.5h-4a.5.5 0 0 1 0-1h4a.5.5 0 0 0 .5-.5v-4a.5.5 0 0 1 .5-.5z')
    if (iframe) iframe.style.height = '680px'
  }
}

onMounted(() => {
  document.addEventListener('fullscreenchange', onFullscreenChange)
})
onUnmounted(() => {
  document.removeEventListener('fullscreenchange', onFullscreenChange)
})
</script>

::: tip 提示
以上为交互式演示，使用模拟数据展示。您可以点击导航栏切换不同的监控页面，点击「查看」按钮查看详细信息。
:::

## 界面特性

### 实时监控仪表板
- **系统状态**：CPU、内存使用率实时显示
- **Goroutine 状态**：活跃、完成、失败的 goroutine 数量统计
- **请求状态**：当前处理中的 HTTP 请求数量
- **连接状态**：WebSocket 连接数和状态

### Goroutine 监控
- **列表视图**：显示所有 goroutine 的基本信息
- **详细视图**：查看特定 goroutine 的完整调用栈
- **实时更新**：通过 WebSocket 实时更新 goroutine 状态
- **过滤功能**：按状态（运行中、已完成、失败）过滤
- **搜索功能**：按名称或 ID 搜索特定 goroutine

### 请求监控
- **请求列表**：显示所有 HTTP 请求的详细信息
- **请求详情**：查看请求头、响应状态、处理时间等
- **请求链路**：追踪请求在不同组件中的处理过程
- **性能分析**：识别慢请求和性能瓶颈

### 内存分析
- **堆内存视图**：可视化内存分配和使用情况
- **垃圾回收统计**：GC 次数、暂停时间等指标
- **内存泄漏检测**：识别可能的内存泄漏问题