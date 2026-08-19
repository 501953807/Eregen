const { request } = require('../../utils/api')

Page({
  data: {
    activeElderly: 0,
    elderlyList: [],
    healthData: { hr: 0, spo2: 0, steps: 0, battery: 0 },
    location: { address: '', updated: '' },
    medications: [],
    alerts: [],
    loading: true,
    version: '1.0.0',
  },

  onLoad() {},

  onShow() {
    this.refresh()
  },

  async refresh() {
    const token = wx.getStorageSync('token')
    if (!token) {
      this.setData({ loading: false })
      return
    }

    try {
      await Promise.all([
        this.fetchElderlyList(),
        this.fetchHealth(),
        this.fetchLocation(),
        this.fetchMedications(),
        this.fetchAlerts(),
      ])
    } catch (e) {
      console.warn('home refresh failed:', e)
    } finally {
      this.setData({ loading: false })
    }
  },

  /* ---------- Elderly list ---------- */

  async fetchElderlyList() {
    try {
      const res = await request('/api/v1/admin/elderly?page_size=20', {}, 'GET')
      const profiles = (res.data || []).map((p, i) => ({
        id: p.id,
        name: p.name,
        avatar: i % 2 === 0 ? '👴' : '👵',
        online: true,
      }))
      if (profiles.length && this.data.activeElderly >= profiles.length) {
        this.setData({ activeElderly: 0 })
      }
      this.setData({ elderlyList: profiles })
    } catch (e) {
      console.warn('fetchElderlyList failed:', e)
    }
  },

  switchElderly(e) {
    this.setData({ activeElderly: e.currentTarget.dataset.index })
    this.fetchHealth()
    this.fetchLocation()
    this.fetchMedications()
  },

  /* ---------- Health summary ---------- */

  async fetchHealth() {
    const elders = this.data.elderlyList
    if (!elders.length) return
    const elder = elders[this.data.activeElderly]
    if (!elder || !elder.id) return

    try {
      const res = await request(`/api/v1/admin/elderly/${elder.id}/health-stats`, {}, 'GET')
      const d = (res.data || {})
      this.setData({
        healthData: {
          hr: d.avg_hr || d.max_hr || 0,
          spo2: d.avg_spo2 || 0,
          steps: d.total_steps || 0,
          battery: 85,
        },
      })
    } catch (e) {
      console.warn('fetchHealth failed:', e)
    }
  },

  /* ---------- Latest location ---------- */

  async fetchLocation() {
    const elders = this.data.elderlyList
    if (!elders.length) return
    const elder = elders[this.data.activeElderly]
    if (!elder || !elder.id) return

    // Location history endpoint not yet fully implemented for family app view
    // Falls back to demo data
    this.setData({
      location: {
        address: '上海市浦东新区陆家嘴环路1000号',
        updated: '更新于 2分钟前',
      },
    })
  },

  /* ---------- Today's medication ---------- */

  async fetchMedications() {
    const elders = this.data.elderlyList
    if (!elders.length) return
    const elder = elders[this.data.activeElderly]
    if (!elder || !elder.id) return

    try {
      const res = await request(`/api/v1/admin/persons/${elder.id}/medications`, {}, 'GET')
      const items = res.data || []
      const meds = items.slice(0, 4).map(m => {
        const schedTime = m.schedule_time || m.time || '00:00'
        const pillName = m.pill_type || m.name || `药物 (${schedTime})`
        return {
          name: pillName,
          time: schedTime,
          status: 'pending',
          takenTime: '',
        }
      })
      this.setData({ medications: meds })
    } catch (e) {
      console.warn('fetchMedications failed:', e)
    }
  },

  /* ---------- Recent alerts ---------- */

  async fetchAlerts() {
    try {
      const res = await request('/api/v1/admin/alerts?limit=5', {}, 'GET')
      const raw = res.data || []
      const alerts = raw.map(a => ({
        type: a.alert_type,
        title: this._alertTitle(a.alert_type),
        desc: a.description || '',
        time: this._timeAgo(a.created_at),
        level: a.severity === 'high' ? 'critical' : (a.severity === 'medium' ? 'warning' : 'info'),
      }))
      this.setData({ alerts })
    } catch (e) {
      console.warn('fetchAlerts failed:', e)
    }
  },

  /* ---------- Toggle medication status ---------- */

  toggleMed(e) {
    const idx = e.currentTarget.dataset.index
    const meds = [...this.data.medications]
    if (meds[idx]) {
      meds[idx].status = meds[idx].status === 'taken' ? 'pending' : 'taken'
      this.setData({ medications: meds })
    }
  },

  /* ---------- Navigation helpers ---------- */

  goToSettings() {
    wx.showToast({ title: '设置功能开发中', icon: 'none' })
  },

  goHealthReport() {
    wx.navigateTo({ url: '/pages/health/index' })
  },

  goConsult() {
    wx.showToast({ title: '在线咨询功能开发中', icon: 'none' })
  },

  goMedShop() {
    wx.showToast({ title: '药品购买功能开发中', icon: 'none' })
  },

  goDevice() {
    wx.navigateTo({ url: '/pages/bind-device/index' })
  },

  /* ---------- Helpers ---------- */

  _timeAgo(ts) {
    if (!ts) return '未知时间'
    const diff = (Date.now() - new Date(ts).getTime()) / 1000
    if (diff < 60) return '刚刚'
    if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
    if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
    return `${Math.floor(diff / 86400)} 天前`
  },

  _formatTime(ts) {
    const d = new Date(ts)
    return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
  },

  _alertTitle(type) {
    const map = {
      sos: 'SOS 紧急呼叫',
      fall: '跌倒检测触发',
      heart: '心率异常',
      spo2: '血氧偏低',
      geofence_breach: '电子围栏越界',
      med_missed: '用药漏服提醒',
      med_late: '用药延迟提醒',
      device_offline: '设备离线',
      low_battery: '电量不足',
    }
    return map[type] || type
  },

  _formatSteps(steps) {
    if (!steps || steps === 0) return '--'
    if (steps >= 10000) return `${(steps / 10000).toFixed(1)}万`
    if (steps >= 1000) return `${(steps / 1000).toFixed(1)}k`
    return String(steps)
  },

  _medRatio() {
    const meds = this.data.medications
    if (!meds.length) return '0/0'
    const taken = meds.filter(m => m.status === 'taken').length
    return `${taken}/${meds.length}`
  },
})
