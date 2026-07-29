/**
 * Tencent Maps plugin wrapper for WeChat Mini Program.
 * Provides simple map context operations with marker/fence drawing.
 *
 * Usage: In app.json, declare the plugin:
 *   "plugins": {
 *     "weixin-map": {
 *       "version": "1.0.0",
 *       "provider": "your-tencent-plugin-id"
 *     }
 *   }
 */

// Note: Plugin must be imported in app.js before using

const _mapContext = null; // Will be set after plugin initialization

/**
 * Get or create map context for a given map component.
 * @param {string} mapId — The ID of the <map> component (e.g., "elderMap")
 * @returns {Object} Map context object
 */
export function getMapContext(mapId) {
  if (typeof wx === 'undefined') {
    throw new Error('wx is not available — must run in WeChat Mini Program environment');
  }

  // Initialize if not already done
  if (_mapContext === null) {
    throw new Error('Tencent Map plugin not initialized yet. Call initPlugin() first.');
  }

  return wx.createMapContext(mapId);
}

/**
 * Initialize the map plugin (call once in app.onLaunch).
 * @param {string} appId — Tencent Map App ID from console.tencent.com
 * @param {Function} callback — Optional callback after initialization
 */
export function initPlugin(appId, callback) {
  if (typeof Plugin !== 'undefined') {
    Plugin.import('weixin-map', {
      appId,
    }).then(() => {
      console.log('Tencent Map plugin initialized successfully');
      if (callback) callback();
    }).catch(err => {
      console.error('Failed to import Tencent Map plugin:', err);
      if (callback) callback(err);
    });
  } else {
    console.warn('Plugin API not available — using fallback mock mode');
    // In dev mode without plugin, create a mock context
    _mapContext = {
      location: () => Promise.resolve({ lat: 31.2396, lon: 121.4826 }),
      addMarker: () => {},
      setView: () => {},
    };
    if (callback) callback(null);
  }
}

/**
 * Add a marker to the map.
 * @param {Object} mapContext — Map context from getMapContext()
 * @param {number} latitude — Marker latitude
 * @param {number} longitude — Marker longitude
 * @param {string} title — Marker title/label
 * @param {Object} options — Additional options (id, width, height, iconPath, etc.)
 */
export function addMarker(mapContext, latitude, longitude, title, options = {}) {
  const markerOptions = {
    id: options.id || 'current-location',
    latitude,
    longitude,
    title: title || '当前位置',
    width: options.width || 30,
    height: options.height || 30,
    ...options.iconPath && { iconPath: options.iconPath },
  };

  mapContext.addMarker(markerOptions, () => {
    console.log('Marker added successfully');
  });
}

/**
 * Set the map view to include all markers/boundaries.
 * @param {Object} mapContext
 * @param {Array} points — Array of {latitude, longitude} objects
 */
export function setViewWithPoints(mapContext, points) {
  if (points && points.length > 0) {
    mapContext.includePoints({
      success: () => console.log('Map view updated to fit points'),
      fail: err => console.error('includePoints failed:', err),
    });
  }
}

/**
 * Draw an electronic fence (circle polygon) on the map.
 * @param {Object} mapContext
 * @param {number} centerLat — Center latitude
 * @param {number} centerLon — Center longitude
 * @param {number} radiusMeters — Fence radius in meters
 * @param {Object} options — fillColor, strokeColor, strokeWidth, etc.
 */
export function drawFenceCircle(mapContext, centerLat, centerLon, radiusMeters, options = {}) {
  // Convert radius (meters) to approximate degrees (~1 deg ≈ 111km)
  const degRadius = radiusMeters / 111000;
  const points = [];
  const segments = 32; // Number of points around the circle

  for (let i = 0; i <= segments; i++) {
    const angle = (i / segments) * 2 * Math.PI;
    points.push({
      latitude: centerLat + degRadius * Math.cos(angle),
      longitude: centerLon + degRadius * Math.sin(angle),
    });
  }

  mapContext.setPolygons({
    data: [{
      paths: [points],
      fillColor: options.fillColor || 'rgba(74, 144, 217, 0.2)',
      strokeColor: options.strokeColor || '#4A90D9',
      strokeWidth: options.strokeWidth || 2,
      opacity: options.opacity || 0.5,
    }],
    success: () => console.log('Fence drawn successfully'),
    fail: err => console.error('Failed to draw fence:', err),
  });
}

/**
 * Clear all polygons (fences) from the map.
 * @param {Object} mapContext
 */
export function clearPolygons(mapContext) {
  mapContext.clearPolygons({
    success: () => console.log('Fences cleared'),
  });
}

/**
 * Get current user location (with permission).
 * @returns {Promise<{latitude: number, longitude: number, accuracy: number}>}
 */
export function getCurrentLocation() {
  return new Promise((resolve, reject) => {
    wx.getLocation({
      type: 'gcj02', // Use GCJ-02 coordinate system (standard for China maps)
      success: (res) => {
        resolve({
          latitude: res.lat,
          longitude: res.lng,
          accuracy: res.accuracy,
        });
      },
      fail: err => {
        reject(err);
      },
    });
  });
}

/**
 * Convert address coordinates to reverse geocoding (address string).
 * @param {number} lat
 * @param {number} lng
 * @returns {Promise<string>}
 */
export function reverseGeocode(lat, lng) {
  return new Promise((resolve, reject) => {
    wx.getSystemInfo({
      success: (sysInfo) => {
        // Note: Real reverse geocode requires Tencent Map API service call
        // This is a stub — in production, call your backend geocoding service
        resolve(`纬度 ${lat.toFixed(4)}, 经度 ${lng.toFixed(4)}`);
      },
      fail: reject,
    });
  });
}

/**
 * Open system map at specified location.
 * @param {number} lat
 * @param {number} lng
 * @param {string} name — Location name
 */
export function openSystemMap(lat, lng, name) {
  wx.openLocation({
    latitude: lat,
    longitude: lng,
    name: name,
    success: () => console.log('System map opened'),
    fail: err => console.error('Failed to open system map:', err),
  });
}
