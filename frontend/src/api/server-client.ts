import * as grpc from '@grpc/grpc-js';
import { UserServiceClient } from './generated/v1/user_grpc_pb';
import { KnowledgeServiceClient } from './generated/v1/knowledge_grpc_pb';

/**
 * Singleton gRPC client manager for server-side operations
 * These clients are used in server actions and API routes
 */
class GRPCClientManager {
  private static userInstance: UserServiceClient | null = null;
  private static knowledgeInstance: KnowledgeServiceClient | null = null;

  static getUserInstance(): UserServiceClient {
    if (!this.userInstance) {
      const grpcUrl = process.env.MS_USER_GRPC_URL_INTERNAL || 'localhost:50051';
      
      // Create credentials for internal communication
      // In production, this should use TLS, but for internal services we can use insecure
      const credentials = grpc.credentials.createInsecure();
      
      this.userInstance = new UserServiceClient(grpcUrl, credentials);
    }
    
    return this.userInstance;
  }

  static getKnowledgeInstance(): KnowledgeServiceClient {
    if (!this.knowledgeInstance) {
      const grpcUrl = process.env.MS_KNOWLEDGE_GRPC_URL_INTERNAL || 'localhost:50052';
      
      // Create credentials for internal communication
      const credentials = grpc.credentials.createInsecure();
      
      this.knowledgeInstance = new KnowledgeServiceClient(grpcUrl, credentials);
    }
    
    return this.knowledgeInstance;
  }

  static closeAll(): void {
    if (this.userInstance) {
      this.userInstance.close();
      this.userInstance = null;
    }
    if (this.knowledgeInstance) {
      this.knowledgeInstance.close();
      this.knowledgeInstance = null;
    }
  }
}

export const getUserServiceClient = (): UserServiceClient => {
  return GRPCClientManager.getUserInstance();
};

export const getKnowledgeServiceClient = (): KnowledgeServiceClient => {
  return GRPCClientManager.getKnowledgeInstance();
};

export const closeGRPCClient = (): void => {
  GRPCClientManager.closeAll();
};