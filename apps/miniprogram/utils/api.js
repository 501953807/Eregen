const API_BASE = 'https://api.eregen.com/api/v1'

function request(url, data = {}, method = 'GET') {
  const token = wx.getStorageSync('token')
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE}${url}`,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      success: (res) => {
        if (res.statusCode < 400) {
          resolve(res.data)
        } else if (res.statusCode === 401) {
          wx.removeStorageSync('token')
          wx.reLaunch({ url: '/pages/login/index' })
          reject(new Error('unauthorized'))
        } else {
          reject(new Error(res.data?.message || 'request failed'))
        }
      },
      fail: reject,
    })
  })
}

function login(code) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE}/auth/wechat/login`,
      method: 'POST',
      data: { code },
      success: (res) => {
        if (res.statusCode < 400) {
          wx.setStorageSync('token', res.data.token)
          resolve(res.data)
        } else {
          reject(new Error(res.data?.message || 'login failed'))
        }
      },
      fail: reject,
    })
  })
}

module.exports = { request, login, API_BASE }
