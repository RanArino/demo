

## (1) Document list modal via page `/spaces' is not working.

Observation:
- The dialog opened from the "Documents" column shows a placeholder instead of the actual list.
- Location: `frontend/src/app/spaces/components/ListView.tsx`.
- Exact block to replace: dialog content under the Documents column.

Pin‑point code section:

```338:356:frontend/src/app/spaces/components/ListView.tsx
                    <TableCell className="text-center">
                      <Dialog>
                        <DialogTrigger asChild>
                          <Button variant="ghost" className="h-8 px-2 text-blue-600 hover:text-blue-800 hover:bg-blue-50" aria-label={`View documents for ${space.title}`}>
                            {space.documentCount || 0}
                          </Button>
                        </DialogTrigger>
                        <DialogContent aria-describedby={undefined}>
                          <DialogHeader>
                            <DialogTitle>Documents in {space.title}</DialogTitle>
                          </DialogHeader>
                          <div className="text-center py-8">
                            <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                            <p className="text-gray-500">
                              Document list view will be implemented here
                            </p>
                          </div>
                        </DialogContent>
                      </Dialog>
                    </TableCell>
```

Fix approach:
- On dialog open, fetch documents for the space using `listContentSources(space.id)` from `frontend/src/api/actions/contentActions.ts`.
- Render items (reuse `ContentSourceCard` from `frontend/src/app/spaces/[spaceId]/components/ContentSourcesSection.tsx`, or a compact list) and include loading/error states.

---

## (2) At space creation, `created_by` field is not set. Suppose that the sharing feature among various public users will be implemented in the future; in this case, `owner_id` and `created_by` will be different.

Observation:
- The backend service builds a `domain.Space` in `ms_knowledge/internal/service/space_service.go#CreateSpace` but only sets `OwnerID`; it never sets `CreatedBy` nor `LastUpdatedBy`.
- The repository write `ms_knowledge/internal/repository/space_repository.go#Create` persists `created_by` from `space.CreatedBy`. Since the service leaves it zero, DB ends up with a zero UUID.
- The ent schema requires `created_by` (non-nillable), confirming it should be set: `ms_knowledge/ent/schema/space.go`.

Pin‑point code sections:

```42:58:ms_knowledge/internal/service/space_service.go
func (s *SpaceService) CreateSpace(ctx context.Context, title, description string) (*domain.Space, error) {
    // ...
    ownerUUID, parseErr := MustGetOwnerUUID(ctx)
    // ...
    space := &domain.Space{
        Title:       strings.TrimSpace(title),
        Description: strings.TrimSpace(description),
        OwnerID:     ownerUUID,
    }
    // CreatedBy / LastUpdatedBy NOT set
}
```

```33:48:ms_knowledge/internal/repository/space_repository.go
create := r.client.Space.Create().
    // ...
    SetOwnerID(space.OwnerID).
    SetCreatedAt(space.CreatedAt).
    SetCreatedBy(space.CreatedBy).   // expects value from service
    SetLastUpdatedAt(space.LastUpdatedAt)
```

Root cause:
- Service layer omits assigning `CreatedBy` and `LastUpdatedBy`; repository forwards zero-value to DB.

Fix approach:
- In `CreateSpace`, set `CreatedBy` and `LastUpdatedBy` to the caller’s internal UUID resolved from context.
- Optionally extend proto/response mapping later, but DB write is the critical fix.

Proposed edit:
- In `ms_knowledge/internal/service/space_service.go` within `CreateSpace` add:
  - `CreatedBy: ownerUUID,`
  - `LastUpdatedBy: ownerUUID,`

Follow‑ups:
- Consider returning these fields via gRPC if needed; current `domainSpaceToProto` doesn’t expose them, which is fine for now.

---

## (3) List view inline editing vs navigation (title/description/keywords)

Observation:
- In the `/spaces` list, title, description, and keywords are editable inline.
- Clicking the title currently navigates to the space page, which prevents editing the title directly.
- There is also an Actions column with an edit modal, duplicating the inline edit capability.

Pin‑point code sections:

```259:283:frontend/src/app/spaces/components/ListView.tsx
                    <TableCell>
                      <div
                        onClick={() => handleEdit(space.id, 'title', space.title)}
                        className="cursor-pointer hover:bg-blue-50 rounded px-2 py-1 -mx-2 -my-1 transition-colors"
                      >
                        {editingId === space.id && editedField === 'title' ? (
                          <Input ... />
                        ) : (
                          <div>
                            <button
                              onClick={() => onSelect(space.id)}
                              className="font-medium text-gray-900 hover:text-blue-600 text-left"
                            >
                              {space.title}
                            </button>
                          </div>
                        )}
                      </div>
                    </TableCell>
```

```381:395:frontend/src/app/spaces/components/ListView.tsx
                    <TableCell>
                      <Dialog>
                        <DialogTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8" aria-label={`Edit space ${space.title}`}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                        </DialogTrigger>
                        <DialogContent aria-describedby={undefined}>
                          <DialogHeader>
                            <DialogTitle>Edit Space</DialogTitle>
                          </DialogHeader>
                          <EditSpaceForm space={space} />
                        </DialogContent>
                      </Dialog>
                    </TableCell>
```

Root cause:
- The title cell contains both an outer clickable wrapper that enters edit mode and an inner button that triggers navigation via `onSelect(space.id)`. These conflicting click targets cause navigation to take precedence, making the title effectively non‑editable by click.
- The Actions column provides a modal editor, which is redundant if inline editing is supported; it also encourages editing via a separate workflow, increasing UX inconsistency.

Decision / Change request:
- Prioritize inline editing. Disable navigation when clicking the title and remove the Actions column/edit modal button.
- If removing navigation from the title is not feasible, disable inline editing altogether and keep only the modal editor. The preferred option is the former.

Proposed edit (preferred):
- In `ListView.tsx` title cell, remove the inner `button` that calls `onSelect(space.id)` and render the title as plain text inside the editable wrapper so clicking enters edit mode. Alternatively, keep navigation behind a distinct icon/link outside the editable area if needed later.
- Remove the "Actions" column header and the corresponding `TableCell` that renders the `<Pencil>` dialog trigger and `<EditSpaceForm>` modal.

Acceptance criteria:
- Clicking title enters inline edit state without navigating away.
- Description and keywords remain inline editable as before.
- No Actions column is present in the table; no edit modal button is shown.
- Keyboard focus and blur still commit edits via existing `handleSave` logic.

Out of scope (now):
- Persisting inline edits to backend; current component performs optimistic UI only. Backend wiring can be added separately.

## (4) Statistical numbers are not updated on Spaces table(DB)

Observations:
- ALthough the frontend side correctly reflects the number of documents uploaded, the PostgreSQL DB does not update the number of documents uploaded. In terms of total size bytes, neither frontend and backend(DB) reflects the actual size of the documents uploaded. 
- Once root issue is that Content_Sources table cannot store the actual size of each document uploaded.

Pin‑point code sections:

```18:62:ms_knowledge/ent/schema/space.go
field.Int("document_count").Default(0),
field.Int64("total_size_bytes").Default(0),
```

```101:176:ms_knowledge/internal/repository/space_repository.go
// GetWithStats and getSpaceStats compute counts and total_size_bytes by querying ContentSource
// They do not persist results back to Space; values are calculated on the fly.
```

```91:158:ms_knowledge/internal/service/content_service.go
// CreateUploadURL creates a ContentSource with SizeBytes set and status=UPLOADING
// but does not update Space.document_count or Space.total_size_bytes.
```

```392:426:ms_knowledge/internal/service/content_service.go
// DeleteContentSource removes the row and R2 object but does not decrement Space counters.
```

Root cause:
- `Space.document_count` and `Space.total_size_bytes` are denormalized fields but are never updated in any service/repository path. Stats shown in the UI are derived dynamically via `ListWithStats`/`GetWithStats`, while the `space` table columns remain at their defaults (0). Thus DB values appear stale/incorrect.
- `ContentSource.size_bytes` is set at creation time, so the raw data is available; it is simply not aggregated into `space`.

Fix approaches (choose one):
1. Persist denormalized counters on write paths (recommended for list performance):
   - On content create: `Space.UpdateOneID(spaceID).AddDocumentCount(1).AddTotalSizeBytes(sizeBytes)`.
   - On content delete: `AddDocumentCount(-1).AddTotalSizeBytes(-sizeBytes)` guarded to not go below zero.
   - On content size/status corrections: adjust totals accordingly.
   - Add a best‑effort backfill admin job to recompute and set both fields for existing spaces.
2. Drop/ignore denormalized columns and always compute with aggregation queries (simpler but can be slower at scale). Optionally replace with a DB view/materialized view.
3. Use DB triggers to keep counters in sync when `contentsource` rows are inserted/updated/deleted.

Proposed edit (service‑level, minimal risk):
- In `ContentService.CreateUploadURL`, after `contentRepo.Create`, call a new `spaceRepo.AddStats(ctx, spaceID, +1, +sizeBytes)`.
- In `ContentService.DeleteContentSource`, before/after delete, call `spaceRepo.AddStats(ctx, content.SpaceID, -1, -content.SizeBytes)`.
- Implement `AddStats` in `space_repository.go` using Ent `UpdateOneID(...).AddDocumentCount(...).AddTotalSizeBytes(...)` in a single transaction with the content op if possible.

Acceptance criteria:
- DB columns `space.document_count` and `space.total_size_bytes` reflect current values after upload/delete operations.
- A backfill task updates existing rows to correct values based on current `contentsource` rows.

Note:
- Frontend can continue using `ListWithStats` for correctness; this change ensures DB columns are trustworthy for analytics and filtering.

---

## (5) Spaces table's `access_level` field is not correctly set.

Observations:
- access_level field should be set to "private" by default. However, DB is not set to "private" by default.
- Also, access_level should be easily updated via the following section;
    - on /spaces page, Gallery view, the access_level should be easily updated via left top by clicking the current access level badge.
    - on /spaces page, List view, the access_level should be easily updated via the current access level badge on the column "Access".
    - on /spaces/[spaceId] page, the access_level should be easily updated share icon on the left sidebar. 
    - All in all, once click each component(badge or icon), it should open a common modal to update the access_level.

Pin‑point code sections:

```48:59:ms_knowledge/internal/service/space_service.go
space := &domain.Space{
    Title:       strings.TrimSpace(title),
    Description: strings.TrimSpace(description),
    OwnerID:     ownerUUID,
}
```

```33:48:ms_knowledge/internal/repository/space_repository.go
create := r.client.Space.Create().
    ...
    SetAccessLevel(space.AccessLevel).
    SetGuestAccessEnabled(space.GuestAccessEnabled).
```

```48:59:ms_knowledge/ent/schema/space.go
field.String("access_level").Default("private"),
```

Root cause:
- Ent schema defines a default `"private"` for `access_level`, but the repository always calls `SetAccessLevel(space.AccessLevel)`. When the service constructs `domain.Space` without setting `AccessLevel`, the empty string overrides the Ent default at insert time, resulting in a blank/incorrect DB value.

Backend fix:
- In `space_repository.Create`, stop forcing `access_level` when the value is empty. Let Ent apply its default. Two options:
  1) Preferred: remove `.SetAccessLevel(space.AccessLevel)` entirely and rely on the Ent default unless the service explicitly sets a value.
  2) Or conditionally set only when `space.AccessLevel != ""`.
- Optionally, in `SpaceService.CreateSpace`, set `AccessLevel: "private"` explicitly to be defensive.
- Add a one‑off migration/backfill to set `access_level='private'` where empty/null.

Frontend enhancement (common UX):
- Make access level editable via a shared modal:
  - Gallery view: `frontend/src/app/spaces/components/SpaceCard.tsx` badge at lines ~88‑97 has no click handler.
  - List view: `frontend/src/app/spaces/components/ListView.tsx` badge at lines ~371‑379 is read‑only.
- Implement `AccessLevelModal` component and open it from both places when the badge is clicked; wire to a `updateSpaceAccessLevel(spaceId, level)` action.

Acceptance criteria:
- New spaces persist with `access_level='private'` by default.
- Updating access level from any of the three entry points opens the same modal and persists the change.
- Existing rows with missing access level are backfilled to `private`.

---

## (6) UX/UI Enhancement for left sidebar on /spaces/[spaceId] page
Observations:
- Currently, if user move cursor over the left sidebar, the sidebar will be opened. 
- However, if users do not recognize this funcitonality, they cannnot use this feature.
- so, when the left sidebar is closed, the icon 'panel-left-open' should be displayed on the left top of the page.

Pin‑point code sections:

```20:33:frontend/src/app/spaces/[spaceId]/components/LeftSidebar.tsx
const [isVisible, setIsVisible] = useState(false);
const [isPinned, setIsPinned] = useState(false);
```

```38:59:frontend/src/app/spaces/[spaceId]/components/LeftSidebar.tsx
// Hover trigger zone renders a 8px invisible strip at the left edge that sets isVisible=true on mouse enter when not pinned.
```

Root cause / current behavior:
- Sidebar relies on hover trigger zone (`w-8` fixed strip) to reveal when not pinned. There is no persistent affordance to discover this.

Refactor plan:
- Add a small floating toggle button when the sidebar is hidden and not pinned:
  - Place at top‑left (`fixed left-2 top-2 z-40`).
  - Use `PanelRightOpen` icon; clicking sets `isVisible(true)` for temporary open, or opens a small popover with a “Pin sidebar” control.
- Ensure the button hides when the sidebar is visible or pinned.
- Keep existing hover zone for power users, but discoverability is solved by the visible button.

Pin‑point insertion:
- In `LeftSidebar.tsx` JSX return, before the hover trigger zone, render:
  - `{!isPinned && !isVisible && (<Button variant="ghost" size="icon" className="fixed left-2 top-2 z-40 h-8 w-8" onClick={() => setIsVisible(true)} aria-label="Open sidebar"><PanelRightOpen className="h-4 w-4"/></Button>)}`

Acceptance criteria:
- With sidebar closed and unpinned, a small button is visible at top‑left; clicking it opens the sidebar.
- When the sidebar is pinned or open, the button is hidden.
