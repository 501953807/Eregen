const { ApiClient } = require('../../utils/api')

Page({
  data: {
    name: '',
    birthDate: '',
    today: new Date().toISOString().slice(0, 10),
    loading: false,
  },

  onNameInput(e) {
    this.setData({ name: e.detail.value })
  },

  onBirthDateChange(e) {
    this.setData({ birthDate: e.detail.value })
  },

  submit() {
    const { name, birthDate } = this.data
    if (!name.trim()) {
      wx.showToast({ title: '请输入老人姓名', icon: 'none' })
      return
    }
    if (!birthDate) {
      wx.showToast({ title: '请选择出生日期', icon: 'none' })
      return
    }
    this.setData({ loading: true })
    const api = new ApiClient()
    api.post('/elderly', {
      name: name.trim(),
      birth_date: birthDate,
    })
      .then((res) => {
        wx.setStorageSync('last_elderly', res || {})
        wx.showToast({ title: '添加成功', icon: 'success' })
        setTimeout(() => {
          wx.switchTab({ url: '/pages/home/index' })
        }, 1500)
      })
      .catch((e) => {
        wx.showToast({ title: e.message || '添加失败', icon: 'none' })
      })
      .finally(() => {
        this.setData({ loading: false })
      })
  },
})
