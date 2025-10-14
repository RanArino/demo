import { Vector3 } from 'three';

/**
 * Generate quasi-uniform points on a disk using a Fibonacci spiral.
 * Positions are distributed on the XZ plane (Y taken from the provided center).
 */
export function generateFibonacciDiskPositions(count: number, radius: number, center: Vector3): Vector3[] {
  if (count <= 0) {
    return [];
  }

  if (count === 1) {
    return [center.clone()];
  }

  const positions: Vector3[] = [];
  const goldenAngle = Math.PI * (3 - Math.sqrt(5));
  const baseY = center.y;

  for (let index = 0; index < count; index += 1) {
    const radialFactor = Math.sqrt((index + 0.5) / count);
    const distance = radius * radialFactor;
    const theta = goldenAngle * index;
    const x = Math.cos(theta) * distance;
    const z = Math.sin(theta) * distance;
    positions.push(new Vector3(center.x + x, baseY, center.z + z));
  }

  return positions;
}
