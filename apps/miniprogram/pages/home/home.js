Page({
  data: {
    activeElderly: 0,
    elderlyList: [
      { id: 1, name: '爷爷', avatar: '👴', online: true },
      { id: 2, name: '奶奶', avatar: '👵', online: false },
    ],
    healthData: { hr: 72, spo2: 98, steps: 3456, battery: 85 },
    location: { address: '陆家嘴环路 1000 号', updated: '更新于 1 分钟前 · 在安全区域内' },
    medications: [
      { name: '氨氯地平片 5mg', time: '08:00', status: 'taken', takenTime: '08:12' },
      { name: '阿司匹林肠溶片 100mg', time: '08:00', status: 'taken', takenTime: '08:12' },
      { name: '阿托伐他汀钙片 20mg', time: '13:00', status: 'taken', takenTime: '13:05' },
      { name: '氨氯地平片 5mg', time: '18:00', status: 'pending' },
    ],
    alerts: [
      { type: 'sos', title: 'SOS 紧急呼叫', desc: '爷爷触发 SOS，已通知所有紧急联系人', time: '2 分钟前', level: 'critical' },
      { type: 'med', title: '用药漏服提醒', desc: '奶奶早餐降压药未按时服用', time: '昨天 08:30', level: 'warning' },
    ],
  },
  switchElderly(e) {
    this.setData({ activeElderly: e.currentTarget.dataset.index })
  },
})
