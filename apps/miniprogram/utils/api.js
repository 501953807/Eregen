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

function listFirmware(deviceType, tier) {
  const token = wx.getStorageSync('token')
  return new Promise((resolve, reject) => {
    const qs = []
    if (deviceType) qs.push(`device_type=${deviceType}`)
    if (tier) qs.push(`tier=${tier}`)
    wx.request({
      url: `${API_BASE}/admin/firmware${qs.length ? '?' + qs.join('&') : ''}`,
      method: 'GET',
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      success: (res) => {
        if (res.statusCode < 400) {
          resolve(res.data || [])
        } else {
          resolve([])
        }
      },
      fail: reject,
    })
  })
}

function pushOTA(firmwareId, deviceIds) {
  const token = wx.getStorageSync('token')
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE}/admin/ota/push`,
      method: 'POST',
      data: { firmware_id: firmwareId, device_ids: deviceIds || [] },
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      success: (res) => {
        if (res.statusCode < 400) resolve(res.data)
        else reject(new Error(res.data?.message || 'push failed'))
      },
      fail: reject,
    })
  })
}

class ApiClient {
  constructor() {
    this.baseURL = API_BASE
  }

  _request(url, data = {}, method = 'GET') {
    const token = wx.getStorageSync('token')
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${this.baseURL}${url}`,
        method,
        data,
        header: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        success: (res) => {
          if (res.statusCode < 400) resolve(res.data)
          else if (res.statusCode === 401) {
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

  get(url, opts = {}) {
    const qs = opts.query
      ? Object.entries(opts.query)
          .filter(([, v]) => v !== undefined && v !== null && v !== '')
          .map(([k, v]) => `${k}=${encodeURIComponent(v)}`)
          .join('&')
      : ''
    return this._request(`${url}${qs ? '?' + qs : ''}`, {}, 'GET')
  }

  post(url, data = {}) {
    return this._request(url, data, 'POST')
  }
}

module.exports = { request, login, API_BASE, listFirmware, pushOTA, ApiClient }
