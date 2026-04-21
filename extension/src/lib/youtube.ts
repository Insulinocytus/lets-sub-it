export function extractVideoId(url: string): string {
  try {
    const parsed = new URL(url);
    const isYoutubeWatchHost =
      parsed.hostname === 'www.youtube.com' || parsed.hostname === 'youtube.com';

    if (!isYoutubeWatchHost || parsed.pathname !== '/watch') {
      return '';
    }

    return parsed.searchParams.get('v') ?? '';
  } catch {
    return '';
  }
}
