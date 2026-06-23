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
// tossctl mark — terminal prompt: a ">" chevron of pixels + a green caret dash,
// on a dark rounded tile. Mirrors the design source (Pencil node z3Ufn1).
export function TossctlIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 56 56" fill="none" aria-hidden="true" {...props}>
      <rect width="56" height="56" rx="12" fill="#111111" />
      <rect x="12" y="17" width="5" height="5" fill="#ffffff" />
      <rect x="20" y="25" width="5" height="5" fill="#ffffff" />
      <rect x="12" y="33" width="5" height="5" fill="#ffffff" />
      <rect x="28" y="33" width="16" height="5" rx="1" fill="#34D399" />
    </svg>
  );
}
