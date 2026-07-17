Page({
  data: {
    elderly: {
      name: '李秀英',
      age: 72,
      gender: '女',
      birthday: '1954-03-15',
      tier: 'plus',
    },
    healthStats: [
      { label: '心率', value: '72', unit: 'bpm', emoji: '❤️', bg: 'linear-gradient(135deg, #FF6B6B, #EE5A5A)' },
      { label: '血氧', value: '98', unit: '%', emoji: '🫁', bg: 'linear-gradient(135deg, #4A90D9, #357ABD)' },
      { label: '步数', value: '3,456', unit: '步', emoji: '👣', bg: 'linear-gradient(135deg, #67C23A, #529B2E)' },
      { label: '睡眠', value: '7.5', unit: '小时', emoji: '😴', bg: 'linear-gradient(135deg, #909399, #6B6E72)' },
    ],
    devices: [
      { id: '1', name: '颐贞手环', model: 'GD32E230 中端版', type: 'bracelet', battery: 68, online: true },
      { id: '2', name: '颐贞药盒', model: 'ESP32-C3 智能版', type: 'pillbox', battery: 92, online: true },
    ],
    recentAlerts: [
      { id: '1', type: 'sos', title: 'SOS紧急呼叫', time: '今天 14:32', statusText: '已处理', status: 'resolved', emoji: '🆘' },
      { id: '2', type: 'med', title: '漏服药物提醒', time: '今天 12:05', statusText: '未处理', status: 'pending', emoji: '💊' },
      { id: '3', type: 'health', title: '心率偏高预警', time: '昨天 20:18', statusText: '已处理', status: 'resolved', emoji: '❤️' },
    ],
    medications: [
      { id: '1', name: '氨氯地平片', dose: '5mg × 1', time: '08:00', taken: true },
      { id: '2', name: '阿司匹林肠溶片', dose: '100mg × 1', time: '08:00', taken: true },
      { id: '3', name: '阿托伐他汀钙片', dose: '20mg × 1', time: '20:00', taken: false },
    ],
  },

  onLoad(options) {
    if (options.id) {
      // In production: fetch elderly detail by ID from API
      console.log('Loading elderly detail for:', options.id)
    }
  },

  get tierText() {
    const map = { starter: '入门版', plus: '中端版', pro: '高端版' }
    return map[this.data.elderly.tier] || '标准版'
  },

  goToMedication() {
    wx.switchTab({ url: '/pages/medication/index' })
  },

  goToLocation() {
    wx.showToast({ title: '地图功能开发中', icon: 'none' })
  },

  goToHealth() {
    wx.showToast({ title: '健康报告开发中', icon: 'none' })
  },

  goToSettings() {
    wx.switchTab({ url: '/pages/mine/index' })
  },
})
