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
