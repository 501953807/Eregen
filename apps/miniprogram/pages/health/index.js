// pages/health/index.js
const { ApiClient } = require('../../utils/api');
const util = require('../../utils/util');

Page({
  data: {
    riskScore: 0,
    riskLevel: '加载中...',
    riskLevelColor: '#4CAF50',
    riskSummary: '',
    timeRanges: [
      { label: '今日' },
      { label: '本周' },
      { label: '本月' },
      { label: '自定义' }
    ],
    currentRange: '本周',
    latestHR: null,
    latestSPO2: null,
    latestSteps: 0,
    latestSleep: null,
    latestSys: null,
    latestDia: null,
    hasBP: false,
    hrTrend: '稳定',
    hrTrendColor: '#4CAF50',
    spo2Trend: '正常',
    spo2TrendColor: '#4CAF50',
    stepsTrend: '偏低',
    stepsTrendColor: '#FFA726',
    sleepTrend: '改善',
    sleepTrendColor: '#4CAF50',
    formatSteps: '--',
    todaySteps: '0',
    calories: '0',
    activeMin: '0',
    showCompare: false,
    compareData: [],
    aiInsightText: ''
  },

  onLoad() {
    this._fetchHealthData();
  },

  onShow() {
    // Refresh when coming back
    this._fetchHealthData();
  },

  _fetchHealthData() {
    const api = new ApiClient();
    const elderlyId = wx.getStorageSync('elderly_id') || '';
    if (!elderlyId) return;

    api.get(`/health/history`, {
      query: { elderly_id: elderlyId, days: 7 }
    }).then(res => {
      if (res.code === 'OK' && res.data && res.data.length > 0) {
        this._processHealthData(res.data);
      }
    }).catch(() => {
      // Use demo data for now
      this._loadDemoData();
    });
  },

  _processHealthData(records) {
    const latest = records[0];
    const hr = latest.hr;
    const spo2 = latest.spo2;
    const steps = latest.steps || 0;
    const sleep = latest.sleep_hours;
    const sys = latest.bp_systolic;
    const dia = latest.bp_diastolic;

    // Risk score calculation
    let riskScore = 0;
    if (hr && (hr < 60 || hr > 100)) riskScore += 20;
    if (spo2 && spo2 < 95) riskScore += 30;
    if (sys && sys > 140) riskScore += 25;
    if (dia && dia > 90) riskScore += 15;
    if (sleep && sleep < 6) riskScore += 10;
    riskScore = Math.min(riskScore, 100);

    const riskLevel = riskScore < 30 ? '低风险' : (riskScore < 60 ? '中风险' : '高风险');
    const riskColor = riskScore < 30 ? '#4CAF50' : (riskScore < 60 ? '#FFA726' : '#FF5252');

    const hrTrend = hr ? (hr >= 60 && hr <= 100 ? '稳定' : '偏高') : '暂无数据';
    const spo2Trend = spo2 ? (spo2 >= 95 ? '正常' : '偏低') : '暂无数据';
    const stepsTrend = steps >= 5000 ? '达标' : '偏低';

    const insightParts = [];
    if (hr && hr >= 60 && hr <= 100) insightParts.push('心率稳定在正常范围');
    if (spo2 && spo2 >= 95) insightParts.push('血氧水平良好');
    if (steps && steps < 5000) insightParts.push('今日步数略低于目标（5000步），建议傍晚散步30分钟');
    if (sleep && sleep < 6) insightParts.push('睡眠不足6小时，注意休息');
    if (sys && sys > 140) insightParts.push('收缩压偏高，建议减少盐分摄入');
    if (insightParts.length === 0) insightParts.push('各项指标基本正常，继续保持健康生活方式');

    // Intergenerational comparison data
    const elderAvg = {
      hr: hr || 72,
      spo2: spo2 || 97,
      steps: steps || 3456,
      sys: sys || 120,
      sleep: sleep || 7.2
    };
    const peerAvg = {
      hr: 75,
      spo2: 96,
      steps: 6200,
      sys: 128,
      sleep: 6.5
    };

    const compareData = [
      this._makeCompareRow('心率', elderAvg.hr, peerAvg.hr, 'bpm', true),
      this._makeCompareRow('血氧', elderAvg.spo2, peerAvg.spo2, '%'),
      this._makeCompareRow('步数', elderAvg.steps, peerAvg.steps, '步'),
      this._makeCompareRow('血压', elderAvg.sys, peerAvg.sys, 'mmHg', true),
      this._makeCompareRow('睡眠', elderAvg.sleep, peerAvg.sleep, 'h')
    ];

    this.setData({
      riskScore,
      riskLevel,
      riskLevelColor: riskColor,
      riskSummary: this._getRiskSummary(hr, spo2, steps),
      latestHR: hr,
      latestSPO2: spo2,
      latestSteps: steps,
      latestSleep: sleep ? sleep.toFixed(1) : null,
      latestSys: sys,
      latestDia: dia,
      hasBP: !!(sys && dia),
      hrTrend,
      hrTrendColor: (hr && hr >= 60 && hr <= 100) ? '#4CAF50' : '#FF6B6B',
      spo2Trend,
      spo2TrendColor: (spo2 && spo2 >= 95) ? '#4CAF50' : '#FFA726',
      stepsTrend,
      stepsTrendColor: steps >= 5000 ? '#4CAF50' : '#FFA726',
      formatSteps: util.formatNumber(steps),
      todaySteps: util.formatNumber(steps),
      calories: String(Math.floor(steps * 0.04)),
      activeMin: String(Math.floor(steps / 120)),
      aiInsightText: insightParts.join('；'),
      compareData
    });

    this._drawGauge(riskScore);
    this._drawTrendChart(records);
  },

  _makeCompareRow(label, value, average, unit, lowerIsBetter) {
    const diff = value - average;
    const isBetter = lowerIsBetter
      ? Math.abs(diff) < 5
      : diff > 0;
    const diffPercent = average !== 0 ? ((diff / average) * 100).toFixed(1) : '0';
    const barWidth = Math.min(100, (value / (average * 1.3)) * 100);
    return {
      label,
      value,
      unit,
      isBetter,
      diffText: `${isBetter ? '-' : '+'}${diffPercent}%`,
      barWidth: Math.max(5, barWidth)
    };
  },

  _getRiskSummary(hr, spo2, steps) {
    if (hr && (hr < 60 || hr > 100)) return '心率略高于/低于正常范围，建议持续监测';
    if (spo2 && spo2 < 95) return '血氧偏低，请注意休息并咨询医生';
    return '各项指标基本正常，步数略低于目标值';
  },

  _loadDemoData() {
    const demoRecords = [
      { hr: 72, spo2: 97, steps: 4520, sleep_hours: 7.2, bp_systolic: 118, bp_diastolic: 76, timestamp: new Date().toISOString() }
    ];
    this._processHealthData(demoRecords);
  },

  _drawGauge(score) {
    const ctx = wx.createCanvasContext('riskGauge', this);
    const cx = 100, cy = 60, r = 45;

    // Background arc
    ctx.setStrokeStyle('rgba(255,255,255,0.2)');
    ctx.setLineWidth(10);
    ctx.beginPath();
    ctx.arc(cx, cy, r, Math.PI * 0.8, Math.PI * 2.2);
    ctx.stroke();

    // Score arc
    const color = score < 30 ? '#4CAF50' : (score < 60 ? '#FFA726' : '#FF5252');
    ctx.setStrokeStyle(color);
    ctx.setLineWidth(10);
    ctx.beginPath();
    ctx.arc(cx, cy, r, Math.PI * 0.8, Math.PI * 0.8 + (score / 100) * Math.PI * 1.4);
    ctx.stroke();

    ctx.draw();
  },

  _drawTrendChart(records) {
    if (!records || records.length === 0) return;
    const ctx = wx.createCanvasContext('trendChart', this);
    const w = 700, h = 260;
    const pad = { top: 30, bottom: 40, left: 10, right: 10 };
    const chartW = w - pad.left - pad.right;
    const chartH = h - pad.top - pad.bottom;

    // Background
    ctx.setFillStyle('#FFFFFF');
    ctx.fillRect(0, 0, w, h);

    // Grid lines
    ctx.setStrokeStyle('#F0F0F5');
    ctx.setLineWidth(1);
    for (let i = 0; i <= 4; i++) {
      const y = pad.top + (chartH / 4) * i;
      ctx.beginPath();
      ctx.moveTo(pad.left, y);
      ctx.lineTo(w - pad.right, y);
      ctx.stroke();
    }

    // Heart rate line
    const hrValues = records.slice(0, 7).map(r => r.hr || 70);
    this._drawLine(ctx, hrValues, '#4A90D9', pad, chartW, chartH, 7);

    // X-axis labels
    ctx.setFontSize(20);
    ctx.setFillStyle('#999999');
    const days = ['日', '一', '二', '三', '四', '五', '六'];
    const today = new Date();
    for (let i = 0; i < 7; i++) {
      const d = new Date(today);
      d.setDate(d.getDate() - (6 - i));
      const x = pad.left + (chartW / 6) * i;
      ctx.fillText(days[d.getDay()], x - 6, h - 10);
    }

    ctx.draw();
  },

  _drawLine(ctx, values, color, pad, chartW, chartH, count) {
    const maxVal = Math.max(...values) * 1.2;
    const minVal = Math.min(...values) * 0.8;
    const range = maxVal - minVal || 1;

    ctx.setStrokeStyle(color);
    ctx.setLineWidth(3);
    ctx.beginPath();
    for (let i = 0; i < count; i++) {
      const x = pad.left + (chartW / (count - 1)) * i;
      const y = pad.top + chartH - ((values[i] - minVal) / range) * chartH;
      if (i === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    }
    ctx.stroke();

    // Dots
    for (let i = 0; i < count; i++) {
      const x = pad.left + (chartW / (count - 1)) * i;
      const y = pad.top + chartH - ((values[i] - minVal) / range) * chartH;
      ctx.beginPath();
      ctx.arc(x, y, 4, 0, Math.PI * 2);
      ctx.setFillStyle(color);
      ctx.fill();
    }
  },

  selectTimeRange(e) {
    const range = e.currentTarget.dataset.range;
    this.setData({ currentRange: range });
    this._fetchHealthData();
  },

  toggleCompare() {
    this.setData({ showCompare: !this.data.showCompare });
  },

  goBack() {
    wx.navigateBack();
  },

  shareHealth() {
    wx.showShareMenu({ withShareTicket: true });
  }
});
