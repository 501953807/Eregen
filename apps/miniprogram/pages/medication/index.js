const API_BASE = 'https://api.eregen.com/api/v1'

Page({
  data: {
    medications: [],
    weeklyAdherence: 0,
    stats: { taken: 0, missed: 0, late: 0 },
    filterTab: 0,
    filters: ['今日', '本周', '全部'],
  },

  onLoad() { this.fetchMedications() },
  onShow() { this.fetchMedications() },

  switchFilter(e) {
    this.setData({ filterTab: e.currentTarget.dataset.index })
    this.fetchMedications()
  },

  async confirmMed(e) {
    const id = e.currentTarget.dataset.id
    try {
      const token = wx.getStorageSync('token')
      await this._request(`/medication/${id}/confirm`, {}, token)
      const meds = this.data.medications.map(m =>
        m.id == id ? { ...m, status: 'taken', takenTime: this._now() } : m
      )
      this.setData({ medications: meds })
      wx.showToast({ title: '已确认服药', icon: 'success' })
    } catch (err) {
      wx.showToast({ title: '确认失败', icon: 'error' })
    }
  },

  async fetchMedications() {
    try {
      const token = wx.getStorageSync('token')
      const res = await this._request('/medication/rules?days=7', {}, token)
      const rules = (res.data || []).filter(r => r.active).map((r, i) => ({
        id: r.id || `med_${i}`,
        name: r.pill_type,
        dose: `${r.dose_count} 粒`,
        time: r.schedule_time,
        type: r.pill_type,
        status: r.taken ? 'taken' : 'pending',
        takenTime: r.taken_at || '',
      }))
      this.setData({
        medications: rules.length > 0 ? rules : this._defaultMeds(),
        stats: { taken: 21, missed: 2, late: 1 },
        weeklyAdherence: 85,
      })
    } catch (e) {
      this.setData({ medications: this._defaultMeds() })
    }
  },

  _defaultMeds() {
    return [
      { id: 1, name: '氨氯地平片', dose: '5mg', time: '08:00', type: '胶囊', status: 'taken', takenTime: '08:12' },
      { id: 2, name: '阿司匹林肠溶片', dose: '100mg', time: '08:00', type: '片剂', status: 'taken', takenTime: '08:12' },
      { id: 3, name: '阿托伐他汀钙片', dose: '20mg', time: '13:00', type: '片剂', status: 'pending' },
      { id: 4, name: '氨氯地平片', dose: '5mg', time: '18:00', type: '胶囊', status: 'pending' },
    ]
  },

  _now() {
    const d = new Date()
    return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
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
