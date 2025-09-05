export function markdownToPlainText(markdown: string): string {
  // Remove code blocks
  let text = markdown.replace(/```[\s\S]*?```/g, '');
  // Remove inline code
  text = text.replace(/`([^`]*)`/g, '$1');
  // Remove headings and formatting symbols (#, *, _, >, -, +)
  text = text.replace(/^\s{0,3}#{1,6}\s+/gm, '');
  text = text.replace(/[\*_]{1,3}([^\*_]+)[\*_]{1,3}/g, '$1');
  text = text.replace(/^>\s?/gm, '');
  text = text.replace(/^[\-\+\*]\s+/gm, '');
  // Convert links [text](url) -> text (url)
  text = text.replace(/\[([^\]]+)\]\(([^\)]+)\)/g, '$1 ($2)');
  // Remove images ![alt](url) -> alt (url)
  text = text.replace(/!\[([^\]]*)\]\(([^\)]+)\)/g, '$1 ($2)');
  // Remove remaining markdown artifacts
  text = text.replace(/\*\*|__/g, '');
  text = text.replace(/\*|_/g, '');
  // Normalize whitespace
  text = text.replace(/\s+\n/g, '\n').trim();
  return text;
}

export async function copyMarkdownToClipboard(markdown: string): Promise<void> {
  await navigator.clipboard.writeText(markdown);
}


