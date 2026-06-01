<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import type { TrendPoint } from '../api/stock'

const props = defineProps<{
  data: TrendPoint[]
  color?: string
}>()

const canvasRef = ref<HTMLCanvasElement | null>(null)

function draw() {
  const canvas = canvasRef.value
  if (!canvas || props.data.length < 2) return

  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const dpr = window.devicePixelRatio || 1
  const width = canvas.clientWidth
  const height = canvas.clientHeight
  canvas.width = width * dpr
  canvas.height = height * dpr
  ctx.scale(dpr, dpr)

  const values = props.data.map(d => d.close)
  const min = Math.min(...values)
  const max = Math.max(...values)
  const range = max - min || 1
  const padding = 2

  const xStep = (width - padding * 2) / (props.data.length - 1)
  const getY = (v: number) => height - padding - ((v - min) / range) * (height - padding * 2)
  const getX = (i: number) => padding + i * xStep

  ctx.clearRect(0, 0, width, height)

  // Draw area fill
  ctx.beginPath()
  ctx.moveTo(getX(0), height)
  ctx.lineTo(getX(0), getY(values[0]))
  for (let i = 1; i < values.length; i++) {
    ctx.lineTo(getX(i), getY(values[i]))
  }
  ctx.lineTo(getX(values.length - 1), height)
  ctx.closePath()
  const gradient = ctx.createLinearGradient(0, 0, 0, height)
  const baseColor = props.color || '#3b82f6'
  gradient.addColorStop(0, baseColor + '20')
  gradient.addColorStop(1, baseColor + '05')
  ctx.fillStyle = gradient
  ctx.fill()

  // Draw line
  ctx.beginPath()
  ctx.moveTo(getX(0), getY(values[0]))
  for (let i = 1; i < values.length; i++) {
    ctx.lineTo(getX(i), getY(values[i]))
  }
  ctx.strokeStyle = baseColor
  ctx.lineWidth = 2
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.stroke()
}

onMounted(draw)
watch(() => props.data, draw, { deep: true })
</script>

<template>
  <canvas ref="canvasRef" class="sparkline" />
</template>

<style scoped>
.sparkline {
  width: 100%;
  height: 60px;
}
</style>
