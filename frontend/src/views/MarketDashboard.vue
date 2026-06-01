<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import {
  getMarketDashboard, getMarketGroups, createMarketGroup, updateMarketGroup,
  deleteMarketGroup, reorderMarketGroups, createMarketItem, deleteMarketItem,
  reorderMarketItems, getAvailableIndices,
  type MarketGroup, type MarketDataItem, type AvailableIndex
} from '../api/market'
import { formatPercent, formatNumber } from '../utils/format'

const groups = ref<MarketGroup[]>([])
const marketDataMap = ref<Map<string, MarketDataItem>>(new Map())
const loading = ref(false)
const beijingTime = ref('')
const availableIndices = ref<AvailableIndex[]>([])

// Modals
const showAddGroup = ref(false)
const showEditGroup = ref(false)
const showAddIndex = ref(false)
const selectedGroupForAdd = ref<number | null>(null)
const newGroupName = ref('')
const editGroupId = ref<number | null>(null)
const editGroupName = ref('')
const selectedSymbols = ref<string[]>([])

// Drag state
const dragGroupIndex = ref<number | null>(null)
const dragOverGroupIndex = ref<number | null>(null)
const dragItemKey = ref<string | null>(null)
const dragOverItemKey = ref<string | null>(null)
const showManagement = ref(false)

function toggleManagement() {
  showManagement.value = !showManagement.value
}

function updateBeijingTime() {
  beijingTime.value = new Date().toLocaleString('zh-CN', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

async function fetchGroups() {
  try {
    groups.value = await getMarketGroups()
  } catch (e) {
    console.error(e)
  }
}

async function fetchDashboard() {
  if (groups.value.length === 0) {
    marketDataMap.value = new Map()
    return
  }
  loading.value = true
  try {
    const data = await getMarketDashboard()
    const map = new Map<string, MarketDataItem>()
    for (const cat of data.categories) {
      for (const item of cat.items) {
        map.set(item.symbol, item)
      }
    }
    marketDataMap.value = map
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function loadAll() {
  loading.value = true
  await fetchGroups()
  await fetchDashboard()
  loading.value = false
}

let timer: number
let clockTimer: number
onMounted(() => {
  loadAll()
  updateBeijingTime()
  timer = window.setInterval(tick, 30000)
  clockTimer = window.setInterval(updateBeijingTime, 1000)
})
onUnmounted(() => {
  clearInterval(timer)
  clearInterval(clockTimer)
})

let groupRefreshCounter = 0

function tick() {
  try {
    // Refresh groups every ~2 minutes to pick up changes
    groupRefreshCounter++
    if (groupRefreshCounter >= 4) {
      groupRefreshCounter = 0
      fetchGroups()
    }
    if (shouldAutoRefresh()) {
      fetchDashboard()
    }
  } catch (e) {
    console.error('tick error:', e)
  }
}

function getBeijingParts() {
  const now = new Date()
  const fmt = new Intl.DateTimeFormat('en-US', {
    timeZone: 'Asia/Shanghai',
    hour12: false,
    year: 'numeric', month: 'numeric', day: 'numeric',
    hour: 'numeric', minute: 'numeric', second: 'numeric'
  })
  const parts = fmt.formatToParts(now)
  const get = (type: string) => parseInt(parts.find(p => p.type === type)!.value)
  const year = get('year')
  const month = get('month')
  const dayOfMonth = get('day')
  const hour = get('hour')
  const minute = get('minute')
  const day = new Date(year, month - 1, dayOfMonth).getDay()
  return { year, month, day, dayOfMonth, hour, minute, timeVal: hour * 60 + minute }
}

function isSourceTypeActive(sourceType: string) {
  const { day, timeVal, month } = getBeijingParts()
  const isWeekday = day >= 1 && day <= 5

  switch (sourceType) {
    case 'a_share':
      return isWeekday && ((timeVal >= 570 && timeVal < 690) || (timeVal >= 780 && timeVal < 900))
    case 'hk':
      return isWeekday && ((timeVal >= 570 && timeVal < 720) || (timeVal >= 780 && timeVal < 960))
    case 'us': {
      const dst = month >= 3 && month <= 10
      const usCloseToday = dst ? 240 : 300
      if (isWeekday && timeVal >= (dst ? 1290 : 1350)) return true
      if (day >= 2 && day <= 6 && timeVal < usCloseToday) return true
      return false
    }
    case 'futures':
      return isWeekday
    default:
      return false
  }
}

function getGroupStatus(group: MarketGroup): { text: string; cls: string } | null {
  if (!group.items || group.items.length === 0) return null
  const anyActive = group.items.some(item => isSourceTypeActive(item.source_type))
  if (anyActive) {
    return { text: '交易中', cls: 'status-active' }
  }
  return { text: '已收盘', cls: 'status-closed' }
}

function shouldAutoRefresh() {
  if (!groups.value || groups.value.length === 0) return false
  const allItems = groups.value.flatMap(g => g.items || [])
  if (allItems.length === 0) return false
  return allItems.some(item => isSourceTypeActive(item.source_type))
}

function itemClass(pct: number | string | undefined) {
  const num = typeof pct === 'string' ? parseFloat(pct) : pct
  if (num === undefined || isNaN(num)) return 'flat'
  if (num > 0) return 'up'
  if (num < 0) return 'down'
  return 'flat'
}

function sparklinePath(type: string) {
  if (type === 'up') return 'M0,30 Q15,25 30,20 T60,5'
  if (type === 'down') return 'M0,5 Q15,10 30,15 T60,30'
  return 'M0,15 Q15,18 30,12 T60,15'
}

function formatCardDate(tradeDate?: string) {
  if (!tradeDate) return ''
  return tradeDate
}

// Group CRUD
function openAddGroup() {
  newGroupName.value = ''
  showAddGroup.value = true
}

async function confirmAddGroup() {
  if (!newGroupName.value.trim()) return
  await createMarketGroup(newGroupName.value.trim())
  showAddGroup.value = false
  await fetchGroups()
}

function openEditGroup(group: MarketGroup) {
  editGroupId.value = group.id
  editGroupName.value = group.name
  showEditGroup.value = true
}

async function confirmEditGroup() {
  if (!editGroupName.value.trim() || editGroupId.value === null) return
  await updateMarketGroup(editGroupId.value, editGroupName.value.trim())
  showEditGroup.value = false
  await fetchGroups()
}

async function confirmDeleteGroup(id: number) {
  if (!confirm('确定删除该分组及其所有指标吗？')) return
  await deleteMarketGroup(id)
  await fetchGroups()
  await fetchDashboard()
}

// Item CRUD
async function openAddIndex(groupId: number | null) {
  if (groupId === null) {
    if (groups.value.length === 0) {
      alert('请先新建分组')
      return
    }
    groupId = groups.value[0].id
  }
  selectedGroupForAdd.value = groupId
  selectedSymbols.value = []
  if (availableIndices.value.length === 0) {
    try {
      availableIndices.value = await getAvailableIndices()
    } catch (e) {
      console.error(e)
    }
  }
  showAddIndex.value = true
}

function toggleSymbol(symbol: string) {
  const idx = selectedSymbols.value.indexOf(symbol)
  if (idx >= 0) {
    selectedSymbols.value.splice(idx, 1)
  } else {
    selectedSymbols.value.push(symbol)
  }
}

async function confirmAddIndices() {
  if (selectedGroupForAdd.value === null || selectedSymbols.value.length === 0) {
    showAddIndex.value = false
    return
  }
  const groupId = selectedGroupForAdd.value
  for (const symbol of selectedSymbols.value) {
    const idx = availableIndices.value.find(i => i.symbol === symbol)
    if (!idx) continue
    await createMarketItem(groupId, {
      symbol: idx.symbol,
      name: idx.name,
      source_type: idx.source_type
    })
  }
  showAddIndex.value = false
  await fetchGroups()
  await fetchDashboard()
}

async function confirmDeleteItem(itemId: number) {
  if (!confirm('确定删除该指标吗？')) return
  await deleteMarketItem(itemId)
  await fetchGroups()
  await fetchDashboard()
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
  reorderMarketGroups(groups.value.map(g => g.id))
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

  reorderMarketItems(group.items.map(i => i.id))
}

function onItemDragEnd() {
  dragItemKey.value = null
  dragOverItemKey.value = null
}

// Group available indices by category
const groupedAvailableIndices = computed(() => {
  const map = new Map<string, AvailableIndex[]>()
  for (const idx of availableIndices.value) {
    if (!map.has(idx.category)) {
      map.set(idx.category, [])
    }
    map.get(idx.category)!.push(idx)
  }
  return Array.from(map.entries())
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <h2>全球主要市场实时行情</h2>
        <span class="time-tag">{{ beijingTime || '-' }} (北京时间)</span>
      </div>
      <div class="header-right">
        <button v-if="showManagement" class="btn-primary" @click="openAddGroup">+ 新建分组</button>
        <button class="btn-settings" :class="{ active: showManagement }" @click="toggleManagement">
          {{ showManagement ? '完成' : '设置' }}
        </button>
      </div>
    </div>

    <el-skeleton :rows="6" animated v-if="loading && groups.length === 0" />

    <div v-if="groups.length === 0" class="empty-state">
      <p>暂无分组，请先新建分组并添加指标</p>
      <button v-if="showManagement" class="btn-primary" @click="openAddGroup">新建分组</button>
    </div>

    <div
      v-for="(group, gIndex) in groups"
      :key="group.id"
      class="market-section"
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
        <span
          v-if="getGroupStatus(group)"
          :class="['status-tag', getGroupStatus(group)!.cls]"
        >
          {{ getGroupStatus(group)!.text }}
        </span>
        <div v-show="showManagement" class="section-actions">
          <button class="btn-icon" @click="openAddIndex(group.id)">+ 指标</button>
          <button class="btn-icon" @click="openEditGroup(group)">编辑</button>
          <button class="btn-icon btn-danger" @click="confirmDeleteGroup(group.id)">删除</button>
        </div>
      </div>
      <div class="market-grid">
        <div
          v-for="(item, iIndex) in group.items"
          :key="item.id"
          class="market-card"
          :class="[itemClass(marketDataMap.get(item.symbol)?.change_percent), { 'drag-over': dragOverItemKey === `${group.id}:${iIndex}` }]"
          :draggable="showManagement"
          @dragstart="onItemDragStart(group.id, iIndex)"
          @dragover="onItemDragOver(group.id, iIndex, $event)"
          @drop="onItemDrop(group.id, iIndex)"
          @dragend="onItemDragEnd"
        >
          <div class="market-card-top">
            <span class="market-name">{{ item.name }}</span>
            <button v-show="showManagement" class="btn-delete-item" @click="confirmDeleteItem(item.id)" title="删除">&times;</button>
          </div>
          <div class="market-card-body">
            <div class="market-price-box">
              <div class="market-price">{{ formatNumber(marketDataMap.get(item.symbol)?.price, 2) }}</div>
              <div v-if="marketDataMap.get(item.symbol)?.ma20 != null" class="market-ma20">
                MA20 {{ formatNumber(marketDataMap.get(item.symbol)?.ma20, 2) }}
              </div>
              <div class="market-change">
                <span :class="['change-badge', itemClass(marketDataMap.get(item.symbol)?.change_percent)]">
                  {{ formatPercent(marketDataMap.get(item.symbol)?.change_percent) }}
                </span>
              </div>
            </div>
            <svg class="sparkline" viewBox="0 0 60 35" preserveAspectRatio="none">
              <path
                :d="sparklinePath(itemClass(marketDataMap.get(item.symbol)?.change_percent))"
                fill="none"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </div>
          <div class="market-card-footer">
            <span class="card-date">{{ formatCardDate(marketDataMap.get(item.symbol)?.trade_date) }}</span>
          </div>
        </div>
      </div>
    </div>

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

    <!-- Add Index Dialog -->
    <el-dialog v-model="showAddIndex" title="添加指标" width="600px" :close-on-click-modal="false">
      <div v-if="selectedGroupForAdd" class="selected-group-info">
        添加到分组：<strong>{{ groups.find(g => g.id === selectedGroupForAdd)?.name }}</strong>
      </div>
      <div class="index-list">
        <div v-for="[cat, indices] in groupedAvailableIndices" :key="cat" class="index-category">
          <h4 class="index-category-title">{{ cat }}</h4>
          <div class="index-options">
            <label v-for="idx in indices" :key="idx.symbol" class="index-option">
              <input type="checkbox" :checked="selectedSymbols.includes(idx.symbol)" @change="toggleSymbol(idx.symbol)">
              <span>{{ idx.name }}</span>
            </label>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showAddIndex = false">取消</el-button>
        <el-button type="primary" @click="confirmAddIndices" :disabled="selectedSymbols.length === 0">
          确定 ({{ selectedSymbols.length }})
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
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right {
  display: flex;
  gap: 10px;
}

.time-tag {
  font-size: 13px;
  font-weight: 500;
  color: #6b7280;
  background: #f3f4f6;
  padding: 4px 12px;
  border-radius: 20px;
}

.btn-primary {
  padding: 8px 16px;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

.btn-settings {
  padding: 8px 16px;
  border-radius: 8px;
  background: #f3f4f6;
  color: #6b7280;
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-settings:hover {
  background: #e5e7eb;
}

.btn-settings.active {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
}

.empty-state {
  text-align: center;
  padding: 80px 20px;
  color: #9ca3af;
  font-size: 15px;
}

.empty-state p {
  margin-bottom: 20px;
}

.market-section {
  margin-bottom: 32px;
  transition: all 0.2s;
}

.market-section.drag-over {
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

.status-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 12px;
  margin-left: 8px;
}

.status-active {
  background: #f0fdf4;
  color: #10b981;
}

.status-closed {
  background: #f3f4f6;
  color: #9ca3af;
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

.market-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
}

.market-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
  border: 1px solid rgba(0, 0, 0, 0.04);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  cursor: grab;
}

.market-card:active {
  cursor: grabbing;
}

.market-card.drag-over {
  border: 2px dashed #3b82f6;
  transform: scale(1.02);
}

.market-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 3px;
  background: #e5e7eb;
  transition: background 0.3s;
}

.market-card.up::before {
  background: #ef4444;
}

.market-card.down::before {
  background: #10b981;
}

.market-card.flat::before {
  background: #9ca3af;
}

.market-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.08);
}

.market-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.market-name {
  font-size: 13px;
  font-weight: 500;
  color: #6b7280;
}

.market-card-footer {
  margin-top: 12px;
  padding-top: 10px;
  border-top: 1px solid rgba(0, 0, 0, 0.04);
  text-align: right;
}

.card-date {
  font-size: 11px;
  font-weight: 500;
  color: #9ca3af;
}

.btn-delete-item {
  background: none;
  border: none;
  color: #d1d5db;
  font-size: 18px;
  cursor: pointer;
  line-height: 1;
  padding: 0 2px;
  margin-left: 4px;
  flex-shrink: 0;
}

.btn-delete-item:hover {
  color: #ef4444;
}

.market-card-body {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
}

.market-price-box {
  flex: 1;
  min-width: 0;
}

.market-price {
  font-size: 28px;
  font-weight: 800;
  color: #111827;
  line-height: 1.2;
  margin-bottom: 4px;
}

.market-ma20 {
  font-size: 11px;
  font-weight: 500;
  color: #9ca3af;
  margin-bottom: 8px;
}

.market-change {
  display: flex;
  align-items: center;
}

.change-badge {
  font-size: 12px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 20px;
}

.change-badge.up {
  background: #fef2f2;
  color: #ef4444;
}

.change-badge.down {
  background: #f0fdf4;
  color: #10b981;
}

.change-badge.flat {
  background: #f3f4f6;
  color: #9ca3af;
}

.sparkline {
  width: 60px;
  height: 35px;
  flex-shrink: 0;
}

.market-card.up .sparkline path {
  stroke: #ef4444;
}

.market-card.down .sparkline path {
  stroke: #10b981;
}

.market-card.flat .sparkline path {
  stroke: #9ca3af;
}

.text-up {
  color: #ef4444;
}

.text-down {
  color: #10b981;
}

/* Dialog styles */
.selected-group-info {
  margin-bottom: 16px;
  padding: 10px 14px;
  background: #f3f4f6;
  border-radius: 8px;
  font-size: 14px;
  color: #374151;
}

.index-list {
  max-height: 400px;
  overflow-y: auto;
}

.index-category {
  margin-bottom: 16px;
}

.index-category-title {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
  margin: 0 0 10px 0;
  padding-bottom: 6px;
  border-bottom: 1px solid #e5e7eb;
}

.index-options {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.index-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
  font-size: 13px;
  color: #4b5563;
}

.index-option:hover {
  background: #f3f4f6;
}

.index-option input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: #2563eb;
  cursor: pointer;
}

@media (max-width: 640px) {
  .market-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header-right {
    width: 100%;
  }

  .btn-primary {
    flex: 1;
  }

  .index-options {
    grid-template-columns: 1fr;
  }
}
</style>
