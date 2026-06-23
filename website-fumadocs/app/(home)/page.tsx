import Link from 'fumadocs-core/link';
import { TossctlIcon } from '@/app/layout.client';
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
    label: 'DATA',
    title: '넓은 조회',
    desc: '계좌·시세·호가·체결·수급·시장지수·지수 상세·업종 등락·배당·거래내역까지.',
  },
  {
    icon: ShieldCheck,
    label: 'SAFETY',
    title: '안전한 거래',
    desc: '기본 비활성 + dry-run preview + --execute/--confirm 2단계 게이트.',
  },
  {
    icon: Bot,
    label: 'AGENTS',
    title: 'AI 에이전트 친화',
    desc: '모든 명령 --output json. Claude·Codex·Cursor·OpenClaw 어디든 연결.',
  },
  {
    icon: Sparkles,
    label: 'INTELLIGENCE',
    title: '토스 AI 기능',
    desc: 'AI 시그널·개인화 뉴스 브리핑·조건검색·커뮤니티 랭킹 등 공식 API 미지원.',
  },
  {
    icon: Radio,
    label: 'REALTIME',
    title: '실시간 푸시',
    desc: 'push listen(SSE)로 주문·체결·보유 변경을 실시간 스트림.',
  },
  {
    icon: TerminalSquare,
    label: 'AUTOMATION',
    title: '자동화 우선',
    desc: 'table · JSON · CSV · SSE 출력. 파이프라인·스크립트에 바로 연결.',
  },
];

function Ring({ size, opacity }: { size: number; opacity: number }) {
  return (
    <span
      className="absolute rounded-full border border-white"
      style={{ width: size, height: size, opacity }}
    />
  );
}

export default function HomePage() {
  return (
    <main className="flex flex-1 flex-col bg-[#0a0a0a] text-white">
      {/* ── Hero ───────────────────────────────────────────── */}
      <section className="relative overflow-hidden border-b border-white/10">
        {/* grid */}
        <div
          className="pointer-events-none absolute inset-0"
          style={{
            backgroundImage:
              'linear-gradient(to right, rgba(255,255,255,0.04) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.04) 1px, transparent 1px)',
            backgroundSize: '64px 64px',
            maskImage: 'radial-gradient(70% 60% at 50% 35%, black, transparent)',
            WebkitMaskImage: 'radial-gradient(70% 60% at 50% 35%, black, transparent)',
          }}
        />
        {/* glow */}
        <div
          className="pointer-events-none absolute left-1/2 top-0 -z-0 h-[420px] w-[680px] -translate-x-1/2"
          style={{
            background:
              'radial-gradient(50% 50% at 50% 0%, rgba(52,211,153,0.14), transparent)',
          }}
        />

        <div className="relative mx-auto flex max-w-3xl flex-col items-center px-4 pb-20 pt-16 text-center">
          <div className="mb-12 flex items-center gap-3 font-mono text-xs text-white/45">
            <span className="font-sans text-base font-bold tracking-tight text-white">tossctl</span>
            <span className="text-white/25">×</span>
            <span className="font-sans text-sm font-bold tracking-tight text-white/40">AI Agents</span>
          </div>

          {/* central mark + rings */}
          <div className="relative grid place-items-center" style={{ width: 280, height: 200 }}>
            <Ring size={280} opacity={0.04} />
            <Ring size={210} opacity={0.06} />
            <Ring size={150} opacity={0.1} />
            <Ring size={96} opacity={0.16} />
            <TossctlIcon className="relative size-[84px] drop-shadow-[0_0_40px_rgba(52,211,153,0.25)]" />
          </div>

          <h1 className="mt-2 font-sans text-5xl font-bold tracking-tight md:text-6xl">
            tossinvest-cli
          </h1>
          <p className="mt-4 font-mono text-sm text-white/45">
            connect your AI agents to Toss Securities
          </p>
          <p className="mt-6 max-w-xl text-base text-white/65">
            <code className="font-mono text-white/90">tossctl</code> 하나로 계좌·시세·거래내역
            조회와 제한된 거래까지. 공식 Open API(예정)보다 넓은 범위를, 사람도 에이전트도 동일한
            명령으로.
          </p>

          <div className="mt-9 flex flex-wrap items-center justify-center gap-3">
            <Link
              href="/docs"
              className="rounded-md bg-brand px-5 py-2.5 text-sm font-semibold text-brand-foreground transition-opacity hover:opacity-90"
            >
              시작하기
            </Link>
            <Link
              href="https://github.com/JungHoonGhae/tossinvest-cli"
              className="inline-flex items-center gap-2 rounded-md border border-white/15 px-5 py-2.5 text-sm font-medium text-white/80 transition-colors hover:bg-white/5"
            >
              <Github className="size-4" />
              GitHub
            </Link>
          </div>

          {/* terminal */}
          <div className="mt-14 w-full overflow-hidden rounded-xl border border-white/10 bg-black/40 text-left shadow-2xl backdrop-blur">
            <div className="flex items-center gap-1.5 border-b border-white/10 px-4 py-3">
              <span className="size-3 rounded-full bg-[#ff5f56]" />
              <span className="size-3 rounded-full bg-[#ffbd2e]" />
              <span className="size-3 rounded-full bg-[#27c93f]" />
              <span className="ml-3 font-mono text-xs text-white/35">tossctl</span>
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

      {/* ── Features ───────────────────────────────────────── */}
      <section className="mx-auto w-full max-w-5xl px-4 py-20">
        <div className="mb-8 font-mono text-[11px] uppercase tracking-[0.2em] text-white/35">
          왜 tossctl 인가
        </div>
        <div className="grid gap-px overflow-hidden rounded-xl border border-white/10 bg-white/10 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((f) => (
            <div key={f.title} className="bg-[#0f0f0f] p-6 transition-colors hover:bg-[#141414]">
              <div className="mb-4 grid size-9 place-items-center rounded-lg border border-white/10 bg-white/5">
                <f.icon className="size-4.5 text-brand-200" />
              </div>
              <div className="mb-1 font-mono text-[10px] uppercase tracking-widest text-white/30">
                {f.label}
              </div>
              <h3 className="mb-1.5 font-medium">{f.title}</h3>
              <p className="text-sm text-white/55">{f.desc}</p>
            </div>
          ))}
        </div>

        {/* LLM strip */}
        <div className="mt-12 flex flex-col items-start gap-3 rounded-xl border border-white/10 bg-[#0f0f0f] p-6 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="font-medium">LLM이 바로 읽는 문서</h2>
            <p className="mt-1 text-sm text-white/55">
              모든 페이지에 Copy Markdown · ChatGPT/Claude/Cursor로 열기 + <code className="font-mono text-white/80">/llms.txt</code> 제공.
            </p>
          </div>
          <Link
            href="/docs/guide/agents"
            className="shrink-0 rounded-md border border-white/15 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5"
          >
            AI 에이전트 가이드 →
          </Link>
        </div>
      </section>

      {/* ── Footer ─────────────────────────────────────────── */}
      <footer className="border-t border-white/10">
        <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-3 px-4 py-7 font-mono text-xs sm:flex-row">
          <span className="inline-flex items-center gap-2 text-white/45">
            <Github className="size-3.5" />
            JungHoonGhae/tossinvest-cli
          </span>
          <span className="inline-flex items-center gap-2">
            <span className="font-sans font-bold text-[#00ADD8]">Go</span>
            <span className="text-white/20">/</span>
            <span className="font-bold text-white/40">MIT</span>
            <span className="text-white/20">/</span>
            <span className="font-bold text-[#FF8800]">BETA</span>
          </span>
        </div>
        <p className="px-4 pb-8 text-center font-mono text-[10px] leading-relaxed tracking-wider text-white/25">
          비공식 CLI · 토스증권과 무관 · 투자 손익의 책임은 본인에게 있습니다
        </p>
      </footer>
    </main>
  );
}
