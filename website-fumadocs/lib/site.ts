// Single source of truth for the canonical site URL (used by metadata,
// robots.txt, sitemap.xml, and the llms.txt routes).
export const siteUrl =
  process.env.NEXT_PUBLIC_SITE_URL ??
  (process.env.VERCEL_PROJECT_PRODUCTION_URL
    ? `https://${process.env.VERCEL_PROJECT_PRODUCTION_URL}`
    : 'https://tossinvest-cli.vercel.app');

export const absoluteUrl = (path: string) => new URL(path, siteUrl).toString();
