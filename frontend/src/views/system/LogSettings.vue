<template>
  <div class="log-settings">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="日志设置">
            <template #extra>
              <a-button type="primary" @click="handleSave" :loading="loading">
                保存配置
              </a-button>
            </template>
          </a-page-header>

          <a-card :bordered="false">
            <a-form
              :model="logConfig"
              :rules="logRules"
              @finish="handleSave"
              layout="vertical"
            >
              <a-divider orientation="left">日志级别设置</a-divider>
              
              <a-alert
                message="配置说明"
                description="设置系统日志级别。日志级别从低到高：debug（调试）< info（信息）< warn（警告）< error（错误）。选择某个级别后，系统会记录该级别及以上的所有日志。"
                type="info"
                show-icon
                style="margin-bottom: 24px"
              />

              <a-form-item label="日志级别" name="level">
                <a-select
                  v-model:value="logConfig.level"
                  placeholder="选择日志级别"
                  style="width: 200px"
                >
                  <a-select-option value="debug">Debug（调试）</a-select-option>
                  <a-select-option value="info">Info（信息）</a-select-option>
                  <a-select-option value="warn">Warn（警告）</a-select-option>
                  <a-select-option value="error">Error（错误）</a-select-option>
                </a-select>
                <div style="margin-top: 8px; color: #999; font-size: 12px;">
                  当前日志级别：{{ logConfig.level }}
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

          <a-card :bordered="false" style="margin-top: 24px">
            <template #title>
              <span>日志文件列表</span>
              <a-button
                type="link"
                @click="loadLogFiles"
                :loading="filesLoading"
                style="margin-left: 16px"
              >
                刷新
              </a-button>
            </template>

            <a-table
              :columns="columns"
              :data-source="logFiles"
              :loading="filesLoading"
              :pagination="false"
              row-key="filename"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'action'">
                  <a-button
                    type="link"
                    @click="handleDownload(record.filename)"
                    :loading="downloading === record.filename"
                  >
                    下载
                  </a-button>
                </template>
              </template>
            </a-table>
          </a-card>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  getLogLevel,
  setLogLevel,
  getLogFiles,
  downloadLogFile,
  type LogLevelRequest,
  type LogFileInfo
} from '@/api/system'
import AppHeader from '@/components/AppHeader.vue'

const loading = ref(false)
const filesLoading = ref(false)
const downloading = ref<string | null>(null)

const logConfig = ref({
  level: 'error'
})

const logFiles = ref<LogFileInfo[]>([])

const columns = [
  {
    title: '文件名',
    dataIndex: 'filename',
    key: 'filename'
  },
  {
    title: '大小',
    dataIndex: 'size_formatted',
    key: 'size'
  },
  {
    title: '修改时间',
    dataIndex: 'mod_time',
    key: 'mod_time'
  },
  {
    title: '操作',
    key: 'action',
    width: 100
  }
]

const logRules = {
  level: [
    { required: true, message: '请选择日志级别', trigger: 'change' }
  ]
}

// 加载日志级别
const loadLogLevel = async () => {
  loading.value = true
  try {
    const result = await getLogLevel()
    logConfig.value.level = result.level || 'error'
  } catch (error: any) {
    message.error('加载日志级别失败: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

// 保存日志级别
const handleSave = async () => {
  if (!logConfig.value.level) {
    message.error('请选择日志级别')
    return
  }

  loading.value = true
  try {
    const configToSave: LogLevelRequest = {
      level: logConfig.value.level
    }
    await setLogLevel(configToSave)
    message.success('日志级别保存成功')
    await loadLogLevel()
  } catch (error: any) {
    message.error(error.response?.data?.message || error.message || '保存配置失败')
  } finally {
    loading.value = false
  }
}

// 加载日志文件列表
const loadLogFiles = async () => {
  filesLoading.value = true
  try {
    const result = await getLogFiles()
    logFiles.value = result.files || []
  } catch (error: any) {
    message.error('加载日志文件列表失败: ' + (error.response?.data?.message || error.message))
  } finally {
    filesLoading.value = false
  }
}

// 下载日志文件
const handleDownload = async (filename: string) => {
  downloading.value = filename
  try {
    await downloadLogFile(filename)
    message.success('日志文件下载成功')
  } catch (error: any) {
    message.error('下载日志文件失败: ' + (error.response?.data?.message || error.message))
  } finally {
    downloading.value = null
  }
}

onMounted(async () => {
  await loadLogLevel()
  await loadLogFiles()
})
</script>

<style scoped>
.log-settings {
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

