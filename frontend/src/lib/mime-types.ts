/**
 * Maps common MIME subtypes to user-friendly names.
 * Only the subtype part (after the slash) is used for mapping.
 */
export const mimeTypeToFriendlyName: Record<string, string> = {
  // Application types
  pdf: 'PDF',
  'vnd.openxmlformats-officedocument.wordprocessingml.document': 'Word',
  msword: 'Word',
  json: 'JSON',
  zip: 'ZIP',

  // Text types
  plain: 'Text',
  markdown: 'Markdown',
  csv: 'CSV',
  html: 'HTML',
  css: 'CSS',
  javascript: 'JavaScript',

  // Image types
  jpeg: 'JPEG',
  png: 'PNG',
  gif: 'GIF',
  svg: 'SVG',
  webp: 'WEBP',

  // Audio types
  mpeg: 'MP3',
  wav: 'WAV',

  // Video types
  mp4: 'MP4',
  webm: 'WebM',
};

/**
 * Returns a user-friendly name for a given MIME type.
 * If the MIME type is not recognized, returns the uppercased subtype.
 * If no MIME type is provided, returns 'File'.
 */
export function getFriendlyNameFromMimeType(mimeType: string | null | undefined): string {
  if (!mimeType) return 'File';

  const parts = mimeType.split('/');
  const subType = parts.length > 1 ? parts[1] : parts[0];

  // Handle cases like 'text/plain;charset=UTF-8'
  const cleanSubType = subType.split(';')[0];

  return mimeTypeToFriendlyName[cleanSubType] || cleanSubType.toUpperCase();
}