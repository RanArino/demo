import { Vector3 } from 'three';

/**
 * Generate evenly distributed points on a sphere surface.
 * Based on the Fibonacci sphere algorithm described in canvas-ui docs.
 */
export function generateFibonacciSpherePositions(
  count: number,
  radius: number,
  center: Vector3
): Vector3[] {
  if (count <= 0) {
    return [];
  }

  if (count === 1) {
    return [center.clone()];
  }

  const positions: Vector3[] = [];
  const goldenAngle = Math.PI * (3 - Math.sqrt(5));

  for (let index = 0; index < count; index += 1) {
    const y = 1 - (index / (count - 1)) * 2;
    const radiusAtY = Math.sqrt(Math.max(0, 1 - y * y));
    const theta = goldenAngle * index;

    const x = Math.cos(theta) * radiusAtY;
    const z = Math.sin(theta) * radiusAtY;

    positions.push(
      new Vector3(
        center.x + x * radius,
        center.y + y * radius,
        center.z + z * radius
      )
    );
  }

  return positions;
}
