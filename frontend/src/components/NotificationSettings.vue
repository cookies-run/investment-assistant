<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getNotificationConfig, saveNotificationConfig, type NotificationConfig } from '../api/notification'

const visible = ref(false)
const config = ref<NotificationConfig>({
  id: 0,
  feishu_webhook: '',
  enable_feishu: false,
})
const loading = ref(false)
const saving = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const data = await getNotificationConfig()
    if (data) {
      config.value = data
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})

async function save() {
  saving.value = true
  try {
    await saveNotificationConfig({
      feishu_webhook: config.value.feishu_webhook,
      enable_feishu: config.value.enable_feishu,
    })
    ElMessage.success('保存成功')
    visible.value = false
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

function open() {
  visible.value = true
}

defineExpose({ open })
</script>

<template>
  <el-dialog v-model="visible" title="通知设置" width="500px" :close-on-click-modal="false">
    <div v-loading="loading">
      <div class="setting-section">
        <div class="setting-header">
          <el-icon :size="20" color="#3370ff"><ChatDotRound /></el-icon>
          <span class="setting-title">飞书机器人推送</span>
        </div>
        <div class="setting-desc">
          当策略触发（止盈/止损）时，通过飞书群机器人推送消息到手机
        </div>

        <el-form label-position="top" class="settings-form">
          <el-form-item>
            <template #label>
              <div class="form-label">
                <span>启用飞书推送</span>
                <el-switch v-model="config.enable_feishu" />
              </div>
            </template>
          </el-form-item>

          <el-form-item label="飞书 Webhook 地址">
            <el-input
              v-model="config.feishu_webhook"
              placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxxx"
              :disabled="!config.enable_feishu"
            />
            <div class="input-hint">
              在飞书群设置中添加「自定义机器人」，复制 Webhook 地址填入此处
            </div>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="save" :loading="saving">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.setting-section {
  padding: 8px 4px;
}

.setting-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.setting-title {
  font-size: 16px;
  font-weight: 700;
  color: #1f2937;
}

.setting-desc {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 20px;
  line-height: 1.5;
}

.settings-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #374151;
}

.form-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.input-hint {
  font-size: 12px;
  color: #9ca3af;
  margin-top: 6px;
  line-height: 1.4;
}
</style>
