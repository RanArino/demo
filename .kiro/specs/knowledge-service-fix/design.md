# Design Document: Knowledge Service Fixes

## 1. Overview
This document defines the technical design to resolve the 6 requirements listed in `requirements.md`. The focus is on frontend UX improvements and backend data integrity enhancements.

## 2. Architecture Design
No changes to the existing architecture. Modifications are primarily limited to frontend React components and within the backend Go service.

### 2.1. Technology Stack
- **Frontend:** Next.js, TypeScript, Tailwind CSS, Shadcn UI
- **Backend (ms_knowledge):** Go, Ent, gRPC
- **Infrastructure:** Docker

## 3. Database Design
We will correct and enforce the usage of existing columns in the `spaces` table, but there are no schema changes.

---

## 4. Feature-Specific Design

### Requirement 1: Modal Display for Document List

**Target Files:**
- `frontend/src/app/spaces/components/ListView.tsx`

**Design:**
1. Add new state to `ListView.tsx` to manage document list, loading state, and error state within the modal.
    ```typescript
    const [documents, setDocuments] = useState<ContentSource[]>([]);
    const [isLoadingDocs, setIsLoadingDocs] = useState(false);
    const [errorDocs, setErrorDocs] = useState<string | null>(null);
    ```
2. Use `Dialog`'s `onOpenChange` event or `DialogTrigger`'s `onClick` event as a trigger to call a function that fetches documents.
    ```typescript
    const handleOpenDocumentDialog = async (spaceId: string) => {
      setIsLoadingDocs(true);
      setErrorDocs(null);
      const result = await listContentSources(spaceId);
      if (result.ok) {
        setDocuments(result.data);
      } else {
        setErrorDocs(result.error.message);
      }
      setIsLoadingDocs(false);
    };
    ```
3. Modify the `TableCell` with `DialogTrigger` as follows.
    ```tsx
    <Dialog onOpenChange={(open) => open && handleOpenDocumentDialog(space.id)}>
      <DialogTrigger asChild>
        <Button variant="ghost" ...>
          {space.stats?.contentCount.toString() || 0}
        </Button>
      </DialogTrigger>
      <DialogContent>
        ...
        {/* Add loading, error, and document list rendering logic here */}
      </DialogContent>
    </Dialog>
    ```
4. Within `DialogContent`, render UI based on the state of `isLoadingDocs`, `errorDocs`, and `documents`.
    - Loading: Display spinner.
    - Error: Display error message.
    - Success: Loop through `documents` array and display each document using a component similar to `ContentSourceCard`. Reuse or reference `ContentSourceCard` from `ContentSourcesSection.tsx` to create a more compact list item component within `ListView.tsx`.

### Requirement 1.1: Enhanced Document Modal UI/UX

**Target Files:**
- `frontend/src/app/spaces/components/ListView.tsx`

**Design:**
1. **Modal Size and Positioning:**
    - Modify `DialogContent` to use custom styling for 80% screen coverage with center positioning.
    ```tsx
    <DialogContent className="max-w-[80vw] max-h-[80vh] w-[80vw] h-[80vh]">
      <DialogHeader>
        <DialogTitle>Documents in {space.title}</DialogTitle>
      </DialogHeader>
      <div className="flex-1 overflow-hidden">
        {/* Table content here */}
      </div>
    </DialogContent>
    ```

2. **Table Layout Implementation:**
    - Replace the compact card display with a proper table structure similar to the spaces list.
    - Use `Table`, `TableHeader`, `TableBody`, `TableRow`, `TableHead`, and `TableCell` components from Shadcn UI.
    ```tsx
    <div className="flex-1 overflow-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Title</TableHead>
            <TableHead>Keywords</TableHead>
            <TableHead>Content Summary</TableHead>
            <TableHead>File Type</TableHead>
            <TableHead>Size</TableHead>
            <TableHead>Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {documents.map((doc) => (
            <TableRow key={doc.id}>
              <TableCell className="font-medium">{doc.title}</TableCell>
              <TableCell>{doc.keywords?.join(', ') || 'N/A'}</TableCell>
              <TableCell className="max-w-xs truncate">{doc.contentSummary || 'N/A'}</TableCell>
              <TableCell>{doc.mimeType || 'N/A'}</TableCell>
              <TableCell>{formatFileSize(doc.sizeBytes)}</TableCell>
              <TableCell>
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={() => window.open(doc.previewUrl, '_blank')}
                >
                  Open Preview
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
    ```

3. **Utility Functions:**
    - Add helper function to format file sizes in human-readable format.
    ```typescript
    const formatFileSize = (bytes: number): string => {
      if (bytes === 0) return '0 Bytes';
      const k = 1024;
      const sizes = ['Bytes', 'KB', 'MB', 'GB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };
    ```

4. **Preview Functionality:**
    - Implement preview URL generation or use existing preview functionality.
    - The preview button should open documents in a new tab/window.
    - Handle cases where preview is not available (show disabled state).

### Requirement 2: Setting `created_by` Field During Space Creation

**Target Files:**
- `ms_knowledge/internal/service/space_service.go`

**Design:**
1. Within the `CreateSpace` function, immediately after obtaining `ownerUUID`, set the same UUID to the `CreatedBy` and `LastUpdatedBy` fields of the `domain.Space` struct.
    ```go
    // ms_knowledge/internal/service/space_service.go

    func (s *SpaceService) CreateSpace(ctx context.Context, title, description string, keywords []string, icon string) (*domain.Space, error) {
        // ...
        ownerUUID, parseErr := MustGetOwnerUUID(ctx)
        if parseErr != nil {
            return nil, parseErr
        }

        space := &domain.Space{
            Title:         strings.TrimSpace(title),
            Description:   strings.TrimSpace(description),
            Keywords:      keywords,
            Icon:          strings.TrimSpace(icon),
            OwnerID:       ownerUUID,
            CreatedBy:     ownerUUID, // <--- Add
            LastUpdatedBy: ownerUUID, // <--- Add
        }

        err := s.spaceRepo.Create(ctx, space)
        // ...
    }
    ```
2. No changes needed to the `Create` method in `space_repository.go` as it already calls `SetCreatedBy`. Setting the value at the service layer is sufficient for the repository to persist it.

### Requirement 3: Resolving Conflict Between Inline Editing and Navigation in List View

**Target Files:**
- `frontend/src/app/spaces/components/ListView.tsx`

**Design:**
1. **Remove Navigation from Title Cell:**
    - Remove the `button` element within `TableCell` and replace it with `div` or `span`. Since the `onClick` event is aggregated in the outer `div` that calls `handleEdit`, the code that calls `onSelect` will be removed.
    ```tsx
    // Before change
    // <div>
    //   <button onClick={() => onSelect(space.id)} ...>
    //     {space.title}
    //   </button>
    // </div>

    // After change
    <div className="font-medium text-gray-900">
      {space.title}
    </div>
    ```
2. **Remove Actions Column:**
    - Remove the "Actions" `TableHead` from `TableHeader`.
    - Completely remove the last `TableCell` in `TableRow` within `TableBody` that contains the `Pencil` icon and `EditSpaceForm`.

### Requirement 4: Database Update for Space Statistics

**Target Files:**
- `ms_knowledge/internal/service/content_service.go`
- `ms_knowledge/internal/repository/space_repository.go`
- `ms_knowledge/internal/domain/space.go` (interface definition)

**Design:**
1. **Add Statistics Update Method to Repository Layer:**
    - Add `UpdateStats(ctx context.Context, spaceID uuid.UUID, docCountChange int, sizeChange int64) error` to the `domain.SpaceRepository` interface.
    - Implement this method in `space_repository.go`. Use Ent's `UpdateOneID` with `AddDocumentCount` and `AddTotalSizeBytes` for atomic updates.
    ```go
    // ms_knowledge/internal/repository/space_repository.go
    func (r *spaceRepository) UpdateStats(ctx context.Context, spaceID uuid.UUID, docCountChange int, sizeChange int64) error {
        return r.client.Space.UpdateOneID(spaceID).
            AddDocumentCount(docCountChange).
            AddTotalSizeBytes(sizeChange).
            Exec(ctx)
    }
    ```
2. **Call Statistics Update from Service Layer:**
    - Within `ContentService.CreateUploadURL`, call `spaceRepo.UpdateStats` immediately after `contentRepo.Create` succeeds.
    ```go
    // ms_knowledge/internal/service/content_service.go
    func (s *ContentService) CreateUploadURL(...) {
        // ...
        err = s.contentRepo.Create(ctx, content)
        if err != nil {
            return nil, "", "", time.Time{}, fmt.Errorf("failed to create content source: %w", err)
        }

        // Update statistics
        if err := s.spaceRepo.UpdateStats(ctx, spaceID, 1, sizeBytes); err != nil {
            s.logger.Printf("WARN: failed to update space stats for space %s: %v", spaceID, err)
            // Log error but don't fail the upload process itself
        }

        // ...
    }
    ```
    - Within `ContentService.DeleteContentSource`, call `spaceRepo.UpdateStats` before `contentRepo.Delete`. Get the size from the `content` to be deleted.
    ```go
    // ms_knowledge/internal/service/content_service.go
    func (s *ContentService) DeleteContentSource(ctx context.Context, id uuid.UUID) error {
        content, err := s.contentRepo.GetByID(ctx, id)
        // ...

        // Update statistics
        if err := s.spaceRepo.UpdateStats(ctx, content.SpaceID, -1, -content.SizeBytes); err != nil {
            s.logger.Printf("WARN: failed to update space stats for space %s: %v", content.SpaceID, err)
        }

        // ... Delete R2 object and DB record
    }
    ```
3. **Backfill:**
    - Implement this as a migration script or temporary admin API. For each existing space, aggregate the `contentsource` table and update the `document_count` and `total_size_bytes` in the `spaces` table.

### Requirement 5: Default Value Setting for `access_level` and UI Updates

**Target Files:**
- `ms_knowledge/internal/service/space_service.go`
- `frontend/src/app/spaces/components/SpaceCard.tsx`
- `frontend/src/app/spaces/components/ListView.tsx`
- **New:** `frontend/src/app/spaces/components/AccessLevelModal.tsx`

**Design (Backend):**
1. In `space_service.go`'s `CreateSpace`, explicitly set `AccessLevel` to `"private"` when constructing the `domain.Space`. This prevents the repository layer from overwriting it with an empty string.
    ```go
    // ms_knowledge/internal/service/space_service.go
    space := &domain.Space{
        // ...
        OwnerID:     ownerUUID,
        CreatedBy:   ownerUUID,
        LastUpdatedBy: ownerUUID,
        AccessLevel: "private", // <--- Add
    }
    ```

**Design (Frontend):**
1. **Create New `AccessLevelModal.tsx`:**
    - Create a `Dialog`-based modal.
    - Accept `space` object as a property.
    - Display options like "Private", "Public" using radio buttons or select boxes.
    - When save button is clicked, call the `updateSpaceAccessLevel` server action (newly created).
2. **Modify `SpaceCard.tsx` and `ListView.tsx`:**
    - Wrap the `access_level` badge with `DialogTrigger` and open `AccessLevelModal` when clicked.
    - Pass the `space` object to the modal.
    - Fix `getAccessIcon` and `getAccessColor` to display correctly based on `space.accessLevel` value.
    ```tsx
    // Example: ListView.tsx
    <TableCell>
      <Dialog>
        <DialogTrigger asChild>
          <Badge variant="secondary" ...>
            {getAccessIcon(space.accessLevel)}
            {space.accessLevel}
          </Badge>
        </DialogTrigger>
        <DialogContent>
          <AccessLevelModal space={space} />
        </DialogContent>
      </Dialog>
    </TableCell>
    ```
3. **Create `updateSpaceAccessLevel` Server Action:**
    - Take `spaceId` and `accessLevel` as arguments.
    - Call the backend's `UpdateSpace` gRPC method.

### Requirement 6: UX Improvement for Left Sidebar

**Target Files:**
- `frontend/src/app/spaces/[spaceId]/components/LeftSidebar.tsx`

**Design:**
1. Add a button to open the sidebar at the root of `LeftSidebar.tsx`'s JSX.
2. This button should only render when the condition `!isPinned && !isVisible` is true.
3. Set `fixed` positioning and high `z-index` for the button, placing it in the top-left.
4. Call `setIsVisible(true)` in the `onClick` event to display the sidebar.
    ```tsx
    // frontend/src/app/spaces/[spaceId]/components/LeftSidebar.tsx

    return (
      <>
        {/* Sidebar display button */}
        {!isPinned && !isVisible && (
          <Button
            variant="ghost"
            size="icon"
            className="fixed left-4 top-4 z-50 h-8 w-8 bg-white/80 backdrop-blur-sm rounded-full shadow-md hover:bg-white"
            onClick={() => setIsVisible(true)}
            aria-label="Open sidebar"
          >
            <PanelRightOpen className="h-4 w-4" />
          </Button>
        )}

        {/* Existing hover zone */}
        {!isPinned && (
          <div
            className="fixed left-0 top-0 h-full w-8 z-30"
            onMouseEnter={handleMouseEnter}
          />
        )}

        <aside ...>
          {/* ...existing sidebar content */}
        </aside>
      </>
    );
    ```
    - Need to import `PanelRightOpen` icon from `lucide-react`.
