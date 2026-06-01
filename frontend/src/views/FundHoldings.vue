<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listFunds, createFund, updateFund, deleteFund, type Fund } from '../api/fund'
import {
  getFundGroups, createFundGroup, updateFundGroup, deleteFundGroup,
  reorderFundGroups, createFundGroupItem, deleteFundGroupItem, reorderFundGroupItems,
  type FundGroup
} from '../api/fundGroup'
import FundCard from '../components/FundCard.vue'
import FundDialog from '../components/FundDialog.vue'
import FundDetailDrawer from '../components/FundDetailDrawer.vue'

const funds = ref<Fund[]>([])
const groups = ref<FundGroup[]>([])
const loading = ref(false)
const showManagement = ref(false)
const showOnlyHoldings = ref(false)

// Modals
const showAddGroup = ref(false)
const showEditGroup = ref(false)
const showAddFundToGroup = ref(false)
const newGroupName = ref('')
const editGroupId = ref<number | null>(null)
const editGroupName = ref('')
const selectedGroupForAdd = ref<number | null>(null)
const selectedFundCodes = ref<string[]>([])

// Drag state
const dragGroupIndex = ref<number | null>(null)
const dragOverGroupIndex = ref<number | null>(null)
const dragItemKey = ref<string | null>(null)
const dragOverItemKey = ref<string | null>(null)

const fundDialogVisible = ref(false)
const editingFund = ref<any>(undefined)

const drawerVisible = ref(false)
const selectedFundCode = ref('')
const selectedFundName = ref('')

const fundMap = computed(() => {
  const map = new Map<string, Fund>()
  for (const f of funds.value) {
    map.set(f.fund_code, f)
  }
  return map
})

const filteredFunds = computed(() => {
  if (!showOnlyHoldings.value) return funds.value
  return funds.value.filter(f => (f.hold_quantity ?? 0) > 0)
})

const filteredFundMap = computed(() => {
  const map = new Map<string, Fund>()
  for (const f of filteredFunds.value) {
    map.set(f.fund_code, f)
  }
  return map
})

const groupedFundCodes = computed(() => {
  const codes = new Set<string>()
  for (const g of groups.value) {
    for (const item of g.items || []) {
      codes.add(item.fund_code)
    }
  }
  return codes
})

const ungroupedFunds = computed(() => {
  return filteredFunds.value.filter(f => !groupedFundCodes.value.has(f.fund_code))
})

const filteredGroups = computed(() => {
  if (!showOnlyHoldings.value) return groups.value
  return groups.value.map(g => ({
    ...g,
    items: (g.items || []).filter(item => {
      const fund = fundMap.value.get(item.fund_code)
      return fund && (fund.hold_quantity ?? 0) > 0
    })
  })).filter(g => g.items.length > 0 || !showOnlyHoldings.value)
})

async function fetchData() {
  loading.value = true
  try {
    const [f, g] = await Promise.all([listFunds(), getFundGroups()])
    funds.value = f
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

function openEditFund(fund: Fund) {
  editingFund.value = { ...fund }
  fundDialogVisible.value = true
}

function openAddFund() {
  editingFund.value = undefined
  fundDialogVisible.value = true
}

function onFundClick(fund: Fund) {
  selectedFundCode.value = fund.fund_code
  selectedFundName.value = fund.fund_name
  drawerVisible.value = true
}

// Group CRUD
function openAddGroup() {
  newGroupName.value = ''
  showAddGroup.value = true
}

async function confirmAddGroup() {
  if (!newGroupName.value.trim()) return
  await createFundGroup(newGroupName.value.trim())
  showAddGroup.value = false
  await fetchData()
}

function openEditGroup(group: FundGroup) {
  editGroupId.value = group.id
  editGroupName.value = group.name
  showEditGroup.value = true
}

async function confirmEditGroup() {
  if (!editGroupName.value.trim() || editGroupId.value === null) return
  await updateFundGroup(editGroupId.value, editGroupName.value.trim())
  showEditGroup.value = false
  await fetchData()
}

async function confirmDeleteGroup(id: number) {
  try {
    await ElMessageBox.confirm('确定删除该分组及其中的基金?', '提示', { type: 'warning' })
    await deleteFundGroup(id)
    await fetchData()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.error || '删除失败')
    }
  }
}

// Item CRUD
async function openAddFundToGroup(groupId: number | null) {
  if (groupId === null) {
    if (groups.value.length === 0) {
      ElMessage.warning('请先新建分组')
      return
    }
    groupId = groups.value[0].id
  }
  selectedGroupForAdd.value = groupId
  selectedFundCodes.value = []
  showAddFundToGroup.value = true
}

function toggleFundCode(code: string) {
  const idx = selectedFundCodes.value.indexOf(code)
  if (idx >= 0) {
    selectedFundCodes.value.splice(idx, 1)
  } else {
    selectedFundCodes.value.push(code)
  }
}

async function confirmAddFundsToGroup() {
  if (selectedGroupForAdd.value === null || selectedFundCodes.value.length === 0) {
    showAddFundToGroup.value = false
    return
  }
  const groupId = selectedGroupForAdd.value
  for (const code of selectedFundCodes.value) {
    const fund = fundMap.value.get(code)
    if (!fund) continue
    await createFundGroupItem(groupId, {
      fund_code: fund.fund_code,
      fund_name: fund.fund_name
    })
  }
  showAddFundToGroup.value = false
  await fetchData()
}

async function confirmRemoveItem(itemId: number) {
  try {
    await ElMessageBox.confirm('确定将该基金从分组中移除?', '提示', { type: 'warning' })
    await deleteFundGroupItem(itemId)
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
  reorderFundGroups(groups.value.map(g => g.id))
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

  reorderFundGroupItems(group.items.map(i => i.id))
}

function onItemDragEnd() {
  dragItemKey.value = null
  dragOverItemKey.value = null
}

const availableFundsForGroup = computed(() => {
  if (selectedGroupForAdd.value === null) return []
  const group = groups.value.find(g => g.id === selectedGroupForAdd.value)
  const existingCodes = new Set((group?.items || []).map(i => i.fund_code))
  return funds.value.filter(f => !existingCodes.has(f.fund_code))
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">
        <h2>基金持仓</h2>
        <span class="fund-count">(共 {{ funds.length }} 支基金)</span>
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
        <el-button type="primary" class="add-btn" @click="openAddFund">
          <el-icon><Plus /></el-icon> 新增基金
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
          <span class="group-count">({{ group.items?.length || 0 }} 支)</span>
          <div v-show="showManagement" class="section-actions">
            <button class="btn-icon" @click="openAddFundToGroup(group.id)">+ 添加</button>
            <button class="btn-icon" @click="openEditGroup(group)">编辑</button>
            <button class="btn-icon btn-danger" @click="confirmDeleteGroup(group.id)">删除</button>
          </div>
        </div>
        <div class="card-list">
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
            <FundCard
              v-if="filteredFundMap.get(item.fund_code)"
              :fund="filteredFundMap.get(item.fund_code)!"
              class="clickable-card"
              @click="onFundClick(filteredFundMap.get(item.fund_code)!)"
              @edit="openEditFund"
              @delete="onDeleteFund"
            />
            <button v-show="showManagement" class="btn-remove-item" @click="confirmRemoveItem(item.id)" title="从分组移除">&times;</button>
          </div>
        </div>
        <el-empty v-if="(!group.items || group.items.length === 0) && showManagement" description="分组为空" />
      </div>

      <!-- Ungrouped Section -->
      <div v-if="ungroupedFunds.length > 0 || !showOnlyHoldings" class="group-section">
        <div class="section-header">
          <div class="section-icon" style="background: linear-gradient(135deg, #9ca3af 0%, #6b7280 100%);">未</div>
          <h3 class="section-title">未分组</h3>
          <span class="group-count">({{ ungroupedFunds.length }} 支)</span>
        </div>
        <div class="card-list">
          <FundCard
            v-for="f in ungroupedFunds"
            :key="f.fund_code"
            :fund="f"
            class="clickable-card"
            @click="onFundClick(f)"
            @edit="openEditFund"
            @delete="onDeleteFund"
          />
        </div>
      </div>

      <el-empty v-if="!loading && filteredFunds.length === 0" description="暂无基金持仓" />
    </div>

    <FundDialog v-model:visible="fundDialogVisible" :initial-data="editingFund" @submit="onAddFund" />
    <FundDetailDrawer v-model:visible="drawerVisible" :fund-code="selectedFundCode" :fund-name="selectedFundName" />

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

    <!-- Add Fund to Group Dialog -->
    <el-dialog v-model="showAddFundToGroup" title="添加基金到分组" width="500px" :close-on-click-modal="false">
      <div v-if="selectedGroupForAdd" class="selected-group-info">
        添加到分组：<strong>{{ groups.find(g => g.id === selectedGroupForAdd)?.name }}</strong>
      </div>
      <div class="fund-option-list">
        <label
          v-for="fund in availableFundsForGroup"
          :key="fund.fund_code"
          class="fund-option"
        >
          <input type="checkbox" :checked="selectedFundCodes.includes(fund.fund_code)" @change="toggleFundCode(fund.fund_code)">
          <span>{{ fund.fund_name }} ({{ fund.fund_code }})</span>
        </label>
        <el-empty v-if="availableFundsForGroup.length === 0" description="暂无可添加的基金" />
      </div>
      <template #footer>
        <el-button @click="showAddFundToGroup = false">取消</el-button>
        <el-button type="primary" @click="confirmAddFundsToGroup" :disabled="selectedFundCodes.length === 0">
          确定 ({{ selectedFundCodes.length }})
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

.fund-count {
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

.card-list {
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

.card-wrapper.drag-over :deep(.fund-card) {
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

.clickable-card {
  cursor: pointer;
}

.selected-group-info {
  margin-bottom: 16px;
  padding: 10px 14px;
  background: #f3f4f6;
  border-radius: 8px;
  font-size: 14px;
  color: #374151;
}

.fund-option-list {
  max-height: 400px;
  overflow-y: auto;
}

.fund-option {
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

.fund-option:hover {
  background: #f3f4f6;
}

.fund-option input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: #2563eb;
  cursor: pointer;
}

@media (max-width: 768px) {
  .card-list {
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
