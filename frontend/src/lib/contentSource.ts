export function formatDate(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleDateString();
}

export function fileTypeLabel(mimeType?: string | null, sourceType?: string | null): string {
  const mime = (mimeType || '').toLowerCase();
  if (mime.includes('pdf')) return 'PDF';
  if (mime.includes('markdown')) return 'MD';
  if (mime.includes('plain')) return 'TXT';
  if (mime.includes('msword') || mime.includes('word')) return 'DOC';
  if (mime.includes('officedocument')) return 'DOCX';
  if (mime.includes('audio')) return 'AUDIO';
  return (sourceType || 'file').toString().toUpperCase();
}
