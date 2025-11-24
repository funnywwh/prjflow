<template>
  <div class="resource-utilization">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="资源利用率分析">
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
                  @change="handleSearchProjectChange"
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
                title="平均利用率"
                :value="utilization.avg_utilization"
                suffix="%"
                :precision="2"
              />
            </a-col>
            <a-col :span="8">
              <a-statistic
                title="统计天数"
                :value="utilization.days"
                suffix="天"
              />
            </a-col>
            <a-col :span="8">
              <a-statistic
                title="统计日期范围"
                :value="`${utilization.start_date} ~ ${utilization.end_date}`"
              />
            </a-col>
          </a-row>

          <a-card title="资源利用率详情" :bordered="false">
            <a-table
              :scroll="{ x: 'max-content' }"
              :columns="columns"
              :data-source="utilization.utilization_stats"
              :pagination="false"
              row-key="resource_id"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'user'">
                  <span>{{ record.nickname || record.username }}({{ record.username }})</span>
                </template>
                <template v-else-if="column.key === 'utilization'">
                  <a-progress
                    :percent="record.utilization"
                    :status="getUtilizationStatus(record.utilization)"
                    :format="(percent: number) => `${percent?.toFixed(2)}%`"
                  />
                </template>
                <template v-else-if="column.key === 'total_hours'">
                  <span>{{ record.total_hours.toFixed(2) }} / {{ record.max_hours.toFixed(2) }} 小时</span>
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
import { ref, reactive, onMounted } from 'vue'
import { saveLastSelected, getLastSelected } from '@/utils/storage'
import { message } from 'ant-design-vue'
import type { Dayjs } from 'dayjs'
// import dayjs from 'dayjs'
// import { DatePicker } from 'ant-design-vue'
import AppHeader from '@/components/AppHeader.vue'

// const { RangePicker } = DatePicker
import { getResourceUtilization, type ResourceUtilization } from '@/api/resource'
import { getUsers } from '@/api/user'
import { getProjects } from '@/api/project'
import type { User } from '@/api/user'
import type { Project } from '@/api/project'

const loading = ref(false)
const users = ref<User[]>([])
const projects = ref<Project[]>([])
const utilization = ref<ResourceUtilization>({
  start_date: '',
  end_date: '',
  days: 0,
  utilization_stats: [],
  avg_utilization: 0
})
const dateRange = ref<[Dayjs, Dayjs] | null>(null)

const searchForm = reactive({
  user_id: undefined as number | undefined,
  project_id: undefined as number | undefined
})

const columns = [
  { title: '用户', key: 'user', width: 200 },
  { title: '项目', dataIndex: 'project_name', key: 'project_name', width: 200 },
  { title: '工时', key: 'total_hours', width: 200 },
  { title: '利用率', key: 'utilization', width: 300 }
]

const filterUserOption = (input: string, option: any) => {
  const user = users.value.find(u => u.id === option.value)
  if (!user) return false
  const nickname = user.nickname || ''
  const username = user.username || ''
  return nickname.toLowerCase().includes(input.toLowerCase()) ||
    username.toLowerCase().includes(input.toLowerCase())
}

const getUtilizationStatus = (utilization: number) => {
  if (utilization >= 80) return 'success'
  if (utilization >= 50) return 'active'
  if (utilization > 0) return 'exception'
  return 'normal'
}

const loadUsers = async () => {
  try {
    const res = await getUsers({ size: 1000 })
    users.value = res.list || []
  } catch (error: any) {
    message.error('加载用户列表失败: ' + (error.response?.data?.message || error.message))
  }
}

const loadProjects = async () => {
  try {
    const res = await getProjects({ size: 1000 })
    projects.value = res.list || []
  } catch (error: any) {
    message.error('加载项目列表失败: ' + (error.response?.data?.message || error.message))
  }
}

const loadUtilization = async () => {
  loading.value = true
  try {
    const params: any = {}
    if (searchForm.user_id) params.user_id = searchForm.user_id
    if (searchForm.project_id) params.project_id = searchForm.project_id
    if (dateRange.value && dateRange.value[0] && dateRange.value[1]) {
      params.start_date = dateRange.value[0].format('YYYY-MM-DD')
      params.end_date = dateRange.value[1].format('YYYY-MM-DD')
    }

    const res = await getResourceUtilization(params)
    utilization.value = res
  } catch (error: any) {
    message.error('加载利用率数据失败: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

// 搜索表单项目选择改变
const handleSearchProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_resource_utilization_project_search', value)
}

const handleSearch = () => {
  loadUtilization()
}

const handleReset = () => {
  searchForm.user_id = undefined
  searchForm.project_id = undefined
  dateRange.value = null
  // 清除保存的搜索项目选择
  saveLastSelected('last_selected_resource_utilization_project_search', undefined)
  loadUtilization()
}

const handleDateRangeChange = () => {
  loadUtilization()
}

onMounted(() => {
  // 从 localStorage 恢复最后选择的搜索项目
  const lastSearchProjectId = getLastSelected<number>('last_selected_resource_utilization_project_search')
  if (lastSearchProjectId) {
    searchForm.project_id = lastSearchProjectId
  }
  loadUsers()
  loadProjects()
  loadUtilization()
})
</script>

<style scoped>
.resource-utilization {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
  background: #fff;
  padding: 24px;
  border-radius: 8px;
}
</style>

