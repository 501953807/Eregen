<template>
  <div class="hope-demo-page">
    <div class="hope-page-header">
      <div>
        <h1 class="hope-page-header__title">Hope UI 组件库</h1>
        <p class="hope-page-header__subtitle">Materio × Eregen 设计系统 — 基础公共组件展示</p>
      </div>
      <div class="hope-page-header__actions">
        <HopeBadge color="primary" type="text">v1.0.0</HopeBadge>
      </div>
    </div>

    <!-- 设计 Token -->
    <HopeCard title="设计 Token" subtitle="Design Tokens — Materio × Eregen">
      <div class="demo-section">
        <div class="demo-section__label">颜色系统</div>
        <div class="color-grid">
          <div v-for="(hex, key) in colors" :key="key" class="color-swatch">
            <div class="color-swatch__preview" :style="{ background: hex }"></div>
            <div class="color-swatch__name">{{ key }}</div>
            <div class="color-swatch__hex">{{ hex }}</div>
          </div>
        </div>
      </div>
      <div class="demo-section">
        <div class="demo-section__label">圆角系统</div>
        <div class="radius-grid">
          <div v-for="(val, key) in radii" :key="key" class="radius-item">
            <div class="radius-item__box" :style="{ borderRadius: val }"></div>
            <div class="radius-item__label">{{ key }}: {{ val }}</div>
          </div>
        </div>
      </div>
      <div class="demo-section">
        <div class="demo-section__label">阴影系统</div>
        <div class="shadow-grid">
          <div v-for="(val, key) in shadows" :key="key" class="shadow-item">
            <div class="shadow-item__box" :style="{ boxShadow: val }"></div>
            <div class="shadow-item__label">{{ key }}</div>
            <code>{{ val }}</code>
          </div>
        </div>
      </div>
    </HopeCard>

    <!-- Buttons -->
    <HopeCard title="Buttons" subtitle="按钮 — 5 种语义色 + Accent 紫色">
      <div class="demo-section">
        <div class="demo-section__label">Variants</div>
        <div class="btn-row">
          <HopeBtn variant="filled">Filled</HopeBtn>
          <HopeBtn variant="outlined">Outlined</HopeBtn>
          <HopeBtn variant="text">Text</HopeBtn>
          <HopeBtn variant="ghost">Ghost</HopeBtn>
          <HopeBtn variant="plain">Plain</HopeBtn>
        </div>
      </div>
      <div class="demo-section">
        <div class="demo-section__label">Colors</div>
        <div class="btn-row">
          <HopeBtn variant="filled">Primary</HopeBtn>
          <HopeBtn variant="filled" color="success">Success</HopeBtn>
          <HopeBtn variant="filled" color="warning">Warning</HopeBtn>
          <HopeBtn variant="filled" color="error">Error</HopeBtn>
          <HopeBtn variant="filled" color="info">Info</HopeBtn>
          <HopeBtn variant="filled" color="accent">Accent</HopeBtn>
        </div>
      </div>
      <div class="demo-section">
        <div class="demo-section__label">Sizes</div>
        <div class="btn-row">
          <HopeBtn variant="filled" size="sm">Small</HopeBtn>
          <HopeBtn variant="filled" size="md">Medium</HopeBtn>
          <HopeBtn variant="filled" size="lg">Large</HopeBtn>
          <HopeBtn variant="filled" size="icon"><el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg></el-icon></HopeBtn>
        </div>
      </div>
      <div class="demo-section">
        <div class="demo-section__label">With Icon</div>
        <div class="btn-row">
          <HopeBtn variant="filled"><el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg></el-icon>添加</HopeBtn>
          <HopeBtn variant="outlined"><el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg></el-icon>导出</HopeBtn>
        </div>
      </div>
    </HopeCard>

    <!-- Stat Cards -->
    <HopeCard title="Stat Cards" subtitle="数据卡片 — 带顶部渐变条和趋势标签">
      <div class="hope-grid-4">
        <HopeStatCard value="1,284" label="在线设备" icon-color="primary">
          <template #icon><el-icon :size="24"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg></el-icon></template>
          <template #trend><span class="hope-stat-card__trend hope-stat-card__trend-up">↑ 2.3%</span></template>
        </HopeStatCard>
        <HopeStatCard value="568" label="活跃家属" icon-color="success">
          <template #icon><el-icon :size="24"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/></svg></el-icon></template>
          <template #trend><span class="hope-stat-card__trend hope-stat-card__trend-up">↑ 5.1%</span></template>
        </HopeStatCard>
        <HopeStatCard value="23" label="待处理告警" icon-color="warning">
          <template #icon><el-icon :size="24"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/></svg></el-icon></template>
          <template #trend><span class="hope-stat-card__trend hope-stat-card__trend-down">↓ 12.5%</span></template>
        </HopeStatCard>
        <HopeStatCard value="87%" label="设备在线率" icon-color="accent">
          <template #icon><el-icon :size="24"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg></el-icon></template>
          <template #trend><span class="hope-stat-card__trend hope-stat-card__trend-neutral">→ 0.0%</span></template>
        </HopeStatCard>
      </div>
    </HopeCard>

    <!-- Badges & Chips -->
    <HopeCard title="Badges & Chips" subtitle="徽章与标签">
      <div class="demo-section">
        <div class="demo-section__label">Badges</div>
        <div class="btn-row">
          <HopeBadge color="primary">Primary</HopeBadge>
          <HopeBadge color="success">Success</HopeBadge>
          <HopeBadge color="warning">Warning</HopeBadge>
          <HopeBadge color="error">Error</HopeBadge>
          <HopeBadge color="info">Info</HopeBadge>
          <HopeBadge color="accent">Accent</HopeBadge>
          <HopeBadge color="primary" type="dot" />
          <HopeBadge color="success" type="dot" />
          <HopeBadge color="error" type="dot" />
        </div>
      </div>
      <div class="demo-section">
        <div class="demo-section__label">Chips</div>
        <div class="btn-row">
          <HopeChip>默认</HopeChip>
          <HopeChip color="primary">手环 Starter</HopeChip>
          <HopeChip color="success">已在线</HopeChip>
          <HopeChip color="warning">待升级</HopeChip>
          <HopeChip color="error">故障</HopeChip>
          <HopeChip color="accent">医疗腕带</HopeChip>
        </div>
      </div>
    </HopeCard>

    <!-- Alerts -->
    <HopeCard title="Alerts" subtitle="告警提示">
      <HopeAlert color="primary" title="提示">
        这是一个信息提示告警，用于常规通知。
      </HopeAlert>
      <HopeAlert color="success" title="成功">
        操作已完成，数据已保存。
      </HopeAlert>
      <HopeAlert color="warning" title="警告">
        设备固件版本过旧，建议尽快升级。
      </HopeAlert>
      <HopeAlert color="error" title="错误" :closable="true">
        网络连接失败，请检查网络设置。
      </HopeAlert>
      <HopeAlert color="accent" title="Materio 风格">
        特殊操作需要使用 Accent 紫色样式。
      </HopeAlert>
    </HopeCard>

    <!-- Inputs -->
    <HopeCard title="Inputs" subtitle="输入框 — 浮动标签">
      <div class="hope-grid-2" style="gap:16px">
        <HopeInput label="设备名称" placeholder="请输入设备名称" />
        <HopeInput label="设备ID" placeholder="自动生成的设备ID" disabled />
        <HopeInput label="老人姓名" placeholder="请输入老人姓名" error="姓名不能为空" />
        <HopeInput label="备注" placeholder="选填" helper="支持最多200个字符" />
      </div>
    </HopeCard>

    <!-- Progress -->
    <HopeCard title="Progress" subtitle="进度条">
      <div class="demo-section">
        <div class="demo-section__label">Linear Progress</div>
        <div style="max-width:400px">
          <div class="hope-progress__label"><span>手环固件更新</span><span class="hope-progress__value">72%</span></div>
          <HopeProgress :value="72" color="primary" />
        </div>
      </div>
      <div class="demo-section">
        <div class="demo-section__label">Color Variants</div>
        <div style="max-width:400px;display:flex;flex-direction:column;gap:12px">
          <HopeProgress :value="85" color="success" :show-label="true" label="设备在线" />
          <HopeProgress :value="45" color="warning" :show-label="true" label="内存使用" />
          <HopeProgress :value="23" color="error" :show-label="true" label="告警处理率" />
          <HopeProgress :value="60" color="accent" :show-label="true" label="学习进度" />
        </div>
      </div>
    </HopeCard>

    <!-- Tabs -->
    <HopeCard title="Tabs" subtitle="标签导航 — line & pill">
      <div class="demo-section">
        <div class="demo-section__label">Line Tabs</div>
        <HopeTabs v-model="activeTab" :tabs="lineTabs" :animated="true" />
      </div>
      <div class="demo-section">
        <div class="demo-section__label">Pill Tabs</div>
        <HopeTabs v-model="activePillTab" :tabs="pillTabs" :pill-style="true" />
      </div>
    </HopeCard>

    <!-- Table -->
    <HopeCard title="Table" subtitle="表格 — hover高亮、条纹、排序">
      <HopeTable
        :columns="tableColumns"
        :data="tableData"
        :loading="false"
        :striped="true"
        :row-key="(row: any) => row.id"
      >
        <template #col-status="{ row }">
          <HopeBadge :color="row.status === '在线' ? 'success' : 'error'">{{ row.status }}</HopeBadge>
        </template>
        <template #col-type="{ row }">
          <HopeChip :color="row.type === '手环' ? 'primary' : 'accent'">{{ row.type }}</HopeChip>
        </template>
      </HopeTable>
    </HopeCard>

    <!-- Modal -->
    <HopeCard title="Modal" subtitle="弹窗 — slide-up 动画">
      <div class="btn-row">
        <HopeBtn variant="filled" @click="modalVisible = true">打开弹窗</HopeBtn>
        <HopeBtn variant="outlined" @click="modalLarge = true">大弹窗</HopeBtn>
      </div>
      <HopeModal v-model="modalVisible" title="确认操作" subtitle="此操作不可撤销，请确认">
        <div style="padding:8px 0">
          <p style="color:var(--hope-text-secondary);line-height:1.7">这是弹窗的内容区域。所有内容都可以通过 slot 自定义。</p>
        </div>
        <template #footer>
          <HopeBtn variant="ghost" @click="modalVisible = false">取消</HopeBtn>
          <HopeBtn variant="filled" @click="modalVisible = false">确认</HopeBtn>
        </template>
      </HopeModal>
      <HopeModal v-model="modalLarge" title="大弹窗示例" size="lg">
        <div style="padding:8px 0">
          <p style="color:var(--hope-text-secondary)">宽度 720px 的大弹窗，适合表单和长内容。</p>
        </div>
        <template #footer>
          <HopeBtn variant="ghost" @click="modalLarge = false">取消</HopeBtn>
          <HopeBtn variant="filled" @click="modalLarge = false">保存</HopeBtn>
        </template>
      </HopeModal>
    </HopeCard>

    <!-- Avatars -->
    <HopeCard title="Avatars" subtitle="头像 — 5 种尺寸 + 在线状态">
      <div class="btn-row" style="align-items:center">
        <HopeAvatar name="管" size="xs" status="online" />
        <HopeAvatar name="张" size="sm" status="online" />
        <HopeAvatar name="李" size="md" status="away" />
        <HopeAvatar name="王" size="lg" status="busy" />
        <HopeAvatar name="赵" size="xl" status="offline" />
      </div>
      <div class="demo-section" style="margin-top:16px">
        <div class="demo-section__label">Avatar Group</div>
        <div class="hope-avatar-group">
          <HopeAvatar name="张" />
          <HopeAvatar name="李" />
          <HopeAvatar name="王" />
          <HopeAvatar name="赵" />
          <HopeAvatar name="孙" />
        </div>
      </div>
    </HopeCard>

    <!-- Timeline -->
    <HopeCard title="Timeline" subtitle="时间线 — 垂直（带连接线）">
      <HopeTimeline :items="timelineItems" />
    </HopeCard>

    <!-- Accordion -->
    <HopeCard title="Accordion" subtitle="手风琴 — 图标旋转动画">
      <HopeAccordion :items="accordionItems" />
    </HopeCard>

    <!-- Breadcrumb -->
    <HopeCard title="Breadcrumb" subtitle="面包屑导航">
      <HopeBreadcrumb :crumbs="breadcrumbCrumbs" />
    </HopeCard>

    <!-- Skeleton -->
    <HopeCard title="Skeleton" subtitle="骨架屏加载状态">
      <div style="display:flex;gap:20px;align-items:flex-start">
        <div style="flex:1">
          <div class="hope-skeleton hope-skeleton--title" style="width:60%"></div>
          <div class="hope-skeleton hope-skeleton--text"></div>
          <div class="hope-skeleton hope-skeleton--text" style="width:80%"></div>
          <div class="hope-skeleton hope-skeleton--text"></div>
        </div>
        <div class="hope-skeleton hope-skeleton--circle" style="width:48px;height:48px;flex-shrink:0"></div>
      </div>
    </HopeCard>

    <!-- Tooltip -->
    <HopeCard title="Tooltip" subtitle="浮动提示">
      <div class="btn-row">
        <HopeTooltip content="这是提示内容">
          <HopeBtn variant="plain">悬停查看</HopeBtn>
        </HopeTooltip>
        <HopeTooltip content="快速提示">
          <span style="color:var(--hope-text-muted);cursor:help;font-size:20px">ⓘ</span>
        </HopeTooltip>
      </div>
    </HopeCard>

    <!-- Dropdown -->
    <HopeCard title="Dropdown" subtitle="下拉菜单 — 带阴影和圆角">
      <HopeDropdown>
        <template #trigger>
          <HopeBtn variant="plain">操作菜单 ▼</HopeBtn>
        </template>
        <div class="hope-dropdown-header">用户操作</div>
        <div class="hope-dropdown-item"><el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg></el-icon>个人资料</div>
        <div class="hope-dropdown-item"><el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg></el-icon>系统设置</div>
        <div class="hope-dropdown-divider"></div>
        <div class="hope-dropdown-item danger"><el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg></el-icon>退出登录</div>
      </HopeDropdown>
    </HopeCard>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElIcon } from 'element-plus'
import {
  HopeTheme, HopeBtn, HopeCard, HopeStatCard, HopeBadge, HopeChip,
  HopeAlert, HopeInput, HopeProgress, HopeTabs, HopeTable, HopeModal,
  HopeAvatar, HopeTimeline, HopeAccordion, HopeBreadcrumb, HopeSkeleton,
  HopeTooltip, HopeDropdown
} from '@/components/hope'

// Color tokens
const colors: Record<string, string> = {
  'Primary': '#4A7C5F',
  'Primary Light': '#6FAF8F',
  'Primary Dark': '#365C45',
  'Accent': '#8C57FF',
  'Accent Dark': '#7E4EE6',
  'Success': '#56CA00',
  'Warning': '#C49A2C',
  'Error': '#C04A42',
  'Info': '#4A8FB8',
  'Text Primary': '#1A2E26',
  'Text Secondary': '#3D5A4E',
  'Text Muted': '#94A9A2',
  'Bg Page': '#F4F6F4',
  'Bg Surface': '#FFFFFF',
}

const radii: Record<string, string> = {
  'sm': '4px',
  'md': '6px',
  'lg': '10px',
  'xl': '14px',
  '2xl': '20px',
  'pill': '50px',
}

const shadows: Record<string, string> = {
  'sm': '0 2px 4px rgba(26,46,38,0.08)',
  'md': '0 4px 12px rgba(26,46,38,0.10)',
  'lg': '0 8px 24px rgba(26,46,38,0.12)',
  'xl': '0 16px 48px rgba(26,46,38,0.14)',
  'primary': '0 2px 8px rgba(74,124,95,0.20)',
}

const activeTab = ref('overview')
const activePillTab = ref('all')
const modalVisible = ref(false)
const modalLarge = ref(false)

const lineTabs = [
  { label: '概览', value: 'overview' },
  { label: '设备', value: 'devices', badge: 12 },
  { label: '告警', value: 'alerts' },
  { label: '设置', value: 'settings', disabled: true },
]

const pillTabs = [
  { label: '全部', value: 'all' },
  { label: '手环', value: 'bracelet' },
  { label: '药盒', value: 'pillbox' },
  { label: '腕带', value: 'wristband' },
]

const tableColumns = [
  { prop: 'id', label: '设备ID', sortable: true },
  { prop: 'name', label: '设备名称' },
  { prop: 'type', label: '类型' },
  { prop: 'status', label: '状态' },
  { prop: 'battery', label: '电量' },
]

const tableData = [
  { id: 'BR-001', name: '爷爷手环', type: '手环', status: '在线', battery: '87%' },
  { id: 'PX-002', name: '奶奶药盒', type: '药盒', status: '在线', battery: '92%' },
  { id: 'BR-003', name: '外公手环', type: '手环', status: '离线', battery: '15%' },
  { id: 'WB-004', name: '住院号腕带', type: '腕带', status: '在线', battery: '100%' },
  { id: 'BR-005', name: '外婆手环', type: '手环', status: '在线', battery: '64%' },
]

const timelineItems = [
  { title: '设备绑定成功', meta: '2026-08-13 14:30', body: '手环 BR-001 已成功绑定到老人张三档案', color: 'success' as const },
  { title: '用药提醒已设置', meta: '2026-08-13 10:00', body: '已为爷爷配置每日早中晚三次用药提醒规则', color: 'primary' as const },
  { title: '告警已处理', meta: '2026-08-12 22:15', body: '心率异常告警 P0 已由值班护士确认并处理', color: 'warning' as const },
  { title: '固件升级完成', meta: '2026-08-12 16:00', body: '药盒 PX-002 已成功升级至 v2.4.1', color: 'info' as const },
]

const accordionItems = [
  { title: '如何绑定设备？', body: '在设备管理页面点击「注册设备」，输入设备 SN 码即可绑定到对应老人档案。' },
  { title: '告警优先级说明', body: 'P0 紧急告警（SOS/跌倒）会立即推送至家属APP和短信；P1 重要告警（心率异常）通过APP推送；P2 通知类（电量低）纳入每日汇总。' },
  { title: '用药规则配置', body: '可为每位老人配置多条用药规则，包括时间、剂量、药品类型和备注。系统会在用药时间前 5 分钟发送提醒。' },
]

const breadcrumbCrumbs = [
  { text: '首页', href: '#' },
  { text: '设备管理', href: '#' },
  { text: '手环设备' },
]
</script>

<style scoped>
.demo-page {
  max-width: 1200px;
  margin: 0 auto;
}
.demo-section {
  margin-bottom: 20px;
}
.demo-section:last-child { margin-bottom: 0; }
.demo-section__label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--hope-text-muted);
  margin-bottom: 10px;
}
.btn-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}
.color-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 12px;
}
.color-swatch {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}
.color-swatch__preview {
  width: 100%;
  height: 48px;
  border-radius: var(--hope-radius-md);
  box-shadow: var(--hope-shadow-sm);
}
.color-swatch__name {
  font-size: 11px;
  font-weight: 600;
  color: var(--hope-text-secondary);
  text-align: center;
}
.color-swatch__hex {
  font-size: 11px;
  font-family: 'JetBrains Mono', monospace;
  color: var(--hope-text-muted);
}
.radius-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}
.radius-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}
.radius-item__box {
  width: 48px;
  height: 48px;
  background: var(--hope-primary);
}
.radius-item__label {
  font-size: 11px;
  font-family: 'JetBrains Mono', monospace;
  color: var(--hope-text-muted);
}
.shadow-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}
.shadow-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px;
}
.shadow-item__box {
  width: 64px;
  height: 64px;
  background: var(--hope-surface);
  border-radius: var(--hope-radius-lg);
}
.shadow-item__label {
  font-size: 11px;
  font-weight: 600;
  color: var(--hope-text-secondary);
}
.shadow-item code {
  font-size: 10px;
  font-family: 'JetBrains Mono', monospace;
  color: var(--hope-text-muted);
  max-width: 120px;
  text-align: center;
}
</style>
