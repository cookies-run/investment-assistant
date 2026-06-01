<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listStocks, createStock, updateStock, deleteStock, type Stock } from '../api/stock'
import { listFunds, createFund, updateFund, deleteFund, type Fund } from '../api/fund'
import StockCard from '../components/StockCard.vue'
import FundCard from '../components/FundCard.vue'
import StockDialog from '../components/StockDialog.vue'
import FundDialog from '../components/FundDialog.vue'

const stocks = ref<Stock[]>([])
const funds = ref<Fund[]>([])
const loading = ref(false)

const stockDialogVisible = ref(false)
const fundDialogVisible = ref(false)
const editingStock = ref<any>(undefined)
const editingFund = ref<any>(undefined)

async function fetchData() {
  loading.value = true
  try {
    const [s, f] = await Promise.all([listStocks(), listFunds()])
    stocks.value = s
    funds.value = f
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)

async function onAddStock(data: any) {
  try {
    if (editingStock.value) {
      await updateStock(data.stock_code, data)
      ElMessage.success('更新成功')
    } else {
      await createStock(data)
      ElMessage.success('新增成功')
    }
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '操作失败')
  }
  editingStock.value = undefined
}

async function onDeleteStock(code: string) {
  try {
    await ElMessageBox.confirm('确定删除该个股?', '提示', { type: 'warning' })
    await deleteStock(code)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.error || '删除失败')
    }
  }
}

async function onAddFund(data: any) {
  try {
    if (editingFund.value) {
      await updateFund(data.fund_code, data)
      ElMessage.success('更新成功')
    } else {
      await createFund(data)
      ElMessage.success('新增成功')
    }
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '操作失败')
  }
  editingFund.value = undefined
}

async function onDeleteFund(code: string) {
  try {
    await ElMessageBox.confirm('确定删除该基金?', '提示', { type: 'warning' })
    await deleteFund(code)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.error || '删除失败')
    }
  }
}

function openEditStock(stock: Stock) {
  editingStock.value = { ...stock }
  stockDialogVisible.value = true
}

function openEditFund(fund: Fund) {
  editingFund.value = { ...fund }
  fundDialogVisible.value = true
}

function openAddStock() {
  editingStock.value = undefined
  stockDialogVisible.value = true
}

function openAddFund() {
  editingFund.value = undefined
  fundDialogVisible.value = true
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <div>
        <h2>标的配置</h2>
        <p class="section-subtitle">管理您的个股与基金监控标的</p>
      </div>
    </div>

    <div class="config-section">
      <div class="section-header-bar stock-bar">
        <div class="section-left">
          <div class="section-icon-box stock-icon">
            <el-icon :size="18"><Collection /></el-icon>
          </div>
          <div>
            <div class="section-title">个股</div>
            <div class="section-subtitle">共 {{ stocks.length }} 只</div>
          </div>
        </div>
        <el-button type="primary" @click="openAddStock">
          <el-icon><Plus /></el-icon> 新增个股
        </el-button>
      </div>
      <div class="card-list">
        <StockCard
          v-for="s in stocks"
          :key="s.stock_code"
          :stock="s"
          @edit="openEditStock"
          @delete="onDeleteStock"
        />
      </div>
      <el-empty v-if="!loading && stocks.length === 0" description="暂无个股配置" />
    </div>

    <div class="config-section">
      <div class="section-header-bar fund-bar">
        <div class="section-left">
          <div class="section-icon-box fund-icon">
            <el-icon :size="18"><Box /></el-icon>
          </div>
          <div>
            <div class="section-title">基金</div>
            <div class="section-subtitle">共 {{ funds.length }} 只</div>
          </div>
        </div>
        <el-button type="primary" @click="openAddFund">
          <el-icon><Plus /></el-icon> 新增基金
        </el-button>
      </div>
      <div class="card-list">
        <FundCard
          v-for="f in funds"
          :key="f.fund_code"
          :fund="f"
          @edit="openEditFund"
          @delete="onDeleteFund"
        />
      </div>
      <el-empty v-if="!loading && funds.length === 0" description="暂无基金配置" />
    </div>

    <StockDialog v-model:visible="stockDialogVisible" :initial-data="editingStock" @submit="onAddStock" />
    <FundDialog v-model:visible="fundDialogVisible" :initial-data="editingFund" @submit="onAddFund" />
  </div>
</template>

<style scoped>
.config-section {
  margin-bottom: 32px;
}

.section-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  padding: 12px 16px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.section-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.section-icon-box {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stock-icon {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #2563eb;
}

.fund-icon {
  background: linear-gradient(135deg, #f5f3ff 0%, #ede9fe 100%);
  color: #8b5cf6;
}

.section-header-bar .section-title {
  font-size: 16px;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
}

.section-header-bar .section-subtitle {
  font-size: 12px;
  color: #9ca3af;
  margin: 2px 0 0 0;
}
</style>
