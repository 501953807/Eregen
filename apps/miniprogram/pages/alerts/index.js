const API_BASE = 'https://api.eregen.com/api/v1'

Page({
  data: {
    filterTab: 0,
    filters: ['全部', '未处理', 'SOS', '跌倒', '健康'],
    alerts: [],
    loading: true,
  },

  onLoad() {
    this.fetchAlerts()
  },

  onShow() {
    // Refresh when returning to page
    this.fetchAlerts()
  },

  switchFilter(e) {
    const index = e.currentTarget.dataset.index
    this.setData({ filterTab: index })
    this.filterAlerts(index)
  },

  async fetchAlerts() {
    try {
      const token = wx.getStorageSync('token')
      if (!token) {
        this.setData({ alerts: [], loading: false })
        return
      }
      const res = await this._request('/alerts?limit=50', {}, token)
      const raw = Array.isArray(res.data) ? res.data : (res.data?.data || [])
      const alerts = raw.map(a => ({
        id: a.id,
        type: a.alert_type,
        title: this._alertTitle(a.alert_type),
        device: a.device_id || '',
        time: a.created_at?.slice(0, 16) || '未知时间',
        status: a.status === 'pending' ? 'unread' : 'read',
        priority: a.severity || 'P2',
      }))
      this.setData({ alerts, loading: false })
    } catch (e) {
      console.warn('fetchAlerts failed:', e)
      this.setData({ alerts: [], loading: false })
    }
  },

  _alertTitle(type) {
    const map = {
      sos: 'SOS 紧急呼叫',
      fall: '跌倒检测触发',
      heart: '心率异常',
      spo2: '血氧偏低',
      geofence: '电子围栏越界',
      med_missed: '用药漏服提醒',
      med_late: '用药延迟提醒',
      high_temp: '药盒温度异常',
    }
    return map[type] || type
  },

  filterAlerts(index) {
    const keyword = this.data.filters[index].toLowerCase()
    let filtered = this.data.alerts
    if (index > 0) {
      if (keyword === '未处理') filtered = filtered.filter(a => a.status === 'unread')
      else if (keyword === 'sos') filtered = filtered.filter(a => a.type === 'sos')
      else if (keyword === '跌倒') filtered = filtered.filter(a => a.type === 'fall')
      else if (keyword === '健康') filtered = filtered.filter(a => ['heart', 'spo2'].includes(a.type))
    }
    this.setData({ alerts: filtered })
  },

  onAlertTap(e) {
    const alertId = e.currentTarget.dataset.id
    if (!alertId) return
    // Mark as read via API
    const token = wx.getStorageSync('token')
    if (token) {
      this._request(`/alerts/${alertId}/resolve`, {}, token).catch(() => {})
    }
    // Navigate to detail
    wx.navigateTo({ url: `/pages/alert-detail/index?id=${alertId}` })
  },

  _request(url, data, token) {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${API_BASE}${url}`,
        method: 'POST',
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
