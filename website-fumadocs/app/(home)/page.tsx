import Link from 'fumadocs-core/link';
import { buttonVariants } from 'fumadocs-ui/components/ui/button';
import { cn } from '@/lib/cn';
import {
  Bot,
  ChartCandlestick,
  Github,
  Radio,
  ShieldCheck,
  Sparkles,
  TerminalSquare,
} from 'lucide-react';

const features = [
  {
    icon: ChartCandlestick,
    title: '넓은 조회',
    desc: '계좌·시세·호가·체결·수급·시장지수·지수 상세·업종 등락·배당·거래내역까지.',
  },
  {
    icon: ShieldCheck,
    title: '안전한 거래',
    desc: '기본 비활성 + dry-run preview + --execute/--confirm 2단계 게이트.',
  },
  {
    icon: Bot,
    title: 'AI 에이전트 친화',
    desc: '모든 명령 --output json. Claude·Codex·Cursor·OpenClaw 어디든 연결.',
  },
  {
    icon: Sparkles,
    title: '토스 AI 기능',
    desc: 'AI 시그널·개인화 뉴스 브리핑·조건검색·커뮤니티 랭킹 등 공식 API 미지원 기능.',
  },
  {
    icon: Radio,
    title: '실시간',
    desc: 'push listen(SSE)로 주문·체결·보유 변경을 실시간 스트림.',
  },
  {
    icon: TerminalSquare,
    title: '자동화 우선',
    desc: 'table · JSON · CSV · SSE 출력. 파이프라인·스크립트에 바로 연결.',
  },
];

export default function HomePage() {
  return (
    <main className="flex flex-1 flex-col">
      {/* Hero */}
      <section className="relative overflow-hidden border-b">
        <div
          className="pointer-events-none absolute inset-0 -z-10 opacity-60"
          style={{
            background:
              'radial-gradient(60% 60% at 50% 0%, color-mix(in srgb, var(--color-brand) 22%, transparent), transparent)',
          }}
        />
        <div className="mx-auto flex max-w-3xl flex-col items-center px-4 py-24 text-center">
          <span className="mb-5 rounded-full border px-3 py-1 text-xs text-fd-muted-foreground">
            비공식 · 토스증권 웹 세션 재사용 CLI
          </span>
          <h1 className="text-4xl font-semibold tracking-tight md:text-6xl">
            토스증권을 <span className="text-brand">터미널·AI 에이전트</span>에서
          </h1>
          <p className="mt-5 max-w-2xl text-lg text-fd-muted-foreground">
            <code className="font-mono">tossctl</code> 하나로 계좌·시세·거래내역 조회와 제한된
            거래까지. 공식 Open API(예정)보다 넓은 범위를, 사람도 에이전트도 동일한 명령으로.
          </p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <Link href="/docs" className={cn(buttonVariants())}>
              시작하기
            </Link>
            <Link
              href="https://github.com/JungHoonGhae/tossinvest-cli"
              className={cn(buttonVariants({ color: 'secondary', className: 'gap-2' }))}
            >
              <Github className="size-4" />
              GitHub
            </Link>
          </div>

          {/* Terminal */}
          <div className="mt-12 w-full overflow-hidden rounded-xl border bg-[#0b1220] text-left shadow-2xl">
            <div className="flex items-center gap-1.5 border-b border-white/10 px-4 py-3">
              <span className="size-3 rounded-full bg-[#ff5f56]" />
              <span className="size-3 rounded-full bg-[#ffbd2e]" />
              <span className="size-3 rounded-full bg-[#27c93f]" />
              <span className="ml-3 font-mono text-xs text-white/40">tossctl</span>
            </div>
            <pre className="overflow-auto p-4 font-mono text-[13px] leading-relaxed text-white/80">
              <code>{`$ tossctl auth login          # QR + 폰 승인
$ tossctl account summary --output json
$ tossctl quote get 005930    # 국내 6자리 자동 인식
$ tossctl market index nasdaq # 지수 상세 (OHLC·52주)
$ tossctl order preview --symbol TSLA --side buy --qty 1 --price 250`}</code>
            </pre>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="mx-auto w-full max-w-5xl px-4 py-20">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((f) => (
            <div key={f.title} className="rounded-xl border p-5 transition-colors hover:bg-fd-accent/40">
              <f.icon className="mb-3 size-6 text-brand" />
              <h3 className="mb-1 font-medium">{f.title}</h3>
              <p className="text-sm text-fd-muted-foreground">{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* LLM strip */}
      <section className="border-t bg-fd-card/40">
        <div className="mx-auto flex max-w-5xl flex-col items-start gap-3 px-4 py-12 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="text-xl font-semibold">LLM이 바로 읽는 문서</h2>
            <p className="mt-1 text-sm text-fd-muted-foreground">
              모든 페이지에 Copy Markdown · ChatGPT/Claude/Cursor로 열기 + <code>/llms.txt</code> 제공.
            </p>
          </div>
          <Link href="/docs/guide/agents" className={cn(buttonVariants({ color: 'secondary' }))}>
            AI 에이전트 가이드
          </Link>
        </div>
      </section>

      <footer className="border-t py-8 text-center text-sm text-fd-muted-foreground">
        MIT License · 비공식 프로젝트 (토스증권과 무관)
      </footer>
    </main>
  );
}
