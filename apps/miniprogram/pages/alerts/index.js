Page({
  data: {
    filterTab: 0,
    filters: ['全部', '未处理', 'SOS', '跌倒', '健康'],
    alerts: [
      { id: 1, type: 'sos', icon: '🆘', title: 'SOS 紧急呼叫', device: 'BR-0042', time: '2026-07-16 14:32', status: 'unread', priority: 'P0' },
      { id: 2, type: 'fall', icon: '⚠️', title: '跌倒检测触发', device: 'BR-0017', time: '2026-07-16 13:18', status: 'processing', priority: 'P0' },
      { id: 3, type: 'heart', icon: '💓', title: '心率异常偏高', device: 'BR-0089', time: '2026-07-16 12:05', status: 'resolved', priority: 'P1' },
      { id: 4, type: 'geofence', icon: '📍', title: '电子围栏越界', device: 'BR-0033', time: '2026-07-16 11:42', status: 'unread', priority: 'P1' },
      { id: 5, type: 'med', icon: '💊', title: '用药漏服提醒', device: 'PX-0012', time: '2026-07-16 10:15', status: 'resolved', priority: 'P2' },
    ],
  },
  switchFilter(e) {
    this.setData({ filterTab: e.currentTarget.dataset.index })
  },
})
