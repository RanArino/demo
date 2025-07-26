# GEMINI.md

This file provides guidance to Gemini CLI when working with code in this repository.

## Project Overview

This is a Knowledge Exploration and Information Structuring Platform that hierarchically organizes and visualizes large amounts of text/document information in interactive 2D/3D spaces. The project aims to solve information overload problems from linear text formats (PDFs, Markdown, AI chat histories).

## Engineering Philosophy

The engineering philosophy for this project is defined in `@/notes/dev/guides/coding-philosophy.md`. The core principles are:
1.  **Write human-understandable code.**
2.  **Let the code speak for itself.**
3.  **The core question of code review: "Can I maintain this code without any issues?"**
4.  **Embrace pure functions and log only what is necessary.**

## Architecture

This is a monorepo with a microservices architecture. Each microservice is located in a directory prefixed with `ms_` (e.g., `/ms_user`). The `/backend-go` and `/backend-py` directories serve as templates.

For detailed architecture information, please visit the `@/notes/dev/guides` directory which contains comprehensive guides for each component:

- **Frontend (`/frontend`)**: A Next.js 15 application. See `@/notes/dev/guides/frontend-guide.md` for detailed architecture.
- **Backend Services**: Follow a gRPC-First architecture for both Go and Python. See `@/notes/dev/guides/backend-guide.md` for implementation details.
- **Design & Architecture**: See `@/notes/dev/guides/design.md` for overall system design and architecture patterns.
- **Infrastructure**: Docker Compose for orchestration.

## Implementation Checklists & Completion Reports

All implementation checklists and completion reports for development are managed centrally in the `@/notes/dev/implementations/` directory as a single source of truth. Each file in this folder is labeled with a 3-digit number prefix (e.g., `001_...`, `002_...`).

- Each file starts with a checklist of tasks.
- When user asks any revision of the implementation, you should first check the checklist, then add/modify/delete the checkboxes as needed.
- As each task is completed, its checkbox is checked and the paths of the modified files are listed under the checkbox.
- This provides a clear, step-by-step record of implementation progress and file changes.

## Error Reporting Policy

If the system fails to fix the same error more than three times, it must:
- Analyze what happened and summarize the approaches tried so far.
- Identify the root errors.
- Write this analysis to `@/notes/dev/implementations/999_temp_error_report.md` for review.

This ensures persistent errors are documented for human intervention.

## Development Commands

### Frontend Development
```bash
cd frontend
npm install              # Install dependencies
npm run dev             # Start dev server with Turbopack (http://localhost:3000)
npm run build           # Production build
npm run lint            # Run ESLint
npm start               # Start production server
```

### Backend Go Development
To work on a Go microservice, navigate to its directory (e.g., `cd ms_user`).
```bash
go mod download         # Install dependencies
go run ./cmd/server/main.go # Run the server
go build -o app ./cmd/server/main.go # Build binary
go test ./...          # Run tests
```

### Backend Python Development
To work on a Python microservice, navigate to its directory (e.g., `cd ms_ml`).
```bash
cd ms_ml # (or other Python microservice)
pip install -r requirements.txt
python app/main.py
```

### Full Stack Development
```bash
# Start all services with Docker Compose
docker-compose up -d

# Start specific services
docker-compose up backend-go frontend

# Rebuild containers after changes
docker-compose build
docker-compose up --build

# View logs
docker-compose logs -f [service-name]

# Stop all services
docker-compose down
```

### Frontend Development
```bash
cd frontend
npm install
npm run dev
```

### Go Microservice Development
```bash
cd ms_user # (or other Go microservice)
go mod download
go run ./cmd/server/main.go
```

### Python Microservice Development
```bash
cd ms_ml # (or other Python microservice)
pip install -r requirements.txt
python app/main.py
```

### Full Stack Development
```bash
# Start all services with Docker Compose
docker-compose up -d

# View logs for a specific service
docker-compose logs -f [service-name]

# Stop all services
docker-compose down
```
