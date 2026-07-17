const API_BASE = 'https://api.eregen.com/api/v1'

Page({
  data: {
    medications: [],
    weeklyAdherence: 0,
    stats: { taken: 0, missed: 0, late: 0 },
    loading: true,
  },

  onLoad() {
    this.fetchMedications()
  },

  onShow() {
    this.fetchMedications()
  },

  async fetchMedications() {
    try {
      const token = wx.getStorageSync('token')
      if (!token) {
        this.setData({ medications: this._defaultMeds(), loading: false })
        return
      }
      // Fetch today's medication rules
      const res = await this._request('/medication/rules?active=true&limit=20', {}, token)
      const rawRules = Array.isArray(res.data) ? res.data : (res.data?.data || [])

      // Fetch today's medication log
      const logRes = await this._request('/medication/log?date=' + this._todayStr(), {}, token).catch(() => ({ data: [] }))
      const takenIds = new Set(Array.isArray(logRes.data) ? logRes.data.map(l => l.rule_id) : [])

      const meds = rawRules.slice(0, 8).map((r, i) => {
        const isTaken = takenIds.has(r.id)
        const now = this._nowMinutes()
        const schedTime = this._timeToMinutes(r.schedule_time || r.time || '08:00')
        const status = isTaken ? 'taken' : (now >= schedTime ? 'pending' : 'soon')
        return {
          id: r.id || i,
          name: r.pill_type || '药物',
          dose: r.dose || '',
          time: r.schedule_time || r.time || '08:00',
          type: r.dose ? `${r.dose}` : '',
          status,
          takenTime: isTaken ? this._formatTime(new Date()) : '',
        }
      })

      // Calculate adherence stats
      const taken = meds.filter(m => m.status === 'taken').length
      const missed = meds.filter(m => m.status === 'missed').length
      const late = meds.filter(m => m.status === 'pending' && this._nowMinutes() > this._timeToMinutes(m.time)).length

      this.setData({
        medications: meds,
        'stats.taken': taken,
        'stats.missed': missed,
        'stats.late': late,
        weeklyAdherence: meds.length > 0 ? Math.round(taken / meds.length * 100) : 0,
        loading: false,
      })
    } catch (e) {
      console.warn('fetchMedications failed:', e)
      this.setData({ medications: this._defaultMeds(), loading: false })
    }
  },

  markTaken(e) {
    const id = e.currentTarget.dataset.id
    const token = wx.getStorageSync('token')
    if (!token) return

    this._request(`/medication/log`, { rule_id: id, action: 'taken' }, token)
      .then(() => {
        const meds = this.data.medications.map(m =>
          m.id == id ? { ...m, status: 'taken', takenTime: this._formatTime(new Date()) } : m
        )
        this.setData({ medications: meds })
        wx.showToast({ title: '已记录', icon: 'success' })
      })
      .catch(() => {
        wx.showToast({ title: '记录失败', icon: 'error' })
      })
  },

  _defaultMeds() {
    return [
      { id: 1, name: '氨氯地平片', dose: '5mg', time: '08:00', type: '胶囊', status: 'taken', takenTime: '08:12' },
      { id: 2, name: '阿司匹林肠溶片', dose: '100mg', time: '08:00', type: '片剂', status: 'taken', takenTime: '08:12' },
      { id: 3, name: '阿托伐他汀钙片', dose: '20mg', time: '13:00', type: '片剂', status: 'pending' },
      { id: 4, name: '氨氯地平片', dose: '5mg', time: '18:00', type: '胶囊', status: 'pending' },
      { id: 5, name: '维生素D', dose: '400IU', time: '18:00', type: '软胶囊', status: 'pending' },
    ]
  },

  _todayStr() {
    const d = new Date()
    return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`
  },

  _nowMinutes() {
    const d = new Date()
    return d.getHours() * 60 + d.getMinutes()
  },

  _timeToMinutes(t) {
    const parts = t.split(':')
    return parseInt(parts[0]) * 60 + parseInt(parts[1])
  },

  _formatTime(d) {
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
