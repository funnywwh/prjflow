<template>
  <div class="report-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="工作报告">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增{{ activeTab === 'daily' ? '日报' : '周报' }}
              </a-button>
            </template>
          </a-page-header>

          <a-tabs v-model:activeKey="activeTab" @change="handleTabChange">
            <a-tab-pane key="daily" tab="日报">
              <a-card :bordered="false" style="margin-bottom: 16px">
                <a-form layout="inline" :model="dailySearchForm">
                  <a-form-item label="状态">
                    <a-select
                      v-model:value="dailySearchForm.status"
                      placeholder="选择状态"
                      allow-clear
                      style="width: 120px"
                    >
                      <a-select-option value="draft">草稿</a-select-option>
                      <a-select-option value="submitted">已提交</a-select-option>
                      <a-select-option value="approved">已审批</a-select-option>
                      <a-select-option value="rejected">已拒绝</a-select-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item label="开始日期">
                    <a-date-picker
                      v-model:value="dailySearchForm.start_date"
                      placeholder="选择开始日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item label="结束日期">
                    <a-date-picker
                      v-model:value="dailySearchForm.end_date"
                      placeholder="选择结束日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleDailySearch">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleDailyReset">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false" class="table-card">
                <a-table
                  :scroll="{ x: 'max-content', y: tableScrollHeight }"
                  :columns="dailyColumns"
                  :data-source="dailyReports"
                  :loading="dailyLoading"
                  :pagination="dailyPagination"
                  row-key="id"
                  @change="handleDailyTableChange"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <a-tag :color="getStatusColor(record.status)">
                        {{ getStatusText(record.status) }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'approval_status'">
                      <div v-if="record.approvers && record.approvers.length > 0">
                        <a-space size="small" wrap>
                          <span v-for="approver in record.approvers" :key="approver.id">
                            <a-tag :color="getApproverStatusColor(record, approver.id)">
                              {{ approver.nickname || approver.username }}: {{ getApproverStatusText(record, approver.id) }}
                            </a-tag>
                          </span>
                        </a-space>
                      </div>
                      <span v-else>-</span>
                    </template>
                    <template v-else-if="column.key === 'date'">
                      {{ formatDate(record.date) }}
                    </template>
                    <template v-else-if="column.key === 'content'">
                      <div 
                        v-if="record.content" 
                        class="markdown-preview-2lines" 
                        v-html="renderMarkdown(record.content)"
                      ></div>
                      <span v-else>-</span>
                    </template>
                    <template v-else-if="column.key === 'created_at'">
                      {{ formatDateTime(record.created_at) }}
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleDailyEdit(record)">
                          编辑
                        </a-button>
                        <a-button
                          v-if="record.status === 'draft'"
                          type="link"
                          size="small"
                          @click="handleDailySubmit(record)"
                        >
                          提交
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个日报吗？"
                          @confirm="handleDailyDelete(record.id)"
                        >
                          <a-button type="link" size="small" danger>删除</a-button>
                        </a-popconfirm>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-tab-pane>

            <a-tab-pane key="weekly" tab="周报">
              <a-card :bordered="false" style="margin-bottom: 16px">
                <a-form layout="inline" :model="weeklySearchForm">
                  <a-form-item label="状态">
                    <a-select
                      v-model:value="weeklySearchForm.status"
                      placeholder="选择状态"
                      allow-clear
                      style="width: 120px"
                    >
                      <a-select-option value="draft">草稿</a-select-option>
                      <a-select-option value="submitted">已提交</a-select-option>
                      <a-select-option value="approved">已审批</a-select-option>
                      <a-select-option value="rejected">已拒绝</a-select-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item label="开始日期">
                    <a-date-picker
                      v-model:value="weeklySearchForm.start_date"
                      placeholder="选择开始日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item label="结束日期">
                    <a-date-picker
                      v-model:value="weeklySearchForm.end_date"
                      placeholder="选择结束日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleWeeklySearch">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleWeeklyReset">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false" class="table-card">
                <a-table
                  :scroll="{ x: 'max-content', y: tableScrollHeight }"
                  :columns="weeklyColumns"
                  :data-source="weeklyReports"
                  :loading="weeklyLoading"
                  :pagination="weeklyPagination"
                  row-key="id"
                  @change="handleWeeklyTableChange"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <a-tag :color="getStatusColor(record.status)">
                        {{ getStatusText(record.status) }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'approval_status'">
                      <div v-if="record.approvers && record.approvers.length > 0">
                        <a-space size="small" wrap>
                          <span v-for="approver in record.approvers" :key="approver.id">
                            <a-tag :color="getApproverStatusColor(record, approver.id)">
                              {{ approver.nickname || approver.username }}: {{ getApproverStatusText(record, approver.id) }}
                            </a-tag>
                          </span>
                        </a-space>
                      </div>
                      <span v-else>-</span>
                    </template>
                    <template v-else-if="column.key === 'week'">
                      {{ formatDate(record.week_start) }} ~ {{ formatDate(record.week_end) }}
                    </template>
                    <template v-else-if="column.key === 'summary'">
                      <div 
                        v-if="record.summary" 
                        class="markdown-preview-2lines" 
                        v-html="renderMarkdown(record.summary)"
                      ></div>
                      <span v-else>-</span>
                    </template>
                    <template v-else-if="column.key === 'next_week_plan'">
                      <div 
                        v-if="record.next_week_plan" 
                        class="markdown-preview-2lines" 
                        v-html="renderMarkdown(record.next_week_plan)"
                      ></div>
                      <span v-else>-</span>
                    </template>
                    <template v-else-if="column.key === 'created_at'">
                      {{ formatDateTime(record.created_at) }}
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleWeeklyEdit(record)">
                          编辑
                        </a-button>
                        <a-button
                          v-if="record.status === 'draft'"
                          type="link"
                          size="small"
                          @click="handleWeeklySubmit(record)"
                        >
                          提交
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个周报吗？"
                          @confirm="handleWeeklyDelete(record.id)"
                        >
                          <a-button type="link" size="small" danger>删除</a-button>
                        </a-popconfirm>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-tab-pane>

            <a-tab-pane key="approval">
              <template #tab>
                <a-badge 
                  :count="pendingApprovalCount" 
                  :number-style="{ backgroundColor: '#ff4d4f' }"
                  :show-zero="false"
                >
                  <span>审批</span>
                </a-badge>
              </template>
              <a-card :bordered="false" style="margin-bottom: 16px">
                <a-form layout="inline" :model="approvalSearchForm">
                  <a-form-item label="审批状态">
                    <a-select
                      v-model:value="approvalSearchForm.approval_status"
                      placeholder="选择审批状态"
                      allow-clear
                      style="width: 150px"
                    >
                      <a-select-option value="pending">待审批</a-select-option>
                      <a-select-option value="approved">已审批</a-select-option>
                      <a-select-option value="rejected">已拒绝</a-select-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item label="提交人">
                    <a-select
                      v-model:value="approvalSearchForm.user_id"
                      placeholder="选择提交人"
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
                        {{ user.nickname || user.username }}
                      </a-select-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item label="开始日期">
                    <a-date-picker
                      v-model:value="approvalSearchForm.start_date"
                      placeholder="选择开始日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item label="结束日期">
                    <a-date-picker
                      v-model:value="approvalSearchForm.end_date"
                      placeholder="选择结束日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleApprovalSearch">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleApprovalReset">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false">
                <a-table
                  :scroll="{ x: 'max-content' }"
                  :columns="approvalColumns"
                  :data-source="approvalReports"
                  :loading="approvalLoading"
                  :pagination="approvalPagination"
                  row-key="rowKey"
                  @change="handleApprovalTableChange"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'type'">
                      <a-tag :color="record.reportType === 'daily' ? 'blue' : 'green'">
                        {{ record.reportType === 'daily' ? '日报' : '周报' }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'status'">
                      <a-tag :color="getStatusColor(record.status)">
                        {{ getStatusText(record.status) }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'date'">
                      <span v-if="record.reportType === 'daily'">
                        {{ formatDate(record.date) }}
                      </span>
                      <span v-else>
                        {{ formatDate(record.week_start) }} ~ {{ formatDate(record.week_end) }}
                      </span>
                    </template>
                    <template v-else-if="column.key === 'user'">
                      {{ record.user?.nickname || record.user?.username || '-' }}
                    </template>
                    <template v-else-if="column.key === 'content'">
                      <div 
                        v-if="record.reportType === 'daily' && record.content" 
                        class="markdown-preview-2lines" 
                        v-html="renderMarkdown(record.content)"
                      ></div>
                      <div 
                        v-else-if="record.reportType === 'weekly' && record.summary" 
                        class="markdown-preview-2lines" 
                        v-html="renderMarkdown(record.summary)"
                      ></div>
                      <span v-else>-</span>
                    </template>
                    <template v-else-if="column.key === 'approval_status'">
                      <a-tag v-if="getApprovalStatusText(record)" :color="getApprovalStatusColor(record)">
                        {{ getApprovalStatusText(record) }}
                      </a-tag>
                      <span v-else>-</span>
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleApprovalView(record)">
                          查看
                        </a-button>
                        <a-button
                          v-if="record.reportType === 'daily' && canApproveDaily(record) && !isApprovalCompleted(record)"
                          type="link"
                          size="small"
                          style="color: #52c41a"
                          @click="handleDailyApproveClick(record)"
                        >
                          审批
                        </a-button>
                        <a-button
                          v-if="record.reportType === 'weekly' && canApproveWeekly(record) && !isApprovalCompleted(record)"
                          type="link"
                          size="small"
                          style="color: #52c41a"
                          @click="handleWeeklyApproveClick(record)"
                        >
                          审批
                        </a-button>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-tab-pane>
          </a-tabs>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 日报编辑/创建模态框 -->
    <a-modal
      :mask-closable="true"
      v-model:open="dailyModalVisible"
      :title="dailyModalTitle"
      :width="800"
      ok-text="提交"
      @ok="handleDailySubmitForm"
      @cancel="handleDailyCancel"
    >
      <a-form
        ref="dailyFormRef"
        :model="dailyFormData"
        :rules="dailyFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="日期" name="date">
          <a-date-picker
            v-model:value="dailyFormData.date"
            placeholder="选择日期"
            style="width: 100%"
            format="YYYY-MM-DD"
            :disabled="!!dailyFormData.id"
            @change="loadDailyWorkSummary"
          />
        </a-form-item>
        <a-form-item label="工作内容" name="content">
          <MarkdownEditor
            ref="dailyContentEditorRef"
            v-model="dailyFormData.content"
            placeholder="请输入工作内容（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="审批人" name="approver_ids">
          <a-select
            v-model:value="dailyFormData.approver_ids"
            mode="multiple"
            placeholder="选择审批人（可选，支持多选）"
            allow-clear
            show-search
            :filter-option="filterUserOption"
          >
            <a-select-option
              v-for="user in availableApprovers"
              :key="user.id"
              :value="user.id"
            >
              {{ user.nickname || user.username }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 周报编辑/创建模态框 -->
    <a-modal
      :mask-closable="true"
      v-model:open="weeklyModalVisible"
      :title="weeklyModalTitle"
      :width="800"
      ok-text="提交"
      @ok="handleWeeklySubmitForm"
      @cancel="handleWeeklyCancel"
    >
      <a-form
        ref="weeklyFormRef"
        :model="weeklyFormData"
        :rules="weeklyFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="周开始日期" name="week_start">
          <a-date-picker
            v-model:value="weeklyFormData.week_start"
            placeholder="选择周开始日期"
            style="width: 100%"
            format="YYYY-MM-DD"
            @change="loadWeeklyWorkSummary"
          />
        </a-form-item>
        <a-form-item label="周结束日期" name="week_end">
          <a-date-picker
            v-model:value="weeklyFormData.week_end"
            placeholder="选择周结束日期"
            style="width: 100%"
            format="YYYY-MM-DD"
            @change="loadWeeklyWorkSummary"
          />
        </a-form-item>
        <a-form-item label="工作总结" name="summary">
          <MarkdownEditor
            ref="weeklySummaryEditorRef"
            v-model="weeklyFormData.summary"
            placeholder="请输入工作总结（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="下周计划" name="next_week_plan">
          <MarkdownEditor
            ref="weeklyPlanEditorRef"
            v-model="weeklyFormData.next_week_plan"
            placeholder="请输入下周计划（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="审批人" name="approver_ids">
          <a-select
            v-model:value="weeklyFormData.approver_ids"
            mode="multiple"
            placeholder="选择审批人（可选，支持多选）"
            allow-clear
            show-search
            :filter-option="filterUserOption"
          >
            <a-select-option
              v-for="user in availableApprovers"
              :key="user.id"
              :value="user.id"
            >
              {{ user.nickname || user.username }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 日报审批弹窗 -->
    <a-modal
      :mask-closable="false"
      v-model:open="dailyApproveModalVisible"
      title="审批日报"
      :width="900"
      @ok="handleDailyApproveSubmit"
      @cancel="handleDailyApproveCancel"
    >
      <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }" v-if="dailyApproveData">
        <a-form-item label="日期">
          <span>{{ dailyApproveData.date ? formatDate(dailyApproveData.date) : '-' }}</span>
        </a-form-item>
        <a-form-item label="工作内容">
          <div v-html="renderMarkdown(dailyApproveData.content || '')" style="max-height: 300px; overflow-y: auto; border: 1px solid #d9d9d9; padding: 12px; border-radius: 4px;"></div>
        </a-form-item>
        <a-form-item label="审批人">
          <div v-if="dailyApproveData.approvers && dailyApproveData.approvers.length > 0">
            <a-space size="small" wrap>
              <span v-for="approver in dailyApproveData.approvers" :key="approver.id">
                <a-tag :color="getApproverStatusColor(dailyApproveData, approver.id)">
                  {{ approver.nickname || approver.username }}: {{ getApproverStatusText(dailyApproveData, approver.id) }}
                </a-tag>
              </span>
            </a-space>
          </div>
          <span v-else>-</span>
        </a-form-item>
        <a-form-item label="审批记录" v-if="dailyApproveData.approval_records && dailyApproveData.approval_records.length > 0">
          <a-list :data-source="dailyApproveData.approval_records" size="small" bordered>
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta>
                  <template #title>
                    <span>{{ item.approver?.nickname || item.approver?.username || '未知' }}</span>
                    <a-tag :color="getApproverStatusColor(dailyApproveData, item.approver_id)" style="margin-left: 8px;">
                      {{ getApproverStatusText(dailyApproveData, item.approver_id) }}
                    </a-tag>
                  </template>
                  <template #description>
                    <div v-if="item.comment" style="margin-top: 4px; color: #666;">
                      <strong>批注：</strong>{{ item.comment }}
                    </div>
                    <div style="margin-top: 4px; font-size: 12px; color: #999;">
                      {{ item.updated_at ? formatDateTime(item.updated_at) : (item.created_at ? formatDateTime(item.created_at) : '') }}
                    </div>
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </a-form-item>
        <a-form-item label="批注" name="comment">
          <a-textarea
            v-model:value="dailyApproveComment"
            placeholder="请输入批注（可选）"
            :rows="4"
          />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-space>
          <a-button @click="handleDailyApproveCancel">取消</a-button>
          <a-button type="primary" danger @click="handleDailyApproveReject">拒绝</a-button>
          <a-button type="primary" @click="handleDailyApproveSubmit">通过</a-button>
        </a-space>
      </template>
    </a-modal>

    <!-- 周报审批弹窗 -->
    <a-modal
      :mask-closable="false"
      v-model:open="weeklyApproveModalVisible"
      title="审批周报"
      :width="900"
      @ok="handleWeeklyApproveSubmit"
      @cancel="handleWeeklyApproveCancel"
    >
      <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }" v-if="weeklyApproveData">
        <a-form-item label="周期">
          <span>{{ weeklyApproveData.week_start ? formatDate(weeklyApproveData.week_start) : '-' }} 至 {{ weeklyApproveData.week_end ? formatDate(weeklyApproveData.week_end) : '-' }}</span>
        </a-form-item>
        <a-form-item label="工作总结">
          <div v-html="renderMarkdown(weeklyApproveData.summary || '')" style="max-height: 300px; overflow-y: auto; border: 1px solid #d9d9d9; padding: 12px; border-radius: 4px;"></div>
        </a-form-item>
        <a-form-item label="下周计划">
          <div v-html="renderMarkdown(weeklyApproveData.next_week_plan || '')" style="max-height: 300px; overflow-y: auto; border: 1px solid #d9d9d9; padding: 12px; border-radius: 4px;"></div>
        </a-form-item>
        <a-form-item label="审批人">
          <div v-if="weeklyApproveData.approvers && weeklyApproveData.approvers.length > 0">
            <a-space size="small" wrap>
              <span v-for="approver in weeklyApproveData.approvers" :key="approver.id">
                <a-tag :color="getApproverStatusColor(weeklyApproveData, approver.id)">
                  {{ approver.nickname || approver.username }}: {{ getApproverStatusText(weeklyApproveData, approver.id) }}
                </a-tag>
              </span>
            </a-space>
          </div>
          <span v-else>-</span>
        </a-form-item>
        <a-form-item label="审批记录" v-if="weeklyApproveData.approval_records && weeklyApproveData.approval_records.length > 0">
          <a-list :data-source="weeklyApproveData.approval_records" size="small" bordered>
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta>
                  <template #title>
                    <span>{{ item.approver?.nickname || item.approver?.username || '未知' }}</span>
                    <a-tag :color="getApproverStatusColor(weeklyApproveData, item.approver_id)" style="margin-left: 8px;">
                      {{ getApproverStatusText(weeklyApproveData, item.approver_id) }}
                    </a-tag>
                  </template>
                  <template #description>
                    <div v-if="item.comment" style="margin-top: 4px; color: #666;">
                      <strong>批注：</strong>{{ item.comment }}
                    </div>
                    <div style="margin-top: 4px; font-size: 12px; color: #999;">
                      {{ item.updated_at ? formatDateTime(item.updated_at) : (item.created_at ? formatDateTime(item.created_at) : '') }}
                    </div>
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </a-form-item>
        <a-form-item label="批注" name="comment">
          <a-textarea
            v-model:value="weeklyApproveComment"
            placeholder="请输入批注（可选）"
            :rows="4"
          />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-space>
          <a-button @click="handleWeeklyApproveCancel">取消</a-button>
          <a-button type="primary" danger @click="handleWeeklyApproveReject">拒绝</a-button>
          <a-button type="primary" @click="handleWeeklyApproveSubmit">通过</a-button>
        </a-space>
      </template>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch, computed, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { formatDateTime, formatDate } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import { usePermissionStore } from '@/stores/permission'
import { useAuthStore } from '@/stores/auth'
import { getUsers } from '@/api/user'
import {
  getDailyReports,
  getDailyReport,
  createDailyReport,
  updateDailyReport,
  deleteDailyReport,
  updateDailyReportStatus,
  getWeeklyReports,
  getWeeklyReport,
  createWeeklyReport,
  updateWeeklyReport,
  deleteWeeklyReport,
  updateWeeklyReportStatus,
  approveDailyReport,
  approveWeeklyReport,
  getWorkSummary,
  type DailyReport,
  type WeeklyReport,
  type CreateDailyReportRequest,
  type CreateWeeklyReportRequest
} from '@/api/report'
import { getDashboard } from '@/api/dashboard'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

// 配置marked
marked.setOptions({
  highlight: function(code: string, lang: string) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(code, { language: lang }).value
      } catch (err) {
        console.error('Highlight error:', err)
      }
    }
    return hljs.highlightAuto(code).value
  },
  breaks: true,
  gfm: true
} as any)

const route = useRoute()
const router = useRouter()
const activeTab = ref<'daily' | 'weekly' | 'approval'>('daily')

// 日报相关
const dailyLoading = ref(false)
const dailyReports = ref<DailyReport[]>([])
const dailySearchForm = reactive({
  status: undefined as string | undefined,
  start_date: undefined as Dayjs | undefined,
  end_date: undefined as Dayjs | undefined
})
// 计算表格滚动高度
const tableScrollHeight = computed(() => {
  return 'calc(100vh - 500px)'
})

const dailyPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})
const dailyColumns = [
  { title: '日期', key: 'date', width: 120 },
  { title: '工作内容', dataIndex: 'content', key: 'content', ellipsis: true },
  { title: '状态', key: 'status', width: 100 },
  { title: '审批状态', key: 'approval_status', width: 200 },
  { title: '创建时间', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const }
]

const dailyModalVisible = ref(false)
const dailyModalTitle = ref('新增日报')
const dailyFormRef = ref()
const dailyContentEditorRef = ref()
const dailyFormData = reactive<{
  id?: number
  date?: Dayjs
  content?: string
  approver_ids?: number[] // 审批人ID数组（多选）
}>({
  date: undefined,
  content: '',
  approver_ids: []
})
const dailyFormRules = {
  date: [{ required: true, message: '请选择日期', trigger: 'change' }]
}

// 周报相关
const weeklyLoading = ref(false)
const weeklyReports = ref<WeeklyReport[]>([])
const weeklySearchForm = reactive({
  status: undefined as string | undefined,
  start_date: undefined as Dayjs | undefined,
  end_date: undefined as Dayjs | undefined
})
const weeklyPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})
const weeklyColumns = [
  { title: '周期', key: 'week', width: 200 },
  { title: '工作总结', dataIndex: 'summary', key: 'summary', ellipsis: true },
  { title: '下周计划', dataIndex: 'next_week_plan', key: 'next_week_plan', ellipsis: true },
  { title: '状态', key: 'status', width: 100 },
  { title: '审批状态', key: 'approval_status', width: 200 },
  { title: '创建时间', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const }
]

const weeklyModalVisible = ref(false)
const weeklyModalTitle = ref('新增周报')
const weeklyFormRef = ref()
const weeklySummaryEditorRef = ref()
const weeklyPlanEditorRef = ref()
const weeklyFormData = reactive<{
  id?: number
  week_start?: Dayjs
  week_end?: Dayjs
  summary?: string
  next_week_plan?: string
  approver_ids?: number[] // 审批人ID数组（多选）
}>({
  week_start: undefined,
  week_end: undefined,
  summary: '',
  next_week_plan: '',
  approver_ids: []
})
const weeklyFormRules = {
  week_start: [{ required: true, message: '请选择周开始日期', trigger: 'change' }],
  week_end: [{ required: true, message: '请选择周结束日期', trigger: 'change' }]
}

// 审批相关（合并日报和周报）
type ApprovalReport = (DailyReport | WeeklyReport) & {
  reportType: 'daily' | 'weekly'
  rowKey: string
}
const approvalLoading = ref(false)
const approvalReports = ref<ApprovalReport[]>([])
const approvalSearchForm = reactive({
  approval_status: undefined as string | undefined,
  user_id: undefined as number | undefined,
  start_date: undefined as Dayjs | undefined,
  end_date: undefined as Dayjs | undefined
})
const approvalPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})
const approvalColumns = [
  { title: '类型', key: 'type', width: 80 },
  { title: '日期/周期', key: 'date', width: 200 },
  { title: '提交人', key: 'user', width: 120 },
  { title: '内容', key: 'content', ellipsis: true },
  { title: '状态', key: 'status', width: 100 },
  { title: '审批状态', key: 'approval_status', width: 120 },
  { title: '操作', key: 'action', width: 150, fixed: 'right' as const }
]

// 公共数据
const users = ref<any[]>([]) // 用户列表（用于审批人选择）
const pendingApprovalCount = ref(0) // 待审批数量

// 权限管理
const permissionStore = usePermissionStore()
const authStore = useAuthStore()
const isAdmin = computed(() => permissionStore.hasRole('admin'))

// 加载日报列表
const loadDailyReports = async () => {
  dailyLoading.value = true
  try {
    const params: any = {
      page: dailyPagination.current,
      size: dailyPagination.pageSize
    }
    if (dailySearchForm.status) {
      params.status = dailySearchForm.status
    }
    if (dailySearchForm.start_date && dailySearchForm.start_date.isValid()) {
      params.start_date = dailySearchForm.start_date.format('YYYY-MM-DD')
    }
    if (dailySearchForm.end_date && dailySearchForm.end_date.isValid()) {
      params.end_date = dailySearchForm.end_date.format('YYYY-MM-DD')
    }
    const response = await getDailyReports(params)
    dailyReports.value = response.list
    dailyPagination.total = response.total
  } catch (error: any) {
    message.error(error.message || '加载日报列表失败')
  } finally {
    dailyLoading.value = false
  }
}

// 加载周报列表
const loadWeeklyReports = async () => {
  weeklyLoading.value = true
  try {
    const params: any = {
      page: weeklyPagination.current,
      size: weeklyPagination.pageSize
    }
    if (weeklySearchForm.status) {
      params.status = weeklySearchForm.status
    }
    if (weeklySearchForm.start_date && weeklySearchForm.start_date.isValid()) {
      params.start_date = weeklySearchForm.start_date.format('YYYY-MM-DD')
    }
    if (weeklySearchForm.end_date && weeklySearchForm.end_date.isValid()) {
      params.end_date = weeklySearchForm.end_date.format('YYYY-MM-DD')
    }
    const response = await getWeeklyReports(params)
    weeklyReports.value = response.list
    weeklyPagination.total = response.total
  } catch (error: any) {
    message.error(error.message || '加载周报列表失败')
  } finally {
    weeklyLoading.value = false
  }
}

// 加载用户列表（用于审批人选择）
const loadUsers = async () => {
  try {
    const response = await getUsers({ size: 1000 })
    users.value = response.list || []
  } catch (error: any) {
    console.error('加载用户列表失败:', error)
  }
}

// 可用的审批人列表（过滤掉当前用户自己）
const availableApprovers = computed(() => {
  const currentUserId = authStore.user?.id
  if (!currentUserId) return users.value
  return users.value.filter(user => user.id !== currentUserId)
})

// 用户选择器过滤函数
const filterUserOption = (input: string, option: any) => {
  const user = availableApprovers.value.find(u => u.id === option.value)
  if (!user) return false
  const keyword = input.toLowerCase()
  return (
    (user.nickname || '').toLowerCase().includes(keyword) ||
    (user.username || '').toLowerCase().includes(keyword)
  )
}

// 标签页切换
const handleTabChange = (key: string) => {
  activeTab.value = key as 'daily' | 'weekly' | 'approval'
  if (key === 'daily') {
    loadDailyReports()
  } else if (key === 'weekly') {
    loadWeeklyReports()
  } else if (key === 'approval') {
    loadApprovalReports()
  }
}

// 日报搜索
const handleDailySearch = () => {
  dailyPagination.current = 1
  loadDailyReports()
}

// 日报重置
const handleDailyReset = () => {
  dailySearchForm.status = undefined
  dailySearchForm.start_date = undefined
  dailySearchForm.end_date = undefined
  dailyPagination.current = 1
  loadDailyReports()
}

// 日报表格变化
const handleDailyTableChange = (pag: any) => {
  dailyPagination.current = pag.current
  dailyPagination.pageSize = pag.pageSize
  loadDailyReports()
}

// 加载日报工作汇总
const loadDailyWorkSummary = async () => {
  // 编辑模式下不自动汇总
  if (dailyFormData.id) {
    return
  }
  
  if (!dailyFormData.date || !dailyFormData.date.isValid()) {
    return
  }
  
  try {
    const dateStr = dailyFormData.date.format('YYYY-MM-DD')
    const summary = await getWorkSummary({
      start_date: dateStr,
      end_date: dateStr
    })
    
    // 只有在内容为空时才自动填充（允许用户修改）
    if (!dailyFormData.content || dailyFormData.content.trim() === '') {
      dailyFormData.content = summary.content
    }
  } catch (error: any) {
    console.error('加载工作汇总失败:', error)
    // 不显示错误提示，因为可能是没有工作记录
  }
}

// 加载周报工作汇总
const loadWeeklyWorkSummary = async () => {
  // 编辑模式下不自动汇总
  if (weeklyFormData.id) {
    return
  }
  
  if (!weeklyFormData.week_start || !weeklyFormData.week_end || 
      !weeklyFormData.week_start.isValid() || !weeklyFormData.week_end.isValid()) {
    return
  }
  
  try {
    const summary = await getWorkSummary({
      start_date: weeklyFormData.week_start.format('YYYY-MM-DD'),
      end_date: weeklyFormData.week_end.format('YYYY-MM-DD')
    })
    
    // 只有在内容为空时才自动填充（允许用户修改）
    if (!weeklyFormData.summary || weeklyFormData.summary.trim() === '') {
      weeklyFormData.summary = summary.content
    }
  } catch (error: any) {
    console.error('加载工作汇总失败:', error)
    // 不显示错误提示，因为可能是没有工作记录
  }
}

// 创建日报
const handleCreate = () => {
  if (activeTab.value === 'daily') {
    dailyModalTitle.value = '新增日报'
    dailyFormData.id = undefined
    dailyFormData.date = dayjs()
    dailyFormData.content = ''
    dailyFormData.approver_ids = []
    dailyModalVisible.value = true
    // 打开界面后自动加载汇总
    nextTick(() => {
      loadDailyWorkSummary()
    })
  } else {
    weeklyModalTitle.value = '新增周报'
    weeklyFormData.id = undefined
    // 默认设置为本周
    const today = dayjs()
    weeklyFormData.week_start = today.startOf('week').add(1, 'day') // 周一
    weeklyFormData.week_end = today.endOf('week').add(1, 'day') // 周日
    weeklyFormData.summary = ''
    weeklyFormData.next_week_plan = ''
    weeklyFormData.approver_ids = []
    weeklyModalVisible.value = true
    // 打开界面后自动加载汇总
    nextTick(() => {
      loadWeeklyWorkSummary()
    })
  }
}

// 清理内容中的失效 blob URL
const cleanBlobUrls = (content: string): string => {
  if (!content) return content
  // 移除所有 blob URL 的图片标记
  // 匹配格式: ![alt](blob:...)
  return content.replace(/!\[([^\]]*)\]\(blob:[^)]+\)/g, '')
}

// 编辑日报
const handleDailyEdit = (record: DailyReport) => {
  dailyModalTitle.value = '编辑日报'
  dailyFormData.id = record.id
  dailyFormData.date = dayjs(record.date)
  // 清理失效的 blob URL
  dailyFormData.content = cleanBlobUrls(record.content || '')
  dailyFormData.approver_ids = record.approvers?.map(a => a.id) || []
  dailyModalVisible.value = true
}

// 提交日报表单
const handleDailySubmitForm = async () => {
  try {
    await dailyFormRef.value.validate()
    
    const data: CreateDailyReportRequest = {
      date: dailyFormData.date!.format('YYYY-MM-DD'),
      content: dailyFormData.content || '',
      status: 'submitted',
      approver_ids: dailyFormData.approver_ids && dailyFormData.approver_ids.length > 0 ? dailyFormData.approver_ids : undefined
    }
    if (dailyFormData.id) {
      await updateDailyReport(dailyFormData.id, data)
      message.success('更新成功')
    } else {
      await createDailyReport(data)
      message.success('创建成功')
    }
    dailyModalVisible.value = false
    // 如果是从写日报路由打开的，关闭后跳转回报告页面
    if (route.name === 'CreateDailyReport') {
      router.push({ name: 'Report', query: { tab: 'daily' } })
    }
    loadDailyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  }
}

// 提交日报（状态改为已提交）
const handleDailySubmit = async (record: DailyReport) => {
  try {
    await updateDailyReportStatus(record.id, { status: 'submitted' })
    message.success('提交成功')
    loadDailyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    message.error(error.message || '提交失败')
  }
}

// 删除日报
const handleDailyDelete = async (id: number) => {
  try {
    await deleteDailyReport(id)
    message.success('删除成功')
    loadDailyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 取消日报表单
const handleDailyCancel = () => {
  dailyFormRef.value?.resetFields()
  // 如果是从写日报路由打开的，关闭后跳转回报告页面
  if (route.name === 'CreateDailyReport') {
    router.push({ name: 'Report', query: { tab: 'daily' } })
  }
}

// 周报搜索
const handleWeeklySearch = () => {
  weeklyPagination.current = 1
  loadWeeklyReports()
}

// 周报重置
const handleWeeklyReset = () => {
  weeklySearchForm.status = undefined
  weeklySearchForm.start_date = undefined
  weeklySearchForm.end_date = undefined
  weeklyPagination.current = 1
  loadWeeklyReports()
}

// 周报表格变化
const handleWeeklyTableChange = (pag: any) => {
  weeklyPagination.current = pag.current
  weeklyPagination.pageSize = pag.pageSize
  loadWeeklyReports()
}

// 编辑周报
const handleWeeklyEdit = (record: WeeklyReport) => {
  weeklyModalTitle.value = '编辑周报'
  weeklyFormData.id = record.id
  weeklyFormData.week_start = dayjs(record.week_start)
  weeklyFormData.week_end = dayjs(record.week_end)
  // 清理失效的 blob URL
  weeklyFormData.summary = cleanBlobUrls(record.summary || '')
  weeklyFormData.next_week_plan = cleanBlobUrls(record.next_week_plan || '')
  weeklyFormData.approver_ids = record.approvers?.map(a => a.id) || []
  weeklyModalVisible.value = true
}

// 提交周报表单
const handleWeeklySubmitForm = async () => {
  try {
    await weeklyFormRef.value.validate()
    
    const data: CreateWeeklyReportRequest = {
      week_start: weeklyFormData.week_start!.format('YYYY-MM-DD'),
      week_end: weeklyFormData.week_end!.format('YYYY-MM-DD'),
      summary: weeklyFormData.summary || '',
      next_week_plan: weeklyFormData.next_week_plan || '',
      status: 'submitted',
      approver_ids: weeklyFormData.approver_ids && weeklyFormData.approver_ids.length > 0 ? weeklyFormData.approver_ids : undefined
    }
    if (weeklyFormData.id) {
      await updateWeeklyReport(weeklyFormData.id, data)
      message.success('更新成功')
    } else {
      await createWeeklyReport(data)
      message.success('创建成功')
    }
    weeklyModalVisible.value = false
    loadWeeklyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  }
}

// 提交周报（状态改为已提交）
const handleWeeklySubmit = async (record: WeeklyReport) => {
  try {
    await updateWeeklyReportStatus(record.id, { status: 'submitted' })
    message.success('提交成功')
    loadWeeklyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    message.error(error.message || '提交失败')
  }
}


// 删除周报
const handleWeeklyDelete = async (id: number) => {
  try {
    await deleteWeeklyReport(id)
    message.success('删除成功')
    loadWeeklyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 取消周报表单
const handleWeeklyCancel = () => {
  weeklyFormRef.value?.resetFields()
}

// 审批相关状态
const dailyApproveModalVisible = ref(false)
const dailyApproveData = ref<DailyReport | null>(null)
const dailyApproveComment = ref('')

const weeklyApproveModalVisible = ref(false)
const weeklyApproveData = ref<WeeklyReport | null>(null)
const weeklyApproveComment = ref('')

// 判断是否可以审批日报
const canApproveDaily = (record: DailyReport): boolean => {
  if (record.status !== 'submitted') return false
  // 检查当前用户是否是审批人
  const currentUser = authStore.user
  if (!currentUser) return false
  if (isAdmin.value) return true
  return record.approvers?.some(a => a.id === currentUser.id) || false
}

// 判断是否可以审批周报
const canApproveWeekly = (record: WeeklyReport): boolean => {
  if (record.status !== 'submitted') return false
  // 检查当前用户是否是审批人
  const currentUser = authStore.user
  if (!currentUser) return false
  if (isAdmin.value) return true
  return record.approvers?.some(a => a.id === currentUser.id) || false
}

// 打开日报审批弹窗
const handleDailyApproveClick = async (record: DailyReport) => {
  try {
    // 获取完整的报告详情
    const fullRecord = await getDailyReport(record.id)
    dailyApproveData.value = fullRecord
    // 检查是否已有审批记录
    const currentUser = authStore.user
    if (currentUser) {
      const existingApproval = fullRecord.approval_records?.find(r => r.approver_id === currentUser.id)
      if (existingApproval) {
        dailyApproveComment.value = existingApproval.comment || ''
      } else {
        dailyApproveComment.value = ''
      }
    }
    dailyApproveModalVisible.value = true
  } catch (error: any) {
    message.error(error.message || '加载报告详情失败')
  }
}

// 提交日报审批（通过）
const handleDailyApproveSubmit = async () => {
  if (!dailyApproveData.value) return
  try {
    await approveDailyReport(dailyApproveData.value.id, {
      status: 'approved',
      comment: dailyApproveComment.value
    })
    message.success('审批通过')
    dailyApproveModalVisible.value = false
    loadDailyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    message.error(error.message || '审批失败')
  }
}

// 拒绝日报审批
const handleDailyApproveReject = async () => {
  if (!dailyApproveData.value) return
  try {
    await approveDailyReport(dailyApproveData.value.id, {
      status: 'rejected',
      comment: dailyApproveComment.value
    })
    message.success('已拒绝')
    dailyApproveModalVisible.value = false
    loadDailyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    message.error(error.message || '操作失败')
  }
}

// 取消日报审批
const handleDailyApproveCancel = () => {
  dailyApproveModalVisible.value = false
  dailyApproveData.value = null
  dailyApproveComment.value = ''
}

// 打开周报审批弹窗
const handleWeeklyApproveClick = async (record: WeeklyReport) => {
  try {
    // 获取完整的报告详情
    const fullRecord = await getWeeklyReport(record.id)
    weeklyApproveData.value = fullRecord
    // 检查是否已有审批记录
    const currentUser = authStore.user
    if (currentUser) {
      const existingApproval = fullRecord.approval_records?.find(r => r.approver_id === currentUser.id)
      if (existingApproval) {
        weeklyApproveComment.value = existingApproval.comment || ''
      } else {
        weeklyApproveComment.value = ''
      }
    }
    weeklyApproveModalVisible.value = true
  } catch (error: any) {
    message.error(error.message || '加载报告详情失败')
  }
}

// 提交周报审批（通过）
const handleWeeklyApproveSubmit = async () => {
  if (!weeklyApproveData.value) return
  try {
    await approveWeeklyReport(weeklyApproveData.value.id, {
      status: 'approved',
      comment: weeklyApproveComment.value
    })
    message.success('审批通过')
    weeklyApproveModalVisible.value = false
    loadWeeklyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    message.error(error.message || '审批失败')
  }
}

// 拒绝周报审批
const handleWeeklyApproveReject = async () => {
  if (!weeklyApproveData.value) return
  try {
    await approveWeeklyReport(weeklyApproveData.value.id, {
      status: 'rejected',
      comment: weeklyApproveComment.value
    })
    message.success('已拒绝')
    weeklyApproveModalVisible.value = false
    loadWeeklyReports()
    if (activeTab.value === 'approval') {
      loadApprovalReports()
    }
    loadPendingApprovalCount()
  } catch (error: any) {
    message.error(error.message || '操作失败')
  }
}

// 取消周报审批
const handleWeeklyApproveCancel = () => {
  weeklyApproveModalVisible.value = false
  weeklyApproveData.value = null
  weeklyApproveComment.value = ''
}

// 渲染Markdown
const renderMarkdown = (content: string): string => {
  if (!content || content.trim() === '') {
    return '<p class="empty-text">暂无内容</p>'
  }
  
  // 先清理失效的 blob URL（避免显示错误）
  const cleanedContent = cleanBlobUrls(content)
  
  let html = marked.parse(cleanedContent) as string
  
  // 处理图片URL，确保相对路径的图片能正确显示
  // 将相对路径的图片URL转换为绝对路径
  html = html.replace(/<img([^>]*)\ssrc=["']([^"']+)["']([^>]*)>/gi, (match, before, src, after) => {
    // 如果是相对路径（以 /uploads/ 开头），保持不变（Vite代理会处理）
    // 如果是 blob: URL，移除（因为已经失效）
    if (src.startsWith('blob:')) {
      return '' // 移除失效的 blob URL 图片
    }
    // 如果是完整的 HTTP/HTTPS URL，保持不变
    if (src.startsWith('http://') || src.startsWith('https://')) {
      return match
    }
    // 如果是相对路径（以 /uploads/ 开头），保持不变
    if (src.startsWith('/uploads/')) {
      return match
    }
    // 如果是其他相对路径，可能需要添加 /uploads/ 前缀
    return `<img${before} src="${src.startsWith('/') ? src : `/uploads/${src}`}"${after}>`
  })
  
  return html
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    draft: 'default',
    submitted: 'processing',
    approved: 'success',
    rejected: 'error'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    draft: '草稿',
    submitted: '已提交',
    approved: '已审批',
    rejected: '已拒绝'
  }
  return texts[status] || status
}

// 加载审批列表（合并日报和周报）
const loadApprovalReports = async () => {
  approvalLoading.value = true
  try {
    const currentUserId = authStore.user?.id
    const allReports: ApprovalReport[] = []
    
    // 加载日报
    const dailyParams: any = {
      page: 1,
      size: 1000, // 获取所有数据，前端分页
      for_approval: true
    }
    if (approvalSearchForm.user_id) {
      dailyParams.user_id = approvalSearchForm.user_id
    }
    if (approvalSearchForm.start_date && approvalSearchForm.start_date.isValid()) {
      dailyParams.start_date = approvalSearchForm.start_date.format('YYYY-MM-DD')
    }
    if (approvalSearchForm.end_date && approvalSearchForm.end_date.isValid()) {
      dailyParams.end_date = approvalSearchForm.end_date.format('YYYY-MM-DD')
    }
    const dailyResponse = await getDailyReports(dailyParams)
    
    // 加载周报
    const weeklyParams: any = {
      page: 1,
      size: 1000, // 获取所有数据，前端分页
      for_approval: true
    }
    if (approvalSearchForm.user_id) {
      weeklyParams.user_id = approvalSearchForm.user_id
    }
    if (approvalSearchForm.start_date && approvalSearchForm.start_date.isValid()) {
      weeklyParams.start_date = approvalSearchForm.start_date.format('YYYY-MM-DD')
    }
    if (approvalSearchForm.end_date && approvalSearchForm.end_date.isValid()) {
      weeklyParams.end_date = approvalSearchForm.end_date.format('YYYY-MM-DD')
    }
    const weeklyResponse = await getWeeklyReports(weeklyParams)
    
    // 合并日报，添加类型标识
    dailyResponse.list.forEach((report: DailyReport) => {
      allReports.push({
        ...report,
        reportType: 'daily',
        rowKey: `daily-${report.id}`
      } as ApprovalReport)
    })
    
    // 合并周报，添加类型标识
    weeklyResponse.list.forEach((report: WeeklyReport) => {
      allReports.push({
        ...report,
        reportType: 'weekly',
        rowKey: `weekly-${report.id}`
      } as ApprovalReport)
    })
    
    // 根据审批状态过滤
    let filtered = allReports
    if (approvalSearchForm.approval_status) {
      filtered = filtered.filter((report: ApprovalReport) => {
        if (!report.approval_records || report.approval_records.length === 0) {
          return approvalSearchForm.approval_status === 'pending'
        }
        const myApproval = report.approval_records.find((r: any) => r.approver_id === currentUserId)
        if (!myApproval) {
          return approvalSearchForm.approval_status === 'pending'
        }
        return myApproval.status === approvalSearchForm.approval_status
      })
    }
    
    // 按创建时间倒序排序
    filtered.sort((a, b) => {
      const aTime = new Date(a.created_at || 0).getTime()
      const bTime = new Date(b.created_at || 0).getTime()
      return bTime - aTime
    })
    
    // 前端分页
    const start = (approvalPagination.current - 1) * approvalPagination.pageSize
    const end = start + approvalPagination.pageSize
    approvalReports.value = filtered.slice(start, end)
    approvalPagination.total = filtered.length
  } catch (error: any) {
    message.error(error.message || '加载审批列表失败')
  } finally {
    approvalLoading.value = false
  }
}

// 获取审批状态颜色（用于审批列表，显示当前用户的审批状态）
const getApprovalStatusColor = (record: DailyReport | WeeklyReport) => {
  const currentUserId = authStore.user?.id
  // 如果当前用户不是审批人，返回空（显示空白）
  if (!record.approval_records || record.approval_records.length === 0) {
    return '' // 没有审批记录，返回空
  }
  const myApproval = record.approval_records.find((r: any) => r.approver_id === currentUserId)
  if (!myApproval) {
    return '' // 当前用户不是审批人，返回空
  }
  // 当前用户是审批人，显示审批状态
  if (myApproval.status === 'approved') {
    return 'success' // 已通过
  } else if (myApproval.status === 'rejected') {
    return 'error' // 已拒绝
  }
  return 'orange' // 待审批
}

// 获取审批状态文本（用于审批列表，显示当前用户的审批状态）
const getApprovalStatusText = (record: DailyReport | WeeklyReport) => {
  const currentUserId = authStore.user?.id
  // 如果当前用户不是审批人，返回空（显示空白）
  if (!record.approval_records || record.approval_records.length === 0) {
    return '' // 没有审批记录，返回空
  }
  const myApproval = record.approval_records.find((r: any) => r.approver_id === currentUserId)
  if (!myApproval) {
    return '' // 当前用户不是审批人，返回空
  }
  // 当前用户是审批人，显示审批状态
  if (myApproval.status === 'approved') {
    return '已通过'
  } else if (myApproval.status === 'rejected') {
    return '已拒绝'
  }
  return '待审批'
}

// 获取指定审批人的状态颜色
const getApproverStatusColor = (record: DailyReport | WeeklyReport, approverId: number) => {
  if (!record.approval_records || record.approval_records.length === 0) {
    return 'orange' // 待审批
  }
  const approval = record.approval_records.find((r: any) => r.approver_id === approverId)
  if (!approval) {
    return 'orange' // 待审批
  }
  if (approval.status === 'approved') {
    return 'success' // 已通过
  } else if (approval.status === 'rejected') {
    return 'error' // 已拒绝
  }
  return 'orange' // 待审批
}

// 获取指定审批人的状态文本
const getApproverStatusText = (record: DailyReport | WeeklyReport, approverId: number) => {
  if (!record.approval_records || record.approval_records.length === 0) {
    return '待审批'
  }
  const approval = record.approval_records.find((r: any) => r.approver_id === approverId)
  if (!approval) {
    return '待审批'
  }
  if (approval.status === 'approved') {
    return '已通过'
  } else if (approval.status === 'rejected') {
    return '已拒绝'
  }
  return '待审批'
}

// 判断当前用户的审批是否已完成（已通过或已拒绝）
const isApprovalCompleted = (record: DailyReport | WeeklyReport): boolean => {
  const currentUserId = authStore.user?.id
  if (!currentUserId) return false
  if (!record.approval_records || record.approval_records.length === 0) {
    return false
  }
  const myApproval = record.approval_records.find((r: any) => r.approver_id === currentUserId)
  if (!myApproval) {
    return false
  }
  return myApproval.status === 'approved' || myApproval.status === 'rejected'
}

// 审批搜索
const handleApprovalSearch = () => {
  approvalPagination.current = 1
  loadApprovalReports()
}

// 审批重置
const handleApprovalReset = () => {
  approvalSearchForm.approval_status = undefined
  approvalSearchForm.user_id = undefined
  approvalSearchForm.start_date = undefined
  approvalSearchForm.end_date = undefined
  approvalPagination.current = 1
  loadApprovalReports()
}

// 审批表格变化
const handleApprovalTableChange = (pag: any) => {
  approvalPagination.current = pag.current
  approvalPagination.pageSize = pag.pageSize
  loadApprovalReports()
}

// 查看审批详情
const handleApprovalView = async (record: ApprovalReport) => {
  try {
    if (record.reportType === 'daily') {
      const report = await getDailyReport(record.id)
      dailyApproveData.value = report
      dailyApproveComment.value = ''
      dailyApproveModalVisible.value = true
    } else {
      const report = await getWeeklyReport(record.id)
      weeklyApproveData.value = report
      weeklyApproveComment.value = ''
      weeklyApproveModalVisible.value = true
    }
  } catch (error: any) {
    message.error(error.message || '加载报告详情失败')
  }
}

// 加载待审批数量
const loadPendingApprovalCount = async () => {
  try {
    const dashboardData = await getDashboard()
    pendingApprovalCount.value = dashboardData.reports.pending_approval || 0
  } catch (error: any) {
    console.error('加载待审批数量失败:', error)
  }
}

onMounted(() => {
  // 读取路由查询参数
  // 注意：状态字段默认保持为空，不从路由参数自动设置
  if (route.query.tab) {
    activeTab.value = route.query.tab as 'daily' | 'weekly' | 'approval'
  }
  
  // 如果是写日报路由，自动打开新增日报对话框
  if (route.name === 'CreateDailyReport') {
    activeTab.value = 'daily'
    nextTick(() => {
      handleCreate()
    })
  }
  
  loadDailyReports()
  loadUsers()
  loadPendingApprovalCount()
})

// 监听标签页切换，刷新列表和待审批数量
watch(activeTab, () => {
  if (activeTab.value === 'daily') {
    loadDailyReports()
  } else if (activeTab.value === 'weekly') {
    loadWeeklyReports()
  } else if (activeTab.value === 'approval') {
    loadApprovalReports()
    loadPendingApprovalCount()
  }
})

// 监听路由变化，从其他页面返回时刷新列表
watch(() => route.query, () => {
  if (route.query.tab) {
    const tab = route.query.tab as string
    if (tab === 'approval') {
      activeTab.value = 'approval'
      loadApprovalReports()
    } else if (tab === 'daily' || tab === 'weekly') {
      activeTab.value = tab as 'daily' | 'weekly'
    }
  }
}, { immediate: false })
</script>

<style scoped>
.report-management {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.report-management :deep(.ant-layout) {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content {
  padding: 24px;
  background: #f0f2f5;
  flex: 1;
  height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content-inner {
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: 0;
}

.table-card {
  margin-top: 16px;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.table-card :deep(.ant-card-body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 16px;
}

.table-card :deep(.ant-table-wrapper) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.table-card :deep(.ant-spin-nested-loading) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.table-card :deep(.ant-spin-container) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.table-card :deep(.ant-table) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-card :deep(.ant-table-container) {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.content-inner {
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
}

.table-card {
  margin-top: 16px;
}

/* 审批弹窗中的Markdown渲染样式 */
:deep(.markdown-preview) {
  word-wrap: break-word;
  line-height: 1.6;
}

:deep(.markdown-preview p) {
  margin-bottom: 8px;
}

:deep(.markdown-preview h1),
:deep(.markdown-preview h2),
:deep(.markdown-preview h3) {
  margin-top: 16px;
  margin-bottom: 8px;
  font-weight: 600;
}

:deep(.markdown-preview code) {
  background-color: #f6f8fa;
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 85%;
}

:deep(.markdown-preview pre) {
  background-color: #f6f8fa;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
}

:deep(.markdown-preview ul),
:deep(.markdown-preview ol) {
  padding-left: 24px;
  margin-bottom: 8px;
}

:deep(.empty-text) {
  color: #999;
  font-style: italic;
}

/* 审批弹窗中的图片样式 */
:deep(.markdown-preview img),
div[v-html] :deep(img) {
  max-width: 100%;
  height: auto;
  display: block;
  margin: 16px 0;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 确保审批弹窗中的图片容器可以滚动 */
div[v-html] {
  word-wrap: break-word;
}

/* Markdown 预览限制为 2 行 */
.markdown-preview-2lines {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-wrap: break-word;
  word-break: break-word;
  white-space: normal;
  line-height: 1.5;
  max-height: 3em; /* 2行的高度，每行约1.5em */
}

/* Markdown 预览样式优化 */
.markdown-preview-2lines :deep(p) {
  margin: 0;
  margin-bottom: 0.25em;
  word-wrap: break-word;
  word-break: break-word;
}

.markdown-preview-2lines :deep(p:last-child) {
  margin-bottom: 0;
}

.markdown-preview-2lines :deep(h1),
.markdown-preview-2lines :deep(h2),
.markdown-preview-2lines :deep(h3),
.markdown-preview-2lines :deep(h4),
.markdown-preview-2lines :deep(h5),
.markdown-preview-2lines :deep(h6) {
  margin: 0;
  margin-bottom: 0.25em;
  font-size: 1em;
  font-weight: 600;
  word-wrap: break-word;
  word-break: break-word;
}

.markdown-preview-2lines :deep(ul),
.markdown-preview-2lines :deep(ol) {
  margin: 0;
  margin-bottom: 0.25em;
  padding-left: 1.2em;
  word-wrap: break-word;
  word-break: break-word;
}

.markdown-preview-2lines :deep(li) {
  margin: 0;
  word-wrap: break-word;
  word-break: break-word;
}

.markdown-preview-2lines :deep(code) {
  background-color: #f6f8fa;
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 0.9em;
}

.markdown-preview-2lines :deep(pre) {
  margin: 0;
  padding: 0;
  background: transparent;
  display: inline;
}

.markdown-preview-2lines :deep(img) {
  display: none; /* 表格中不显示图片 */
}

.markdown-preview-2lines :deep(blockquote) {
  margin: 0;
  padding-left: 0.5em;
  border-left: 2px solid #ddd;
  display: inline;
}
</style>


