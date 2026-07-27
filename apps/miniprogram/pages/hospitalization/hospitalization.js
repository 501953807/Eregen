// pages/hospitalization/index.js
const { request } = require('../../utils/api')

Page({
  data: {
    loading: true,
    currentStay: null,
    treatments: [],
    verifications: [],
  },

  onLoad() {
    this._fetchData()
  },

  onShow() {
    // Refresh when navigating back to this page
    this._fetchData()
  },

  async _fetchData() {
    const token = wx.getStorageSync('token')
    if (!token) {
      this.setData({ loading: false })
      return
    }

    try {
      await Promise.all([
        this._fetchHospitalization(),
        this._fetchTreatments(),
        this._fetchVerifications(),
      ])
    } catch (e) {
      console.warn('hospitalization fetch failed:', e)
    } finally {
      this.setData({ loading: false })
    }
  },

  /* ---------- Hospitalization info ---------- */

  async _fetchHospitalization() {
    const elders = this.data.elderlyList || []
    let elderlyId = ''
    if (elders.length) {
      const activeIdx = this.data.activeElderly || 0
      elderlyId = elders[activeIdx]?.id || ''
    } else {
      elderlyId = wx.getStorageSync('elderly_id') || ''
    }
    if (!elderlyId) return

    try {
      const res = await request(`/medical/hospitalization?elderly_id=${elderlyId}&status=active`, {}, 'GET')
      const stay = res.data || {}
      this.setData({
        currentStay: {
          hospital: stay.hospital || '市第一人民医院',
          bed: stay.bed || '3床',
          admissionDate: stay.admission_date || '2026-07-20',
          doctor: stay.doctor || '王主任',
          diagnosis: stay.diagnosis || '高血压三级（极高危）',
          wristbandType: stay.wristband_type || '医用腕带 (Plus)',
          wristbandId: stay.wristband_id || 'WB-H-20260720-0042',
          dischargeDate: stay.discharge_date || '待定',
        },
      })
    } catch (e) {
      console.warn('_fetchHospitalization failed:', e)
      // Use demo data when API unavailable
      this.setData({
        currentStay: {
          hospital: '市第一人民医院',
          bed: '3床',
          admissionDate: '2026-07-20',
          doctor: '王主任',
          diagnosis: '高血压三级（极高危）',
          wristbandType: '医用腕带 (Plus)',
          wristbandId: 'WB-H-20260720-0042',
          dischargeDate: '待定',
        },
      })
    }
  },

  /* ---------- Treatment records ---------- */

  async _fetchTreatments() {
    const token = wx.getStorageSync('token')
    if (!token) return

    try {
      const res = await request('/medical/treatments?status=active', {}, 'GET')
      const items = (res.data || []).map((t, i) => ({
        id: t.id || i,
        icon: this._treatmentIcon(t.type),
        title: t.title || t.type || '诊疗记录',
        desc: t.description || '',
        time: this._formatTime(t.timestamp || t.created_at),
      }))
      this.setData({ treatments: items })
    } catch (e) {
      console.warn('_fetchTreatments failed:', e)
      // Demo data
      this.setData({
        treatments: [
          { id: 1, icon: '💉', title: '静脉输液', desc: '0.9%氯化钠 250ml + 头孢曲松 2g', time: '08:30' },
          { id: 2, icon: '🩺', title: '医生查房', desc: '王主任查房，血压控制良好，继续当前方案', time: '09:15' },
          { id: 3, icon: '❤️', title: '心电监测', desc: '24小时动态心电图，未见明显异常', time: '10:00' },
          { id: 4, icon: '💊', title: '用药调整', desc: '降压药剂量微调，增加利尿剂', time: '10:30' },
        ],
      })
    }
  },

  /* ---------- Verification history ---------- */

  async _fetchVerifications() {
    const token = wx.getStorageSync('token')
    if (!token) return

    try {
      const res = await request('/medical/verifications?status=active', {}, 'GET')
      const items = (res.data || []).map((v, i) => ({
        id: v.id || i,
        verified: v.verified !== false,
        purpose: v.purpose || v.description || '腕带核验',
        verifier: v.verifier || '护士站',
        time: this._formatTime(v.created_at || v.timestamp),
      }))
      this.setData({ verifications: items })
    } catch (e) {
      console.warn('_fetchVerifications failed:', e)
      // Demo data
      this.setData({
        verifications: [
          { id: 1, verified: true, purpose: '入院身份核验', verifier: '护士站-李护士', time: '07:45' },
          { id: 2, verified: true, purpose: '用药扫码核对', verifier: '护士站-李护士', time: '08:25' },
          { id: 3, verified: true, purpose: '腕带功能检测', verifier: '设备科-张工', time: '09:00' },
          { id: 4, verified: false, purpose: '外出审批核验', verifier: '值班医生', time: '14:30' },
        ],
      })
    }
  },

  /* ---------- Helpers ---------- */

  _treatmentIcon(type) {
    const map = {
      infusion: '💉',
      checkup: '🩺',
      ecg: '❤️',
      medication: '💊',
      surgery: '🔬',
      rehab: '🏃',
    }
    return map[type] || '📋'
  },

  _formatTime(ts) {
    if (!ts) return ''
    const d = new Date(ts)
    if (isNaN(d.getTime())) return ts
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  },

  /* ---------- Navigation ---------- */

  goBack() {
    wx.navigateBack()
  },

  viewAllVerifications() {
    wx.showToast({ title: '全部核验记录开发中', icon: 'none' })
  },
})
