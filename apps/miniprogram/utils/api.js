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

/**
 * Fetch firmware list with optional filters.
 * @param {'bracelet'|'pillbox'} [deviceType]
 * @param {string} [tier]
 * @returns {Promise<{id:string,version:string,device_type:string,tier:string}[]>}
 */
function listFirmware(deviceType, tier) {
  const token = wx.getStorageSync('token')
  return new Promise((resolve, reject) => {
    const qs = []
    if (deviceType) qs.push(`device_type=${deviceType}`)
    if (tier) qs.push(`tier=${tier}`)
    const url = `${API_BASE}/admin/firmware${qs.length ? '?' + qs.join('&') : ''}`
    wx.request({
      url,
      method: 'GET',
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      success: (res) => {
        if (res.statusCode < 400) {
          const items = (res.data?.data || [])
            .filter(f => f.active !== false)
            .sort((a, b) => (b.version || '').localeCompare(a.version || ''))
          resolve(items)
        } else if (res.statusCode === 401) {
          wx.removeStorageSync('token')
          wx.reLaunch({ url: '/pages/login/index' })
          reject(new Error('unauthorized'))
        } else {
          resolve([]) // non-fatal: firmware endpoint may not be ready
        }
      },
      fail: reject,
    })
  })
}

/**
 * Trigger OTA push for a firmware release.
 * @param {string} firmwareId
 * @param {string[]} [deviceIds]
 */
function pushOTA(firmwareId, deviceIds) {
  const token = wx.getStorageSync('token')
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE}/admin/ota/push`,
      method: 'POST',
      data: {
        firmware_id: firmwareId,
        ...(deviceIds && deviceIds.length ? { device_ids: deviceIds } : {}),
      },
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
          reject(new Error(res.data?.message || 'push failed'))
        }
      },
      fail: reject,
    })
  })
}
