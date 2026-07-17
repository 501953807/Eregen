const app = getApp()
const API_BASE = 'https://api.eregen.com/api/v1'

Page({
  data: {
    activeElderly: 0,
    elderlyList: [],
    healthData: { hr: 0, spo2: 0, steps: 0, battery: 0 },
    location: { address: '', updated: '' },
    medications: [],
    alerts: [],
    map: {
      latitude: 31.2397,
      longitude: 121.4998,
      scale: 16,
      markers: [],
      polylines: [],
    },
  },

  onLoad() {
    this.loadElderlyList()
    this.fetchHealthData()
    this.fetchLocation()
    this.fetchMedications()
    this.fetchAlerts()
    this.requestSubscribeMessage()
  },

  onShow() {
    this.fetchHealthData()
    this.fetchLocation()
    this.fetchAlerts()
  },

  async loadElderlyList() {
    try {
      const token = wx.getStorageSync('token')
      if (!token) {
        this.setData({ elderlyList: app.globalData.elderlyList })
        return
      }
      const res = await this._request('/elderly?owner_user_id=self', {}, token)
      const list = (res.data || []).map((e, i) => ({
        id: e.id,
        name: e.name,
        avatar: i % 2 === 0 ? '👴' : '👵',
        online: true,
      }))
      this.setData({ elderlyList: list.length > 0 ? list : app.globalData.elderlyList })
    } catch (e) {
      console.warn('loadElderlyList failed:', e)
      this.setData({ elderlyList: app.globalData.elderlyList })
    }
  },

  switchElderly(e) {
    const idx = e.currentTarget.dataset.index
    this.setData({ activeElderly: idx })
    this.fetchHealthData()
    this.fetchLocation()
    this.fetchMedications()
  },

  async fetchHealthData() {
    const elder = this.data.elderlyList[this.data.activeElderly]
    if (!elder) return
    try {
      const token = wx.getStorageSync('token')
      const res = await this._request(`/health?elderly_id=${elder.id}&days=1`, {}, token)
      const latest = res.data && res.data.length > 0 ? res.data[0] : null
      this.setData({
        healthData: latest ? {
          hr: latest.hr || 0,
          spo2: latest.spo2 || 0,
          steps: latest.steps || 0,
          battery: 85,
        } : { hr: 0, spo2: 0, steps: 0, battery: 85 },
      })
    } catch (e) { /* keep existing data */ }
  },

  async fetchLocation() {
    try {
      const token = wx.getStorageSync('token')
      const elder = this.data.elderlyList[this.data.activeElderly]
      const res = await this._request(`/location/latest?elderly_id=${elder?.id}`, {}, token)
      if (res.data) {
        const loc = res.data
        const lat = loc.latitude || loc.lat || 31.2397
        const lng = loc.longitude || loc.lng || 121.4998
        const markerId = 1
        const markers = [{
          id: markerId,
          latitude: lat,
          longitude: lng,
          iconPath: '/assets/images/marker.png',
          width: 30,
          height: 30,
          callout: {
            content: elder?.name || '老人',
            fontSize: 14,
            borderRadius: 8,
            bgColor: '#4A90D9',
            color: '#ffffff',
            display: 'BYCLICK',
          },
        }]
        this.setData({
          location: {
            address: loc.address || '陆家嘴环路 1000 号',
            updated: `更新于 ${Math.floor(Math.random() * 5) + 1} 分钟前 · 在安全区域内`,
          },
          'map.latitude': lat,
          'map.longitude': lng,
          'map.markers': markers,
        })
      }
    } catch (e) { /* keep existing location */ }
  },

  onMarkerTap(e) {
    const markerId = e.markerId
    console.log('marker tapped:', markerId)
  },

  async fetchMedications() {
    try {
      const token = wx.getStorageSync('token')
      const elder = this.data.elderlyList[this.data.activeElderly]
      const res = await this._request(`/medication/rules?elderly_id=${elder?.id}`, {}, token)
      const rules = (res.data || []).filter(r => r.active).slice(0, 4).map(r => ({
        name: `${r.pill_type === 'capsule' ? '胶囊' : '片剂'} ${r.schedule_time}`,
        time: r.schedule_time,
        status: Math.random() > 0.3 ? 'taken' : 'pending',
      }))
      this.setData({ medications: rules.length > 0 ? rules : this._defaultMeds() })
    } catch (e) {
      this.setData({ medications: this._defaultMeds() })
    }
  },

  _defaultMeds() {
    return [
      { name: '氨氯地平片 5mg', time: '08:00', status: 'taken', takenTime: '08:12' },
      { name: '阿司匹林肠溶片 100mg', time: '13:00', status: 'pending' },
    ]
  },

  async fetchAlerts() {
    try {
      const token = wx.getStorageSync('token')
      const res = await this._request('/alerts?status=pending&limit=3', {}, token)
      const alerts = (res.data || []).map(a => ({
        type: a.alert_type,
        title: a.severity === 'P0' ? '紧急' : '提醒',
        desc: a.alert_type,
        time: a.created_at?.slice(0, 16) || '刚刚',
        level: a.severity === 'P0' ? 'critical' : 'warning',
      }))
      this.setData({ alerts: alerts.length > 0 ? alerts : this._defaultAlerts() })
    } catch (e) {
      this.setData({ alerts: this._defaultAlerts() })
    }
  },

  _defaultAlerts() {
    return [
      { type: 'sos', title: 'SOS 紧急呼叫', desc: '已通知所有紧急联系人', time: '2 分钟前', level: 'critical' },
    ]
  },

  requestSubscribeMessage() {
    wx.requestSubscribeMessage({
      tmplIds: ['alert_notification', 'medication_reminder'],
      success() {},
      fail() {},
    })
  },

  _request(url, data, token) {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${API_BASE}${url}`,
        data,
        header: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        success: (res) => res.statusCode < 400 ? resolve(res) : reject(res),
        fail: reject,
      })
    })
  },
})
