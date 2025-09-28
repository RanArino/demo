import { createPromiseClient, PromiseClient } from '@bufbuild/connect';
import { createGrpcTransport } from '@bufbuild/connect-node';
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

  static getUserInstance(): PromiseClient<typeof UserService> {
    if (!this.userInstance) {
      const grpcUrl = process.env.MS_USER_GRPC_URL_INTERNAL || 'http://localhost:50051';
      const fullUrl = grpcUrl.startsWith('http') ? grpcUrl : `http://${grpcUrl}`;
      
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
      const grpcUrl = process.env.MS_KNOWLEDGE_GRPC_URL_INTERNAL || 'http://localhost:50052';
      const fullUrl = grpcUrl.startsWith('http') ? grpcUrl : `http://${grpcUrl}`;

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
      const grpcUrl = process.env.MS_CANVAS_GRPC_URL_INTERNAL || 'http://localhost:50055';
      const fullUrl = grpcUrl.startsWith('http') ? grpcUrl : `http://${grpcUrl}`;

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
