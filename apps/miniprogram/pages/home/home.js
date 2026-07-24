const { request } = require('../../utils/api')

Page({
  data: {
    activeElderly: 0,
    elderlyList: [],
    healthData: { hr: 0, spo2: 0, steps: 0, battery: 0 },
    location: { address: '暂无定位数据', updated: '' },
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
      const res = await request('/elderly', {}, 'GET')
      const profiles = (res.data?.profiles || []).map((p, i) => ({
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
      const res = await request(`/elderly/${elder.id}/health/summary`, {}, 'GET')
      const d = res.data || {}
      this.setData({
        healthData: {
          hr: d.hr || 0,
          spo2: d.spo2 || 0,
          steps: d.steps || 0,
          battery: d.battery_pct || 0,
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

    try {
      const res = await request(`/elderly/${elder.id}/location/latest`, {}, 'GET')
      const loc = res.data || {}
      const addr = loc.address || `${loc.lat?.toFixed(4)}°, ${loc.lon?.toFixed(4)}°`
      const ts = loc.updated_at || loc.timestamp || loc.created_at
      this.setData({
        location: {
          address: addr,
          updated: ts ? `更新于 ${this._timeAgo(ts)}` : '',
        },
      })
    } catch (e) {
      console.warn('fetchLocation failed:', e)
    }
  },

  /* ---------- Today's medication ---------- */

  async fetchMedications() {
    const elders = this.data.elderlyList
    if (!elders.length) return
    const elder = elders[this.data.activeElderly]
    if (!elder || !elder.id) return

    try {
      const res = await request(`/elderly/${elder.id}/medication/today`, {}, 'GET')
      const items = res.data || []
      const meds = items.map(m => {
        const schedTime = m.schedule_time || m.time || '00:00'
        const pillName = m.pill_name || m.rule_name || `药物 (${schedTime})`
        const status = m.taken ? 'taken' : (m.missed_at ? 'missed' : 'pending')
        return {
          name: pillName,
          time: schedTime,
          status,
          takenTime: m.taken_at ? this._formatTime(m.taken_at) : '',
        }
      })
      this.setData({ medications: meds.slice(0, 4) })
    } catch (e) {
      console.warn('fetchMedications failed:', e)
    }
  },

  /* ---------- Recent alerts ---------- */

  async fetchAlerts() {
    try {
      const res = await request('/alerts?page_size=5', {}, 'GET')
      const raw = res.data?.alerts || []
      const alerts = raw.map(a => ({
        type: a.severity === 'P0' ? 'sos' : (a.severity === 'P1' ? 'warning' : 'info'),
        title: this._alertTitle(a.alert_type),
        desc: a.description || '',
        time: this._timeAgo(a.created_at),
        level: a.severity === 'P0' ? 'critical' : (a.severity === 'P1' ? 'warning' : 'info'),
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
    wx.navigateTo({ url: '/pages/settings/index' })
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
    wx.navigateTo({ url: '/pages/device/index' })
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
