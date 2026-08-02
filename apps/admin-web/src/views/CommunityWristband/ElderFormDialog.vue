<template>
  <el-dialog v-model="visible" :title="editing ? '编辑老人' : '新增老人'" width="640px" destroy-on-close>
    <el-form :model="form" label-width="100px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="姓名"><el-input v-model="form.name" placeholder="请输入姓名" /></el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="性别">
            <el-radio-group v-model="form.gender">
              <el-radio :value="1">男</el-radio>
              <el-radio :value="2">女</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="身份证号"><el-input v-model="form.id_card" placeholder="18 位身份证号码" /></el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="年龄"><el-input-number v-model="form.age" :min="0" :max="150" style="width:100%;" /></el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="地址"><el-input v-model="form.address" type="textarea" :rows="2" placeholder="居住地址" /></el-form-item>
      <el-form-item label="紧急联系人"><el-input v-model="form.emergency_contact" placeholder="姓名 + 电话" /></el-form-item>
      <el-form-item label="所属医院">
        <el-select v-model="form.hospital_id" placeholder="请选择" style="width:100%;">
          <el-option label="社区医院 A" value="hospital-a" />
          <el-option label="社区医院 B" value="hospital-b" />
          <el-option label="社区医院 C" value="hospital-c" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-radio-group v-model="form.status">
          <el-radio value="active">正常</el-radio>
          <el-radio value="deactivated">停用</el-radio>
          <el-radio value="deceased">deceased</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="$emit('save', form)">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{ modelValue: boolean; editing: boolean; initialForm: Record<string, any> }>()
const emit = defineEmits<{ 'update:modelValue': [v: boolean]; save: [form: Record<string, any>] }>()

const visible = ref(false)
const form = ref<Record<string, any>>({ gender: 1, status: 'active' })

watch(() => props.modelValue, v => { visible.value = v; if (v) form.value = { ...props.initialForm } })
watch(visible, v => { if (!v) emit('update:modelValue', false) })
</script>
