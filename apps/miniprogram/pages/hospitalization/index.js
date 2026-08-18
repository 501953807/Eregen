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

  async _fetchHospitalization() {
    const token = wx.getStorageSync('token')
    if (!token) return
    // Hospitalization is a B2B feature; elderly in family app won't have admissions
    this._setMockStay()
  },

  _setMockStay() {
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
  },

  async _fetchTreatments() {
    const token = wx.getStorageSync('token')
    if (!token) return
    // Hospital treatment records require admin API; use mock data
    this.setData({
      treatments: [
        { id: 1, icon: '💉', title: '静脉输液', desc: '0.9%氯化钠 250ml + 头孢曲松 2g', time: '08:30' },
        { id: 2, icon: '🩺', title: '医生查房', desc: '王主任查房，血压控制良好，继续当前方案', time: '09:15' },
        { id: 3, icon: '❤️', title: '心电监测', desc: '24小时动态心电图，未见明显异常', time: '10:00' },
        { id: 4, icon: '💊', title: '用药调整', desc: '降压药剂量微调，增加利尿剂', time: '10:30' },
      ],
    })
  },

  async _fetchVerifications() {
    const token = wx.getStorageSync('token')
    if (!token) return
    // Verification records are managed in nurse terminal
    this.setData({
      verifications: [
        { id: 1, verified: true, purpose: '入院身份核验', verifier: '护士站-李护士', time: '07:45' },
        { id: 2, verified: true, purpose: '用药扫码核对', verifier: '护士站-李护士', time: '08:25' },
        { id: 3, verified: true, purpose: '腕带功能检测', verifier: '设备科-张工', time: '09:00' },
        { id: 4, verified: false, purpose: '外出审批核验', verifier: '值班医生', time: '14:30' },
      ],
    })
  },

  _formatTime(ts) {
    if (!ts) return ''
    const d = new Date(ts)
    if (isNaN(d.getTime())) return ts
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  },

  goBack() {
    wx.navigateBack()
  },

  viewAllVerifications() {
    wx.showToast({ title: '全部核验记录开发中', icon: 'none' })
  },
})
