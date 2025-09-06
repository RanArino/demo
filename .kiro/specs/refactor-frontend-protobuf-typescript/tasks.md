# Implementation Tasks

## Phase 1: Cleanup and Preparation

- [x] Delete the old JavaScript generated files (`*.js` and `*_grpc_pb.js`) from `frontend/src/api/generated/v1`.
- [x] Add a `proto:gen` script to `frontend/package.json` to execute `buf generate` from the `frontend` directory.

## Phase 2: Refactoring

- [x] Search the entire `frontend/src` directory for imports from `knowledge_pb.js` and `user_pb.js`.
- [x] For each file found, refactor the code to use the new TypeScript generated files (`knowledge_pb.ts`, `user_pb.ts`, `knowledge_connectweb.ts`, `user_connectweb.ts`).
  - [x] Update import paths.
  - [x] Update message creation and field access to be compatible with the new TypeScript types.
  - [x] Update gRPC client usage to use the new `connect-web` clients.
- [x] **Component Refactoring:** Update components to use generated types.
  - [x] Refactor `frontend/src/components/content/ContentSourceCard.tsx`.
  - [x] Refactor `frontend/src/app/spaces/components/SpaceCard.tsx`.
  - [x] Refactor `frontend/src/app/spaces/[spaceId]/components/ContentSourcesSection.tsx`.
- [x] **Action Refactoring:** Update server actions to use generated types.
  - [x] Refactor `frontend/src/api/actions/contentActions.ts`.
  - [x] Refactor `frontend/src/api/actions/spaceActions.ts`.
- [x] **Hook Refactoring:** Update hooks to use generated types.
  - [x] Refactor `frontend/src/app/spaces/hooks/useProcessingPoller.ts`.
  - [x] Refactor `frontend/src/app/spaces/hooks/useUpload.ts`.
- [x] **Type Definition Cleanup:** Remove all files in `frontend/src/app/spaces/types/` and update `frontend/src/app/spaces/types/index.ts` to export only the remaining necessary types.

- [x] **Global:** Delete the `frontend/src/app/spaces/types` directory.

## Phase 3: Verification

- [x] Run `npm install` in the `frontend` directory to make sure all dependencies are up to date.
- [x] Run the `proto:gen` script to ensure the TypeScript files are generated correctly.
- [x] Run `npm run dev` to build and start the frontend application.
- [x] Manually test the application to ensure all features that depend on the protobuf objects are working correctly.
- [x] Run `npm run lint` to check for any linting errors.
