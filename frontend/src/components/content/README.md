# Content (Shared) Components

This directory contains reusable components and utilities for rendering and interacting with content sources across the app.

## Components

- **ContentSourceCard.tsx**
  - Presentational card for a single `ContentSource`
  - Props:
    - `contentSource`: the item to render
    - `onView(contentSource)`
    - `onDownload(contentSource, kind: 'original' | 'processed')`
    - `onDelete(contentSource)`
    - `selected: boolean`
    - `onSelectChange(id, checked)`
    - `progress?: number`
  - Stateless UI: no routing/toast; callers control side effects
  - Used in:
    - `frontend/src/app/spaces/[spaceId]/components/ContentSourcesSection.tsx`

## Utilities

Utilities are pure functions and live in `frontend/src/lib/contentSource.ts`.

- `formatDate(date: Date | string): string`
  - Formats a date into a locale string
- `fileTypeLabel(mimeType?: string, sourceType?: string): string`
  - Produces a concise label (e.g., `PDF`, `DOCX`, `TXT`) from mime/source type
