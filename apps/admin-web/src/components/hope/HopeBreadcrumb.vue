<template>
  <div class="hope-breadcrumb">
    <template v-for="(crumb, i) in crumbs" :key="i">
      <span v-if="i > 0" class="hope-breadcrumb-sep">/</span>
      <span v-if="i < crumbs.length - 1" class="hope-breadcrumb-item">
        <a :href="crumb.href || '#'" @click.prevent="crumb.click && crumb.click()">{{ crumb.text }}</a>
      </span>
      <span v-else class="hope-breadcrumb-item active">{{ crumb.text }}</span>
    </template>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  crumbs: Array<{ text: string; href?: string; click?: () => void }>
}>()
</script>

<style scoped>
.hope-breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--hope-text-muted);
  padding: 8px 0;
}
.hope-breadcrumb-sep {
  color: var(--hope-text-muted);
  opacity: 0.5;
  margin: 0 2px;
}
.hope-breadcrumb-item a {
  color: var(--hope-text-secondary);
  text-decoration: none;
  transition: color 0.15s;
}
.hope-breadcrumb-item a:hover {
  color: var(--hope-primary);
}
.hope-breadcrumb-item.active {
  color: var(--hope-text);
  font-weight: 500;
}
</style>
