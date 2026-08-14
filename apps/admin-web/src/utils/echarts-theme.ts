/**
 * ECharts Theme Config — Materio (紫色) × Eregen (绿色) 混合
 * 适配 Eregen 项目中的 ECharts 图表配色
 *
 * 使用方式：在 init echarts 时传入此 theme 配置
 */
export const materioEChartsTheme = {
  // ── 颜色系统 ──
  color: [
    '#8C57FF', // primary (Materio 紫色)
    '#56CA00', // success (绿色)
    '#FFB400', // warning (黄色)
    '#FF4C51', // error (红色)
    '#16B1FF', // info (蓝色)
    '#8A8D93', // secondary
    '#C04A42', // danger (Eregen 红色)
    '#4A8FB8', // info (Eregen 蓝色)
  ],

  // ── 图表背景 ──
  backgroundColor: 'transparent',

  // ── 文字 ──
  textStyle: {
    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans SC", Roboto, sans-serif',
    fontSize: 13,
    fontWeight: 400,
  },

  // ── 标题 ──
  title: {
    textStyle: {
      fontSize: 16,
      fontWeight: 600,
      color: '#1A2E26',
    },
    subtextStyle: {
      fontSize: 13,
      color: '#94A9A2',
    },
  },

  // ── 图例 ──
  legend: {
    textStyle: {
      fontSize: 13,
      color: '#6B8980',
      fontWeight: 500,
    },
    icon: 'roundRect',
    itemWidth: 14,
    itemHeight: 14,
    itemGap: 16,
  },

  // ── 提示框 ──
  tooltip: {
    backgroundColor: '#FFFFFF',
    borderColor: 'rgba(26,46,38,0.12)',
    borderWidth: 1,
    borderRadius: 10,
    padding: [10, 14],
    textStyle: {
      color: '#1A2E26',
      fontSize: 13,
    },
    extraCssText: 'box-shadow: 0 4px 16px rgba(26,46,38,0.14);',
  },

  // ── 轴 ──
  axis: {
    axisLine: {
      lineStyle: {
        color: 'rgba(26,46,38,0.10)',
        width: 1,
      },
    },
    axisLabel: {
      textStyle: {
        color: '#8A8D93',
        fontSize: 12,
      },
    },
    splitLine: {
      lineStyle: {
        color: 'rgba(26,46,38,0.06)',
        type: 'solid',
      },
    },
    splitArea: {
      areaStyle: {
        color: ['rgba(244,246,244,0.5)', 'transparent'],
      },
    },
  },

  // ── 网格 ──
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true,
  },

  // ── 数据区域缩放 ──
  dataZoom: [
    {
      type: 'inside',
      start: 0,
      end: 100,
    },
    {
      start: 0,
      end: 100,
      handleSize: '80%',
      handleStyle: {
        color: '#8C57FF',
        borderColor: '#8C57FF',
      },
      textStyle: {
        color: '#8A8D93',
        fontSize: 11,
      },
    },
  ],

  // ── 标记 ──
  markLine: {
    symbol: ['none', 'none'],
    label: {
      textStyle: {
        color: '#616161',
        fontSize: 12,
      },
    },
    lineStyle: {
      type: 'dashed',
      color: '#BDBDBD',
    },
    emphasis: {
      disabled: true,
    },
  },

  // ── 视觉映射 ──
  visualMap: {
    textStyle: {
      color: '#616161',
      fontSize: 12,
    },
    outOfRange: {
      color: '#8A8D93',
    },
  },

  // ── 时间轴 ──
  timeline: {
    lineStyle: {
      color: '#8C57FF',
      borderWidth: 1,
    },
    controlStyle: {
      color: '#8C57FF',
      borderColor: '#8C57FF',
      borderWidth: 1,
    },
    checkpointStyle: {
      color: '#8C57FF',
      borderColor: '#FFFFFF',
      borderWidth: 2,
    },
    label: {
      textStyle: {
        color: '#616161',
        fontSize: 12,
      },
    },
  },

  // ── 图表类型默认配置 ──
  line: {
    symbol: 'circle',
    symbolSize: 6,
    lineStyle: {
      width: 2.5,
      type: 'solid',
    },
    itemStyle: {
      borderWidth: 2,
    },
    areaStyle: {},
    emphasis: {
      focus: 'series',
    },
  },
  bar: {
    categoryAxis: {
      axisLine: {
        lineStyle: {
          color: 'rgba(26,46,38,0.10)',
        },
      },
      splitLine: {
        show: false,
      },
    },
    valueAxis: {
      axisLine: {
        show: false,
      },
      splitLine: {
        lineStyle: {
          color: 'rgba(26,46,38,0.06)',
        },
      },
    },
    emphasis: {
      focus: 'series',
    },
  },
  pie: {
    roseType: false,
    avoidLabelOverlap: true,
    itemStyle: {
      borderRadius: 6,
      borderColor: '#FFFFFF',
      borderWidth: 2,
    },
    label: {
      formatter: '{b}: {c}\n({d}%)',
      fontSize: 12,
      color: '#616161',
    },
    emphasis: {
      label: {
        fontSize: 13,
        fontWeight: 600,
      },
      itemStyle: {
        shadowBlur: 10,
        shadowOffsetX: 0,
        shadowColor: 'rgba(0, 0, 0, 0.2)',
      },
    },
  },
  scatter: {
    symbolSize: 8,
  },
  radar: {
    symbol: 'circle',
    symbolSize: 4,
  },
  map: {
    label: {
      show: false,
      fontSize: 10,
      color: '#616161',
    },
    itemStyle: {
      borderColor: 'rgba(26,46,38,0.2)',
      borderWidth: 1,
      areaColor: '#E8F4EC',
    },
    emphasis: {
      label: {
        show: true,
        color: '#1A2E26',
        fontSize: 12,
        fontWeight: 600,
      },
      itemStyle: {
        areaColor: '#D5E6DA',
      },
    },
  },
  force: {
    symbol: 'circle',
    symbolSize: 6,
    edgeStyle: {
      color: 'rgba(26,46,38,0.2)',
      width: 1,
      type: 'solid',
    },
  },
  chord: {
    padding: 4,
    itemStyle: {
      borderWidth: 1,
      borderColor: 'rgba(26,46,38,0.1)',
    },
    lineStyle: {
      borderWidth: 1,
      borderColor: 'rgba(26,46,38,0.1)',
    },
  },
  arc: {
    label: {
      show: true,
      position: 'outside',
    },
  },
}

/**
 * 快速初始化工具函数
 */
export function initMaterioChart(dom: HTMLElement | string, option: any) {
  const chart = (window as any).echarts?.init(dom)
  if (chart) {
    chart.setOption(option, true)
  }
  return chart
}

/**
 * Eregen 绿色版 ECharts 主题（与 Materio 紫色互补）
 */
export const eregenGreenEChartsTheme = {
  color: [
    '#4A7C5F', // Eregen primary 绿色
    '#8C57FF', // Materio accent 紫色（点缀）
    '#56CA00', // success
    '#FFB400', // warning
    '#C04A42', // error
    '#4A8FB8', // info
    '#6FAF8F', // primary light
    '#D5E6DA', // primary lighter
  ],
}
