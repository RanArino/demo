export const SUPPORTED_FILE_TYPES = {
  'application/pdf': '.pdf',
  'text/plain': '.txt',
  'text/markdown': '.md',
  'audio/mpeg': '.mp3',
  'audio/wav': '.wav',
  'audio/mp4': '.m4a',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': '.docx',
  'application/msword': '.doc'
} as const

export type SupportedMimeType = keyof typeof SUPPORTED_FILE_TYPES

export const MAX_FILE_SIZE = 100 * 1024 * 1024 // 100MB
export const MAX_FILES_PER_UPLOAD = 10
export const UPLOAD_CHUNK_SIZE = 5 * 1024 * 1024 // 5MB chunks
