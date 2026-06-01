import { createRouter, createWebHashHistory } from 'vue-router'
import MarketDashboard from '../views/MarketDashboard.vue'
import StockHoldings from '../views/StockHoldings.vue'
import FundHoldings from '../views/FundHoldings.vue'
import AlertRecords from '../views/AlertRecords.vue'

const routes = [
  { path: '/', redirect: '/market' },
  { path: '/market', component: MarketDashboard },
  { path: '/stocks', component: StockHoldings },
  { path: '/funds', component: FundHoldings },
  { path: '/alerts', component: AlertRecords },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
