const { login: apiLogin } = require('./api')

function wxLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: async (res) => {
        if (res.code) {
          try {
            const data = await apiLogin(res.code)
            resolve(data)
          } catch (e) {
            reject(e)
          }
        } else {
          reject(new Error('wx.login failed'))
        }
      },
      fail: reject,
    })
  })
}

function getToken() {
  return wx.getStorageSync('token') || ''
}

function isLoggedIn() {
  return !!wx.getStorageSync('token')
}

module.exports = { wxLogin, getToken, isLoggedIn }
