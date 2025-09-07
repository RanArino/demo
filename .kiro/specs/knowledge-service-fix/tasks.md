# Implementation Plan: Knowledge Service Fixes

> This document identifies and manages the development tasks to be performed based on the design document.

## Feature 1: Modal Display for Document List (Requirement 1)

- [x] **1.1. Add state variables to `ListView.tsx`**
  > Add three states (`documents`, `isLoadingDocs`, `errorDocs`) to the component.
  > **Related Requirements:** 1
  > - `frontend/src/app/spaces/components/ListView.tsx`

- [x] **1.2. Implement document fetching function**
  > Implement `handleOpenDocumentDialog` function that calls the `listContentSources` server action and updates state logic.
  > **Related Requirements:** 1
  > - `frontend/src/app/spaces/components/ListView.tsx`

- [x] **1.3. Modify `Dialog` trigger and content**
  > Add `onOpenChange` to the `Dialog` component to call `handleOpenDocumentDialog`.
  > **Related Requirements:** 1
  > - `frontend/src/app/spaces/components/ListView.tsx`

- [x] **1.4. Implement modal UI**
  > Implement UI within `DialogContent` that switches between loading, error, and document list displays. Create a compact card component for document display.
  > **Related Requirements:** 1
  > - `frontend/src/app/spaces/components/ListView.tsx`

- [x] **1.5. Enhance modal size and positioning**
  > Modify `DialogContent` to use 80% screen width and height with center positioning.
  > **Related Requirements:** 1.1
  > - `frontend/src/app/spaces/components/ListView.tsx`

- [x] **1.6. Implement table layout for documents**
  > Replace compact card display with proper table structure using Shadcn UI Table components.
  > **Related Requirements:** 1.1
  > - `frontend/src/app/spaces/components/ListView.tsx`

- [x] **1.7. Add document table columns**
  > Implement table columns: Title, Keywords, Content Summary, MIME Type, Size, and Preview Button.
  > **Related Requirements:** 1.1
  > - `frontend/src/app/spaces/components/ListView.tsx`

- [x] **1.8. Implement file size formatting utility**
  > Add helper function to format file sizes in human-readable format (Bytes, KB, MB, GB).
  > **Related Requirements:** 1.1
  > - `frontend/src/app/spaces/components/ListView.tsx`

- [x] **1.9. Implement preview functionality**
  > Add preview button that opens documents in new tab/window with proper error handling.
  > **Related Requirements:** 1.1
  > - `frontend/src/app/spaces/components/ListView.tsx`

## Feature 2: Backend Fixes (Requirements 2, 4, 5)

- [ ] **2.1. Fix `space_service.go` (created_by)**
  > Add code to set `CreatedBy` and `LastUpdatedBy` in the `domain.Space` struct within the `CreateSpace` function.
  > **Related Requirements:** 2

- [ ] **2.2. Fix `space_service.go` (access_level)**
  > Explicitly set `AccessLevel` to `"private"` in the `domain.Space` struct within the `CreateSpace` function.
  > **Related Requirements:** 5

- [ ] **2.3. Update `domain.SpaceRepository` interface**
  > Add `UpdateStats` method definition to `domain/space.go`.
  > **Related Requirements:** 4

- [ ] **2.4. Implement `UpdateStats` in `space_repository.go`**
  > Implement logic to atomically update statistics using Ent's `AddDocumentCount` and `AddTotalSizeBytes`.
  > **Related Requirements:** 4

- [ ] **2.5. Fix `content_service.go`'s `CreateUploadURL`**
  > Implement processing to call `spaceRepo.UpdateStats`, increment document count by +1, and add size.
  > **Related Requirements:** 4

- [ ] **2.6. Fix `content_service.go`'s `DeleteContentSource`**
  > Implement processing to call `spaceRepo.UpdateStats`, decrement document count by -1, and subtract size.
  > **Related Requirements:** 4

- [ ] **2.7. Create database backfill script**
  > Prepare a one-time script or admin functionality to fix statistics (`document_count`, `total_size_bytes`) and `access_level` for existing spaces.
  > **Related Requirements:** 4, 5

- [ ] **2.8. Fix `content_repository.go` missing `size_bytes` field**
  > Add missing `SetSizeBytes(content.SizeBytes)` in Create method and `SizeBytes` field mapping in GetByID and List methods.
  > **Related Requirements:** 7
  > - `ms_knowledge/internal/repository/content_repository.go`

## Feature 3: Frontend UI/UX Improvements (Requirements 3, 5, 6)

- [ ] **3.1. Fix inline editing in `ListView.tsx`**
  > Remove navigation `button` from title cell and make it enter edit mode only on click.
  > **Related Requirements:** 3

- [ ] **3.2. Remove Actions column from `ListView.tsx`**
  > Remove `TableHead` and `TableCell` related to Actions column from `TableHeader` and `TableBody`.
  > **Related Requirements:** 3

- [ ] **3.3. Create new `AccessLevelModal.tsx` component**
  > Implement modal UI and logic for changing access level.
  > **Related Requirements:** 5

- [ ] **3.4. Create `updateSpaceAccessLevel` server action**
  > Implement server action called from frontend that hits the backend gRPC endpoint.
  > **Related Requirements:** 5

- [ ] **3.5. Integrate modal into `ListView.tsx` and `SpaceCard.tsx`**
  > Modify to open `AccessLevelModal` when access level badge is clicked.
  > **Related Requirements:** 5

- [ ] **3.6. Add display button to `LeftSidebar.tsx`**
  > Implement floating button that appears when sidebar is hidden and not pinned.
  > **Related Requirements:** 6

## General Tasks

- [ ] **4.1. Verification and Testing**
  > Verify that all fixes work as per requirements in the local environment.

- [ ] **4.2. Create Pull Request and Review**
  > Create a pull request summarizing the fixes and receive team review.
