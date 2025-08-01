# **Development Workflows**

## **Go Backend Development**

### **Project Structure & Development Flow:**

1. Define the proto file under api/proto/v{version\_num}/.  
2. Generate gRPC and protobuf code.  
   protoc \--go\_out=. \--go\_opt=paths=source\_relative \\  
       \--go-grpc\_out=. \--go-grpc\_opt=paths=source\_relative \\  
       api/proto/v1/{service}.proto

3. Define the Ent schema under ent/schema/.  
4. Generate the Ent ORM code.  
   go generate ./ent

5. Define repository interfaces under internal/repository/.  
6. Define service interfaces and logic under internal/service/.  
7. Implement gRPC and/or WebSocket handlers under internal/server/.  
8. Define the main application entry point under cmd/server/main.go.  
9. Write tests for each layer.

### **Development Commands:**

cd go-service-name  
go mod tidy             \# Install/tidy dependencies  
go run ./cmd/server/main.go \# Run the server  
go build \-o app ./cmd/server/main.go \# Build binary  
go test ./...          \# Run tests

## **Python Backend Development (FastAPI)**

### **Project Structure & Development Flow:**

1. Define the proto file under api/proto/v{version\_num}/ (if gRPC is needed).  
2. Define Pydantic models under app/domain/ (for business entities).  
3. Define database schemas under app/repository/ (e.g., SQLAlchemy models).  
4. Implement the repository layer under app/repository/.  
5. Implement the service layer under app/service/.  
6. Implement API handlers (gRPC, WebSocket, or HTTP Routes) under app/server/.  
7. Define the main application under app/main.py.  
8. Define configuration under app/config.py.  
9. Write tests under tests/.

### **Development Commands:**

cd python-service-name  
python \-m venv venv     \# Create virtual environment  
source venv/bin/activate \# Activate virtual environment (Linux/Mac)  
\# venv\\Scripts\\activate  \# Activate virtual environment (Windows)  
pip install \-r requirements.txt \# Install dependencies  
uvicorn app.main:app \--reload \--host 0.0.0.0 \--port 8000 \# Run with hot reload  
pytest tests/          \# Run tests  
pip freeze \> requirements.txt \# Update dependencies

## **Frontend Development (Next.js)**

### **Project Structure & Development Flow:**

1. Define components under src/components/ and src/features/.  
2. Define pages/routes under src/app/ (App Router).  
3. Define all API integration logic under src/api/.  
   * api/generated/: Output for generated Protobuf TS code.  
   * api/actions/: All Next.js Server Actions.  
   * api/server-client.ts: Server-side gRPC client singleton.  
   * api/client.ts: Client-side factories for gRPC-web and WebSockets.  
4. Define middleware under middleware.ts.  
5. Define global styles under src/app/globals.css.

### **Development Commands:**

cd frontend  
npm install              \# Install dependencies  
npm run dev             \# Start dev server (http://localhost:3000)  
npm run build           \# Production build  
npm run start           \# Start production server  
npm run lint            \# Run ESLint  
npm run proto:gen       \# Generate gRPC/Protobuf code

### **Protocol Buffer Code Generation:**

This command should be run from the frontend directory. It generates universal TypeScript code that can be used by both the server and client.

\# In package.json  
"scripts": {  
  "proto:gen": "npx @bufbuild/buf generate"  
}

\# Assumes a buf.gen.yaml file in the frontend root:  
\# version: v1  
\# plugins:  
\#   \- plugin: es  
\#     out: src/api/generated  
\#   \- plugin: grpc-web  
\#     out: src/api/generated  
\#     opt:  
\#       \- import\_style=typescript

## **Multi-Service Development**

### **Full Stack Development Commands:**

\# Start all services with Docker Compose in detached mode  
docker-compose up \-d

\# Start specific services  
docker-compose up \-d ms\_user frontend envoy

\# Rebuild containers after changes  
docker-compose up \-d \--build

\# View logs for a specific service  
docker-compose logs \-f \[service-name\]

\# Stop all services  
docker-compose down

## **Testing Workflows**

### **Go Backend Testing:**

go test ./... \-v        \# Run all tests with verbose output  
go test ./internal/service \-v \# Test specific package  
go test \-race ./...     \# Test with race condition detection  
go test \-cover ./...    \# Test with coverage report

### **Python Backend Testing:**

pytest tests/ \-v       \# Run all tests with verbose output  
pytest tests/test\_service.py \-v \# Test specific file  
pytest \--cov=app tests/ \# Test with coverage report

### **Frontend Testing:**

npm test                \# Run Jest/Vitest tests  
npm run test:watch     \# Run tests in watch mode  
npm run test:coverage  \# Run tests with coverage  
npm run e2e            \# Run end-to-end tests (e.g., Playwright/Cypress)  
