import { describe, expect, it } from 'vitest';

import { extractVideoId } from './youtube';

describe('extractVideoId', () => {
  it('returns the v param from a watch URL', () => {
    expect(extractVideoId('https://www.youtube.com/watch?v=abc123xyz00')).toBe('abc123xyz00');
  });

  it('accepts youtube.com watch URLs without www', () => {
    expect(extractVideoId('https://youtube.com/watch?v=abc123xyz00')).toBe('abc123xyz00');
  });

  it('accepts youtu.be short links', () => {
    expect(extractVideoId('https://youtu.be/abc123xyz00?t=42')).toBe('abc123xyz00');
  });

  it('rejects non-youtube domains that happen to use /watch?v=', () => {
    expect(extractVideoId('https://example.com/watch?v=abc123xyz00')).toBe('');
  });

  it('rejects youtube URLs that are not watch pages', () => {
    expect(extractVideoId('https://www.youtube.com/embed/abc123xyz00')).toBe('');
  });
});
