<template>
  <div class="product-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-tabs v-model:activeKey="activeTab">
            <!-- 产品线管理 -->
            <a-tab-pane key="productLines" tab="产品线管理">
              <a-page-header title="产品线管理">
                <template #extra>
                  <a-button type="primary" @click="handleCreateProductLine">
                    <template #icon><PlusOutlined /></template>
                    新增产品线
                  </a-button>
                </template>
              </a-page-header>

              <a-card :bordered="false" style="margin-bottom: 16px">
                <a-form layout="inline" :model="productLineSearchForm">
                  <a-form-item label="关键词">
                    <a-input
                      v-model:value="productLineSearchForm.keyword"
                      placeholder="产品线名称/描述"
                      allow-clear
                      style="width: 200px"
                    />
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleSearchProductLine">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleResetProductLine">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false">
                <a-table
                  :columns="productLineColumns"
                  :data-source="productLines"
                  :loading="productLineLoading"
                  row-key="id"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <a-tag :color="record.status === 1 ? 'green' : 'red'">
                        {{ record.status === 1 ? '正常' : '禁用' }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleEditProductLine(record)">
                          编辑
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个产品线吗？"
                          @confirm="handleDeleteProductLine(record.id)"
                        >
                          <a-button type="link" size="small" danger>删除</a-button>
                        </a-popconfirm>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-tab-pane>

            <!-- 产品管理 -->
            <a-tab-pane key="products" tab="产品管理">
              <a-page-header title="产品管理">
                <template #extra>
                  <a-button type="primary" @click="handleCreateProduct">
                    <template #icon><PlusOutlined /></template>
                    新增产品
                  </a-button>
                </template>
              </a-page-header>

              <a-card :bordered="false" style="margin-bottom: 16px">
                <a-form layout="inline" :model="productSearchForm">
                  <a-form-item label="关键词">
                    <a-input
                      v-model:value="productSearchForm.keyword"
                      placeholder="产品名称/编码"
                      allow-clear
                      style="width: 200px"
                    />
                  </a-form-item>
                  <a-form-item label="产品线">
                    <a-select
                      v-model:value="productSearchForm.product_line_id"
                      placeholder="选择产品线"
                      allow-clear
                      style="width: 200px"
                    >
                      <a-select-option
                        v-for="line in productLines"
                        :key="line.id"
                        :value="line.id"
                      >
                        {{ line.name }}
                      </a-select-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleSearchProduct">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleResetProduct">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false" class="table-card">
                <a-table
                  :columns="productColumns"
                  :data-source="products"
                  :loading="productLoading"
                  :pagination="productPagination"
                  :scroll="{ y: tableScrollHeight }"
                  @change="handleProductTableChange"
                  row-key="id"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <a-tag :color="record.status === 1 ? 'green' : 'red'">
                        {{ record.status === 1 ? '正常' : '禁用' }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'product_line'">
                      {{ record.product_line?.name || '-' }}
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleEditProduct(record)">
                          编辑
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个产品吗？"
                          @confirm="handleDeleteProduct(record.id)"
                        >
                          <a-button type="link" size="small" danger>删除</a-button>
                        </a-popconfirm>
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

    <!-- 产品线编辑对话框 -->
    <a-modal
      v-model:open="productLineModalVisible"
      :title="productLineModalTitle"
      :mask-closable="true"
      @ok="handleProductLineSubmit"
      @cancel="handleProductLineCancel"
      :confirm-loading="productLineSubmitting"
      
      >
      <a-form
        ref="productLineFormRef"
        :model="productLineFormData"
        :rules="productLineFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="产品线名称" name="name">
          <a-input v-model:value="productLineFormData.name" placeholder="请输入产品线名称" />
        </a-form-item>
        <a-form-item label="描述" name="description">
          <a-textarea v-model:value="productLineFormData.description" placeholder="请输入描述" :rows="3" />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="productLineFormData.status" placeholder="选择状态">
            <a-select-option :value="1">正常</a-select-option>
            <a-select-option :value="0">禁用</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 产品编辑对话框 -->
    <a-modal
      v-model:open="productModalVisible"
      :title="productModalTitle"
      :mask-closable="true"
      @ok="handleProductSubmit"
      @cancel="handleProductCancel"
      :confirm-loading="productSubmitting"
      
      >
      <a-form
        ref="productFormRef"
        :model="productFormData"
        :rules="productFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="产品名称" name="name">
          <a-input v-model:value="productFormData.name" placeholder="请输入产品名称" />
        </a-form-item>
        <a-form-item label="产品编码" name="code">
          <a-input v-model:value="productFormData.code" placeholder="请输入产品编码" />
        </a-form-item>
        <a-form-item label="产品线" name="product_line_id">
          <a-select
            v-model:value="productFormData.product_line_id"
            placeholder="选择产品线"
          >
            <a-select-option
              v-for="line in productLines"
              :key="line.id"
              :value="line.id"
            >
              {{ line.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="描述" name="description">
          <a-textarea v-model:value="productFormData.description" placeholder="请输入描述" :rows="3" />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="productFormData.status" placeholder="选择状态">
            <a-select-option :value="1">正常</a-select-option>
            <a-select-option :value="0">禁用</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
// import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import {
  getProductLines,
  createProductLine,
  updateProductLine,
  deleteProductLine,
  getProducts,
  createProduct,
  updateProduct,
  deleteProduct,
  type ProductLine,
  type Product,
  type CreateProductLineRequest,
  type CreateProductRequest
} from '@/api/product'

// const route = useRoute()
// const router = useRouter()
const activeTab = ref('productLines')

const productLineLoading = ref(false)
const productLoading = ref(false)
const productLineSubmitting = ref(false)
const productSubmitting = ref(false)

const productLines = ref<ProductLine[]>([])
const products = ref<Product[]>([])

const productLineSearchForm = reactive({
  keyword: ''
})

const productSearchForm = reactive({
  keyword: '',
  product_line_id: undefined as number | undefined
})

const productPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})

const productLineColumns = [
  { title: '产品线名称', dataIndex: 'name', key: 'name' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 150, fixed: 'right' as const }
]

// 计算表格滚动高度
const tableScrollHeight = computed(() => {
  return 'calc(100vh - 450px)'
})

const productColumns = [
  { title: '产品名称', dataIndex: 'name', key: 'name' },
  { title: '产品编码', dataIndex: 'code', key: 'code' },
  { title: '产品线', key: 'product_line', width: 120 },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 150, fixed: 'right' as const }
]

const productLineModalVisible = ref(false)
const productLineModalTitle = ref('新增产品线')
const productLineFormRef = ref()
const productLineFormData = reactive<CreateProductLineRequest & { id?: number }>({
  name: '',
  description: '',
  status: 1
})

const productLineFormRules = {
  name: [{ required: true, message: '请输入产品线名称', trigger: 'blur' }]
}

const productModalVisible = ref(false)
const productModalTitle = ref('新增产品')
const productFormRef = ref()
const productFormData = reactive<CreateProductRequest & { id?: number }>({
  name: '',
  code: '',
  description: '',
  status: 1,
  product_line_id: 0
})

const productFormRules = {
  name: [{ required: true, message: '请输入产品名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入产品编码', trigger: 'blur' }],
  product_line_id: [{ required: true, message: '请选择产品线', trigger: 'change' }]
}

// 加载产品线列表
const loadProductLines = async () => {
  productLineLoading.value = true
  try {
    productLines.value = await getProductLines()
  } catch (error: any) {
    message.error(error.message || '加载产品线列表失败')
  } finally {
    productLineLoading.value = false
  }
}

// 加载产品列表
const loadProducts = async () => {
  productLoading.value = true
  try {
    const params: any = {
      page: productPagination.current,
      size: productPagination.pageSize
    }
    if (productSearchForm.keyword) {
      params.keyword = productSearchForm.keyword
    }
    if (productSearchForm.product_line_id) {
      params.product_line_id = productSearchForm.product_line_id
    }
    const response = await getProducts(params)
    products.value = response.list
    productPagination.total = response.total
  } catch (error: any) {
    message.error(error.message || '加载产品列表失败')
  } finally {
    productLoading.value = false
  }
}

// 产品线搜索
const handleSearchProductLine = () => {
  loadProductLines()
}

// 产品线重置
const handleResetProductLine = () => {
  productLineSearchForm.keyword = ''
  loadProductLines()
}

// 产品搜索
const handleSearchProduct = () => {
  productPagination.current = 1
  loadProducts()
}

// 产品重置
const handleResetProduct = () => {
  productSearchForm.keyword = ''
  productSearchForm.product_line_id = undefined
  handleSearchProduct()
}

// 产品表格变化
const handleProductTableChange = (pag: any) => {
  productPagination.current = pag.current
  productPagination.pageSize = pag.pageSize
  loadProducts()
}

// 新增产品线
const handleCreateProductLine = () => {
  productLineModalTitle.value = '新增产品线'
  Object.assign(productLineFormData, {
    name: '',
    description: '',
    status: 1
  })
  delete productLineFormData.id
  productLineModalVisible.value = true
}

// 编辑产品线
const handleEditProductLine = (record: ProductLine) => {
  productLineModalTitle.value = '编辑产品线'
  Object.assign(productLineFormData, {
    id: record.id,
    name: record.name,
    description: record.description || '',
    status: record.status
  })
  productLineModalVisible.value = true
}

// 提交产品线
const handleProductLineSubmit = async () => {
  try {
    await productLineFormRef.value.validate()
    productLineSubmitting.value = true

    if (productLineFormData.id) {
      await updateProductLine(productLineFormData.id, productLineFormData)
      message.success('更新成功')
    } else {
      await createProductLine(productLineFormData)
      message.success('创建成功')
    }

    productLineModalVisible.value = false
    loadProductLines()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  } finally {
    productLineSubmitting.value = false
  }
}

// 取消产品线
const handleProductLineCancel = () => {
  productLineModalVisible.value = false
  productLineFormRef.value?.resetFields()
}

// 删除产品线
const handleDeleteProductLine = async (id: number) => {
  try {
    await deleteProductLine(id)
    message.success('删除成功')
    loadProductLines()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 新增产品
const handleCreateProduct = () => {
  productModalTitle.value = '新增产品'
  Object.assign(productFormData, {
    name: '',
    code: '',
    description: '',
    status: 1,
    product_line_id: productLines.value[0]?.id || 0
  })
  delete productFormData.id
  productModalVisible.value = true
}

// 编辑产品
const handleEditProduct = (record: Product) => {
  productModalTitle.value = '编辑产品'
  Object.assign(productFormData, {
    id: record.id,
    name: record.name,
    code: record.code,
    description: record.description || '',
    status: record.status,
    product_line_id: record.product_line_id
  })
  productModalVisible.value = true
}

// 提交产品
const handleProductSubmit = async () => {
  try {
    await productFormRef.value.validate()
    productSubmitting.value = true

    if (productFormData.id) {
      await updateProduct(productFormData.id, productFormData)
      message.success('更新成功')
    } else {
      await createProduct(productFormData)
      message.success('创建成功')
    }

    productModalVisible.value = false
    loadProducts()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  } finally {
    productSubmitting.value = false
  }
}

// 取消产品
const handleProductCancel = () => {
  productModalVisible.value = false
  productFormRef.value?.resetFields()
}

// 删除产品
const handleDeleteProduct = async (id: number) => {
  try {
    await deleteProduct(id)
    message.success('删除成功')
    loadProducts()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

onMounted(() => {
  loadProductLines()
  loadProducts()
})
</script>

<style scoped>
.product-management {
  min-height: 100vh;
}

.product-management {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.product-management :deep(.ant-layout) {
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
  background: white;
  padding: 24px;
  border-radius: 4px;
}
</style>

