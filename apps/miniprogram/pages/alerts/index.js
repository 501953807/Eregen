const API_BASE = 'https://api.eregen.com/api/v1'

Page({
  data: {
    filterTab: 0,
    filters: ['全部', '未处理', 'SOS', '跌倒', '健康'],
    alerts: [],
  },

  onLoad() { this.fetchAlerts() },
  onShow() { this.fetchAlerts() },

  switchFilter(e) {
    this.setData({ filterTab: e.currentTarget.dataset.index })
    this.fetchAlerts()
  },

  async handleAlert(e) {
    const id = e.currentTarget.dataset.id
    const action = e.currentTarget.dataset.action
    try {
      const token = wx.getStorageSync('token')
      await this._request(`/alerts/${id}/${action}`, {}, token)
      const alerts = this.data.alerts.map(a =>
        a.id == id ? { ...a, status: action === 'resolve' ? 'resolved' : 'read' } : a
      )
      this.setData({ alerts })
      wx.showToast({ title: action === 'resolve' ? '已标记处理' : '已读', icon: 'success' })
    } catch (err) {
      wx.showToast({ title: '操作失败', icon: 'error' })
    }
  },

  async callEmergency() {
    wx.makePhoneCall({ phoneNumber: '120', fail: () => {} })
  },

  async fetchAlerts() {
    try {
      const token = wx.getStorageSync('token')
      const filterMap = { 0: '', 1: 'pending', 2: 'sos', 3: 'fall', 4: 'health' }
      const severity = filterMap[this.data.filterTab] || ''
      const res = await this._request(`/alerts?severity=${severity}&limit=20`, {}, token)
      const alerts = (res.data || []).map((a, i) => ({
        id: a.id || `alert_${i}`,
        type: a.alert_type,
        icon: a.alert_type === 'sos' ? '🆘' : (a.alert_type === 'fall' ? '⚠️' : '💓'),
        title: a.alert_type === 'sos' ? 'SOS 紧急呼叫' : (a.alert_type === 'fall' ? '跌倒检测触发' : '健康异常'),
        device: a.metadata?.device_id || 'BR-0042',
        time: a.created_at?.slice(0, 16) || '今天',
        status: a.status || 'unread',
        priority: a.severity || 'P1',
      }))
      this.setData({ alerts: alerts.length > 0 ? alerts : this._defaultAlerts() })
    } catch (e) {
      this.setData({ alerts: this._defaultAlerts() })
    }
  },

  _defaultAlerts() {
    return [
      { id: 1, type: 'sos', icon: '🆘', title: 'SOS 紧急呼叫', device: 'BR-0042', time: '2026-07-16 14:32', status: 'unread', priority: 'P0' },
      { id: 2, type: 'fall', icon: '⚠️', title: '跌倒检测触发', device: 'BR-0017', time: '2026-07-16 13:18', status: 'processing', priority: 'P0' },
      { id: 3, type: 'heart', icon: '💓', title: '心率异常偏高', device: 'BR-0089', time: '2026-07-16 12:05', status: 'resolved', priority: 'P1' },
    ]
  },

  _request(url, data, token) {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${API_BASE}${url}`,
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
