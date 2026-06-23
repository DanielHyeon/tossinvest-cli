import type { MetadataRoute } from 'next';
import { source } from '@/lib/source';
import { absoluteUrl } from '@/lib/site';

// Enumerate landing + all docs pages (ko + en) and the LLM entry points so
// crawlers and AI agents can discover the full corpus.
export default function sitemap(): MetadataRoute.Sitemap {
  const docPages = [...source.getPages('ko'), ...source.getPages('en')].map((page) => ({
    url: absoluteUrl(page.url),
    changeFrequency: 'weekly' as const,
    priority: 0.7,
  }));

  return [
    { url: absoluteUrl('/'), changeFrequency: 'weekly', priority: 1 },
    { url: absoluteUrl('/en'), changeFrequency: 'weekly', priority: 0.8 },
    { url: absoluteUrl('/llms.txt'), changeFrequency: 'weekly', priority: 0.5 },
    { url: absoluteUrl('/llms-full.txt'), changeFrequency: 'weekly', priority: 0.5 },
    ...docPages,
  ];
}
