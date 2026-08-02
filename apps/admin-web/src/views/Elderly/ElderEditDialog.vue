<template>
  <el-dialog v-model="visible" :title="'编辑老人档案 — ' + (row?.name || '')" width="720px" destroy-on-close>
    <el-form :model="form" label-width="100px" label-position="right">
      <el-row :gutter="24">
        <el-col :span="12">
          <el-form-item label="姓名" prop="name">
            <el-input v-model="form.name" placeholder="请输入老人姓名" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="性别" prop="gender">
            <el-select v-model="form.gender" placeholder="请选择性别">
              <el-option label="男" value="男" />
              <el-option label="女" value="女" />
              <el-option label="未知" value="" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="24">
        <el-col :span="12">
          <el-form-item label="身份证号" prop="id_card">
            <el-input v-model="form.id_card" placeholder="请输入18位身份证号" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="出生日期" prop="birth_date">
            <el-date-picker v-model="form.birth_date" type="date" placeholder="选择出生日期" value-format="YYYY-MM-DD" format="YYYY-MM-DD" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="紧急联系人" prop="emergency_contact">
        <el-input v-model="form.emergency_contact" placeholder="紧急联系人姓名及关系（如：张三（子））" />
      </el-form-item>
      <el-form-item label="地址" prop="address">
        <el-input v-model="form.address" placeholder="请输入家庭住址" style="width: 100%;" />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-select v-model="form.status" placeholder="请选择状态">
          <el-option label="正常" value="正常" />
          <el-option label="停用" value="停用" />
          <el-option label="失联" value="失联" />
        </el-select>
      </el-form-item>
      <div style="margin-top: 24px; padding: 16px; background: #f5f7fa; border-radius: 8px;">
        <h4 style="margin: 0 0 12px 0; font-size: 14px; color: #555;">关联福利标签（编辑后自动保存）</h4>
        <div style="display: flex; flex-wrap: wrap; gap: 8px;">
          <el-tag v-for="tag in (row?.welfare_tags || [])" :key="tag.code" type="info" size="small">{{ tag.name }}</el-tag>
          <el-tag type="plain" size="small" style="cursor: pointer;" @click="$emit('add-tag')">＋ 添加标签</el-tag>
        </div>
      </div>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="$emit('save', form)">保存更改</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface WelfareTag { code: string; name: string }
interface ElderlyRow { id: string; name: string; id_card?: string; birth_date?: string; gender?: string; emergency_contact?: string; welfare_tags: WelfareTag[]; status: string; address?: string }
interface EditForm { name: string; id_card: string; birth_date: string; gender: string; emergency_contact: string; address: string; status: string }

const props = defineProps<{ modelValue: boolean; row: ElderlyRow | null; initialForm: EditForm }>()
const emit = defineEmits<{ 'update:modelValue': [v: boolean]; save: [form: EditForm]; 'add-tag': [] }>()

const visible = ref(false)
const form = ref<EditForm>({ ...props.initialForm })

watch(() => props.modelValue, v => { visible.value = v; if (v) form.value = { ...props.initialForm } })
watch(visible, v => { if (!v) emit('update:modelValue', false) })
</script>
