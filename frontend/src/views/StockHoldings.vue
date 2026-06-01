<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listStocks, createStock, updateStock, deleteStock, type Stock } from '../api/stock'
import {
  getStockGroups, createStockGroup, updateStockGroup, deleteStockGroup,
  reorderStockGroups, createStockGroupItem, deleteStockGroupItem, reorderStockGroupItems,
  type StockGroup
} from '../api/stockGroup'
import StockCard from '../components/StockCard.vue'
import StockDialog from '../components/StockDialog.vue'
import StockDetailDrawer from '../components/StockDetailDrawer.vue'

const stocks = ref<Stock[]>([])
const groups = ref<StockGroup[]>([])
const loading = ref(false)
const showManagement = ref(false)
const showOnlyHoldings = ref(false)

// Modals
const showAddGroup = ref(false)
const showEditGroup = ref(false)
const showAddStockToGroup = ref(false)
const newGroupName = ref('')
const editGroupId = ref<number | null>(null)
const editGroupName = ref('')
const selectedGroupForAdd = ref<number | null>(null)
const selectedStockCodes = ref<string[]>([])

// Drag state
const dragGroupIndex = ref<number | null>(null)
const dragOverGroupIndex = ref<number | null>(null)
const dragItemKey = ref<string | null>(null)
const dragOverItemKey = ref<string | null>(null)

const stockDialogVisible = ref(false)
const editingStock = ref<any>(undefined)

const drawerVisible = ref(false)
const selectedStockCode = ref('')
const selectedStockName = ref('')

const stockMap = computed(() => {
  const map = new Map<string, Stock>()
  for (const s of stocks.value) {
    map.set(s.stock_code, s)
  }
  return map
})

const filteredStocks = computed(() => {
  if (!showOnlyHoldings.value) return stocks.value
  return stocks.value.filter(s => (s.hold_quantity ?? 0) > 0)
})

const filteredStockMap = computed(() => {
  const map = new Map<string, Stock>()
  for (const s of filteredStocks.value) {
    map.set(s.stock_code, s)
  }
  return map
})

const groupedStockCodes = computed(() => {
  const codes = new Set<string>()
  for (const g of groups.value) {
    for (const item of g.items || []) {
      codes.add(item.stock_code)
    }
  }
  return codes
})

const ungroupedStocks = computed(() => {
  return filteredStocks.value.filter(s => !groupedStockCodes.value.has(s.stock_code))
})

const filteredGroups = computed(() => {
  if (!showOnlyHoldings.value) return groups.value
  return groups.value.map(g => ({
    ...g,
    items: (g.items || []).filter(item => {
      const stock = stockMap.value.get(item.stock_code)
      return stock && (stock.hold_quantity ?? 0) > 0
    })
  })).filter(g => g.items.length > 0 || !showOnlyHoldings.value)
})

async function fetchData() {
  loading.value = true
  try {
    const [s, g] = await Promise.all([listStocks(), getStockGroups()])
    stocks.value = s
    groups.value = g
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)

function toggleManagement() {
  showManagement.value = !showManagement.value
}

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

function openEditStock(stock: Stock) {
  editingStock.value = { ...stock }
  stockDialogVisible.value = true
}

function openSettingsStock(stock: Stock) {
  editingStock.value = { ...stock }
  stockDialogVisible.value = true
}

function openAddStock() {
  editingStock.value = undefined
  stockDialogVisible.value = true
}

function openStockDetail(stock: Stock) {
  selectedStockCode.value = stock.stock_code
  selectedStockName.value = stock.stock_name
  drawerVisible.value = true
}

// Group CRUD
function openAddGroup() {
  newGroupName.value = ''
  showAddGroup.value = true
}

async function confirmAddGroup() {
  if (!newGroupName.value.trim()) return
  await createStockGroup(newGroupName.value.trim())
  showAddGroup.value = false
  await fetchData()
}

function openEditGroup(group: StockGroup) {
  editGroupId.value = group.id
  editGroupName.value = group.name
  showEditGroup.value = true
}

async function confirmEditGroup() {
  if (!editGroupName.value.trim() || editGroupId.value === null) return
  await updateStockGroup(editGroupId.value, editGroupName.value.trim())
  showEditGroup.value = false
  await fetchData()
}

async function confirmDeleteGroup(id: number) {
  try {
    await ElMessageBox.confirm('确定删除该分组及其中的个股?', '提示', { type: 'warning' })
    await deleteStockGroup(id)
    await fetchData()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.error || '删除失败')
    }
  }
}

// Item CRUD
async function openAddStockToGroup(groupId: number | null) {
  if (groupId === null) {
    if (groups.value.length === 0) {
      ElMessage.warning('请先新建分组')
      return
    }
    groupId = groups.value[0].id
  }
  selectedGroupForAdd.value = groupId
  selectedStockCodes.value = []
  showAddStockToGroup.value = true
}

function toggleStockCode(code: string) {
  const idx = selectedStockCodes.value.indexOf(code)
  if (idx >= 0) {
    selectedStockCodes.value.splice(idx, 1)
  } else {
    selectedStockCodes.value.push(code)
  }
}

async function confirmAddStocksToGroup() {
  if (selectedGroupForAdd.value === null || selectedStockCodes.value.length === 0) {
    showAddStockToGroup.value = false
    return
  }
  const groupId = selectedGroupForAdd.value
  for (const code of selectedStockCodes.value) {
    const stock = stockMap.value.get(code)
    if (!stock) continue
    await createStockGroupItem(groupId, {
      stock_code: stock.stock_code,
      stock_name: stock.stock_name
    })
  }
  showAddStockToGroup.value = false
  await fetchData()
}

async function confirmRemoveItem(itemId: number) {
  try {
    await ElMessageBox.confirm('确定将该个股从分组中移除?', '提示', { type: 'warning' })
    await deleteStockGroupItem(itemId)
    await fetchData()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.error || '移除失败')
    }
  }
}

// Drag and Drop - Groups
function onGroupDragStart(index: number) {
  dragGroupIndex.value = index
}

function onGroupDragOver(index: number, e: DragEvent) {
  e.preventDefault()
  dragOverGroupIndex.value = index
}

function onGroupDrop(index: number) {
  if (dragGroupIndex.value === null || dragGroupIndex.value === index) {
    dragGroupIndex.value = null
    dragOverGroupIndex.value = null
    return
  }
  const from = dragGroupIndex.value
  const to = index
  const item = groups.value.splice(from, 1)[0]
  groups.value.splice(to, 0, item)
  dragGroupIndex.value = null
  dragOverGroupIndex.value = null
  reorderStockGroups(groups.value.map(g => g.id))
}

function onGroupDragEnd() {
  dragGroupIndex.value = null
  dragOverGroupIndex.value = null
}

// Drag and Drop - Items
function onItemDragStart(groupId: number, itemIndex: number) {
  dragItemKey.value = `${groupId}:${itemIndex}`
}

function onItemDragOver(groupId: number, itemIndex: number, e: DragEvent) {
  e.preventDefault()
  dragOverItemKey.value = `${groupId}:${itemIndex}`
}

function onItemDrop(targetGroupId: number, targetItemIndex: number) {
  if (!dragItemKey.value) return
  const [srcGroupIdStr, srcItemIndexStr] = dragItemKey.value.split(':')
  const srcGroupId = parseInt(srcGroupIdStr)
  const srcItemIndex = parseInt(srcItemIndexStr)

  dragItemKey.value = null
  dragOverItemKey.value = null

  if (srcGroupId !== targetGroupId) return

  const group = groups.value.find(g => g.id === srcGroupId)
  if (!group) return

  const from = srcItemIndex
  const to = targetItemIndex
  if (from === to) return

  const item = group.items.splice(from, 1)[0]
  group.items.splice(to, 0, item)

  reorderStockGroupItems(group.items.map(i => i.id))
}

function onItemDragEnd() {
  dragItemKey.value = null
  dragOverItemKey.value = null
}

const availableStocksForGroup = computed(() => {
  if (selectedGroupForAdd.value === null) return []
  const group = groups.value.find(g => g.id === selectedGroupForAdd.value)
  const existingCodes = new Set((group?.items || []).map(i => i.stock_code))
  return stocks.value.filter(s => !existingCodes.has(s.stock_code))
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">
        <h2>股票监控</h2>
        <span class="stock-count">(共 {{ stocks.length }} 只个股)</span>
      </div>
      <div class="header-right">
        <el-switch
          v-model="showOnlyHoldings"
          active-text="只看持仓"
          style="margin-right: 12px;"
        />
        <el-button v-if="showManagement" type="primary" @click="openAddGroup">
          <el-icon><Plus /></el-icon> 新建分组
        </el-button>
        <el-button :type="showManagement ? 'primary' : 'default'" @click="toggleManagement">
          {{ showManagement ? '完成' : '管理分组' }}
        </el-button>
        <el-button type="primary" class="add-btn" @click="openAddStock">
          <el-icon><Plus /></el-icon> 新增个股
        </el-button>
      </div>
    </div>

    <div class="portfolio-section" v-loading="loading">
      <!-- Custom Groups -->
      <div
        v-for="(group, gIndex) in filteredGroups"
        :key="group.id"
        class="group-section"
        :class="{ 'drag-over': dragOverGroupIndex === gIndex }"
      >
        <div
          class="section-header"
          :draggable="showManagement"
          @dragstart="onGroupDragStart(gIndex)"
          @dragover="onGroupDragOver(gIndex, $event)"
          @drop="onGroupDrop(gIndex)"
          @dragend="onGroupDragEnd"
        >
          <span v-show="showManagement" class="drag-handle" title="拖动排序">&#x2630;</span>
          <div class="section-icon">{{ group.name.charAt(0) }}</div>
          <h3 class="section-title">{{ group.name }}</h3>
          <span class="group-count">({{ group.items?.length || 0 }} 只)</span>
          <div v-show="showManagement" class="section-actions">
            <button class="btn-icon" @click="openAddStockToGroup(group.id)">+ 添加</button>
            <button class="btn-icon" @click="openEditGroup(group)">编辑</button>
            <button class="btn-icon btn-danger" @click="confirmDeleteGroup(group.id)">删除</button>
          </div>
        </div>
        <div class="card-list stock-card-grid">
          <div
            v-for="(item, iIndex) in group.items"
            :key="item.id"
            class="card-wrapper"
            :class="{ 'drag-over': dragOverItemKey === `${group.id}:${iIndex}` }"
            :draggable="showManagement"
            @dragstart="onItemDragStart(group.id, iIndex)"
            @dragover="onItemDragOver(group.id, iIndex, $event)"
            @drop="onItemDrop(group.id, iIndex)"
            @dragend="onItemDragEnd"
          >
            <StockCard
              v-if="filteredStockMap.get(item.stock_code)"
              :stock="filteredStockMap.get(item.stock_code)!"
              @detail="openStockDetail"
              @settings="openSettingsStock"
              @edit="openEditStock"
              @delete="onDeleteStock"
            />
            <button v-show="showManagement" class="btn-remove-item" @click="confirmRemoveItem(item.id)" title="从分组移除">&times;</button>
          </div>
        </div>
        <el-empty v-if="(!group.items || group.items.length === 0) && showManagement" description="分组为空" />
      </div>

      <!-- Ungrouped Section -->
      <div v-if="ungroupedStocks.length > 0 || !showOnlyHoldings" class="group-section">
        <div class="section-header">
          <div class="section-icon" style="background: linear-gradient(135deg, #9ca3af 0%, #6b7280 100%);">未</div>
          <h3 class="section-title">未分组</h3>
          <span class="group-count">({{ ungroupedStocks.length }} 只)</span>
        </div>
        <div class="card-list stock-card-grid">
          <StockCard
            v-for="s in ungroupedStocks"
            :key="s.stock_code"
            :stock="s"
            @detail="openStockDetail"
            @settings="openSettingsStock"
            @edit="openEditStock"
            @delete="onDeleteStock"
          />
        </div>
      </div>

      <el-empty v-if="!loading && filteredStocks.length === 0" description="暂无个股持仓" />
    </div>

    <StockDialog v-model:visible="stockDialogVisible" :initial-data="editingStock" @submit="onAddStock" />
    <StockDetailDrawer
      v-model:visible="drawerVisible"
      :stock-code="selectedStockCode"
      :stock-name="selectedStockName"
    />

    <!-- Add Group Dialog -->
    <el-dialog v-model="showAddGroup" title="新建分组" width="400px" :close-on-click-modal="false">
      <el-input v-model="newGroupName" placeholder="请输入分组名称" maxlength="20" show-word-limit />
      <template #footer>
        <el-button @click="showAddGroup = false">取消</el-button>
        <el-button type="primary" @click="confirmAddGroup" :disabled="!newGroupName.trim()">确定</el-button>
      </template>
    </el-dialog>

    <!-- Edit Group Dialog -->
    <el-dialog v-model="showEditGroup" title="编辑分组" width="400px" :close-on-click-modal="false">
      <el-input v-model="editGroupName" placeholder="请输入分组名称" maxlength="20" show-word-limit />
      <template #footer>
        <el-button @click="showEditGroup = false">取消</el-button>
        <el-button type="primary" @click="confirmEditGroup" :disabled="!editGroupName.trim()">确定</el-button>
      </template>
    </el-dialog>

    <!-- Add Stock to Group Dialog -->
    <el-dialog v-model="showAddStockToGroup" title="添加个股到分组" width="500px" :close-on-click-modal="false">
      <div v-if="selectedGroupForAdd" class="selected-group-info">
        添加到分组：<strong>{{ groups.find(g => g.id === selectedGroupForAdd)?.name }}</strong>
      </div>
      <div class="stock-option-list">
        <label
          v-for="stock in availableStocksForGroup"
          :key="stock.stock_code"
          class="stock-option"
        >
          <input type="checkbox" :checked="selectedStockCodes.includes(stock.stock_code)" @change="toggleStockCode(stock.stock_code)">
          <span>{{ stock.stock_name }} ({{ stock.stock_code }})</span>
        </label>
        <el-empty v-if="availableStocksForGroup.length === 0" description="暂无可添加的个股" />
      </div>
      <template #footer>
        <el-button @click="showAddStockToGroup = false">取消</el-button>
        <el-button type="primary" @click="confirmAddStocksToGroup" :disabled="selectedStockCodes.length === 0">
          确定 ({{ selectedStockCodes.length }})
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 12px;
}

.page-title {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.page-title h2 {
  font-size: 28px;
  font-weight: 800;
  color: #111827;
  margin: 0;
  letter-spacing: -0.5px;
}

.stock-count {
  font-size: 14px;
  color: #9ca3af;
  font-weight: 500;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.add-btn {
  border-radius: 8px;
  padding: 10px 20px;
  font-weight: 600;
}

.portfolio-section {
  margin-bottom: 32px;
}

.group-section {
  margin-bottom: 32px;
  transition: all 0.2s;
}

.group-section.drag-over {
  background: #eff6ff;
  border-radius: 12px;
  padding: 8px;
  margin: -8px;
  margin-bottom: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding: 4px;
  border-radius: 8px;
  cursor: grab;
}

.section-header:active {
  cursor: grabbing;
}

.drag-handle {
  color: #9ca3af;
  font-size: 16px;
  cursor: grab;
  user-select: none;
  padding: 2px 4px;
}

.drag-handle:active {
  cursor: grabbing;
}

.section-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
}

.section-title {
  font-size: 18px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.group-count {
  font-size: 13px;
  color: #9ca3af;
  font-weight: 500;
}

.section-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.btn-icon {
  padding: 4px 10px;
  border-radius: 6px;
  background: #f3f4f6;
  color: #6b7280;
  border: none;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: #e5e7eb;
  color: #374151;
}

.btn-danger {
  color: #ef4444;
}

.btn-danger:hover {
  background: #fef2f2;
  color: #dc2626;
}

.stock-card-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.card-wrapper {
  position: relative;
  transition: all 0.2s;
}

.card-wrapper.drag-over {
  transform: scale(1.02);
}

.card-wrapper.drag-over :deep(.stock-card) {
  border: 2px dashed #3b82f6;
}

.btn-remove-item {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #fef2f2;
  color: #ef4444;
  border: 1px solid #fecaca;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  opacity: 0;
  transition: opacity 0.2s;
}

.card-wrapper:hover .btn-remove-item {
  opacity: 1;
}

.btn-remove-item:hover {
  background: #ef4444;
  color: white;
}

.selected-group-info {
  margin-bottom: 16px;
  padding: 10px 14px;
  background: #f3f4f6;
  border-radius: 8px;
  font-size: 14px;
  color: #374151;
}

.stock-option-list {
  max-height: 400px;
  overflow-y: auto;
}

.stock-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
  font-size: 14px;
  color: #4b5563;
}

.stock-option:hover {
  background: #f3f4f6;
}

.stock-option input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: #2563eb;
  cursor: pointer;
}

@media (max-width: 768px) {
  .stock-card-grid {
    grid-template-columns: 1fr;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-right {
    width: 100%;
    flex-wrap: wrap;
  }
}
</style>
