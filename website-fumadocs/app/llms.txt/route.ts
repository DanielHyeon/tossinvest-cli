import { source } from '@/lib/source';
import { llms } from 'fumadocs-core/source';

export const revalidate = false;

// Curated index (llmstxt.org style). Prepend a one-line project summary and a
// pointer to the single-file corpus so an agent landing here has immediate
// context before the per-page links.
export function GET() {
  const header =
    '> tossctl — 토스증권을 AI 에이전트·터미널에서 다루는 비공식 CLI. 공식 Open API(예정)보다 넓은 조회·거래 범위(수급·시장지수·AI 시그널·배당·실시간 푸시·소수점 주문 등). 전체 문서를 한 파일로: /llms-full.txt\n\n';

  return new Response(header + llms(source).index(), {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  });
}
