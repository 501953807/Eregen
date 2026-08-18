const { ApiClient } = require('../../utils/api')

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
      const api = new ApiClient()
      const elderlyId = wx.getStorageSync('elderly_id') || ''
      if (!elderlyId) {
        this.setData({ medications: this._defaultMeds(), loading: false })
        return
      }

      const res = await api.get(`/elderly/${elderlyId}/medication/today`)
      const items = Array.isArray(res?.data) ? res.data : []

      const meds = items.slice(0, 8).map(m => ({
        id: m.id,
        name: m.pill_name || m.rule_name || '药物',
        dose: m.dose || '',
        time: m.schedule_time || m.time || '08:00',
        type: m.type || '',
        status: m.taken ? 'taken' : (m.missed_at ? 'missed' : 'pending'),
        takenTime: m.taken_at ? this._formatTime(m.taken_at) : '',
      }))

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
    const ruleId = e.currentTarget.dataset.id
    const api = new ApiClient()
    api.post(`/medication/${ruleId}/take`)
      .then(() => {
        const meds = this.data.medications.map(m =>
          m.id == ruleId ? { ...m, status: 'taken', takenTime: this._formatTime(new Date()) } : m
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

  _nowMinutes() {
    const d = new Date()
    return d.getHours() * 60 + d.getMinutes()
  },

  _timeToMinutes(t) {
    const parts = t.split(':')
    return parseInt(parts[0]) * 60 + parseInt(parts[1])
  },

  _formatTime(d) {
    if (typeof d === 'string') d = new Date(d)
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  },
})
