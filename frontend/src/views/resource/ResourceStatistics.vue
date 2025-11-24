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

          <!-- 资源冲突检测 -->
          <a-card title="资源冲突检测" :bordered="false" style="margin-bottom: 16px">
            <a-form layout="inline" :model="conflictForm">
              <a-form-item label="用户">
                <a-select
                  v-model:value="conflictForm.user_id"
                  placeholder="选择用户"
                  allow-clear
                  show-search
                  :filter-option="filterUserOption"
                  style="width: 200px"
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
              <a-form-item label="日期">
                <a-date-picker
                  v-model:value="conflictForm.date"
                  placeholder="选择日期"
                  format="YYYY-MM-DD"
                  style="width: 150px"
                />
              </a-form-item>
              <a-form-item>
                <a-button type="primary" @click="handleCheckConflict">检测冲突</a-button>
              </a-form-item>
            </a-form>
            <div v-if="conflictResult" style="margin-top: 16px">
              <a-alert
                :type="conflictResult.has_conflict ? 'error' : conflictResult.has_warning ? 'warning' : 'success'"
                :message="conflictResult.has_conflict ? '发现冲突' : conflictResult.has_warning ? '存在警告' : '无冲突'"
                :description="`总工时: ${conflictResult.total_hours.toFixed(2)} 小时`"
                style="margin-bottom: 16px"
              />
              <div v-if="conflictResult.conflicts && conflictResult.conflicts.length > 0">
                <a-list size="small" :data-source="conflictResult.conflicts">
                  <template #renderItem="{ item }">
                    <a-list-item>
                      <a-tag color="red">{{ item }}</a-tag>
                    </a-list-item>
                  </template>
                </a-list>
              </div>
              <a-table
                :scroll="{ x: 'max-content' }"
                v-if="conflictResult.allocations && conflictResult.allocations.length > 0"
                :columns="allocationColumns"
                :data-source="conflictResult.allocations"
                :pagination="false"
                size="small"
                style="margin-top: 16px"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'hours'">
                    <span>{{ record.hours.toFixed(2) }} 小时</span>
                  </template>
                </template>
              </a-table>
            </div>
          </a-card>

          <a-row :gutter="16">
            <a-col :span="12">
              <a-card title="按项目统计" :bordered="false">
                <a-table
                :scroll="{ x: 'max-content' }"
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
                :scroll="{ x: 'max-content' }"
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

          <!-- 资源利用率分析 -->
          <a-card title="资源利用率分析" :bordered="false" style="margin-top: 16px">
            <a-form layout="inline" :model="utilizationForm" style="margin-bottom: 16px">
              <a-form-item label="日期范围">
                <a-range-picker
                  v-model:value="utilizationForm.dateRange"
                  format="YYYY-MM-DD"
                />
              </a-form-item>
              <a-form-item>
                <a-button type="primary" @click="handleLoadUtilization">查询</a-button>
              </a-form-item>
            </a-form>
            <div v-if="utilizationData">
              <a-statistic
                title="平均利用率"
                :value="utilizationData.avg_utilization || 0"
                suffix="%"
                :precision="2"
                style="margin-bottom: 16px"
              />
              <a-table
                :scroll="{ x: 'max-content' }"
                :columns="utilizationColumns"
                :data-source="utilizationData.utilization_stats"
                :pagination="false"
                size="small"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'user'">
                    <span>{{ record.nickname || record.username }}({{ record.username }})</span>
                  </template>
                  <template v-else-if="column.key === 'total_hours'">
                    <span>{{ record.total_hours.toFixed(2) }} 小时</span>
                  </template>
                  <template v-else-if="column.key === 'max_hours'">
                    <span>{{ record.max_hours.toFixed(2) }} 小时</span>
                  </template>
                  <template v-else-if="column.key === 'utilization'">
                    <a-progress
                      :percent="record.utilization"
                      :stroke-color="record.utilization >= 80 ? '#52c41a' : record.utilization >= 60 ? '#faad14' : '#ff4d4f'"
                      :format="(percent: number) => `${percent?.toFixed(2)}%`"
                    />
                  </template>
                </template>
              </a-table>
            </div>
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
import { getResourceStatistics, getResourceUtilization, checkResourceConflict, type ResourceStatistics, type ResourceUtilization, type ResourceConflict } from '@/api/resource'
import { getUsers } from '@/api/user'
import { getProjects } from '@/api/project'
import type { User } from '@/api/user'
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

const allocationColumns = [
  { title: '项目', dataIndex: 'project_name', key: 'project_name' },
  { title: '任务', dataIndex: 'task_title', key: 'task_title' },
  { title: 'Bug', dataIndex: 'bug_title', key: 'bug_title' },
  { title: '工时', key: 'hours', width: 100 },
  { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true }
]

const utilizationColumns = [
  { title: '人员', key: 'user' },
  { title: '项目', dataIndex: 'project_name', key: 'project_name' },
  { title: '总工时', key: 'total_hours', width: 120 },
  { title: '最大工时', key: 'max_hours', width: 120 },
  { title: '利用率', key: 'utilization', width: 200 }
]

const conflictForm = reactive({
  user_id: undefined as number | undefined,
  date: undefined as Dayjs | undefined
})

const conflictResult = ref<ResourceConflict | null>(null)

const utilizationForm = reactive({
  dateRange: null as [Dayjs, Dayjs] | null
})

const utilizationData = ref<ResourceUtilization | null>(null)

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

// 搜索表单项目选择改变
const handleSearchProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_resource_statistics_project_search', value)
}

const handleSearch = () => {
  loadStatistics()
}

const handleReset = () => {
  searchForm.user_id = undefined
  searchForm.project_id = undefined
  dateRange.value = null
  // 清除保存的搜索项目选择
  saveLastSelected('last_selected_resource_statistics_project_search', undefined)
  loadStatistics()
}

const handleDateRangeChange = () => {
  loadStatistics()
}

const handleCheckConflict = async () => {
  if (!conflictForm.user_id || !conflictForm.date) {
    message.warning('请选择用户和日期')
    return
  }
  try {
    const res = await checkResourceConflict({
      user_id: conflictForm.user_id,
      date: conflictForm.date.format('YYYY-MM-DD')
    })
    conflictResult.value = res
  } catch (error: any) {
    message.error('检测冲突失败: ' + (error.response?.data?.message || error.message))
  }
}

const handleLoadUtilization = async () => {
  if (!utilizationForm.dateRange || !utilizationForm.dateRange[0] || !utilizationForm.dateRange[1]) {
    message.warning('请选择日期范围')
    return
  }
  try {
    const params: any = {
      start_date: utilizationForm.dateRange[0].format('YYYY-MM-DD'),
      end_date: utilizationForm.dateRange[1].format('YYYY-MM-DD')
    }
    if (searchForm.user_id) params.user_id = searchForm.user_id
    if (searchForm.project_id) params.project_id = searchForm.project_id
    const res = await getResourceUtilization(params)
    utilizationData.value = res
  } catch (error: any) {
    message.error('加载利用率数据失败: ' + (error.response?.data?.message || error.message))
  }
}

onMounted(() => {
  // 从 localStorage 恢复最后选择的搜索项目
  const lastSearchProjectId = getLastSelected<number>('last_selected_resource_statistics_project_search')
  if (lastSearchProjectId) {
    searchForm.project_id = lastSearchProjectId
  }
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
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
  background: #fff;
  padding: 24px;
  border-radius: 8px;
}
</style>

