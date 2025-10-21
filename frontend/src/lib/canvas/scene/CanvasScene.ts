import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { CSS2DRenderer, CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
import { EffectComposer } from 'three/examples/jsm/postprocessing/EffectComposer.js';
import { RenderPass } from 'three/examples/jsm/postprocessing/RenderPass.js';
import { UnrealBloomPass } from 'three/examples/jsm/postprocessing/UnrealBloomPass.js';
import type { CanvasRenderableNode, CanvasContentNode, CanvasChunkNode, CanvasNodeKind } from '../types';
import { generateFibonacciSpherePositions } from '../positioning/fibonacciSphere';
import { generateFibonacciDiskPositions } from '../positioning/fibonacciDisk';

type OnNodeSelect = (node: CanvasRenderableNode | null) => void;

interface CanvasSceneOptions {
  container: HTMLDivElement;
  onNodeSelect?: OnNodeSelect;
  onFpsUpdate?: (fps: number) => void;
}

interface MeshWithData extends THREE.Mesh {
  userData: {
    node: CanvasRenderableNode;
  };
}

type MultiViewType = 'top' | 'side' | 'iso' | 'front';

interface MultiViewConfig {
  id: string;
  type: MultiViewType;
  container: HTMLDivElement | null;
}

interface MultiViewEntry {
  config: MultiViewConfig;
  renderer: THREE.WebGLRenderer;
  camera: THREE.PerspectiveCamera;
  helper: THREE.CameraHelper;
  width: number;
  height: number;
}

export class CanvasScene {
  private static readonly BLOOM_LAYER = 1;
  private static readonly CURSOR_TRAIL_LIFETIME = 0.9;

  private readonly container: HTMLDivElement;
  private readonly scene: THREE.Scene;
  private readonly camera: THREE.PerspectiveCamera;
  private readonly renderer: THREE.WebGLRenderer;
  private readonly controls: OrbitControls;
  private readonly raycaster: THREE.Raycaster;
  private readonly pointer: THREE.Vector2;
  private readonly nodeGroup: THREE.Group;
  private readonly gridHelper: THREE.GridHelper;
  private readonly ambientLight: THREE.AmbientLight;
  private readonly directionalLight: THREE.DirectionalLight;
  private readonly labelRenderer: CSS2DRenderer;
  private readonly labels: CSS2DObject[] = [];
  private readonly bloomLayer = new THREE.Layers();
  private readonly darkMaterial = new THREE.MeshBasicMaterial({ color: 0x000000 });
  private readonly storedMaterials = new Map<string, THREE.Material | THREE.Material[]>();
  private readonly composer: EffectComposer;
  private readonly bloomPass: UnrealBloomPass;
  private readonly clock = new THREE.Clock();
  private readonly trailGeometry = new THREE.SphereGeometry(0.6, 10, 10);
  private readonly edgePanVector = new THREE.Vector2();
  private readonly nodeMeshes = new Map<string, MeshWithData>();
  private readonly originalPositions = new Map<string, THREE.Vector3>();
  private hiddenKinds = new Set<CanvasNodeKind>();
  private layeredView = false;
  private readonly positionAnimations = new Map<string, { start: THREE.Vector3; end: THREE.Vector3; elapsed: number; duration: number }>();
  private readonly scaleAnimations = new Map<string, { start: number; end: number; elapsed: number; duration: number }>();
  private animationFrame?: number;
  private onNodeSelect?: OnNodeSelect;
  private currentNodes: CanvasRenderableNode[] = [];
  private isDragging = false;
  private lastPointerDown = 0;
  private minimapContainer?: HTMLDivElement;
  private minimapRenderer?: THREE.WebGLRenderer;
  private minimapCamera?: THREE.OrthographicCamera;
  private minimapHelper?: THREE.CameraHelper;
  private readonly multiViewEntries = new Map<string, MultiViewEntry>();
  private readonly multiViewHelpers: THREE.CameraHelper[] = [];
  private selectedNodeId?: string;
  private hoveredNodeId?: string;
  private hoverLinkLines?: THREE.LineSegments;
  private readonly contentToChunkMeshes = new Map<string, MeshWithData[]>();
  private readonly nodeContentSourceIndex = new Map<string, string>();
  private readonly chunkFallbackAnchors = new Map<string, THREE.Vector3>();
  private onFpsUpdate?: (fps: number) => void;
  private cursorLight: THREE.PointLight;
  private cursorTrail: Array<{ mesh: THREE.Mesh; life: number }> = [];
  private lastTrailSpawn = 0;
  private pendingZoomIn?: number;
  private edgePanThreshold = 0.12;
  private edgePanSpeed = 120;
  private zoomInFactor = 0.7;
  private zoomOutFactor = 1.3;
  private zoomBlendIn = 0.85;
  private zoomBlendOut = 0.35;
  private smoothedFps = 0;
  private fpsAccumulator = 0;
  private readonly frustum = new THREE.Frustum();
  private readonly viewProjectionMatrix = new THREE.Matrix4();
  private chunkVisibilityDistance = 260;
  // Advanced link visualization state
  private showClusterEdges = false;
  private showSemanticEdges = false;
  private semanticSimilarityThreshold = 0.45;
  private clusterEdgeLines?: THREE.LineSegments;
  private semanticEdgeLines?: THREE.LineSegments;
  private semanticEdgeMeta: Array<{ aId: string; bId: string; similarity: number }> = [];
  private linkTooltip?: CSS2DObject;
  private readonly nodeSizeDefaults: Record<CanvasNodeKind, number> = {
    cluster: 20,
    content: 15,
    chunk: 4,
  };
  private readonly fallbackColors: Record<CanvasNodeKind, string> = {
    cluster: '#38bdf8',
    content: '#f97316',
    chunk: '#a3e635',
  };
  private isInteracting = false;
  private interactionEndTimeout?: number;
  private readonly interactionCooldown = 140;
  private readonly minChunkVisibilityDistance = 160;
  private readonly maxChunkVisibilityDistance = 340;
  private cursorTrailEnabled = false;
  private readonly cursorTrailInterval = 80;
  private readonly cursorTrailMaxEntries = 80;
  private readonly labelHideDistances: Record<CanvasNodeKind, number> = {
    cluster: 520,
    content: 380,
    chunk: Infinity,
  };
  private readonly tempVector = new THREE.Vector3();
  private hasBloomTargets = false;
  private hoverCard?: CSS2DObject;
  private hoverCardNodeId?: string;

  constructor(options: CanvasSceneOptions) {
    const { container, onNodeSelect, onFpsUpdate } = options;
    this.container = container;
    this.onNodeSelect = onNodeSelect;
    this.onFpsUpdate = onFpsUpdate;
    this.bloomLayer.set(CanvasScene.BLOOM_LAYER);

    const width = container.clientWidth || 1;
    const height = container.clientHeight || 1;

    this.scene = new THREE.Scene();
    this.scene.background = new THREE.Color(0x0f172a);

    this.camera = new THREE.PerspectiveCamera(60, width / height, 0.1, 2000);
    this.camera.position.set(0, 80, 160);

    this.renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    this.renderer.setPixelRatio(window.devicePixelRatio);
    this.renderer.setSize(width, height);
    this.renderer.outputColorSpace = THREE.SRGBColorSpace;
    this.renderer.autoClear = false;
    container.appendChild(this.renderer.domElement);

    this.labelRenderer = new CSS2DRenderer();
    this.labelRenderer.setSize(width, height);
    this.labelRenderer.domElement.style.position = 'absolute';
    this.labelRenderer.domElement.style.top = '0px';
    this.labelRenderer.domElement.style.pointerEvents = 'none';
    this.labelRenderer.domElement.style.zIndex = '1';
    container.appendChild(this.labelRenderer.domElement);

    this.controls = new OrbitControls(this.camera, this.renderer.domElement);
    this.controls.enableDamping = true;
    this.controls.dampingFactor = 0.045;
    this.controls.enablePan = true;
    this.controls.panSpeed = 0.9;
    this.controls.minDistance = 10;
    this.controls.maxDistance = 600;
    this.controls.maxPolarAngle = Math.PI - Math.PI / 12;
    this.controls.addEventListener('start', this.handleControlsStart);
    this.controls.addEventListener('end', this.handleControlsEnd);

    this.raycaster = new THREE.Raycaster();
    this.pointer = new THREE.Vector2();
    this.nodeGroup = new THREE.Group();
    this.scene.add(this.nodeGroup);

    this.gridHelper = new THREE.GridHelper(400, 40, 0x334155, 0x1f2937);
    this.gridHelper.position.y = -20;
    this.scene.add(this.gridHelper);

    this.ambientLight = new THREE.AmbientLight(0xffffff, 0.85);
    this.directionalLight = new THREE.DirectionalLight(0xffffff, 0.6);
    this.directionalLight.position.set(120, 180, 100);
    this.scene.add(this.ambientLight);
    this.scene.add(this.directionalLight);
    this.cursorLight = new THREE.PointLight(0x8ec5ff, 1.6, 220, 2);
    this.cursorLight.position.set(0, 40, 40);
    this.scene.add(this.cursorLight);

    const renderScene = new RenderPass(this.scene, this.camera);
    renderScene.clear = false;
    this.bloomPass = new UnrealBloomPass(
      new THREE.Vector2(width, height),
      1.2,
      0.4,
      0.0
    );
    this.bloomPass.threshold = 0;
    this.bloomPass.strength = 1.15;
    this.bloomPass.radius = 0.55;

    this.composer = new EffectComposer(this.renderer);
    this.composer.setSize(width, height);
    this.composer.addPass(renderScene);
    this.composer.addPass(this.bloomPass);

    this.bindEvents();
    this.animate();
  }

  // Accessibility helpers
  panByDirection(dir: 'left' | 'right' | 'up' | 'down'): void {
    const offset = new THREE.Vector3().subVectors(this.camera.position, this.controls.target);
    const distance = Math.max(20, offset.length());
    const step = Math.max(10, distance * 0.06);
    const dx = dir === 'left' ? -step : dir === 'right' ? step : 0;
    const dy = dir === 'up' ? step : dir === 'down' ? -step : 0;
    this.controls.pan(dx, dy);
  }

  zoomInCenter(): void {
    const rect = this.renderer.domElement.getBoundingClientRect();
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    this.zoomAtScreenPoint(cx, cy, 'in');
  }

  zoomOutCenter(): void {
    const rect = this.renderer.domElement.getBoundingClientRect();
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    this.zoomAtScreenPoint(cx, cy, 'out');
  }

  focusSelected(): void {
    if (!this.selectedNodeId) {
      return;
    }
    const mesh = this.nodeMeshes.get(this.selectedNodeId);
    if (!mesh) {
      return;
    }
    const offset = new THREE.Vector3().subVectors(this.camera.position, this.controls.target);
    const distance = Math.max(40, offset.length());
    const dir = offset.normalize();
    const newTarget = mesh.position.clone();
    const newPos = new THREE.Vector3().addVectors(newTarget, dir.multiplyScalar(distance));
    this.controls.target.copy(newTarget);
    this.camera.position.copy(newPos);
    this.controls.update();
  }

  clearSelection(): void {
    this.setSelection(null);
    this.onNodeSelect?.(null);
  }

  selectNextVisible(delta: 1 | -1 = 1): void {
    const visibleIds: string[] = [];
    this.nodeMeshes.forEach((mesh) => {
      if (mesh.visible) {
        visibleIds.push(mesh.userData.node.id);
      }
    });
    if (visibleIds.length === 0) {
      return;
    }
    visibleIds.sort();
    const currentIndex = this.selectedNodeId ? visibleIds.indexOf(this.selectedNodeId) : -1;
    const nextIndex = currentIndex < 0 ? 0 : (currentIndex + (delta === 1 ? 1 : -1) + visibleIds.length) % visibleIds.length;
    const nextId = visibleIds[nextIndex];
    const mesh = this.nodeMeshes.get(nextId) || null;
    if (mesh) {
      this.setSelection(mesh);
      this.onNodeSelect?.(mesh.userData.node);
    }
  }

  setMultiViewConfigs(configs: MultiViewConfig[]): void {
    const nextById = new Map<string, MultiViewConfig>();
    configs.forEach((config) => {
      nextById.set(config.id, config);
    });

    this.multiViewEntries.forEach((entry, id) => {
      const nextConfig = nextById.get(id);
      if (!nextConfig || !nextConfig.container) {
        this.disposeMultiViewEntry(entry);
        this.multiViewEntries.delete(id);
      }
    });

    configs.forEach((config) => {
      if (!config.container) {
        return;
      }

      let entry = this.multiViewEntries.get(config.id);
      if (!entry) {
        entry = this.createMultiViewEntry(config);
        this.multiViewEntries.set(config.id, entry);
      } else {
        entry.config = config;
        if (entry.renderer.domElement.parentElement !== config.container) {
          entry.renderer.domElement.remove();
          config.container.appendChild(entry.renderer.domElement);
        }
      }

      this.updateMultiViewEntryDimensions(entry);
      this.positionMultiViewCamera(entry);
    });
  }

  jumpToMultiView(type: MultiViewType): void {
    const offset = new THREE.Vector3().subVectors(this.camera.position, this.controls.target);
    const distance = offset.length() || 120;
    const target = this.controls.target.clone();

    this.controls.target.copy(target);
    switch (type) {
      case 'top':
        this.camera.position.set(target.x, target.y + distance, target.z + 0.01 * distance);
        this.camera.up.set(0, 0, -1);
        break;
      case 'side': {
        const height = distance * 0.25;
        this.camera.position.set(target.x + distance, target.y + height, target.z);
        this.camera.up.set(0, 1, 0);
        break;
      }
      case 'front':
        this.camera.position.set(target.x, target.y + distance * 0.1, target.z + distance);
        this.camera.up.set(0, 1, 0);
        break;
      case 'iso': {
        const diag = distance / Math.sqrt(3);
        this.camera.position.set(target.x + diag, target.y + diag, target.z + diag);
        this.camera.up.set(0, 1, 0);
        break;
      }
    }

    this.camera.lookAt(this.controls.target);
    this.controls.update();
    this.edgePanVector.set(0, 0);
    this.updateMultiViewEntries();
  }

  private createMultiViewEntry(config: MultiViewConfig): MultiViewEntry {
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setPixelRatio(window.devicePixelRatio);
    renderer.setClearColor(0x000000, 0);
    renderer.domElement.style.pointerEvents = 'none';

    const camera = new THREE.PerspectiveCamera(60, 1, 0.1, 2000);
    camera.position.copy(this.camera.position);
    camera.lookAt(this.controls.target);

    const helper = new THREE.CameraHelper(camera);
    helper.visible = false;
    const helperMaterial = helper.material as THREE.LineBasicMaterial;
    switch (config.type) {
      case 'top':
        helperMaterial.color.set(0x38bdf8);
        break;
      case 'side':
        helperMaterial.color.set(0xf97316);
        break;
      case 'iso':
        helperMaterial.color.set(0x6366f1);
        break;
      case 'front':
        helperMaterial.color.set(0x22d3ee);
        break;
    }
    this.scene.add(helper);
    this.multiViewHelpers.push(helper);

    if (config.container) {
      config.container.appendChild(renderer.domElement);
    }

    const entry: MultiViewEntry = {
      config,
      renderer,
      camera,
      helper,
      width: 0,
      height: 0,
    };

    this.updateMultiViewEntryDimensions(entry);
    this.positionMultiViewCamera(entry);

    return entry;
  }

  private disposeMultiViewEntry(entry: MultiViewEntry): void {
    entry.renderer.dispose();
    entry.renderer.domElement.remove();

    const helperIndex = this.multiViewHelpers.indexOf(entry.helper);
    if (helperIndex >= 0) {
      this.multiViewHelpers.splice(helperIndex, 1);
    }

    this.scene.remove(entry.helper);
    entry.helper.geometry.dispose();
    const helperMaterial = entry.helper.material;
    if (Array.isArray(helperMaterial)) {
      helperMaterial.forEach((material) => material.dispose());
    } else {
      helperMaterial.dispose();
    }
  }

  private clearMultiViewEntries(): void {
    this.multiViewEntries.forEach((entry) => this.disposeMultiViewEntry(entry));
    this.multiViewEntries.clear();
  }

  private updateMultiViewEntryDimensions(entry: MultiViewEntry): void {
    const container = entry.config.container;
    if (!container) {
      return;
    }

    const width = Math.max(container.clientWidth, 1);
    const height = Math.max(container.clientHeight, 1);

    if (width !== entry.width || height !== entry.height) {
      entry.renderer.setSize(width, height);
      entry.camera.aspect = width / height;
      entry.camera.updateProjectionMatrix();
      entry.width = width;
      entry.height = height;
    }
  }

  private positionMultiViewCamera(entry: MultiViewEntry): void {
    const target = this.controls.target;
    const offset = new THREE.Vector3().subVectors(this.camera.position, target);
    const distance = Math.max(offset.length(), 60);
    switch (entry.config.type) {
      case 'top':
        entry.camera.position.set(target.x, target.y + distance, target.z + 0.01 * distance);
        entry.camera.up.set(0, 0, -1);
        break;
      case 'side': {
        const height = distance * 0.25;
        entry.camera.position.set(target.x + distance, target.y + height, target.z);
        entry.camera.up.set(0, 1, 0);
        break;
      }
      case 'front':
        entry.camera.position.set(target.x, target.y + distance * 0.1, target.z + distance);
        entry.camera.up.set(0, 1, 0);
        break;
      case 'iso': {
        const diag = distance / Math.sqrt(3);
        entry.camera.position.set(target.x + diag, target.y + diag, target.z + diag);
        entry.camera.up.set(0, 1, 0);
        break;
      }
    }
    entry.camera.lookAt(target);
    entry.helper.update();
  }

  private updateMultiViewEntries(): void {
    if (this.multiViewEntries.size === 0) {
      return;
    }

    this.multiViewEntries.forEach((entry) => {
      this.updateMultiViewEntryDimensions(entry);
      this.positionMultiViewCamera(entry);
      entry.renderer.clear();
      entry.renderer.render(this.scene, entry.camera);
    });
  }

  updateSize(): void {
    const width = this.container.clientWidth || 1;
    const height = this.container.clientHeight || 1;
    this.camera.aspect = width / height;
    this.camera.updateProjectionMatrix();
    this.renderer.setSize(width, height);
    this.labelRenderer.setSize(width, height);
    this.composer.setSize(width, height);
    this.bloomPass.setSize(width, height);
    this.updateMinimapSize();
  }

  updateNodes(nodes: CanvasRenderableNode[]): void {
    const previousPositions = new Map<string, THREE.Vector3>();
    this.nodeMeshes.forEach((mesh, id) => {
      previousPositions.set(id, mesh.position.clone());
    });
    this.currentNodes = nodes;
    this.clearLabels();
    this.clearHoverLinks();
    this.hideHoverCard();
    this.disposeNodeMeshes();

    const contentPositionOverrides = this.resolveContentPositionOverrides(nodes);
    const chunkLayout = this.resolveChunkLayout(nodes, contentPositionOverrides);
    this.contentToChunkMeshes.clear();
    this.nodeContentSourceIndex.clear();

    nodes.forEach((node) => {
      if (!node.visibility) {
        return;
      }

      const overridePosition =
        node.kind === 'chunk'
          ? chunkLayout.positions.get(node.id)
          : node.kind === 'content'
            ? contentPositionOverrides.get(node.id)
            : undefined;
      const effectiveSize = this.determineNodeSize(node, chunkLayout.sizes);
      const mesh = this.createMeshForNode(node, { position: overridePosition, size: effectiveSize });
      const labelText = this.getNodeLabelText(node);
      const label = this.createLabelForNode(node, labelText, effectiveSize);
      if (label) {
        mesh.add(label);
        this.labels.push(label);
      }
      mesh.layers.disable(CanvasScene.BLOOM_LAYER);
      const targetPosition = mesh.position.clone();
      mesh.scale.setScalar(1);
      this.nodeGroup.add(mesh);
      this.nodeMeshes.set(node.id, mesh);
      this.originalPositions.set(node.id, targetPosition.clone());

      const previous = previousPositions.get(node.id);
      if (previous) {
        mesh.position.copy(previous);
        this.positionAnimations.set(node.id, {
          start: previous.clone(),
          end: targetPosition,
          elapsed: 0,
          duration: 1,
        });
      } else {
        mesh.position.copy(targetPosition);
        const initialScale = 0.001;
        mesh.scale.setScalar(initialScale);
        this.scaleAnimations.set(node.id, {
          start: initialScale,
          end: 1,
          elapsed: 0,
          duration: 1,
        });
      }

      if (node.kind === 'content') {
        const sourceKey = this.normalizeContentSourceId(node.contentSourceId);
        if (sourceKey) {
          this.nodeContentSourceIndex.set(node.id, sourceKey);
          if (!this.contentToChunkMeshes.has(sourceKey)) {
            this.contentToChunkMeshes.set(sourceKey, []);
          }
        }
      } else if (node.kind === 'chunk') {
        const sourceKey = this.normalizeContentSourceId(node.contentSourceId);
        if (sourceKey) {
          this.nodeContentSourceIndex.set(node.id, sourceKey);
          const list = this.contentToChunkMeshes.get(sourceKey) ?? [];
          list.push(mesh);
          this.contentToChunkMeshes.set(sourceKey, list);
          (mesh.userData as MeshWithData['userData'] & { parentContentSourceId?: string }).parentContentSourceId = sourceKey;
        }
      }
    });

    this.applySelectionHighlight();
    this.applyLayerVisibility();
    this.applyLayerLayout();
    this.updateBloomFlag();
    this.updateMultiViewEntries();
    const hoveredMesh = this.hoveredNodeId ? this.nodeMeshes.get(this.hoveredNodeId) : undefined;
    this.updateHoverLinks(hoveredMesh);
    this.rebuildAdvancedEdges();
  }

  toggleGrid(visible: boolean): void {
    this.gridHelper.visible = visible;
  }

  setLayerVisibility(kind: CanvasNodeKind, visible: boolean): void {
    if (visible) {
      this.hiddenKinds.delete(kind);
    } else {
      this.hiddenKinds.add(kind);
    }
    this.applyLayerVisibility();
    this.applyLayerLayout();
    this.updateMultiViewEntries();
    this.rebuildAdvancedEdges();
  }

  setLayeredView(enabled: boolean): void {
    if (this.layeredView === enabled) {
      return;
    }
    this.layeredView = enabled;
    this.applyLayerLayout({ animate: true, duration: 0.9 });
    this.updateMultiViewEntries();
    this.rebuildAdvancedEdges();
  }

  setEdgePanConfig(config: { speed?: number; threshold?: number }): void {
    if (typeof config.speed === 'number' && Number.isFinite(config.speed)) {
      this.edgePanSpeed = config.speed;
    }
    if (typeof config.threshold === 'number' && Number.isFinite(config.threshold)) {
      this.edgePanThreshold = THREE.MathUtils.clamp(config.threshold, 0.02, 0.3);
    }
  }

  setZoomAggressiveness(intensity: number): void {
    if (!Number.isFinite(intensity)) {
      return;
    }
    const clamped = THREE.MathUtils.clamp(intensity, 0.1, 0.6);
    this.zoomInFactor = 1 - clamped;
    this.zoomOutFactor = 1 + clamped;
    this.zoomBlendIn = THREE.MathUtils.clamp(0.5 + (0.6 - clamped), 0.4, 0.9);
    this.zoomBlendOut = THREE.MathUtils.clamp(0.25 + clamped * 0.3, 0.2, 0.6);
  }

  private applyLayerVisibility(): void {
    const hidden = this.hiddenKinds;
    this.nodeMeshes.forEach((mesh) => {
      const node = mesh.userData.node;
      const visible = !hidden.has(node.kind);
      mesh.visible = visible;
      mesh.children.forEach((child) => {
        child.visible = visible;
      });
    });

    if (this.selectedNodeId) {
      const selectedMesh = this.nodeMeshes.get(this.selectedNodeId);
      if (!selectedMesh || !selectedMesh.visible) {
        this.setSelection(null);
        this.onNodeSelect?.(null);
      }
    }

    if (this.hoveredNodeId) {
      const hoveredMesh = this.nodeMeshes.get(this.hoveredNodeId);
      if (!hoveredMesh || !hoveredMesh.visible) {
        this.hoveredNodeId = undefined;
        this.clearHoverLinks();
      }
    }

    this.updateAdvancedEdgesVisibility();
    if (this.hoverCardNodeId) {
      const targetMesh = this.nodeMeshes.get(this.hoverCardNodeId);
      if (!targetMesh || !targetMesh.visible) {
        this.hideHoverCard();
      }
    }
  }

  private applyLayerLayout(options?: { animate?: boolean; duration?: number }): void {
    const animate = options?.animate ?? false;
    const duration = options?.duration ?? 1;

    if (!this.layeredView) {
      const canSnap = animate || this.positionAnimations.size === 0;
      if (!canSnap) {
        return;
      }
      this.nodeMeshes.forEach((mesh, id) => {
        const original = this.originalPositions.get(id);
        if (!original) {
          return;
        }
        if (animate) {
          const current = mesh.position.clone();
          if (current.distanceToSquared(original) < 0.0001) {
            mesh.position.copy(original);
            this.positionAnimations.delete(id);
            return;
          }
          this.positionAnimations.set(id, {
            start: current,
            end: original.clone(),
            elapsed: 0,
            duration,
          });
        } else {
          mesh.position.copy(original);
          this.positionAnimations.delete(id);
        }
      });
      return;
    }

    const targetPositions = this.resolveLayeredTargets();

    this.nodeMeshes.forEach((mesh, id) => {
      const target = targetPositions.get(id);
      if (!target) {
        return;
      }
      if (animate) {
        const current = mesh.position.clone();
        if (current.distanceToSquared(target) < 0.0001) {
          mesh.position.copy(target);
          this.positionAnimations.delete(id);
          return;
        }
        this.positionAnimations.set(id, {
          start: current,
          end: target.clone(),
          elapsed: 0,
          duration,
        });
      } else {
        mesh.position.copy(target);
        this.positionAnimations.delete(id);
      }
    });
  }

  private resolveLayeredTargets(): Map<string, THREE.Vector3> {
    const clusters: MeshWithData[] = [];
    const contents: MeshWithData[] = [];
    const chunksPrimary: MeshWithData[] = [];
    const chunksSecondary: MeshWithData[] = [];

    this.nodeMeshes.forEach((mesh) => {
      const node = mesh.userData.node;
      if (this.hiddenKinds.has(node.kind)) {
        return;
      }
      if (node.kind === 'cluster') {
        clusters.push(mesh);
      } else if (node.kind === 'content') {
        contents.push(mesh);
      } else if (node.contextType === 'chunk') {
        chunksSecondary.push(mesh);
      } else {
        chunksPrimary.push(mesh);
      }
    });

    const clusterBase = 220;
    const clusterSpacing = 25;
    const contentLayer = 80;
    const chunkPrimaryLayer = -20;
    const chunkSecondaryLayer = -70;

    const sortByOriginalX = (a: MeshWithData, b: MeshWithData) => {
      const aOrig = this.originalPositions.get(a.userData.node.id);
      const bOrig = this.originalPositions.get(b.userData.node.id);
      if (!aOrig || !bOrig) {
        return 0;
      }
      return aOrig.x - bOrig.x;
    };

    const targets = new Map<string, THREE.Vector3>();

    clusters
      .sort(sortByOriginalX)
      .forEach((mesh, index) => {
        const nodeId = mesh.userData.node.id;
        const orig = this.originalPositions.get(nodeId);
        if (!orig) {
          return;
        }
        targets.set(nodeId, new THREE.Vector3(orig.x, clusterBase + index * clusterSpacing, orig.z));
      });

    contents.forEach((mesh) => {
      const nodeId = mesh.userData.node.id;
      const orig = this.originalPositions.get(nodeId);
      if (!orig) {
        return;
      }
      targets.set(nodeId, new THREE.Vector3(orig.x, contentLayer, orig.z));
    });

    chunksPrimary.forEach((mesh) => {
      const nodeId = mesh.userData.node.id;
      const orig = this.originalPositions.get(nodeId);
      if (!orig) {
        return;
      }
      targets.set(nodeId, new THREE.Vector3(orig.x, chunkPrimaryLayer, orig.z));
    });

    chunksSecondary.forEach((mesh, index) => {
      const nodeId = mesh.userData.node.id;
      const orig = this.originalPositions.get(nodeId);
      if (!orig) {
        return;
      }
      const offset = index % 2 === 0 ? 10 : -10;
      targets.set(nodeId, new THREE.Vector3(orig.x + offset, chunkSecondaryLayer, orig.z));
    });

    return targets;
  }

  private updateHoverLinks(mesh?: MeshWithData): void {
    this.clearHoverLinks();
    if (this.isInteracting) {
      return;
    }
    if (!mesh) {
      return;
    }

    const node = mesh.userData.node;
    if (node.kind !== 'content') {
      return;
    }

    const sourceId = this.nodeContentSourceIndex.get(node.id);
    if (!sourceId) {
      return;
    }

    const chunkMeshes = this.contentToChunkMeshes.get(sourceId);
    if (!chunkMeshes || chunkMeshes.length === 0) {
      return;
    }

    const visibleChunks = chunkMeshes.filter((chunkMesh) => chunkMesh.visible);
    if (visibleChunks.length === 0) {
      return;
    }

    const positions = new Float32Array(visibleChunks.length * 6);
    const start = mesh.position.clone();
    visibleChunks.forEach((chunkMesh, index) => {
      const end = chunkMesh.position.clone();
      positions.set([start.x, start.y, start.z, end.x, end.y, end.z], index * 6);
    });

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
    const material = new THREE.LineBasicMaterial({ color: 0x60a5fa, transparent: true, opacity: 0.6 });
    this.hoverLinkLines = new THREE.LineSegments(geometry, material);
    this.scene.add(this.hoverLinkLines);
  }

  private clearHoverLinks(): void {
    if (!this.hoverLinkLines) {
      return;
    }
    this.scene.remove(this.hoverLinkLines);
    this.hoverLinkLines.geometry.dispose();
    const material = this.hoverLinkLines.material;
    if (Array.isArray(material)) {
      material.forEach((mat) => mat.dispose());
    } else {
      material.dispose();
    }
    this.hoverLinkLines = undefined;
  }

  // Advanced edges: configuration API
  setLinksConfig(config: { clusterEdges?: boolean; semanticEdges?: boolean; similarityThreshold?: number }): void {
    if (typeof config.clusterEdges === 'boolean') this.showClusterEdges = config.clusterEdges;
    if (typeof config.semanticEdges === 'boolean') this.showSemanticEdges = config.semanticEdges;
    if (typeof config.similarityThreshold === 'number' && Number.isFinite(config.similarityThreshold)) {
      this.semanticSimilarityThreshold = THREE.MathUtils.clamp(config.similarityThreshold, 0.1, 0.9);
    }
    this.rebuildAdvancedEdges();
  }

  private clearAdvancedEdges(): void {
    if (this.clusterEdgeLines) {
      this.scene.remove(this.clusterEdgeLines);
      this.clusterEdgeLines.geometry.dispose();
      const mat = this.clusterEdgeLines.material as THREE.Material | THREE.Material[];
      if (Array.isArray(mat)) mat.forEach((m) => m.dispose()); else mat.dispose();
      this.clusterEdgeLines = undefined;
    }
    if (this.semanticEdgeLines) {
      this.scene.remove(this.semanticEdgeLines);
      this.semanticEdgeLines.geometry.dispose();
      const mat = this.semanticEdgeLines.material as THREE.Material | THREE.Material[];
      if (Array.isArray(mat)) mat.forEach((m) => m.dispose()); else mat.dispose();
      this.semanticEdgeLines = undefined;
    }
    this.semanticEdgeMeta = [];
    this.hideLinkTooltip();
  }

  private updateAdvancedEdgesVisibility(): void {
    if (this.clusterEdgeLines) this.clusterEdgeLines.visible = this.showClusterEdges;
    if (this.semanticEdgeLines) this.semanticEdgeLines.visible = this.showSemanticEdges;
  }

  private rebuildAdvancedEdges(): void {
    this.clearAdvancedEdges();
    if (this.nodeMeshes.size === 0) {
      return;
    }
    if (this.showClusterEdges) {
      this.clusterEdgeLines = this.buildClusterEdges();
      if (this.clusterEdgeLines) this.scene.add(this.clusterEdgeLines);
    }
    if (this.showSemanticEdges) {
      const built = this.buildSemanticEdges();
      this.semanticEdgeLines = built?.lines;
      this.semanticEdgeMeta = built?.meta ?? [];
      if (this.semanticEdgeLines) this.scene.add(this.semanticEdgeLines);
    }
  }

  private buildClusterEdges(): THREE.LineSegments | undefined {
    const clusters: MeshWithData[] = [];
    const contents: MeshWithData[] = [];
    this.nodeMeshes.forEach((mesh) => {
      if (!mesh.visible) return;
      const kind = mesh.userData.node.kind;
      if (kind === 'cluster') clusters.push(mesh);
      if (kind === 'content') contents.push(mesh);
    });
    if (clusters.length === 0 || contents.length === 0) return undefined;

    const segments: number[] = [];
    for (const c of clusters) {
      const cSize = c.userData.node.display.size ?? 20;
      const radius = cSize * 1.3;
      for (const n of contents) {
        if (!n.visible) continue;
        const dist = c.position.distanceTo(n.position);
        if (dist <= radius * 1.6) {
          segments.push(c.position.x, c.position.y, c.position.z, n.position.x, n.position.y, n.position.z);
        }
      }
    }
    if (segments.length === 0) return undefined;

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(new Float32Array(segments), 3));
    const material = new THREE.LineBasicMaterial({ color: 0x94a3b8, transparent: true, opacity: 0.35 });
    return new THREE.LineSegments(geometry, material);
  }

  private buildSemanticEdges(): { lines: THREE.LineSegments; meta: Array<{ aId: string; bId: string; similarity: number }> } | undefined {
    const contents: MeshWithData[] = [];
    this.nodeMeshes.forEach((mesh) => {
      if (mesh.visible && mesh.userData.node.kind === 'content') contents.push(mesh);
    });
    const n = contents.length;
    if (n < 2) return undefined;
    const threshold = this.semanticSimilarityThreshold;
    const positions: number[] = [];
    const meta: Array<{ aId: string; bId: string; similarity: number }> = [];

    // Build top-K neighbors per node to limit overdraw
    const K = 3;
    for (let i = 0; i < n; i++) {
      const a = contents[i];
      const aKw = new Set(a.userData.node.keywords || []);
      const candidates: Array<{ j: number; sim: number }> = [];
      for (let j = 0; j < n; j++) {
        if (i === j) continue;
        const b = contents[j];
        const bKw = new Set(b.userData.node.keywords || []);
        const inter = new Set([...aKw].filter((x) => bKw.has(x)));
        const union = new Set([...aKw, ...bKw]);
        const sim = union.size === 0 ? 0 : inter.size / union.size;
        if (sim >= threshold) candidates.push({ j, sim });
      }
      candidates.sort((x, y) => y.sim - x.sim);
      const picks = candidates.slice(0, K);
      for (const { j, sim } of picks) {
        if (j <= i) continue; // avoid duplicates
        const b = contents[j];
        positions.push(a.position.x, a.position.y, a.position.z, b.position.x, b.position.y, b.position.z);
        meta.push({ aId: a.userData.node.id, bId: b.userData.node.id, similarity: sim });
      }
    }
    if (positions.length === 0) return undefined;
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(new Float32Array(positions), 3));
    const material = new THREE.LineBasicMaterial({ color: 0x22c55e, transparent: true, opacity: 0.35 });
    const lines = new THREE.LineSegments(geometry, material);
    return { lines, meta };
  }

  private intersectSemanticEdgesAtScreenPoint(clientX: number, clientY: number): { point: THREE.Vector3; index?: number } | null {
    if (!this.semanticEdgeLines) return null;
    const rect = this.renderer.domElement.getBoundingClientRect();
    const normalizedX = (clientX - rect.left) / rect.width;
    const normalizedY = (clientY - rect.top) / rect.height;
    this.pointer.x = normalizedX * 2 - 1;
    this.pointer.y = -(normalizedY * 2 - 1);
    this.raycaster.setFromCamera(this.pointer, this.camera);
    const hits = this.raycaster.intersectObject(this.semanticEdgeLines, false);
    if (hits && hits.length > 0) {
      const h = hits[0];
      return { point: h.point.clone(), index: (h as any).index as number | undefined };
    }
    return null;
  }

  private showLinkTooltip(worldPoint: THREE.Vector3, text: string): void {
    if (this.linkTooltip) {
      const el = this.linkTooltip.element as HTMLElement;
      el.textContent = text;
      this.linkTooltip.position.copy(worldPoint);
      return;
    }
    const el = document.createElement('div');
    el.style.padding = '2px 6px';
    el.style.fontSize = '11px';
    el.style.borderRadius = '6px';
    el.style.background = 'rgba(2,6,23,0.9)';
    el.style.color = '#e2e8f0';
    el.style.border = '1px solid rgba(148,163,184,0.4)';
    el.style.pointerEvents = 'none';
    el.style.whiteSpace = 'nowrap';
    el.textContent = text;
    const label = new CSS2DObject(el);
    label.position.copy(worldPoint);
    this.linkTooltip = label;
    this.scene.add(label);
  }

  private hideLinkTooltip(): void {
    if (!this.linkTooltip) return;
    this.linkTooltip.removeFromParent();
    const el = this.linkTooltip.element as HTMLElement | undefined;
    el?.remove();
    this.linkTooltip = undefined;
  }

  setMinimapContainer(container: HTMLDivElement | null): void {
    if (this.minimapRenderer) {
      this.minimapRenderer.dispose();
      this.minimapRenderer.domElement.remove();
      this.minimapRenderer = undefined;
    }

    if (this.minimapHelper) {
      this.scene.remove(this.minimapHelper);
      this.minimapHelper = undefined;
    }

    this.minimapCamera = undefined;
    this.minimapContainer = undefined;

    if (!container) {
      return;
    }

    this.minimapContainer = container;

    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setPixelRatio(window.devicePixelRatio);
    renderer.setClearColor(0x000000, 0);

    const width = container.clientWidth || 1;
    const height = container.clientHeight || 1;
    renderer.setSize(width, height);
    container.appendChild(renderer.domElement);
    this.minimapRenderer = renderer;
    this.updateMinimapSize();

    const range = 400;
    const camera = new THREE.OrthographicCamera(-range, range, range, -range, 1, 2000);
    camera.position.set(0, 600, 0);
    camera.up.set(0, 0, -1);
    camera.lookAt(0, 0, 0);
    this.minimapCamera = camera;
    this.updateMinimapSize();

    const helper = new THREE.CameraHelper(this.camera);
    helper.visible = false;
    const helperMaterial = helper.material as THREE.LineBasicMaterial;
    helperMaterial.color.set(0xffffff);
    this.scene.add(helper);
    this.minimapHelper = helper;
  }

  dispose(): void {
    cancelAnimationFrame(this.animationFrame ?? 0);
    this.controls.removeEventListener('start', this.handleControlsStart);
    this.controls.removeEventListener('end', this.handleControlsEnd);
    this.controls.dispose();
    this.renderer.domElement.removeEventListener('pointerdown', this.handlePointerDown);
    this.renderer.domElement.removeEventListener('pointermove', this.handlePointerMove);
    this.renderer.domElement.removeEventListener('pointerup', this.handlePointerUp);
    this.renderer.domElement.removeEventListener('pointerleave', this.handlePointerLeave);
    this.renderer.domElement.removeEventListener('dblclick', this.handleDoubleClick);
    window.removeEventListener('resize', this.handleResize);
    if (this.pendingZoomIn) {
      window.clearTimeout(this.pendingZoomIn);
      this.pendingZoomIn = undefined;
    }
    if (this.interactionEndTimeout) {
      window.clearTimeout(this.interactionEndTimeout);
      this.interactionEndTimeout = undefined;
    }
    this.clearLabels();
    this.disposeNodeMeshes();
    this.clearHoverLinks();
    this.hideHoverCard();
    this.hoveredNodeId = undefined;
    this.clearCursorTrail();
    this.clearMultiViewEntries();
    this.clearAdvancedEdges();
    this.hideLinkTooltip();
    this.setMinimapContainer(null);
    this.renderer.dispose();
    this.renderer.domElement.remove();
    this.labelRenderer.domElement.remove();
    this.scene.remove(this.cursorLight);
    this.trailGeometry.dispose();
    this.darkMaterial.dispose();
  }

  private bindEvents() {
    this.renderer.domElement.addEventListener('pointerdown', this.handlePointerDown);
    this.renderer.domElement.addEventListener('pointermove', this.handlePointerMove);
    this.renderer.domElement.addEventListener('pointerup', this.handlePointerUp);
    this.renderer.domElement.addEventListener('pointerleave', this.handlePointerLeave);
    this.renderer.domElement.addEventListener('dblclick', this.handleDoubleClick);
    window.addEventListener('resize', this.handleResize);
  }

  private markInteraction(): void {
    this.isInteracting = true;
    if (this.interactionEndTimeout) {
      window.clearTimeout(this.interactionEndTimeout);
      this.interactionEndTimeout = undefined;
    }
  }

  private scheduleInteractionEnd(): void {
    if (this.interactionEndTimeout) {
      window.clearTimeout(this.interactionEndTimeout);
    }
    this.interactionEndTimeout = window.setTimeout(() => {
      this.isInteracting = false;
      this.interactionEndTimeout = undefined;
    }, this.interactionCooldown);
  }

  private handleControlsStart = () => {
    this.markInteraction();
  };

  private handleControlsEnd = () => {
    this.scheduleInteractionEnd();
  };

  private handleResize = () => {
    this.updateSize();
  };

  private handlePointerDown = () => {
    this.markInteraction();
    this.isDragging = false;
    this.lastPointerDown = performance.now();
    this.edgePanVector.set(0, 0);
    this.hideHoverCard();
  };

  private handlePointerMove = (event: PointerEvent) => {
    if (!this.isDragging) {
      const elapsed = performance.now() - this.lastPointerDown;
      if (elapsed > 150) {
        this.isDragging = true;
      }
    }

    const rect = this.renderer.domElement.getBoundingClientRect();
    const normalizedX = (event.clientX - rect.left) / rect.width;
    const normalizedY = (event.clientY - rect.top) / rect.height;

    this.pointer.x = normalizedX * 2 - 1;
    this.pointer.y = -(normalizedY * 2 - 1);
    this.raycaster.setFromCamera(this.pointer, this.camera);

    const planeNormal = this.camera.getWorldDirection(new THREE.Vector3()).normalize();
    const planePoint = this.controls.target.clone();
    const interactionPlane = new THREE.Plane().setFromNormalAndCoplanarPoint(planeNormal, planePoint);
    const intersectionPoint = new THREE.Vector3();

    if (event.buttons !== 0) {
      this.markInteraction();
    }

    if (this.raycaster.ray.intersectPlane(interactionPlane, intersectionPoint)) {
      this.cursorLight.position.copy(intersectionPoint);
      if (this.cursorTrailEnabled) {
        const now = performance.now();
        if (now - this.lastTrailSpawn > this.cursorTrailInterval) {
          this.lastTrailSpawn = now;
          this.spawnCursorTrail(intersectionPoint.clone());
        }
      }
    }

    const threshold = this.edgePanThreshold;
    let horizontal = 0;
    let vertical = 0;

    if (normalizedX >= 0 && normalizedX <= 1 && normalizedY >= 0 && normalizedY <= 1) {
      if (normalizedX < threshold) {
        horizontal = -(threshold - normalizedX) / threshold;
      } else if (normalizedX > 1 - threshold) {
        horizontal = (normalizedX - (1 - threshold)) / threshold;
      }

      if (normalizedY < threshold) {
        vertical = (threshold - normalizedY) / threshold;
      } else if (normalizedY > 1 - threshold) {
        vertical = -(normalizedY - (1 - threshold)) / threshold;
      }
    }

    if (this.isDragging) {
      horizontal = 0;
      vertical = 0;
    }

    this.edgePanVector.set(horizontal, vertical);

    if (!this.isDragging) {
      const hovered = this.intersectNodesAtScreenPoint(event.clientX, event.clientY);
      const hoveredContentId = hovered && hovered.userData.node.kind === 'content' ? hovered.userData.node.id : undefined;
      if (hoveredContentId !== this.hoveredNodeId) {
        this.hoveredNodeId = hoveredContentId;
        this.updateHoverLinks(hoveredContentId ? hovered : undefined);
      }
      this.updateHoverCard(hovered);

      if (this.showSemanticEdges && this.semanticEdgeLines) {
        (this.raycaster.params as any).Line = { threshold: 3 };
        const intersection = this.intersectSemanticEdgesAtScreenPoint(event.clientX, event.clientY);
        if (intersection) {
          const segIndex = Math.max(0, Math.floor((intersection.index ?? 0) / 2));
          const meta = this.semanticEdgeMeta[segIndex];
          if (meta) {
            const label = `Semantic link: ${Math.round(meta.similarity * 100)}%`;
            this.showLinkTooltip(intersection.point, label);
          }
        } else {
          this.hideLinkTooltip();
        }
      } else {
        this.hideLinkTooltip();
      }
    }
    if (this.isInteracting || this.isDragging) {
      this.hideLinkTooltip();
      this.clearHoverLinks();
      this.hideHoverCard();
    }
  };

  private handlePointerLeave = () => {
    this.edgePanVector.set(0, 0);
    this.hoveredNodeId = undefined;
    this.clearHoverLinks();
    this.hideHoverCard();
    this.scheduleInteractionEnd();
  };

  private handleDoubleClick = (event: MouseEvent) => {
    this.markInteraction();
    if (this.pendingZoomIn) {
      window.clearTimeout(this.pendingZoomIn);
      this.pendingZoomIn = undefined;
    }

    const hit = this.intersectNodesAtScreenPoint(event.clientX, event.clientY);
    if (hit) {
      return;
    }

    this.setSelection(null);
    this.onNodeSelect?.(null);
    this.zoomAtScreenPoint(event.clientX, event.clientY, 'out');
  };

  private handlePointerUp = (event: PointerEvent) => {
    this.scheduleInteractionEnd();
    if (this.pendingZoomIn) {
      window.clearTimeout(this.pendingZoomIn);
      this.pendingZoomIn = undefined;
    }

    if (this.isDragging) {
      return;
    }

    const hit = this.intersectNodesAtScreenPoint(event.clientX, event.clientY);

    if (hit) {
      this.setSelection(hit);
      this.onNodeSelect?.(hit.userData.node);
      return;
    }

    this.setSelection(null);
    this.onNodeSelect?.(null);

    const { clientX, clientY } = event;
    this.pendingZoomIn = window.setTimeout(() => {
      this.zoomAtScreenPoint(clientX, clientY, 'in');
      this.pendingZoomIn = undefined;
    }, 180);
  };

  private animate = () => {
    const delta = this.clock.getDelta();

    if (delta > 0) {
      const currentFps = 1 / delta;
      this.smoothedFps = this.smoothedFps === 0 ? currentFps : this.smoothedFps * 0.9 + currentFps * 0.1;
      this.fpsAccumulator += delta;
      if (this.fpsAccumulator >= 0.5) {
        this.onFpsUpdate?.(Math.round(this.smoothedFps));
        this.fpsAccumulator = 0;
      }
    }

    if (this.edgePanVector.lengthSq() > 0.0001) {
      const panX = this.edgePanVector.x * this.edgePanSpeed * delta;
      const panY = this.edgePanVector.y * this.edgePanSpeed * delta;
      this.controls.pan(panX, panY);
    }

    this.controls.update();
    this.updateCursorTrail(delta);
    this.updateAnimations(delta);
    this.updateChunkVisibilityBudget(delta);
    this.updatePerformanceCulling();
    this.updateLabelVisibility();
    this.updateMultiViewEntries();

    this.renderer.clear();
    this.renderer.render(this.scene, this.camera);
    this.renderer.clearDepth();

    if (this.hasBloomTargets) {
      const previousMask = this.camera.layers.mask;
      this.camera.layers.set(CanvasScene.BLOOM_LAYER);
      this.composer.render();
      this.camera.layers.mask = previousMask;
    }

    this.labelRenderer.render(this.scene, this.camera);

    if (this.minimapRenderer && this.minimapCamera) {
      if (this.minimapHelper) {
        this.minimapHelper.visible = true;
        this.minimapHelper.update();
      }
      this.multiViewHelpers.forEach((helper) => {
        helper.visible = true;
        helper.update();
      });
      this.minimapRenderer.render(this.scene, this.minimapCamera);
      if (this.minimapHelper) {
        this.minimapHelper.visible = false;
      }
      this.multiViewHelpers.forEach((helper) => {
        helper.visible = false;
      });
    }

    this.animationFrame = requestAnimationFrame(this.animate);
  };

  private updateAnimations(delta: number): void {
    if (this.positionAnimations.size === 0 && this.scaleAnimations.size === 0) {
      return;
    }

    this.positionAnimations.forEach((animation, id) => {
      const mesh = this.nodeMeshes.get(id);
      if (!mesh) {
        this.positionAnimations.delete(id);
        return;
      }
      animation.elapsed = Math.min(animation.elapsed + delta, animation.duration);
      const progress = this.easeInOut(animation.elapsed / animation.duration);
      mesh.position.lerpVectors(animation.start, animation.end, progress);
      if (animation.elapsed >= animation.duration) {
        mesh.position.copy(animation.end);
        this.positionAnimations.delete(id);
      }
    });

    this.scaleAnimations.forEach((animation, id) => {
      const mesh = this.nodeMeshes.get(id);
      if (!mesh) {
        this.scaleAnimations.delete(id);
        return;
      }
      animation.elapsed = Math.min(animation.elapsed + delta, animation.duration);
      const progress = this.easeInOut(animation.elapsed / animation.duration);
      const value = animation.start + (animation.end - animation.start) * progress;
      mesh.scale.setScalar(value);
      if (animation.elapsed >= animation.duration) {
        mesh.scale.setScalar(animation.end);
        this.scaleAnimations.delete(id);
      }
    });
  }

  private easeInOut(t: number): number {
    return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
  }

  private updateChunkVisibilityBudget(delta: number): void {
    if (!Number.isFinite(this.smoothedFps) || this.smoothedFps <= 0) {
      return;
    }
    const current = this.chunkVisibilityDistance;
    const decreaseStep = delta * 240;
    const increaseStep = delta * 160;

    if (this.smoothedFps < 45) {
      this.chunkVisibilityDistance = Math.max(this.minChunkVisibilityDistance, current - decreaseStep);
    } else if (this.smoothedFps > 58) {
      this.chunkVisibilityDistance = Math.min(this.maxChunkVisibilityDistance, current + increaseStep);
    }
  }

  private updatePerformanceCulling(): void {
    if (this.nodeMeshes.size === 0) {
      return;
    }

    this.viewProjectionMatrix.multiplyMatrices(this.camera.projectionMatrix, this.camera.matrixWorldInverse);
    this.frustum.setFromProjectionMatrix(this.viewProjectionMatrix);
    const distanceLimit = this.chunkVisibilityDistance;

    this.nodeMeshes.forEach((mesh) => {
      const node = mesh.userData.node;
      if (node.kind !== 'chunk') {
        return;
      }

      if (this.hiddenKinds.has(node.kind)) {
        return;
      }

      const distance = this.camera.position.distanceTo(mesh.position);
      const inFrustum = this.frustum.containsPoint(mesh.position);
      const visible = inFrustum && distance <= distanceLimit;
      mesh.visible = visible;
      mesh.children.forEach((child) => {
        child.visible = visible;
      });
    });

    if (this.hoverLinkLines && this.hoveredNodeId) {
      const sourceId = this.nodeContentSourceIndex.get(this.hoveredNodeId);
      if (sourceId) {
        const visibleChunks = (this.contentToChunkMeshes.get(sourceId) ?? []).filter((chunk) => chunk.visible);
        if (visibleChunks.length === 0) {
          this.clearHoverLinks();
        }
      }
    }
  }

  private updateLabelVisibility(): void {
    if (this.labels.length === 0) {
      return;
    }
    const cameraPosition = this.camera.position;
    this.labels.forEach((label) => {
      const kind = label.userData?.kind as CanvasNodeKind | undefined;
      if (!kind) {
        label.visible = true;
        return;
      }
      const limit = this.labelHideDistances[kind];
      if (!Number.isFinite(limit)) {
        label.visible = true;
        return;
      }
      const worldPosition = label.getWorldPosition(this.tempVector);
      const distance = cameraPosition.distanceTo(worldPosition);
      label.visible = distance <= limit;
    });
  }

  private disposeNodeMeshes(): void {
    this.nodeMeshes.clear();
    this.originalPositions.clear();
    this.contentToChunkMeshes.clear();
    this.nodeContentSourceIndex.clear();
    this.chunkFallbackAnchors.clear();
    this.positionAnimations.clear();
    this.scaleAnimations.clear();
    this.nodeGroup.children.forEach((child) => {
      const mesh = child as THREE.Mesh;
      this.storedMaterials.delete(mesh.uuid);
      mesh.geometry.dispose();
      if (Array.isArray(mesh.material)) {
        mesh.material.forEach((material) => material.dispose());
      } else {
        mesh.material.dispose();
      }
    });
    this.nodeGroup.clear();
  }

  private resolveContentPositionOverrides(nodes: CanvasRenderableNode[]): Map<string, THREE.Vector3> {
    const overrides = new Map<string, THREE.Vector3>();
    const contents = nodes.filter(
      (node): node is CanvasContentNode => node.kind === 'content' && node.visibility
    );

    if (contents.length <= 1) {
      return overrides;
    }

    const originalPositions = contents.map(
      (node) => new THREE.Vector3(node.position.x, node.position.y, node.position.z)
    );
    const radii = contents.map(
      (node) => node.display.size ?? this.nodeSizeDefaults.content
    );
    let positions = originalPositions.map((pos) => pos.clone());
    let autoLaidOut = false;
    const xs = positions.map((pos) => pos.x);
    const zs = positions.map((pos) => pos.z);
    const spreadX = Math.max(...xs) - Math.min(...xs);
    const spreadZ = Math.max(...zs) - Math.min(...zs);
    const maxRadius = radii.length > 0 ? Math.max(...radii) : this.nodeSizeDefaults.content;
    const minSpacing = Math.max(40, maxRadius * 2.8);
    const defaultPositions = contents.every((node) => Math.abs(node.position.x) < 1e-3 && Math.abs(node.position.z) < 1e-3);
    const lineLike = Math.min(spreadX, spreadZ) < minSpacing * 0.45 && Math.max(spreadX, spreadZ) > minSpacing * 1.5;

    if (defaultPositions || lineLike) {
      const average = originalPositions.reduce((acc, vec) => acc.add(vec), new THREE.Vector3()).multiplyScalar(1 / contents.length);
      const layoutRadius = Math.max(minSpacing, minSpacing * Math.sqrt(contents.length) * 0.75);
      const generated = generateFibonacciDiskPositions(contents.length, layoutRadius, new THREE.Vector3(average.x, average.y, average.z));
      positions = generated.map((pos, index) => pos.setY(originalPositions[index].y));
      autoLaidOut = true;
    }

    let needsAdjustment = false;
    for (let i = 0; i < contents.length && !needsAdjustment; i += 1) {
      for (let j = i + 1; j < contents.length; j += 1) {
        const dx = positions[j].x - positions[i].x;
        const dz = positions[j].z - positions[i].z;
        const dist = Math.hypot(dx, dz);
        const minDistance = (radii[i] + radii[j]) * 1.05;
        if (dist < minDistance) {
          needsAdjustment = true;
          break;
        }
      }
    }

    if (!needsAdjustment && !autoLaidOut) {
      return overrides;
    }

    if (!autoLaidOut) {
      const maxIterations = 14;
      for (let iteration = 0; iteration < maxIterations; iteration += 1) {
        let changed = false;
        for (let i = 0; i < contents.length; i += 1) {
          for (let j = i + 1; j < contents.length; j += 1) {
            const posA = positions[i];
            const posB = positions[j];
            const dx = posB.x - posA.x;
            const dz = posB.z - posA.z;
            const dist = Math.hypot(dx, dz);
            const minDistance = (radii[i] + radii[j]) * 1.1;

            if (dist < 1e-4) {
              const angle = ((i * 47 + j * 89 + 13) % 360) * (Math.PI / 180);
              const nx = Math.cos(angle);
              const nz = Math.sin(angle);
              const push = minDistance * 0.5;
              posA.x -= nx * push;
              posA.z -= nz * push;
              posB.x += nx * push;
              posB.z += nz * push;
              changed = true;
              continue;
            }

            if (dist < minDistance) {
              const push = (minDistance - dist) / 2;
              const nx = dx / (dist || 1);
              const nz = dz / (dist || 1);
              posA.x -= nx * push;
              posA.z -= nz * push;
              posB.x += nx * push;
              posB.z += nz * push;
              changed = true;
            }
          }
        }

        if (!changed) {
          break;
        }
      }
    }

    const originalAvg = positions.reduce(
      (acc, _, index) => {
        acc.x += contents[index].position.x;
        acc.z += contents[index].position.z;
        return acc;
      },
      { x: 0, z: 0 }
    );
    originalAvg.x /= contents.length;
    originalAvg.z /= contents.length;

    const adjustedAvg = positions.reduce(
      (acc, pos) => {
        acc.x += pos.x;
        acc.z += pos.z;
        return acc;
      },
      { x: 0, z: 0 }
    );
    adjustedAvg.x /= positions.length;
    adjustedAvg.z /= positions.length;

    const offsetX = adjustedAvg.x - originalAvg.x;
    const offsetZ = adjustedAvg.z - originalAvg.z;

    positions.forEach((pos) => {
      pos.x -= offsetX;
      pos.z -= offsetZ;
    });

    positions.forEach((pos, index) => {
      const original = contents[index];
      const originalVector = new THREE.Vector3(original.position.x, original.position.y, original.position.z);
      if (pos.distanceToSquared(originalVector) > 1e-4) {
        overrides.set(original.id, new THREE.Vector3(pos.x, original.position.y, pos.z));
      }
    });

    return overrides;
  }

  private resolveChunkLayout(
    nodes: CanvasRenderableNode[],
    contentOverrides: Map<string, THREE.Vector3>
  ): { positions: Map<string, THREE.Vector3>; sizes: Map<string, number> } {
    const positions = new Map<string, THREE.Vector3>();
    const sizes = new Map<string, number>();
    const contentBySourceId = new Map<string, CanvasContentNode>();

    nodes.forEach((node) => {
      if (node.kind === 'content') {
        const sourceKey = this.normalizeContentSourceId(node.contentSourceId);
        if (sourceKey) {
          contentBySourceId.set(sourceKey, node);
        }
      }
    });

    const chunkGroups = new Map<string, CanvasChunkNode[]>();
    nodes.forEach((node) => {
      if (node.kind !== 'chunk') {
        return;
      }
      const key = this.normalizeContentSourceId(node.contentSourceId) ?? `__chunk:${node.id}`;
      const list = chunkGroups.get(key);
      if (list) {
        list.push(node);
      } else {
        chunkGroups.set(key, [node]);
      }
    });

    const groupKeys = new Set(chunkGroups.keys());
    for (const key of Array.from(this.chunkFallbackAnchors.keys())) {
      if (!groupKeys.has(key)) {
        this.chunkFallbackAnchors.delete(key);
      }
    }

    chunkGroups.forEach((chunks, key) => {
      if (chunks.length === 0) {
        return;
      }

      const parent = contentBySourceId.get(key);
      let baseCenter: THREE.Vector3;
      let parentSize = parent?.display.size ?? this.nodeSizeDefaults.content;

      if (parent) {
        const override = contentOverrides.get(parent.id);
        baseCenter = override ? override.clone() : new THREE.Vector3(parent.position.x, parent.position.y, parent.position.z);
        this.chunkFallbackAnchors.delete(key);
      } else {
        baseCenter = this.computeChunkFallbackAnchor(key, chunks);
        const maxChunkSize = chunks.reduce((acc, chunk) => Math.max(acc, chunk.display.size ?? this.nodeSizeDefaults.chunk), this.nodeSizeDefaults.chunk);
        parentSize = Math.max(parentSize, maxChunkSize * 5);
      }

      const sortedChunks = [...chunks].sort((a, b) => a.sequenceIndex - b.sequenceIndex);
      const radiusFactor = Math.min(0.85, 0.45 + Math.log2(sortedChunks.length + 1) * 0.1);
      const radius = Math.max(parentSize * 0.45, parentSize * radiusFactor);
      const densityFactor = Math.max(1, Math.sqrt(sortedChunks.length) / 4);

      let generated: THREE.Vector3[];
      if (sortedChunks.length === 1) {
        generated = [this.computeSingleChunkPosition(key, baseCenter, radius)];
      } else {
        generated = generateFibonacciSpherePositions(sortedChunks.length, radius, baseCenter.clone());
      }

      sortedChunks.forEach((chunk, index) => {
        const position = generated[index] ?? baseCenter;
        positions.set(chunk.id, position.clone());
        const baseSize = chunk.display.size ?? this.nodeSizeDefaults.chunk;
        const capByParent = parentSize * 0.18 / densityFactor;
        const size = Math.max(0.6, Math.min(baseSize, capByParent));
        sizes.set(chunk.id, size);
      });
    });

    return { positions, sizes };
  }

  private normalizeContentSourceId(value?: string | null): string | null {
    if (!value) {
      return null;
    }
    const trimmed = value.trim();
    return trimmed.length > 0 ? trimmed : null;
  }

  private computeDeterministicPolar(seed: string): { angle: number; radiusFactor: number } {
    if (!seed) {
      return { angle: Math.PI * 0.25, radiusFactor: 0.5 };
    }

    let hash = 2166136261;
    for (let index = 0; index < seed.length; index += 1) {
      hash ^= seed.charCodeAt(index);
      hash = Math.imul(hash, 16777619);
    }
    const unsigned = hash >>> 0;
    const angle = ((unsigned & 0xffff) / 0xffff) * Math.PI * 2;
    const radiusFactor = (((unsigned >>> 16) & 0xffff) / 0xffff);
    return { angle, radiusFactor: Number.isFinite(radiusFactor) ? radiusFactor : 0.5 };
  }

  private computeChunkFallbackAnchor(key: string, chunks: CanvasChunkNode[]): THREE.Vector3 {
    const cached = this.chunkFallbackAnchors.get(key);
    if (cached) {
      return cached.clone();
    }

    const centroid = chunks.reduce((acc, chunk) => {
      const pos = new THREE.Vector3(chunk.position.x, chunk.position.y, chunk.position.z);
      return acc.add(pos);
    }, new THREE.Vector3());

    if (chunks.length > 0) {
      centroid.multiplyScalar(1 / chunks.length);
    }

    const hasValidCentroid = Number.isFinite(centroid.x) && Number.isFinite(centroid.y) && Number.isFinite(centroid.z);
    let anchor = centroid;

    if (!hasValidCentroid || centroid.lengthSq() < 400) {
      const { angle, radiusFactor } = this.computeDeterministicPolar(key);
      const baseRadius = 180 + radiusFactor * 140;
      const y = hasValidCentroid ? centroid.y : -24;
      anchor = new THREE.Vector3(Math.cos(angle) * baseRadius, y, Math.sin(angle) * baseRadius);
    }

    this.chunkFallbackAnchors.set(key, anchor.clone());
    return anchor.clone();
  }

  private computeSingleChunkPosition(key: string, center: THREE.Vector3, radius: number): THREE.Vector3 {
    const { angle } = this.computeDeterministicPolar(key);
    const effectiveRadius = Math.max(radius, this.nodeSizeDefaults.content * 0.5);
    const offset = new THREE.Vector3(Math.cos(angle), 0, Math.sin(angle)).multiplyScalar(effectiveRadius);
    return center.clone().add(offset);
  }

  private determineNodeSize(node: CanvasRenderableNode, chunkSizes: Map<string, number>): number {
    if (node.kind === 'chunk') {
      const override = chunkSizes.get(node.id);
      if (typeof override === 'number') {
        return override;
      }
      const base = node.display.size ?? this.nodeSizeDefaults.chunk;
      return Math.max(0.6, Math.min(base, this.nodeSizeDefaults.chunk * 0.7));
    }

    return node.display.size ?? this.nodeSizeDefaults[node.kind];
  }

  private resolveNodeColor(node: CanvasRenderableNode): THREE.Color {
    const fallback = this.fallbackColors[node.kind];
    let color: THREE.Color;
    try {
      color = new THREE.Color(node.display.color ?? fallback);
    } catch (error) {
      color = new THREE.Color(fallback);
    }

    const luminance = 0.299 * color.r + 0.587 * color.g + 0.114 * color.b;
    if (!node.display.color || luminance < 0.08) {
      color.set(fallback);
    }

    return color;
  }

  private clearLabels(): void {
    this.labels.forEach((label) => {
      label.removeFromParent();
      const element = label.element as HTMLElement | undefined;
      element?.remove();
    });
    this.labels.length = 0;
  }

  private hideHoverCard(): void {
    if (!this.hoverCard) {
      return;
    }
    this.hoverCard.removeFromParent();
    const element = this.hoverCard.element as HTMLElement | undefined;
    element?.remove();
    this.hoverCard = undefined;
    this.hoverCardNodeId = undefined;
  }

  private getNodeLabelText(node: CanvasRenderableNode): string | null {
    if (node.kind === 'chunk') {
      return null;
    }

    const clean = (value?: string | null) => {
      if (!value) {
        return null;
      }
      const trimmed = value.trim();
      return trimmed.length > 0 ? trimmed : null;
    };

    const base =
      clean(node.title) ??
      clean(node.displayContent) ??
      clean(node.chatContent);

    let resolved = base;

    if (!resolved) {
      if (node.kind === 'content') {
        resolved =
          clean(node.documentTitle) ??
          clean(node.displayContent) ??
          clean(node.chatContent) ??
          (node.keywords.length > 0 ? node.keywords.slice(0, 3).join(', ') : null) ??
          clean(node.contentSourceId);
      } else if (node.kind === 'cluster') {
        resolved =
          clean(node.clusterScope) ??
          clean(node.displayContent) ??
          clean(node.chatContent) ??
          (node.keywords.length > 0 ? node.keywords.slice(0, 3).join(', ') : null);
      }
    }

    if (!resolved) {
      resolved = clean(node.id) ?? null;
    }

    if (!resolved) {
      return null;
    }

    const limit = node.kind === 'content' ? 72 : 64;
    return resolved.length > limit ? `${resolved.slice(0, limit - 1)}…` : resolved;
  }

  private createLabelForNode(node: CanvasRenderableNode, text: string | null, effectiveSize: number): CSS2DObject | null {
    if (!text) {
      return null;
    }

    const element = document.createElement('div');
    element.className = 'canvas-node-label';
    element.textContent = text;
    element.style.padding = '2px 6px';
    element.style.fontSize = '11px';
    element.style.borderRadius = '6px';
    element.style.background = 'rgba(15,23,42,0.85)';
    element.style.color = '#e2e8f0';
    element.style.border = '1px solid rgba(148,163,184,0.4)';
    element.style.whiteSpace = 'nowrap';
    element.style.pointerEvents = 'none';
    element.style.backdropFilter = 'blur(2px)';
    element.style.textShadow = '0 1px 2px rgba(15,15,15,0.6)';

    const label = new CSS2DObject(element);
    label.userData = { nodeId: node.id, kind: node.kind };
    label.position.set(0, effectiveSize * 1.3, 0);
    return label;
  }

  private getMeshRadius(mesh: MeshWithData): number {
    if (!mesh.geometry.boundingSphere) {
      mesh.geometry.computeBoundingSphere();
    }
    return mesh.geometry.boundingSphere?.radius ?? this.nodeSizeDefaults[mesh.userData.node.kind];
  }

  private truncate(text: string, maxLength: number): string {
    if (text.length <= maxLength) {
      return text;
    }
    return `${text.slice(0, Math.max(0, maxLength - 1))}…`;
  }

  private updateHoverCard(mesh?: MeshWithData): void {
    if (this.isInteracting) {
      this.hideHoverCard();
      return;
    }

    if (!mesh) {
      this.hideHoverCard();
      return;
    }

    const node = mesh.userData.node;
    if (node.kind !== 'content' && node.kind !== 'chunk' && node.kind !== 'cluster') {
      this.hideHoverCard();
      return;
    }

    if (this.hoverCardNodeId === node.id && this.hoverCard) {
      this.positionHoverCard(mesh, this.hoverCard);
      return;
    }

    this.hideHoverCard();
    const card = this.createHoverCard(mesh);
    if (!card) {
      return;
    }
    mesh.add(card);
    this.positionHoverCard(mesh, card);
    this.hoverCard = card;
    this.hoverCardNodeId = node.id;
  }

  private positionHoverCard(mesh: MeshWithData, card: CSS2DObject): void {
    const radius = this.getMeshRadius(mesh);
    card.position.set(0, radius * 1.8, 0);
  }

  private createHoverCard(mesh: MeshWithData): CSS2DObject | null {
    const node = mesh.userData.node;
    const element = document.createElement('div');
    element.className = 'canvas-hover-card';
    element.style.padding = '10px 12px 12px';
    element.style.borderRadius = '10px';
    element.style.background = 'rgba(15,23,42,0.92)';
    element.style.border = '1px solid rgba(148,163,184,0.45)';
    element.style.color = '#e2e8f0';
    element.style.minWidth = '200px';
    element.style.maxWidth = '280px';
    element.style.fontSize = '12px';
    element.style.lineHeight = '1.45';
    element.style.pointerEvents = 'auto';
    element.style.boxShadow = '0 16px 30px rgba(15,23,42,0.4)';
    element.style.backdropFilter = 'blur(6px)';

    const header = document.createElement('div');
    header.style.display = 'flex';
    header.style.alignItems = 'center';
    header.style.justifyContent = 'space-between';
    header.style.marginBottom = '8px';
    element.appendChild(header);

    const title = document.createElement('div');
    title.style.fontWeight = '600';
    title.style.fontSize = '13px';
    let headingText = this.getNodeLabelText(node);
    if (!headingText) {
      if (node.kind === 'chunk' && typeof node.sequenceIndex === 'number') {
        headingText = `Chunk ${node.sequenceIndex + 1}`;
      } else if (node.kind === 'cluster') {
        headingText = node.clusterScope ?? 'Cluster';
      } else {
        headingText = node.title ?? node.displayContent ?? node.chatContent ?? node.kind.toUpperCase();
      }
    }
    title.textContent = this.truncate(headingText ?? node.kind.toUpperCase(), 70);
    header.appendChild(title);

    const detailButton = document.createElement('button');
    detailButton.type = 'button';
    detailButton.textContent = 'Details';
    detailButton.style.fontSize = '11px';
    detailButton.style.fontWeight = '600';
    detailButton.style.padding = '4px 8px';
    detailButton.style.borderRadius = '6px';
    detailButton.style.border = '1px solid rgba(148,163,184,0.65)';
    detailButton.style.background = 'rgba(30,41,59,0.9)';
    detailButton.style.color = '#bfdbfe';
    detailButton.style.cursor = 'pointer';
    detailButton.style.transition = 'background 0.15s ease, color 0.15s ease';
    detailButton.addEventListener('mouseenter', () => {
      detailButton.style.background = 'rgba(37,99,235,0.25)';
      detailButton.style.color = '#f8fafc';
    });
    detailButton.addEventListener('mouseleave', () => {
      detailButton.style.background = 'rgba(30,41,59,0.9)';
      detailButton.style.color = '#bfdbfe';
    });
    detailButton.addEventListener('click', (event) => {
      event.stopPropagation();
      this.setSelection(mesh);
      this.onNodeSelect?.(node);
    });
    header.appendChild(detailButton);

    const body = document.createElement('div');
    body.style.display = 'flex';
    body.style.flexDirection = 'column';
    body.style.gap = '6px';
    element.appendChild(body);

    if (node.kind === 'content') {
      if (node.keywords.length > 0) {
        const keywords = document.createElement('div');
        keywords.style.display = 'flex';
        keywords.style.flexWrap = 'wrap';
        keywords.style.gap = '4px';
        keywords.style.marginBottom = '2px';
        const tokens = node.keywords.slice(0, 6);
        tokens.forEach((kw) => {
          const chip = document.createElement('span');
          chip.textContent = kw;
          chip.style.fontSize = '10px';
          chip.style.fontWeight = '600';
          chip.style.padding = '2px 6px';
          chip.style.borderRadius = '999px';
          chip.style.background = 'rgba(59,130,246,0.14)';
          chip.style.color = '#93c5fd';
          keywords.appendChild(chip);
        });
        body.appendChild(keywords);
      }

      const contentBlock = document.createElement('div');
      contentBlock.style.whiteSpace = 'pre-wrap';
      contentBlock.style.fontSize = '11px';
      contentBlock.style.color = '#cbd5f5';
      contentBlock.style.maxHeight = '160px';
      contentBlock.style.overflowY = 'auto';
      contentBlock.style.paddingRight = '4px';
      const sourceText = node.chatContent || node.displayContent || 'No content available';
      contentBlock.textContent = this.truncate(sourceText, 240);
      body.appendChild(contentBlock);
    } else if (node.kind === 'chunk') {
      const chunkBlock = document.createElement('div');
      chunkBlock.style.whiteSpace = 'pre-wrap';
      chunkBlock.style.fontSize = '11px';
      chunkBlock.style.color = '#cbd5f5';
      chunkBlock.style.maxHeight = '160px';
      chunkBlock.style.overflowY = 'auto';
      chunkBlock.style.paddingRight = '4px';
      const sourceText = node.chatContent || node.displayContent || 'No content available';
      chunkBlock.textContent = this.truncate(sourceText, 240);
      body.appendChild(chunkBlock);
    } else if (node.kind === 'cluster') {
      const scope = document.createElement('div');
      scope.style.fontSize = '11px';
      scope.style.color = '#cbd5f5';
      scope.textContent = node.clusterScope ? `Scope: ${node.clusterScope}` : 'Cluster summary unavailable';
      body.appendChild(scope);
      if (typeof node.memberCount === 'number') {
        const members = document.createElement('div');
        members.style.fontSize = '11px';
        members.style.color = '#94a3b8';
        members.textContent = `Members: ${node.memberCount}`;
        body.appendChild(members);
      }
      if (node.keywords.length > 0) {
        const keywords = document.createElement('div');
        keywords.style.display = 'flex';
        keywords.style.flexWrap = 'wrap';
        keywords.style.gap = '4px';
        keywords.style.marginTop = '4px';
        const tokens = node.keywords.slice(0, 6);
        tokens.forEach((kw) => {
          const chip = document.createElement('span');
          chip.textContent = kw;
          chip.style.fontSize = '10px';
          chip.style.fontWeight = '600';
          chip.style.padding = '2px 6px';
          chip.style.borderRadius = '999px';
          chip.style.background = 'rgba(34,197,94,0.16)';
          chip.style.color = '#bbf7d0';
          keywords.appendChild(chip);
        });
        body.appendChild(keywords);
      }
    }

    const card = new CSS2DObject(element);
    return card;
  }

  private intersectNodesAtScreenPoint(clientX: number, clientY: number): MeshWithData | undefined {
    const rect = this.renderer.domElement.getBoundingClientRect();
    const normalizedX = (clientX - rect.left) / rect.width;
    const normalizedY = (clientY - rect.top) / rect.height;

    if (Number.isNaN(normalizedX) || Number.isNaN(normalizedY)) {
      return undefined;
    }

    this.pointer.x = normalizedX * 2 - 1;
    this.pointer.y = -(normalizedY * 2 - 1);
    this.raycaster.setFromCamera(this.pointer, this.camera);
    const intersects = this.raycaster.intersectObjects(this.nodeGroup.children, false);
    if (intersects.length > 0) {
      return intersects[0].object as MeshWithData;
    }
    return undefined;
  }

  private zoomAtScreenPoint(clientX: number, clientY: number, direction: 'in' | 'out'): void {
    const rect = this.renderer.domElement.getBoundingClientRect();
    const normalizedX = (clientX - rect.left) / rect.width;
    const normalizedY = (clientY - rect.top) / rect.height;

    this.pointer.x = normalizedX * 2 - 1;
    this.pointer.y = -(normalizedY * 2 - 1);
    this.raycaster.setFromCamera(this.pointer, this.camera);

    const planeNormal = this.camera.getWorldDirection(new THREE.Vector3()).normalize();
    const planePoint = this.controls.target.clone();
    const interactionPlane = new THREE.Plane().setFromNormalAndCoplanarPoint(planeNormal, planePoint);
    const intersectionPoint = new THREE.Vector3();
    if (!this.raycaster.ray.intersectPlane(interactionPlane, intersectionPoint)) {
      intersectionPoint.copy(this.controls.target);
    }

    const offset = new THREE.Vector3().subVectors(this.camera.position, this.controls.target);
    const currentDistance = offset.length();
    if (currentDistance < 1e-4) {
      return;
    }

    const zoomFactor = direction === 'in' ? this.zoomInFactor : this.zoomOutFactor;
    const minDistance = this.controls.minDistance ?? 1;
    const maxDistance = this.controls.maxDistance ?? Number.POSITIVE_INFINITY;
    const desiredDistance = THREE.MathUtils.clamp(currentDistance * zoomFactor, minDistance, maxDistance);
    const directionVector = offset.normalize().multiplyScalar(desiredDistance);

    const blend = direction === 'in' ? this.zoomBlendIn : this.zoomBlendOut;
    const newTarget = this.controls.target.clone().lerp(intersectionPoint, blend);
    const newCameraPosition = new THREE.Vector3().addVectors(newTarget, directionVector);

    this.controls.target.copy(newTarget);
    this.camera.position.copy(newCameraPosition);
    this.controls.update();
  }

  private applySelectionHighlight(): void {
    const selectedId = this.selectedNodeId;
    this.nodeGroup.children.forEach((child) => {
      const mesh = child as MeshWithData;
      if (!mesh.userData?.node) {
        return;
      }
      if (selectedId && mesh.userData.node.id === selectedId) {
        mesh.layers.enable(CanvasScene.BLOOM_LAYER);
      } else {
        mesh.layers.disable(CanvasScene.BLOOM_LAYER);
      }
    });
    this.updateBloomFlag();
  }

  private setSelection(mesh: MeshWithData | null): void {
    this.selectedNodeId = mesh?.userData.node.id;
    this.applySelectionHighlight();
  }

  private spawnCursorTrail(position: THREE.Vector3): void {
    const material = new THREE.MeshBasicMaterial({
      color: 0x7c3aed,
      transparent: true,
      opacity: 0.55,
    });
    const mesh = new THREE.Mesh(this.trailGeometry, material);
    mesh.position.copy(position);
    this.scene.add(mesh);

    this.cursorTrail.push({ mesh, life: CanvasScene.CURSOR_TRAIL_LIFETIME });

    if (this.cursorTrail.length > this.cursorTrailMaxEntries) {
      const oldest = this.cursorTrail.shift();
      if (oldest) {
        this.disposeTrailEntry(oldest);
      }
    }
  }

  private clearCursorTrail(): void {
    this.cursorTrail.forEach((entry) => this.disposeTrailEntry(entry));
    this.cursorTrail = [];
  }

  private disposeTrailEntry(entry: { mesh: THREE.Mesh; life: number }): void {
    this.scene.remove(entry.mesh);
    const material = entry.mesh.material as THREE.Material;
    material.dispose();
  }

  private updateCursorTrail(delta: number): void {
    for (let index = this.cursorTrail.length - 1; index >= 0; index -= 1) {
      const entry = this.cursorTrail[index];
      entry.life -= delta;
      if (entry.life <= 0) {
        this.disposeTrailEntry(entry);
        this.cursorTrail.splice(index, 1);
        continue;
      }

      const scale = 1 + (CanvasScene.CURSOR_TRAIL_LIFETIME - entry.life) * 1.4;
      entry.mesh.scale.setScalar(scale);

      const material = entry.mesh.material as THREE.MeshBasicMaterial;
      material.opacity = Math.max(0, entry.life / CanvasScene.CURSOR_TRAIL_LIFETIME);
    }
  }

  private darkenNonBloomed = (object: THREE.Object3D) => {
    const mesh = object as THREE.Mesh;
    if (!mesh.isMesh) {
      return;
    }

    if (this.bloomLayer.test(mesh.layers)) {
      return;
    }

    this.storedMaterials.set(mesh.uuid, mesh.material);
    mesh.material = this.darkMaterial;
  };

  private restoreMaterial = (object: THREE.Object3D) => {
    const mesh = object as THREE.Mesh;
    if (!mesh.isMesh) {
      return;
    }

    const material = this.storedMaterials.get(mesh.uuid);
    if (material) {
      mesh.material = material;
      this.storedMaterials.delete(mesh.uuid);
    }
  };

  private updateBloomFlag(): void {
    this.hasBloomTargets = this.nodeGroup.children.some((child) => {
      const mesh = child as MeshWithData;
      return mesh.isMesh && this.bloomLayer.test(mesh.layers);
    });
  }

  private updateMinimapSize(): void {
    if (!this.minimapRenderer || !this.minimapContainer) {
      return;
    }

    const width = this.minimapContainer.clientWidth || 1;
    const height = this.minimapContainer.clientHeight || 1;
    this.minimapRenderer.setSize(width, height);
    if (this.minimapCamera) {
      const range = 400;
      const aspect = width / height;
      this.minimapCamera.left = -range * aspect;
      this.minimapCamera.right = range * aspect;
      this.minimapCamera.top = range;
      this.minimapCamera.bottom = -range;
      this.minimapCamera.updateProjectionMatrix();
    }
  }

  private createMeshForNode(
    node: CanvasRenderableNode,
    overrides?: { position?: THREE.Vector3; size?: number }
  ): MeshWithData {
    let mesh: MeshWithData;
    const color = this.resolveNodeColor(node);
    const effectiveSize = overrides?.size ?? node.display.size ?? this.nodeSizeDefaults[node.kind];

    if (node.kind === 'cluster') {
      const geometry = new THREE.SphereGeometry(effectiveSize, 32, 32);
      const material = new THREE.MeshStandardMaterial({
        color,
        emissive: color.clone().multiplyScalar(0.32),
        emissiveIntensity: 0.85,
        metalness: 0.35,
        roughness: 0.5,
        transparent: true,
        opacity: node.display.opacity ?? 0.85,
      });
      mesh = new THREE.Mesh(geometry, material) as MeshWithData;
    } else if (node.kind === 'content') {
      const geometry = new THREE.IcosahedronGeometry(effectiveSize, 1);
      const material = new THREE.MeshBasicMaterial({
        color,
        wireframe: true,
        transparent: true,
        opacity: node.display.opacity ?? 0.78,
      });
      mesh = new THREE.Mesh(geometry, material) as MeshWithData;
    } else {
      const geometry = new THREE.SphereGeometry(effectiveSize, 16, 16);
      const material = new THREE.MeshStandardMaterial({
        color,
        emissive: color.clone().multiplyScalar(0.45),
        emissiveIntensity: 1.1,
        roughness: 0.4,
        metalness: 0.1,
        opacity: node.display.opacity ?? 0.9,
        transparent: node.display.opacity !== undefined && node.display.opacity < 1,
      });
      mesh = new THREE.Mesh(geometry, material) as MeshWithData;
    }

    const position = overrides?.position ?? new THREE.Vector3(node.position.x, node.position.y, node.position.z);
    mesh.position.copy(position);
    mesh.userData = { node };

    return mesh;
  }
}
