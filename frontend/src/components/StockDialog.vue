<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { searchStocks, type SearchResult } from '../api/stock'

const props = defineProps<{
  visible: boolean
  initialData?: any
}>()
const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'submit', data: any): void
}>()

const form = ref<any>({
  stock_code: '',
  stock_name: '',
  buy_price: 0,
  hold_quantity: 0,
  daily_profit_line: 3,
  daily_loss_line: 3,
  cumulative_profit_line: 8,
  cumulative_loss_line: 8,
  cumulative_days: 5,
  monitor_interval: 60,
  is_active: true,
})

const searching = ref(false)
const searchResults = ref<SearchResult[]>([])

watch(() => props.visible, (v) => {
  if (v) {
    if (props.initialData) {
      form.value = { ...props.initialData }
    } else {
      resetForm()
    }
  }
})

function resetForm() {
  form.value = {
    stock_code: '',
    stock_name: '',
    buy_price: 0,
    hold_quantity: 0,
    daily_profit_line: 3,
    daily_loss_line: 3,
    cumulative_profit_line: 8,
    cumulative_loss_line: 8,
    cumulative_days: 5,
    monitor_interval: 60,
    is_active: true,
  }
  searchResults.value = []
}

async function onSearchInput(q: string) {
  if (!q || q.length < 1) {
    searchResults.value = []
    return
  }
  searching.value = true
  try {
    searchResults.value = await searchStocks(q)
  } catch (e) {
    searchResults.value = []
  } finally {
    searching.value = false
  }
}

function onSelectStock(item: SearchResult) {
  form.value.stock_code = item.code
  form.value.stock_name = item.name
}

function onStockCodeChange(val: any) {
  const found = searchResults.value.find((r: SearchResult) => r.code === val)
  if (found) onSelectStock(found)
}

function submit() {
  if (!form.value.stock_code || !form.value.stock_name) {
    ElMessage.warning('请选择股票')
    return
  }
  const payload = {
    stock_code: form.value.stock_code,
    stock_name: form.value.stock_name,
    buy_price: form.value.buy_price,
    hold_quantity: form.value.hold_quantity,
    daily_profit_line: form.value.daily_profit_line,
    daily_loss_line: form.value.daily_loss_line,
    cumulative_profit_line: form.value.cumulative_profit_line,
    cumulative_loss_line: form.value.cumulative_loss_line,
    cumulative_days: form.value.cumulative_days,
    monitor_interval: form.value.monitor_interval,
    is_active: form.value.is_active,
  }
  emit('submit', payload)
  emit('update:visible', false)
  resetForm()
}
</script>

<template>
  <el-dialog :model-value="visible" @update:model-value="$emit('update:visible', $event)" :title="initialData ? '编辑个股' : '新增个股'" width="520px">
    <el-form :model="form" label-width="110px">
      <el-form-item label="股票代码" required>
        <el-select
          v-model="form.stock_code"
          filterable
          remote
          reserve-keyword
          placeholder="输入代码或名称搜索"
          :remote-method="onSearchInput"
          :loading="searching"
          style="width: 100%;"
          :disabled="!!initialData"
          @change="onStockCodeChange"
        >
          <el-option
            v-for="item in searchResults"
            :key="item.code"
            :label="`${item.name} (${item.code})`"
            :value="item.code"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="股票名称" required>
        <el-input v-model="form.stock_name" disabled />
      </el-form-item>

      <el-form-item label="买入价">
        <el-input-number v-model="form.buy_price" :precision="4" :min="0" style="width: 100%;" />
      </el-form-item>

      <el-form-item label="持仓量">
        <el-input-number v-model="form.hold_quantity" :min="0" :step="100" style="width: 100%;" />
      </el-form-item>

      <div style="margin: 16px 0 8px; font-size: 13px; font-weight: 600; color: #374151;">观察指标设置</div>

      <div class="settings-row">
        <el-form-item label="单日止盈线%">
          <el-input-number v-model="form.daily_profit_line" :precision="2" :min="0" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="单日止损线%">
          <el-input-number v-model="form.daily_loss_line" :precision="2" :min="0" style="width: 100%;" />
        </el-form-item>
      </div>

      <div class="settings-row">
        <el-form-item label="累计止盈线%">
          <el-input-number v-model="form.cumulative_profit_line" :precision="2" :min="0" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="累计止损线%">
          <el-input-number v-model="form.cumulative_loss_line" :precision="2" :min="0" style="width: 100%;" />
        </el-form-item>
      </div>

      <div class="settings-row">
        <el-form-item label="累计天数">
          <el-input-number v-model="form.cumulative_days" :min="1" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="监控频率(秒)">
          <el-input-number v-model="form.monitor_interval" :min="10" style="width: 100%;" />
        </el-form-item>
      </div>

      <el-form-item label="是否激活">
        <el-switch v-model="form.is_active" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" @click="submit">确定</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.settings-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
</style>
