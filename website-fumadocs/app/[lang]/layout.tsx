import type { Metadata } from 'next';
import type { Viewport } from 'next';
import { Space_Grotesk, JetBrains_Mono } from 'next/font/google';
import { Provider } from '@/components/provider';
import { Body } from '@/app/layout.client';
import { source } from '@/lib/source';
import { NextProvider } from 'fumadocs-core/framework/next';
import { TreeContextProvider } from 'fumadocs-ui/contexts/tree';
import { provider } from '@/components/layouts/shared';
import { siteUrl } from '@/lib/site';
import '../global.css';

const geist = Space_Grotesk({
  variable: '--font-sans',
  subsets: ['latin'],
});

const mono = JetBrains_Mono({
  variable: '--font-mono',
  subsets: ['latin'],
});

export const metadata: Metadata = {
  title: {
    default: 'tossinvest-cli',
    template: '%s | tossinvest-cli',
  },
  description:
    '토스증권을 AI 에이전트·터미널에서 다루는 비공식 CLI (tossctl). 공식 Open API(예정)보다 넓은 조회·거래 범위 — 수급·시장지수·AI 시그널·배당·실시간 푸시·소수점 주문 등.',
  metadataBase: new URL(siteUrl),
  icons: {
    icon: '/favicon.svg',
  },
  // Machine-discoverable pointer to the LLM-friendly markdown index.
  alternates: {
    types: {
      'text/markdown': '/llms.txt',
    },
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: dark)', color: '#0A0A0A' },
    { media: '(prefers-color-scheme: light)', color: '#fff' },
  ],
};

export default async function Layout({ params, children }: LayoutProps<'/[lang]'>) {
  const { lang } = await params;
  return (
    <html lang={lang} className={`${geist.variable} ${mono.variable}`} suppressHydrationWarning>
      <Body>
        <NextProvider>
          <TreeContextProvider tree={source.getPageTree(lang)}>
            <Provider i18n={provider(lang)}>{children}</Provider>
          </TreeContextProvider>
        </NextProvider>
      </Body>
    </html>
  );
}
