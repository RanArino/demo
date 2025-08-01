# **Frontend Development Guide (Hybrid Architecture)**

This document outlines a high-performance, hybrid architecture for the Next.js frontend. It leverages **Next.js Server Actions** for secure data fetching and mutations, combined with the best real-time technology—**gRPC-web** or **WebSockets**—for each specific use case.

## **Architecture Principles**

1. **Server-First for Security**: Handle all initial data loads and mutations via Server Actions. This keeps backend addresses and credentials off the client, reducing the attack surface.  
2. **The Right Tool for the Job**: Use request-response patterns (Server Actions) for discrete events and choose the best persistent streaming connection (gRPC-web or WebSockets) for continuous data flows.  
3. **Performance & User Experience**: Optimize for fast initial loads using server-side logic, and smooth, real-time interactions using low-latency streaming.  
4. **End-to-End Type Safety**: Use code generated from .proto files across the entire stack—from the backend to Server Actions to the client.

## **Architecture Overview**

This model separates communication into two distinct patterns, each with its preferred technology.

* **Request-Response (The Default)**  
  * **Technology**: Next.js Server Actions with a server-side @grpc/grpc-js client.  
  * **Use Cases**: Initial page loads, form submissions, creating/updating/deleting data.  
* **Real-time Streaming (The Exception)**  
  * **Technology**: Client-side **gRPC-web** for one-way streams or **WebSockets** for two-way streams.  
  * **Use Cases**: Live visualizations, collaborative editing, chat.

### **Directory Structure**

This structure centralizes all external communication logic under a single api/ directory.

frontend/src/  
├── api/  
│   ├── generated/         # Universal TS code from .proto files  
│   ├── actions/           # All Server Actions  
│   │   └── canvasActions.ts  
│   ├── client.ts          # Exports CLIENT-side factories (gRPC-web, WS)  
│   └── server-client.ts   # Exports SERVER-side gRPC client singleton  
├── app/                   # Next.js App Router (Pages & Layouts)  
├── components/            # Reusable UI Components  
├── features/              # Components & logic for specific features  
└── hooks/                 # Custom React Hooks (Client-side)

## **Development Flow**

### **1. Protocol Buffer Code Generation**

Use protobuf-ts to generate universal TypeScript code. The output should be configured to go into src/api/generated/.

# In your package.json scripts  
"scripts": {  
  "proto:gen": "npx @bufbuild/buf generate"  
}

### **2. Configure Server-Side gRPC Client**

Create a singleton instance of the gRPC client for your Server Actions to use.

**src/api/server-client.ts**

import { GrpcTransport } from " @protobuf-ts/grpc-transport";  
import { YourServiceClient } from "./generated/service.client"; // Generated client

// Singleton to prevent creating new connections on every action  
let client: YourServiceClient | null = null;

export function getGrpcClient() {  
  if (!client) {  
    const transport = new GrpcTransport({  
      host: process.env.GRPC_BACKEND_URL!, // e.g., "localhost:50051"  
    });  
    client = new YourServiceClient(transport);  
  }
  return client;  
}

### **3. Implement Server Actions**

Create actions that use the gRPC client to fetch or mutate data.

**src/api/actions/canvasActions.ts**

"use server";  
import { getGrpcClient } from " @/api/server-client";  
import { revalidatePath } from "next/cache";

// Action to get the initial state of a canvas  
export async function getInitialCanvasState(canvasId: string) {  
  const client = getGrpcClient();  
  try {  
    const response = await client.getCanvas({ canvasId });  
    return { success: true, data: response.response };  
  } catch (error: any) {  
    return { success: false, error: error.message };  
  }
}

// Action to create a new object on the canvas  
export async function createObject(prevState: any, formData: FormData) {  
    const client = getGrpcClient();  
    const canvasId = formData.get('canvasId') as string;  
    // ... get other form data

    try {  
        await client.createObject({ canvasId /* ...other data */ });  
        revalidatePath(`/canvas/${canvasId}`);  
        return { success: true };  
    } catch (error: any) {  
        return { success: false, error: error.message };  
    }  
}

### **4. Configure Client-Side Streaming**

Set up clients for gRPC-web and WebSockets.

**src/api/client.ts**

import { GrpcWebFetchTransport } from " @protobuf-ts/grpcweb-transport";  
import { YourServiceClient } from "./generated/service.client";  
import { io, Socket } from "socket.io-client";

// For ONE-WAY server push (Read-Only Canvas)  
export function getGrpcWebClient() {  
  const transport = new GrpcWebFetchTransport({  
    baseUrl: process.env.NEXT_PUBLIC_GRPC_WEB_URL!, // e.g., "http://localhost:8080"  
  });  
  return new YourServiceClient(transport);  
}

// For TWO-WAY interaction (Collaborative Canvas, Chat)  
export function getWebSocket(path: string): Socket {  
    // In a real app, manage connection state to avoid multiple sockets  
    return io(process.env.NEXT_PUBLIC_WEBSOCKET_URL!, { path });  
}

### **5. Build a Hybrid Component**

This example shows a read-only canvas that uses a Server Action for the initial load and a gRPC-web stream for live updates.

**src/features/canvas/ReadOnlyCanvas.tsx**

"use client";

import { useEffect, useState } from "react";  
import { getInitialCanvasState } from " @/api/actions/canvasActions";  
import { getGrpcWebClient } from " @/api/client";  
import { ThreeScene } from "./ThreeScene"; // Your 3D component

export function ReadOnlyCanvas({ canvasId }: { canvasId: string }) {  
  const [initialData, setInitialData] = useState(null);

  // 1. Initial Load via Server Action  
  useEffect(() => {  
    getInitialCanvasState(canvasId).then(result => {  
      if (result.success) {  
        setInitialData(result.data);  
      }  
    });  
  }, [canvasId]);

  // 2. Real-time Updates via gRPC-web Stream  
  useEffect(() => {  
    if (!initialData) return; // Don't connect until initial data is loaded

    const client = getGrpcWebClient();  
    const stream = client.streamCanvasUpdates({ canvasId });

    const listen = async () => {  
      for await (const update of stream.responses) {  
        // Update your 3D scene's state with the new data  
        console.log("Received position update:", update);  
      }  
    };

    listen();

    return () => stream.cancel(); // Clean up on unmount  
  }, [initialData, canvasId]);

  if (!initialData) {  
    return <div>Loading Canvas...</div>;  
  }

  return <ThreeScene data={initialData} />;
}

## **Backend Dependencies**

Your frontend choices have direct implications for the backend.

* **To support grpc-web:** You must configure a proxy (like **Envoy** or Nginx) or a server middleware to translate grpc-web requests into standard gRPC for your service. Your core service logic does not change.  
* **To support WebSockets:** This requires a **new, dedicated service on your backend**. It must handle WebSocket protocol upgrades, manage connection state (e.g., rooms, users), and contain the logic to process and broadcast messages.

## **Environment Variables**

# Used ONLY on the server (in Server Actions). Not prefixed with NEXT_PUBLIC_.  
GRPC_BACKEND_URL=localhost:50051

# Exposed to the browser for client-side streaming connections.  
NEXT_PUBLIC_GRPC_WEB_URL=http://localhost:8080  
NEXT_PUBLIC_WEBSOCKET_URL=http://localhost:3001

## **Required Dependencies**

* @protobuf-ts/grpc-transport: For server-side gRPC calls.  
* @protobuf-ts/grpcweb-transport: For client-side gRPC-web calls.  
* @bufbuild/buf: For code generation from .proto files.  
* socket.io-client: For WebSocket communication.  
* three.js: For 3D rendering.