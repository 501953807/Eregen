App({
  globalData: {
    elderlyList: [
      { id: 1, name: '爷爷', avatar: '👴', online: true },
      { id: 2, name: '奶奶', avatar: '👵', online: false },
    ],
  },

  /**
   * App lifecycle: onLaunch — called when app starts.
   * Check for stored JWT token and initialize map plugin.
   */
  onLaunch: function() {
    // Initialize from secure storage (WeChat usesStorageSync for simple persistence)
    const token = wx.getStorageSync('token');
    if (token) {
      this.globalData.currentUserToken = token;
    }

    // Initialize Tencent Map plugin
    this.initMapPlugin();
  },

  /**
   * Get current auth token.
   */
  getToken: function() {
    return wx.getStorageSync('token') || this.globalData.currentUserToken || '';
  },

  /**
   * Set and store auth token.
   * @param {string} token
   */
  setToken: function(token) {
    wx.setStorageSync('token', token);
    this.globalData.currentUserToken = token;
  },

  /**
   * Clear token (logout).
   */
  logout: function() {
    wx.removeStorageSync('token');
    this.globalData.currentUserToken = '';
  },

  /**
   * Initialize Tencent Maps plugin.
   */
  initMapPlugin: function() {
    try {
      // Import Tencent Map plugin (configured in project app.json)
      if (typeof Plugin !== 'undefined') {
        Plugin.import('weixin-map', {
          appId: 'YOUR_TENCENT_MAP_APP_ID', // Replace with actual App ID from Tencent console
        });
        console.log('Tencent Map plugin initialized');
      }
    } catch (e) {
      console.warn('Map plugin initialization failed:', e);
    }
  },

  /**
   * Global error handler.
   */
  onError: function(msg) {
    console.error('App error:', msg);
    wx.showToast({ title: '系统出错，请稍后重试', icon: 'none' });
  },
})

