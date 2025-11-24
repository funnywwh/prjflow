<template>
  <div class="resource-calendar">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="资源日历">
            <template #extra>
              <a-space>
                <a-button @click="handlePrevMonth">上个月</a-button>
                <a-button @click="handleCurrentMonth">本月</a-button>
                <a-button @click="handleNextMonth">下个月</a-button>
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
              </a-form-item>
            </a-form>
          </a-card>

          <a-card :bordered="false">
            <div class="calendar-container">
              <div class="calendar-header">
                <div class="calendar-cell header-cell">日期</div>
                <div class="calendar-cell header-cell">工时</div>
                <div class="calendar-cell header-cell">详情</div>
              </div>
              <div
                v-for="(allocations, date) in calendarData"
                :key="date"
                class="calendar-row"
              >
                <div class="calendar-cell date-cell">{{ formatDate(date) }}</div>
                <div class="calendar-cell hours-cell">
                  {{ getTotalHours(allocations) }} 小时
                </div>
                <div class="calendar-cell detail-cell">
                  <a-space wrap>
                    <a-tag
                      v-for="allocation in allocations"
                      :key="allocation.id"
                      :color="getHoursColor(allocation.hours)"
                    >
                      {{ allocation.resource?.user?.nickname || allocation.resource?.user?.username }}:
                      {{ allocation.hours }}h
                      <span v-if="allocation.task">({{ allocation.task.name }})</span>
                    </a-tag>
                  </a-space>
                </div>
              </div>
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
import dayjs from 'dayjs'
import AppHeader from '@/components/AppHeader.vue'
import { getResourceCalendar } from '@/api/resource'
import { getUsers } from '@/api/user'
import { getProjects } from '@/api/project'
import type { User } from '@/api/user'
import type { Project } from '@/api/project'

const loading = ref(false)
const users = ref<User[]>([])
const projects = ref<Project[]>([])
const calendarData = ref<Record<string, any[]>>({})
const currentMonth = ref(dayjs())

const searchForm = reactive({
  user_id: undefined as number | undefined,
  project_id: undefined as number | undefined
})

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

const loadCalendar = async () => {
  loading.value = true
  try {
    const startDate = currentMonth.value.startOf('month')
    const endDate = currentMonth.value.endOf('month')
    const params: any = {
      start_date: startDate.format('YYYY-MM-DD'),
      end_date: endDate.format('YYYY-MM-DD')
    }
    if (searchForm.user_id) params.user_id = searchForm.user_id
    if (searchForm.project_id) params.project_id = searchForm.project_id

    const res = await getResourceCalendar(params)
    calendarData.value = res.data || {}
  } catch (error: any) {
    message.error('加载日历数据失败: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

// 搜索表单项目选择改变
const handleSearchProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_resource_calendar_project_search', value)
}

const handleSearch = () => {
  loadCalendar()
}

const handlePrevMonth = () => {
  currentMonth.value = currentMonth.value.subtract(1, 'month')
  loadCalendar()
}

const handleCurrentMonth = () => {
  currentMonth.value = dayjs()
  loadCalendar()
}

const handleNextMonth = () => {
  currentMonth.value = currentMonth.value.add(1, 'month')
  loadCalendar()
}

const formatDate = (dateStr: string) => {
  return dayjs(dateStr).format('YYYY-MM-DD (dddd)')
}

const getTotalHours = (allocations: any[]) => {
  return allocations.reduce((sum, a) => sum + (a.hours || 0), 0).toFixed(2)
}

const getHoursColor = (hours: number) => {
  if (hours >= 8) return 'green'
  if (hours >= 4) return 'blue'
  if (hours > 0) return 'orange'
  return 'default'
}

onMounted(() => {
  // 从 localStorage 恢复最后选择的搜索项目
  const lastSearchProjectId = getLastSelected<number>('last_selected_resource_calendar_project_search')
  if (lastSearchProjectId) {
    searchForm.project_id = lastSearchProjectId
  }
  loadUsers()
  loadProjects()
  loadCalendar()
})
</script>

<style scoped>
.resource-calendar {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 100%;
  width: 100%;
  margin: 0 auto;
  background: #fff;
  padding: 24px;
  border-radius: 8px;
}

.calendar-container {
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  overflow: hidden;
}

.calendar-header {
  display: flex;
  background: #fafafa;
  border-bottom: 2px solid #e8e8e8;
}

.calendar-row {
  display: flex;
  border-bottom: 1px solid #e8e8e8;
}

.calendar-row:hover {
  background: #f5f5f5;
}

.calendar-cell {
  padding: 12px;
  border-right: 1px solid #e8e8e8;
}

.calendar-cell:last-child {
  border-right: none;
}

.header-cell {
  font-weight: bold;
  text-align: center;
}

.date-cell {
  width: 200px;
  font-weight: 500;
}

.hours-cell {
  width: 100px;
  text-align: center;
}

.detail-cell {
  flex: 1;
}
</style>

