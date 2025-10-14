import { createPromiseClient, PromiseClient } from '@bufbuild/connect';
import { createGrpcTransport } from '@bufbuild/connect-node';
import fs from 'node:fs';
import { UserService } from './generated/v1/user_connectweb';
import { KnowledgeService } from './generated/v1/knowledge_connectweb';
import { CanvasPublic } from './generated/v1/canvas_connectweb';

/**
 * Singleton gRPC client manager for server-side operations
 * These clients are used in server actions and API routes
 */
class GRPCClientManager {
  private static userInstance: PromiseClient<typeof UserService> | null = null;
  private static knowledgeInstance: PromiseClient<typeof KnowledgeService> | null = null;
  private static canvasInstance: PromiseClient<typeof CanvasPublic> | null = null;

  // Resolve a gRPC base URL with sensible fallbacks and IPv6-safe localhost handling
  private static resolveBaseUrl(
    envExternalKey: string,
    envInternalKey: string,
    defaultPort: number,
    defaultServiceHost: string,
  ): string {
    // Prefer external first (when running frontend outside Docker), then internal (docker-compose service name)
    const external = (process.env as Record<string, string | undefined>)[envExternalKey];
    const internal = (process.env as Record<string, string | undefined>)[envInternalKey];
    const raw =
      external ||
      internal ||
      // If neither env is provided, pick a sensible default depending on containerization
      (GRPCClientManager.isInDocker()
        ? `http://${defaultServiceHost}:${defaultPort}` // default service name in docker-compose
        : `http://127.0.0.1:${defaultPort}`); // IPv4 loopback to avoid ::1

    const ensureHttp = (url: string) => (url.startsWith('http://') || url.startsWith('https://') ? url : `http://${url}`);
    const normalized = ensureHttp(raw);
    try {
      let u = new URL(normalized);
      // Avoid IPv6 localhost (::1) which often rejects when servers bind IPv4 only
      if (u.hostname === 'localhost' || u.hostname === '::1' || u.hostname === '[::1]') {
        u.hostname = '127.0.0.1';
      }
      // If running in Docker and external env pointed to loopback, prefer internal service name
      if (GRPCClientManager.isInDocker() && external && (u.hostname === '127.0.0.1')) {
        const internalUrl = internal ? ensureHttp(internal) : `http://${defaultServiceHost}:${defaultPort}`;
        u = new URL(internalUrl);
      }
      // Strip trailing slash for connect-node transport expectations
      return u.toString().replace(/\/$/, '');
    } catch {
      // Fallback conservatively
      return normalized;
    }
  }

  private static isInDocker(): boolean {
    try {
      return fs.existsSync('/.dockerenv');
    } catch {
      return false;
    }
  }

  static getUserInstance(): PromiseClient<typeof UserService> {
    if (!this.userInstance) {
      const fullUrl = this.resolveBaseUrl('MS_USER_GRPC_URL', 'MS_USER_GRPC_URL_INTERNAL', 50051, 'ms_user');

      const transport = createGrpcTransport({
        httpVersion: '2',
        baseUrl: fullUrl,
      });
      
      this.userInstance = createPromiseClient(UserService, transport);
    }
    
    return this.userInstance;
  }

  static getKnowledgeInstance(): PromiseClient<typeof KnowledgeService> {
    if (!this.knowledgeInstance) {
      const fullUrl = this.resolveBaseUrl('MS_KNOWLEDGE_GRPC_URL', 'MS_KNOWLEDGE_GRPC_URL_INTERNAL', 50052, 'ms_knowledge');

      const transport = createGrpcTransport({
        httpVersion: '2',
        baseUrl: fullUrl,
      });

      this.knowledgeInstance = createPromiseClient(KnowledgeService, transport);
    }

    return this.knowledgeInstance;
  }

  static getCanvasInstance(): PromiseClient<typeof CanvasPublic> {
    if (!this.canvasInstance) {
      const fullUrl = this.resolveBaseUrl('MS_CANVAS_GRPC_URL', 'MS_CANVAS_GRPC_URL_INTERNAL', 50055, 'ms_canvas');

      const transport = createGrpcTransport({
        httpVersion: '2',
        baseUrl: fullUrl,
      });

      this.canvasInstance = createPromiseClient(CanvasPublic, transport);
    }

    return this.canvasInstance;
  }
}

export const getUserServiceClient = (): PromiseClient<typeof UserService> => {
  return GRPCClientManager.getUserInstance();
};

export const getKnowledgeServiceClient = (): PromiseClient<typeof KnowledgeService> => {
  return GRPCClientManager.getKnowledgeInstance();
};

export const getCanvasServiceClient = (): PromiseClient<typeof CanvasPublic> => {
  return GRPCClientManager.getCanvasInstance();
};

// No close method is needed for the new clients.
