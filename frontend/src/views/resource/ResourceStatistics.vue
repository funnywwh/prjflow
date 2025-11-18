<template>
  <div class="resource-statistics">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="资源统计">
            <template #extra>
              <a-space>
                <a-range-picker
                  v-model:value="dateRange"
                  @change="handleDateRangeChange"
                />
                <a-button type="primary" @click="handleSearch">查询</a-button>
              </a-space>
            </template>
          </a-page-header>

          <a-card :bordered="false" style="margin-bottom: 16px">
            <a-form layout="inline" :model="searchForm">
              <a-form-item label="用户">
                <a-select
                  v-model:value="searchForm.user_id"
                  placeholder="选择用户"
                  allow-clear
                  show-search
                  :filter-option="filterUserOption"
                  style="width: 150px"
                >
                  <a-select-option
                    v-for="user in users"
                    :key="user.id"
                    :value="user.id"
                  >
                    {{ user.nickname || user.username }}({{ user.username }})
                  </a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item label="项目">
                <a-select
                  v-model:value="searchForm.project_id"
                  placeholder="选择项目"
                  allow-clear
                  style="width: 150px"
                >
                  <a-select-option
                    v-for="project in projects"
                    :key="project.id"
                    :value="project.id"
                  >
                    {{ project.name }}
                  </a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item>
                <a-button type="primary" @click="handleSearch">查询</a-button>
                <a-button style="margin-left: 8px" @click="handleReset">重置</a-button>
              </a-form-item>
            </a-form>
          </a-card>

          <a-row :gutter="16" style="margin-bottom: 16px">
            <a-col :span="8">
              <a-statistic
                title="总工时"
                :value="statistics.total_hours"
                suffix="小时"
                :precision="2"
              />
            </a-col>
          </a-row>

          <a-row :gutter="16">
            <a-col :span="12">
              <a-card title="按项目统计" :bordered="false">
                <a-table
                  :columns="projectColumns"
                  :data-source="statistics.project_stats"
                  :pagination="false"
                  size="small"
                  row-key="project_id"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'total_hours'">
                      <span>{{ record.total_hours.toFixed(2) }} 小时</span>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-col>
            <a-col :span="12">
              <a-card title="按人员统计" :bordered="false">
                <a-table
                  :columns="userColumns"
                  :data-source="statistics.user_stats"
                  :pagination="false"
                  size="small"
                  row-key="user_id"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'user'">
                      <span>{{ record.nickname || record.username }}({{ record.username }})</span>
                    </template>
                    <template v-else-if="column.key === 'total_hours'">
                      <span>{{ record.total_hours.toFixed(2) }} 小时</span>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-col>
          </a-row>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import type { Dayjs } from 'dayjs'
import dayjs from 'dayjs'
import { DatePicker } from 'ant-design-vue'
import AppHeader from '@/components/AppHeader.vue'

const { RangePicker } = DatePicker
import { getResourceStatistics, type ResourceStatistics } from '@/api/resource'
import { getUsers } from '@/api/user'
import { getProjects } from '@/api/project'
import type { User } from '@/types/user'
import type { Project } from '@/api/project'

const loading = ref(false)
const users = ref<User[]>([])
const projects = ref<Project[]>([])
const statistics = ref<ResourceStatistics>({
  total_hours: 0,
  project_stats: [],
  user_stats: []
})
const dateRange = ref<[Dayjs, Dayjs] | null>(null)

const searchForm = reactive({
  user_id: undefined as number | undefined,
  project_id: undefined as number | undefined
})

const projectColumns = [
  { title: '项目名称', dataIndex: 'project_name', key: 'project_name' },
  { title: '总工时', key: 'total_hours', width: 150 }
]

const userColumns = [
  { title: '人员', key: 'user' },
  { title: '总工时', key: 'total_hours', width: 150 }
]

const filterUserOption = (input: string, option: any) => {
  const user = users.value.find(u => u.id === option.value)
  if (!user) return false
  const nickname = user.nickname || ''
  const username = user.username || ''
  return nickname.toLowerCase().includes(input.toLowerCase()) ||
    username.toLowerCase().includes(input.toLowerCase())
}

const loadUsers = async () => {
  try {
    const res = await getUsers({ page_size: 1000 })
    users.value = res.list || []
  } catch (error: any) {
    message.error('加载用户列表失败: ' + (error.response?.data?.message || error.message))
  }
}

const loadProjects = async () => {
  try {
    const res = await getProjects({ page_size: 1000 })
    projects.value = res.list || []
  } catch (error: any) {
    message.error('加载项目列表失败: ' + (error.response?.data?.message || error.message))
  }
}

const loadStatistics = async () => {
  loading.value = true
  try {
    const params: any = {}
    if (searchForm.user_id) params.user_id = searchForm.user_id
    if (searchForm.project_id) params.project_id = searchForm.project_id
    if (dateRange.value && dateRange.value[0] && dateRange.value[1]) {
      params.start_date = dateRange.value[0].format('YYYY-MM-DD')
      params.end_date = dateRange.value[1].format('YYYY-MM-DD')
    }

    const res = await getResourceStatistics(params)
    statistics.value = res
  } catch (error: any) {
    message.error('加载统计数据失败: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadStatistics()
}

const handleReset = () => {
  searchForm.user_id = undefined
  searchForm.project_id = undefined
  dateRange.value = null
  loadStatistics()
}

const handleDateRangeChange = () => {
  loadStatistics()
}

onMounted(() => {
  loadUsers()
  loadProjects()
  loadStatistics()
})
</script>

<style scoped>
.resource-statistics {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 1400px;
  margin: 0 auto;
  background: #fff;
  padding: 24px;
  border-radius: 8px;
}
</style>

