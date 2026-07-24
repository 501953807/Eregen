const app = getApp()
const API_BASE = 'https://api.eregen.com/api/v1'
const { listFirmware } = require('../../utils/api')

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
    // Firmware version check
    _checkingVersions: false,
    _firmwareVersions: {}, // device_id -> { latestVersion, hasUpdate, firmwareId }
  },

  onSerialInput(e) {
    this.setData({ serialNumber: e.detail.value })
  },

  onShow() {
    const devices = wx.getStorageSync('boundDevices')
    if (devices) {
      this.setData({ boundDevices: devices })
    }
    this._checkFirmwareVersions()
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
        tier: 'starter',
        fwVersion: '0.1',
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
      // Check firmware for newly bound device
      this._checkFirmwareVersions()
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
          this.setData({
            boundDevices: devices,
            [`_firmwareVersions.${id}`]: undefined,
          })
          wx.setStorageSync('boundDevices', devices)
          wx.showToast({ title: '已解绑', icon: 'success' })
        }
      },
    })
  },

  /**
   * Check firmware versions for all bound devices.
   */
  _checkFirmwareVersions() {
    const devices = this.data.boundDevices
    if (!devices || devices.length === 0) return

    this.setData({ _checkingVersions: true })

    // Collect unique device_type+tier combos to query
    const combos = []
    for (const dev of devices) {
      const key = `${dev.type}__${dev.tier || 'starter'}`
      if (!combos.includes(key)) {
        combos.push({ key, type: dev.type, tier: dev.tier || 'starter' })
      }
    }

    Promise.all(
      combos.map(c => listFirmware(c.type, c.tier).then(items => ({ ...c, items })))
    ).then(results => {
      const versions = {}
      for (const r of results) {
        const latest = r.items && r.items.length > 0 ? r.items[0] : null
        for (const dev of devices) {
          if (dev.type === r.type && (dev.tier || 'starter') === r.tier) {
            const currentVer = dev.fwVersion || 'v0.1'
            const latestVer = latest ? latest.version : null
            versions[dev.id] = {
              latestVersion: latestVer,
              hasUpdate: latestVer && this._isNewer(latestVer, currentVer),
              firmwareId: latest ? latest.id : '',
              deviceId: dev.id,
            }
          }
        }
      }
      this.setData({
        _firmwareVersions: versions,
        _checkingVersions: false,
      })
    }).catch(() => {
      this.setData({ _checkingVersions: false })
    })
  },

  /**
   * Simple semver-like comparison: is newer > current?
   */
  _isNewer(newer, current) {
    const parse = v => (v || '').replace(/^v/, '').split('.').map(Number)
    const a = parse(newer), b = parse(current)
    for (let i = 0; i < 3; i++) {
      if ((a[i] || 0) > (b[i] || 0)) return true
      if ((a[i] || 0) < (b[i] || 0)) return false
    }
    return false
  },

  /**
   * Trigger OTA push for a specific device.
   */
  _handlePushOTA(e) {
    const { deviceId, firmwareId } = e.currentTarget.dataset
    const ver = this.data._firmwareVersions[deviceId]
    if (!ver || !ver.hasUpdate) return

    wx.showModal({
      title: '推送OTA升级',
      content: `即将升级到 ${ver.latestVersion}，确定继续？`,
      success: (res) => {
        if (!res.confirm) return
        wx.showToast({ title: '推送中...', icon: 'loading' })
        // Note: pushOTA requires importing from api.js
        const { pushOTA: push } = require('../../utils/api')
        push(firmwareId, [deviceId])
          .then(() => {
            wx.hideLoading()
            wx.showToast({ title: '推送成功', icon: 'success' })
          })
          .catch(() => {
            wx.hideLoading()
            wx.showToast({ title: '推送失败', icon: 'none' })
          })
      },
    })
  },
})
