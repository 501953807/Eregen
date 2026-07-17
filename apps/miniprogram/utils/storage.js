const STORAGE_PREFIX = 'eregen_'

function set(key, value) {
  wx.setStorageSync(STORAGE_PREFIX + key, JSON.stringify(value))
}

function get(key) {
  const raw = wx.getStorageSync(STORAGE_PREFIX + key)
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch (e) {
    return raw
  }
}

function remove(key) {
  wx.removeStorageSync(STORAGE_PREFIX + key)
}

function clear() {
  const keys = wx.getStorageInfoSync().keys
  keys.forEach(k => {
    if (k.startsWith(STORAGE_PREFIX)) wx.removeStorageSync(k)
  })
}

module.exports = { set, get, remove, clear }
