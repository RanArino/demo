// 設定可能パラメータ
export const LAYER_CONFIG = {
  documents: { y: 120, radius: 70, color: '#90EE90', label: 'ドキュメント' },
  clusters: { y: 0, radius: 100, color: '#DDA0DD', label: 'セクション' },
  chunks: { y: -150, radius: 120, color: '#87CEEB', label: 'チャンク' }
};

export const CAMERA_CONFIG = {
  position: [400, 250, 400] as [number, number, number],
  target: [0, -40, 0] as [number, number, number]
};

export const ROTATION_CONFIG = {
  sensitivity: 0.005,
  autoRotateSpeed: 0.1
};

export const NODE_CONFIG = {
  defaultSize: 15,
  hoverScale: 1.1,
  selectedScale: 1.3,
  // Focus view scaling
  focusScaleMultiplier: 2.5
};

export const EDGE_CONFIG = {
  default: { color: '#666666', lineWidth: 1, opacity: 0.4 },
  highlight: { color: '#000000', lineWidth: 2, opacity: 0.8 }
};

export const ANIMATION_CONFIG = {
  cameraTransitionDuration: 1500, // ms
  hoverDelay: 500 // ms
}; 