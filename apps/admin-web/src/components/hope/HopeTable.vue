<template>
  <div class="hope-table-wrap">
    <div v-if="$slots.toolbar" class="hope-table-toolbar">
      <slot name="toolbar" />
    </div>
    <table class="hope-table" :class="{ 'hope-table--striped': striped, 'hope-table--compact': compact }">
      <thead>
        <tr>
          <th v-for="col in columns" :key="col.prop" :class="col.sortable ? 'sortable' : ''" @click="col.sortable && col.prop && sort(col.prop)">
            {{ col.label }}
            <span v-if="col.sortable" class="sort-icon">
              <svg v-if="sortField === col.prop && sortOrder === 'asc'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="18 15 12 9 6 15"/></svg>
              <svg v-else width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
            </span>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td :colspan="columns.length" style="text-align:center;padding:32px;color:var(--hope-text-muted);">
            <el-icon :size="20" style="animation:spin 1s linear infinite"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12a9 9 0 11-6.219-8.56"/></svg></el-icon>
          </td>
        </tr>
        <tr v-else-if="!data.length">
          <td :colspan="columns.length" style="text-align:center;padding:32px;color:var(--hope-text-muted);">
            <slot name="empty">暂无数据</slot>
          </td>
        </tr>
        <tr v-for="row in data" :key="getKey(row)" :class="{ 'hope-row-selected': selected?.includes(String(getKey(row))) }">
          <td v-for="col in columns" :key="col.prop">
            <slot :name="`col-${col.prop}`" :row="row" :value="row[col.prop as keyof typeof row]">
              {{ row[col.prop as keyof typeof row] }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-if="$slots.footer || pagination" class="hope-table-footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts" generic="T extends Record<string, any>">
import { ref } from 'vue'

interface Column {
  prop: string
  label: string
  sortable?: boolean
  width?: string | number
}

const props = defineProps<{
  columns: Column[]
  data: T[]
  rowKey?: (row: T) => string | number
  loading?: boolean
  selected?: (string | number)[]
  pagination?: boolean
  striped?: boolean
  compact?: boolean
}>()

const emit = defineEmits<{ sort: [prop: string, order: 'asc' | 'desc'] }>()

const sortField = ref('')
const sortOrder = ref<'asc' | 'desc'>('asc')

function getKey(row: T): string | number {
  const keyFn = props.rowKey || ((r: T) => (r as any).id as string)
  return keyFn(row)
}

function sort(prop: string) {
  const order = sortField.value === prop && sortOrder.value === 'asc' ? 'desc' : 'asc'
  sortField.value = prop
  sortOrder.value = order
  emit('sort', prop, order)
}
</script>

<style scoped>
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.hope-row-selected { background: rgba(74,124,95,0.06) !important; }
</style>
