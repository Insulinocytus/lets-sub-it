export function extractVideoId(url: string): string {
  try {
    const parsed = new URL(url);
    const isYoutubeWatchHost =
      parsed.hostname === 'www.youtube.com' || parsed.hostname === 'youtube.com';
    const isShortLinkHost = parsed.hostname === 'youtu.be' || parsed.hostname === 'www.youtu.be';

    if (isShortLinkHost) {
      return parsed.pathname.split('/').filter(Boolean)[0] ?? '';
    }

    if (!isYoutubeWatchHost || parsed.pathname !== '/watch') {
      return '';
    }

    return parsed.searchParams.get('v') ?? '';
  } catch {
    return '';
  }
}
