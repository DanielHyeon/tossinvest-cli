import type { MetadataRoute } from 'next';
import { absoluteUrl } from '@/lib/site';

// Allow every crawler (including AI agents) and advertise the sitemap.
// The LLM-oriented entry points (/llms.txt, /llms-full.txt) are reachable
// from the sitemap and linked in <head> as a text/markdown alternate.
export default function robots(): MetadataRoute.Robots {
  return {
    rules: [{ userAgent: '*', allow: '/' }],
    sitemap: absoluteUrl('/sitemap.xml'),
    host: absoluteUrl('/'),
  };
}
