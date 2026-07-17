Page({
  data: {
    serialNumber: '',
    selectedTypeIndex: 0,
    selectedElderlyIndex: 0,
    binding: false,
    deviceTypes: ['智能手环', '智能药盒'],
    elderlyList: [
      { id: '1', name: '李秀英（奶奶）' },
      { id: '2', name: '王建国（爷爷）' },
    ],
    boundDevices: [],
  },

  onSerialInput(e) {
    this.setData({ serialNumber: e.detail.value })
  },

  onShow() {
    const devices = wx.getStorageSync('boundDevices')
    if (devices) {
      this.setData({ boundDevices: devices })
    }
  },

  onTypeChange(e) {
    this.setData({ selectedTypeIndex: e.detail.value })
  },

  onElderlyChange(e) {
    this.setData({ selectedElderlyIndex: e.detail.value })
  },

  scanQRCode() {
    wx.scanCode({
      success: (res) => {
        this.setData({ serialNumber: res.result })
        wx.showToast({ title: '已读取序列号', icon: 'success' })
      },
      fail: () => {},
    })
  },

  handleBind() {
    const sn = this.data.serialNumber.trim()
    if (!sn) {
      wx.showToast({ title: '请输入序列号', icon: 'none' })
      return
    }

    this.setData({ binding: true })

    setTimeout(() => {
      const type = this.data.deviceTypes[this.data.selectedTypeIndex]
      const deviceType = type === '智能手环' ? 'bracelet' : 'pillbox'
      const newDevice = {
        id: String(Date.now()),
        name: `颐贞${type}`,
        type: deviceType,
        model: deviceType === 'bracelet' ? 'GD32E230 标准版' : 'ESP32-C3 智能版',
        serialNumber: sn,
        boundAt: new Date().toLocaleString('zh-CN'),
        online: false,
      }

      const devices = [...this.data.boundDevices, newDevice]
      this.setData({
        boundDevices: devices,
        serialNumber: '',
        binding: false,
      })

      wx.setStorageSync('boundDevices', devices)
      wx.showToast({ title: '绑定成功', icon: 'success' })
    }, 1500)
  },

  handleUnbind(e) {
    const id = e.currentTarget.dataset.id
    if (!id) return

    wx.showModal({
      title: '确认解绑',
      content: '解绑后该设备将不再受监控，确定要解绑吗？',
      success: (res) => {
        if (res.confirm) {
          const devices = this.data.boundDevices.filter(d => d.id !== id)
          this.setData({ boundDevices: devices })
          wx.setStorageSync('boundDevices', devices)
          wx.showToast({ title: '已解绑', icon: 'success' })
        }
      },
    })
  },
})
