<template>
  <div class="backup-settings">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="备份设置">
            <template #extra>
              <a-button type="primary" @click="handleSave" :loading="loading">
                保存配置
              </a-button>
            </template>
          </a-page-header>

          <a-card :bordered="false">
            <a-form
              :model="backupConfig"
              :rules="backupRules"
              @finish="handleSave"
              layout="vertical"
            >
              <a-divider orientation="left">自动备份配置</a-divider>
              
              <a-alert
                message="配置说明"
                description="设置自动备份数据库的时间。备份文件将保存在数据库文件所在目录的 backups 文件夹中，自动压缩并保留最近7天的备份。"
                type="info"
                show-icon
                style="margin-bottom: 24px"
              />

              <a-form-item label="启用自动备份" name="enabled">
                <a-switch
                  v-model:checked="backupConfig.enabled"
                  checked-children="开启"
                  un-checked-children="关闭"
                />
                <div style="margin-top: 8px; color: #999; font-size: 12px;">
                  开启后，系统将在指定时间自动备份数据库
                </div>
              </a-form-item>

              <a-form-item 
                v-if="backupConfig.enabled"
                label="备份时间" 
                name="backup_time"
              >
                <a-time-picker
                  v-model:value="backupTime"
                  format="HH:mm"
                  placeholder="选择备份时间"
                  style="width: 200px"
                  @change="handleTimeChange"
                />
                <div style="margin-top: 8px; color: #999; font-size: 12px;">
                  建议在系统低峰期执行备份（如凌晨2点）
                </div>
              </a-form-item>

              <a-form-item v-if="backupConfig.last_backup_date">
                <div style="color: #666; font-size: 14px;">
                  上次备份时间：{{ backupConfig.last_backup_date }}
                </div>
              </a-form-item>

              <a-divider orientation="left">手动备份</a-divider>

              <a-form-item>
                <a-button
                  type="primary"
                  @click="handleTriggerBackup"
                  :loading="backupLoading"
                  :disabled="backupLoading"
                >
                  立即备份
                </a-button>
                <div style="margin-top: 8px; color: #999; font-size: 12px;">
                  手动触发数据库备份，备份完成后会更新上次备份时间
                </div>
              </a-form-item>

              <a-form-item>
                <a-button
                  type="primary"
                  html-type="submit"
                  size="large"
                  :loading="loading"
                >
                  保存配置
                </a-button>
              </a-form-item>
            </a-form>
          </a-card>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import dayjs, { Dayjs } from 'dayjs'
import { getBackupConfig, saveBackupConfig, triggerBackup, type BackupConfigRequest } from '@/api/system'
import AppHeader from '@/components/AppHeader.vue'

const loading = ref(false)
const backupLoading = ref(false)

const backupConfig = ref({
  enabled: false,
  backup_time: '02:00',
  last_backup_date: ''
})

const backupTime = ref<Dayjs | null>(null)

// 将 backup_time 转换为 dayjs 对象
const updateBackupTime = () => {
  if (backupConfig.value.backup_time) {
    const [hour, minute] = backupConfig.value.backup_time.split(':')
    backupTime.value = dayjs().hour(parseInt(hour)).minute(parseInt(minute)).second(0).millisecond(0)
  } else {
    backupTime.value = dayjs().hour(2).minute(0).second(0).millisecond(0)
  }
}

// 监听 backupTime 变化，更新 backup_time
const handleTimeChange = (time: Dayjs | null) => {
  if (time) {
    backupConfig.value.backup_time = time.format('HH:mm')
  }
}

// 验证备份时间
const validateBackupTime = (_rule: any, value: string) => {
  if (!backupConfig.value.enabled) {
    return Promise.resolve()
  }
  if (!value) {
    return Promise.reject('请选择备份时间')
  }
  const timePattern = /^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$/
  if (!timePattern.test(value)) {
    return Promise.reject('备份时间格式错误，应为 HH:mm (24小时制)')
  }
  return Promise.resolve()
}

const backupRules = {
  backup_time: [
    { validator: validateBackupTime, trigger: 'change' }
  ]
}

// 加载备份配置
const loadBackupConfig = async () => {
  loading.value = true
  try {
    const config = await getBackupConfig()
    backupConfig.value = {
      enabled: config.enabled || false,
      backup_time: config.backup_time || '02:00',
      last_backup_date: config.last_backup_date || ''
    }
    updateBackupTime()
  } catch (error: any) {
    message.error('加载配置失败: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

// 保存备份配置
const handleSave = async () => {
  if (backupConfig.value.enabled && !backupConfig.value.backup_time) {
    message.error('请选择备份时间')
    return
  }

  loading.value = true
  try {
    const configToSave: BackupConfigRequest = {
      enabled: backupConfig.value.enabled,
      backup_time: backupConfig.value.backup_time || '02:00'
    }
    await saveBackupConfig(configToSave)
    message.success('备份配置保存成功')
    await loadBackupConfig()
  } catch (error: any) {
    message.error(error.response?.data?.message || error.message || '保存配置失败')
  } finally {
    loading.value = false
  }
}

// 手动触发备份
const handleTriggerBackup = async () => {
  backupLoading.value = true
  try {
    await triggerBackup()
    message.success('备份已触发，正在执行中...')
    // 等待一下再重新加载配置，给备份一些时间完成
    setTimeout(async () => {
      await loadBackupConfig()
      message.success('备份已完成，请查看服务器日志确认备份状态')
    }, 2000)
  } catch (error: any) {
    const errorMsg = error.response?.data?.message || error.message || '触发备份失败'
    if (errorMsg.includes('正在进行中')) {
      message.warning(errorMsg)
    } else {
      message.error(errorMsg)
    }
  } finally {
    backupLoading.value = false
  }
}

onMounted(async () => {
  await loadBackupConfig()
})
</script>

<style scoped>
.backup-settings {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 1200px;
  margin: 0 auto;
  background: #fff;
  padding: 24px;
  border-radius: 8px;
}

:deep(.ant-divider) {
  margin: 24px 0 16px 0;
}
</style>

