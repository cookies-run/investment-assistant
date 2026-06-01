<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { searchFunds, type FundSearchResult } from '../api/fund'
import { getAvailableIndices, type AvailableIndex } from '../api/market'

const props = defineProps<{
  visible: boolean
  initialData?: any
}>()
const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'submit', data: any): void
}>()

const form = ref<any>({
  fund_code: '',
  fund_name: '',
  hold_cost: 0,
  hold_quantity: 0,
  daily_profit_line: 2,
  daily_loss_line: 2,
  cumulative_profit_line: 8,
  cumulative_loss_line: 8,
  cumulative_days: 5,
  fund_type: 'ACTIVE',
  related_index_symbol: '',
  base_currency: 'CNY',
  is_active: true,
})

const searchLoading = ref(false)
const fundOptions = ref<FundSearchResult[]>([])
const indexOptions = ref<AvailableIndex[]>([])

watch(() => props.visible, async (v) => {
  if (v) {
    if (props.initialData) {
      form.value = { ...props.initialData }
    } else {
      resetForm()
    }
    fundOptions.value = []
    try {
      indexOptions.value = await getAvailableIndices()
    } catch {
      indexOptions.value = []
    }
  }
})

function resetForm() {
  form.value = {
    fund_code: '',
    fund_name: '',
    hold_cost: 0,
    hold_quantity: 0,
    daily_profit_line: 2,
    daily_loss_line: 2,
    cumulative_profit_line: 8,
    cumulative_loss_line: 8,
    cumulative_days: 5,
    fund_type: 'ACTIVE',
    related_index_symbol: '',
    base_currency: 'CNY',
    is_active: true,
  }
  fundOptions.value = []
}

async function handleFundSearch(query: string) {
  if (!query || query.length < 2) {
    fundOptions.value = []
    return
  }
  searchLoading.value = true
  try {
    fundOptions.value = await searchFunds(query)
  } catch {
    fundOptions.value = []
  } finally {
    searchLoading.value = false
  }
}

function onFundSelect(val: string) {
  const selected = fundOptions.value.find(f => f.code === val)
  if (selected) {
    form.value.fund_code = selected.code
    form.value.fund_name = selected.name
  }
}

function submit() {
  if (!form.value.fund_code || !form.value.fund_name) {
    ElMessage.warning('请填写基金代码和名称')
    return
  }
  emit('submit', { ...form.value })
  emit('update:visible', false)
  resetForm()
}
</script>

<template>
  <el-dialog :model-value="visible" @update:model-value="$emit('update:visible', $event)" :title="initialData ? '编辑基金' : '新增基金'" width="520px">
    <el-form :model="form" label-width="110px">
      <el-form-item label="基金代码" required>
        <el-select
          v-if="!initialData"
          v-model="form.fund_code"
          filterable
          clearable
          remote
          reserve-keyword
          placeholder="输入基金名称或代码搜索"
          :remote-method="handleFundSearch"
          :loading="searchLoading"
          style="width: 100%"
          @change="onFundSelect"
        >
          <el-option
            v-for="item in fundOptions"
            :key="item.code"
            :label="`${item.name} (${item.code})`"
            :value="item.code"
          />
        </el-select>
        <el-input v-else v-model="form.fund_code" disabled />
      </el-form-item>

      <el-form-item label="基金名称" required>
        <el-input v-model="form.fund_name" :disabled="!initialData" />
      </el-form-item>

      <el-form-item label="基金类型">
        <el-radio-group v-model="form.fund_type">
          <el-radio label="ACTIVE">主动型</el-radio>
          <el-radio label="INDEX">指数型</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="form.fund_type === 'INDEX'" label="关联指数">
        <el-select
          v-model="form.related_index_symbol"
          filterable
          placeholder="选择关联指数"
          style="width: 100%"
        >
          <el-option
            v-for="item in indexOptions"
            :key="item.symbol"
            :label="`${item.name} (${item.symbol})`"
            :value="item.symbol"
          />
        </el-select>
      </el-form-item>

      <el-form-item v-if="form.fund_type === 'INDEX'" label="计价货币">
        <el-select v-model="form.base_currency" placeholder="选择货币" style="width: 100%">
          <el-option label="人民币 (CNY)" value="CNY" />
          <el-option label="港币 (HKD)" value="HKD" />
          <el-option label="美元 (USD)" value="USD" />
        </el-select>
      </el-form-item>

      <el-form-item label="持仓成本">
        <el-input-number v-model="form.hold_cost" :precision="4" :min="0" style="width: 100%;" />
      </el-form-item>

      <el-form-item label="持仓量">
        <el-input-number v-model="form.hold_quantity" :min="0" :step="100" style="width: 100%;" />
      </el-form-item>

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
