import * as grpc from '@grpc/grpc-js';
import { UserServiceClient } from './generated/v1/user_grpc_pb';

/**
 * Singleton gRPC client for server-side operations
 * This client is used in server actions and API routes
 */
class GRPCClientManager {
  private static instance: UserServiceClient | null = null;

  static getInstance(): UserServiceClient {
    if (!this.instance) {
      const grpcUrl = process.env.MS_USER_GRPC_URL_INTERNAL || 'localhost:50051';
      
      // Create credentials for internal communication
      // In production, this should use TLS, but for internal services we can use insecure
      const credentials = grpc.credentials.createInsecure();
      
      this.instance = new UserServiceClient(grpcUrl, credentials);
    }
    
    return this.instance;
  }

  static close(): void {
    if (this.instance) {
      this.instance.close();
      this.instance = null;
    }
  }
}

export const getUserServiceClient = (): UserServiceClient => {
  return GRPCClientManager.getInstance();
};

export const closeGRPCClient = (): void => {
  GRPCClientManager.close();
};