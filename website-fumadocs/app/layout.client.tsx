'use client';

import { useParams } from 'next/navigation';
import { type ReactNode } from 'react';

function getSection(slug?: string): string | undefined {
  if (!slug) return;
  if (slug === 'getting-started') return 'getting-started';
  if (slug === 'guide') return 'guide';
  if (slug === 'reference') return 'reference';
  return;
}

export function Body({ children }: { children: ReactNode }) {
  const { slug = [] } = useParams();
  const mode = Array.isArray(slug) ? getSection(slug[0]) : undefined;

  return <body className={[mode, 'relative flex min-h-screen flex-col'].filter(Boolean).join(' ')}>{children}</body>;
}

// tossctl mark: a terminal-style rounded square with a ">_" glyph.
export function TossctlIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 180 180" fill="none" aria-hidden="true" {...props}>
      <rect x="6" y="6" width="168" height="168" rx="36" fill="var(--color-brand)" />
      <path
        d="M52 66 L82 90 L52 114 M96 118 H128"
        stroke="var(--color-brand-foreground)"
        strokeWidth="13"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
