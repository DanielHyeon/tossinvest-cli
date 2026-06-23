import Link from 'fumadocs-core/link';
import { TossctlIcon } from '@/app/layout.client';
import {
  Bot,
  ChartCandlestick,
  Check,
  Github,
  Radio,
  ShieldCheck,
  Sparkles,
  TerminalSquare,
} from 'lucide-react';

const FEATURE_ICONS = [ChartCandlestick, ShieldCheck, Bot, Sparkles, Radio, TerminalSquare];

const AGENTS = [
  { name: 'Claude Code', logo: '/logos/claude.svg', sub: 'agent' },
  { name: 'Codex', logo: '/logos/codex.svg', sub: 'agent' },
  { name: 'Cursor', logo: '/logos/cursor.svg', sub: 'agent' },
];
const INTEGRATIONS = [
  { name: 'OpenClaw', logo: '/logos/openclaw.svg', sub: 'agent' },
  { name: 'opencode', logo: '/logos/opencode.svg', sub: 'agent' },
  { name: 'bash', logo: '/logos/bash.svg', sub: 'shell' },
];

const content = {
  ko: {
    sub: 'connect your AI agents to Toss Securities',
    desc: (
      <>
        <code className="font-mono text-white/90">tossctl</code> 명령 하나로 토스증권 계좌·시세·
        거래내역을 읽고, 주문까지 보냅니다. 사람이 치든 AI 에이전트가 호출하든 같은 명령으로.
      </>
    ),
    cta: '5분 만에 시작',
    thesis: {
      label: '왜 지금',
      headline: '공식 API는 천천히 열린다. tossctl은 지금 전부 다룬다.',
      body: '토스증권도 공식 Open API를 사전 신청자에게 단계적으로 열고 있습니다. 다만 조회·주문 기본에 머물고, 신청과 승인을 거쳐야 합니다. tossctl은 그 범위를 100% 포함하고, 공식에 아직 없는 기능까지 — 승인 없이, 지금 — 다룹니다.',
      points: [
        { k: '단계적 · 승인제', v: '공식은 사전 신청자에게 좁은 범위부터 천천히 연다.' },
        { k: '기다리지 않는다', v: 'tossctl은 롤아웃을 기다리지 않고 지금 전 범위를 다룬다.' },
        { k: '에이전트가 다룬다', v: '모든 명령이 JSON. 사람도 AI 에이전트도 같은 인터페이스.' },
      ],
    },
    sectionLabel: '왜 tossctl 인가',
    compareLabel: '공식 OPEN API 의 상위집합',
    compareLead: (
      <>
        공식 Open API(예정)의 조회·거래를 <span className="text-brand-200">100% 커버</span>하고,
        공식엔 없는 기능 <span className="text-brand-200">12개 이상</span>을 더합니다.
      </>
    ),
    official: {
      name: '공식 Open API (예정)',
      note: 'REST 조회·주문 기본 · 사전 신청 단계 롤아웃',
      items: ['계좌·잔고', '시세·호가·체결', '주문·취소·정정'],
    },
    toss: {
      name: 'tossctl',
      note: '공식 전부 100% 커버 + 그 너머',
      items: [
        '수급·시장지수·지수 상세·업종 등락',
        'AI 시그널·뉴스 브리핑·조건검색',
        '배당·커뮤니티 랭킹·관심종목 관리',
        '실시간 푸시·소수점 주문·dry-run preview',
      ],
    },
    llmTitle: 'LLM이 바로 읽는 문서',
    llmDesc: (
      <>
        모든 페이지에 Copy Markdown · ChatGPT/Claude/Cursor로 열기 +{' '}
        <code className="font-mono text-white/80">/llms.txt</code> 제공.
      </>
    ),
    llmCta: 'AI 에이전트 가이드 →',
    disclaimer: '비공식 CLI · 토스증권과 무관 · 투자 손익의 책임은 본인에게 있습니다',
    features: [
      { label: 'DATA', title: '넓은 조회', desc: '계좌·시세·호가·체결·수급·지수·업종·배당·거래내역까지 한 CLI로.' },
      { label: 'SAFETY', title: '안전한 거래', desc: '실수로 주문이 나갈 일 없음 — 기본 비활성, dry-run preview, --execute·--confirm 2단계.' },
      { label: 'AGENTS', title: '에이전트 친화', desc: '모든 명령이 --output json. Claude·Codex·Cursor·OpenClaw 가 바로 파싱한다.' },
      { label: 'INTELLIGENCE', title: '토스 AI 기능', desc: '공식 API엔 없는 AI 시그널·뉴스 브리핑·조건검색·커뮤니티 랭킹.' },
      { label: 'REALTIME', title: '실시간 푸시', desc: '주문·체결·보유 변동을 SSE 로 실시간 스트림.' },
      { label: 'AUTOMATION', title: '자동화 우선', desc: 'table·JSON·CSV·SSE 출력 — 스크립트·파이프라인에 바로 꽂는다.' },
    ],
  },
  en: {
    sub: 'connect your AI agents to Toss Securities',
    desc: (
      <>
        Read accounts, quotes, and transactions — and place orders — from one{' '}
        <code className="font-mono text-white/90">tossctl</code> command. The same command whether a
        human types it or an agent calls it.
      </>
    ),
    cta: 'Start in 5 minutes',
    thesis: {
      label: 'WHY NOW',
      headline: 'The official API opens slowly. tossctl drives all of it today.',
      body: "Toss Securities is opening an official Open API too — in stages, to pre-approved applicants. But it stays at read/order basics and needs an application. tossctl covers 100% of that scope, plus what the official API still lacks — no approval, today.",
      points: [
        { k: 'Staged · gated', v: 'The official API opens a narrow scope slowly, to pre-applicants.' },
        { k: 'No waiting', v: 'tossctl drives the full surface now, without waiting for the rollout.' },
        { k: 'Agent-operable', v: 'Every command is JSON — humans and AI agents, one interface.' },
      ],
    },
    sectionLabel: 'WHY TOSSCTL',
    compareLabel: 'A SUPERSET OF THE OFFICIAL OPEN API',
    compareLead: (
      <>
        Covers <span className="text-brand-200">100%</span> of the official Open API's (upcoming)
        read &amp; trade scope, and adds <span className="text-brand-200">12+</span> features it lacks.
      </>
    ),
    official: {
      name: 'Official Open API (planned)',
      note: 'REST read/order basics · staged rollout',
      items: ['Accounts · balances', 'Quotes · orderbook · ticks', 'Place · cancel · amend'],
    },
    toss: {
      name: 'tossctl',
      note: 'Covers 100% of the official — and beyond',
      items: [
        'Flows · indices · index detail · sectors',
        'AI signals · news briefing · screener',
        'Dividends · community rankings · watchlist',
        'Real-time push · fractional orders · dry-run',
      ],
    },
    llmTitle: 'Docs LLMs can read directly',
    llmDesc: (
      <>
        Every page has Copy Markdown · Open in ChatGPT/Claude/Cursor +{' '}
        <code className="font-mono text-white/80">/llms.txt</code>.
      </>
    ),
    llmCta: 'AI Agent Guide →',
    disclaimer: 'Unofficial CLI · not affiliated with Toss Securities · use at your own risk',
    features: [
      { label: 'DATA', title: 'Broad reads', desc: 'Accounts, quotes, orderbook, ticks, flows, indices, sectors, dividends, ledger — one CLI.' },
      { label: 'SAFETY', title: 'Safe trading', desc: 'Never fire an order by accident — off by default, dry-run preview, two-step --execute/--confirm.' },
      { label: 'AGENTS', title: 'Agent-friendly', desc: 'Every command speaks --output json. Claude, Codex, Cursor, OpenClaw parse it straight away.' },
      { label: 'INTELLIGENCE', title: 'Toss AI features', desc: 'AI signals, news briefing, screener, community rankings — none of which the official API has.' },
      { label: 'REALTIME', title: 'Real-time push', desc: 'Stream orders, fills, and holdings changes live over SSE.' },
      { label: 'AUTOMATION', title: 'Automation-first', desc: 'table · JSON · CSV · SSE output — drop it straight into scripts and pipelines.' },
    ],
  },
} as const;

function Ring({ size, opacity }: { size: number; opacity: number }) {
  return (
    <span className="absolute rounded-full border border-white" style={{ width: size, height: size, opacity }} />
  );
}

// MZ8Ua-style fading data dots radiating left/right from the central mark.
function DotTrails() {
  const rows = [-13, 0, 13];
  const cols = [0, 1, 2, 3, 4, 5];
  return (
    <div className="pointer-events-none absolute inset-0 hidden lg:block">
      {rows.map((dy) =>
        cols.map((i) => {
          const op = 0.5 * (1 - i / 6);
          const dist = 56 + i * 15;
          return (
            <span key={`l${dy}-${i}`}>
              <span
                className="absolute rounded-full bg-emerald-300"
                style={{ width: 3, height: 3, left: 120 - dist, top: 85 + dy, opacity: op }}
              />
              <span
                className="absolute rounded-full bg-emerald-300"
                style={{ width: 3, height: 3, left: 120 + dist, top: 85 + dy, opacity: op }}
              />
            </span>
          );
        }),
      )}
    </div>
  );
}

function SpokeCard({ logo, name, sub }: { logo: string; name: string; sub: string }) {
  return (
    <div className="flex items-center gap-3 rounded-lg border border-white/10 bg-[#111111] p-3 transition-colors hover:border-white/25">
      <span className="grid size-8 shrink-0 place-items-center rounded-md border border-white/10 bg-white/[0.06]">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        {/* 로고는 본연의 색 유지 (codex.svg 만 흰색으로 제작됨) */}
        <img src={logo} alt={name} className="size-4 object-contain" />
      </span>
      <div className="leading-tight">
        <div className="text-sm font-medium">{name}</div>
        <div className="font-mono text-[10px] uppercase tracking-widest text-white/30">{sub}</div>
      </div>
    </div>
  );
}

export default async function HomePage(props: PageProps<'/[lang]'>) {
  const { lang } = await props.params;
  const t = content[lang === 'en' ? 'en' : 'ko'];
  const p = lang === 'en' ? '/en' : '';

  return (
    <main className="flex flex-1 flex-col bg-[#0a0a0a] text-white">
      {/* ── Hero (hub & spoke) ──────────────────────────────── */}
      <section className="relative overflow-hidden border-b border-white/10">
        <div
          className="pointer-events-none absolute inset-0"
          style={{
            backgroundImage:
              'linear-gradient(to right, rgba(255,255,255,0.04) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.04) 1px, transparent 1px)',
            backgroundSize: '64px 64px',
            maskImage: 'radial-gradient(75% 60% at 50% 35%, black, transparent)',
            WebkitMaskImage: 'radial-gradient(75% 60% at 50% 35%, black, transparent)',
          }}
        />
        <div
          className="pointer-events-none absolute left-1/2 top-0 h-[460px] w-[760px] -translate-x-1/2"
          style={{ background: 'radial-gradient(50% 50% at 50% 0%, rgba(52,211,153,0.16), transparent)' }}
        />

        <div className="relative mx-auto w-full max-w-6xl px-4 pb-16 pt-14">
          <div className="mb-10 flex items-center justify-center gap-3 font-mono text-xs text-white/45">
            <span className="font-sans text-base font-bold tracking-tight text-white">tossctl</span>
            <span className="text-white/25">×</span>
            <span className="font-sans text-sm font-bold tracking-tight text-white/40">AI Agents</span>
          </div>

          <div className="relative grid items-center gap-8 lg:grid-cols-[1fr_minmax(280px,auto)_1fr]">
            <div
              className="pointer-events-none absolute inset-x-0 top-1/2 hidden h-px -translate-y-1/2 lg:block"
              style={{
                background:
                  'linear-gradient(90deg, transparent, rgba(52,211,153,0.2) 18%, rgba(52,211,153,0.4) 50%, rgba(52,211,153,0.2) 82%, transparent)',
              }}
            />

            <div className="hidden flex-col gap-3 lg:flex">
              <div className="mb-1 font-mono text-[10px] uppercase tracking-[0.2em] text-white/30">AI AGENTS</div>
              {AGENTS.map((a) => (
                <SpokeCard key={a.name} {...a} />
              ))}
            </div>

            <div className="relative z-1 flex flex-col items-center text-center">
              <div className="relative grid place-items-center" style={{ width: 240, height: 170 }}>
                <DotTrails />
                <Ring size={240} opacity={0.04} />
                <Ring size={180} opacity={0.07} />
                <Ring size={124} opacity={0.12} />
                <Ring size={80} opacity={0.18} />
                <TossctlIcon className="relative size-[80px] drop-shadow-[0_0_44px_rgba(52,211,153,0.3)]" />
              </div>
              <h1 className="mt-2 font-sans text-4xl font-bold tracking-tight md:text-5xl">tossinvest-cli</h1>
              <p className="mt-3 font-mono text-xs text-white/45">{t.sub}</p>
              <p className="mt-5 max-w-md text-sm text-white/65">{t.desc}</p>
              <div className="mt-7 flex flex-wrap items-center justify-center gap-3">
                <Link
                  href={`${p}/docs`}
                  className="rounded-md bg-brand px-5 py-2.5 text-sm font-semibold text-brand-foreground transition-opacity hover:opacity-90"
                >
                  {t.cta}
                </Link>
                <Link
                  href="https://github.com/JungHoonGhae/tossinvest-cli"
                  className="inline-flex items-center gap-2 rounded-md border border-white/15 px-5 py-2.5 text-sm font-medium text-white/80 transition-colors hover:bg-white/5"
                >
                  <Github className="size-4" />
                  GitHub
                </Link>
              </div>
            </div>

            <div className="hidden flex-col gap-3 lg:flex">
              <div className="mb-1 text-right font-mono text-[10px] uppercase tracking-[0.2em] text-white/30">AGENTS · SHELL</div>
              {INTEGRATIONS.map((a) => (
                <SpokeCard key={a.name} {...a} />
              ))}
            </div>
          </div>

          <div className="mx-auto mt-14 w-full max-w-3xl overflow-hidden rounded-xl border border-white/10 bg-black/40 text-left shadow-2xl backdrop-blur">
            <div className="flex items-center gap-1.5 border-b border-white/10 px-4 py-3">
              <span className="size-3 rounded-full bg-[#ff5f56]" />
              <span className="size-3 rounded-full bg-[#ffbd2e]" />
              <span className="size-3 rounded-full bg-[#27c93f]" />
              <span className="ml-3 font-mono text-xs text-white/35">tossctl</span>
            </div>
            <pre className="overflow-auto p-4 font-mono text-[13px] leading-relaxed text-white/80">
              <code>{`$ tossctl auth login
$ tossctl account summary --output json
$ tossctl quote get 005930
$ tossctl market index nasdaq
$ tossctl order preview --symbol TSLA --side buy --qty 1 --price 250`}</code>
            </pre>
          </div>
        </div>
      </section>

      {/* ── Thesis / problem ───────────────────────────────── */}
      <section className="border-b border-white/10">
        <div className="mx-auto w-full max-w-5xl px-4 py-20">
          <div className="mb-3 font-mono text-[11px] uppercase tracking-[0.2em] text-brand-200">
            {t.thesis.label}
          </div>
          <h2 className="max-w-3xl font-sans text-2xl font-bold leading-snug md:text-[2rem]">
            {t.thesis.headline}
          </h2>
          <p className="mt-4 max-w-2xl text-white/60">{t.thesis.body}</p>
          <div className="mt-10 grid gap-4 md:grid-cols-3">
            {t.thesis.points.map((pt, i) => (
              <div key={pt.k} className="rounded-xl border border-white/10 bg-[#0f0f0f] p-5">
                <div className="mb-2 font-mono text-[11px] text-white/30">0{i + 1}</div>
                <div className="font-mono text-xs font-medium text-brand-200">{pt.k}</div>
                <p className="mt-2 text-sm leading-relaxed text-white/55">{pt.v}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Superset comparison ────────────────────────────── */}
      <section className="border-b border-white/10 bg-[#0c0c0c]">
        <div className="mx-auto w-full max-w-5xl px-4 py-16">
          <div className="mb-2 font-mono text-[11px] uppercase tracking-[0.2em] text-white/35">{t.compareLabel}</div>
          <p className="mb-8 max-w-2xl text-lg text-white/80">{t.compareLead}</p>

          <div className="grid items-stretch gap-4 md:grid-cols-2">
            <div className="rounded-xl border border-white/10 bg-[#0f0f0f] p-6 opacity-80">
              <div className="mb-1 text-sm font-medium text-white/70">{t.official.name}</div>
              <div className="mb-4 font-mono text-[11px] text-white/35">{t.official.note}</div>
              <ul className="space-y-2 text-sm text-white/55">
                {t.official.items.map((it) => (
                  <li key={it} className="flex items-start gap-2">
                    <span className="mt-1.5 size-1.5 shrink-0 rounded-full bg-white/25" />
                    {it}
                  </li>
                ))}
              </ul>
            </div>

            <div className="relative rounded-xl border border-brand/40 bg-[#0f1512] p-6 shadow-[0_0_40px_-12px_rgba(52,211,153,0.35)]">
              <div className="mb-1 flex items-center gap-2 text-sm font-semibold">
                <TossctlIcon className="size-4 rounded-[4px]" />
                {t.toss.name}
              </div>
              <div className="mb-4 font-mono text-[11px] text-brand-200">{t.toss.note}</div>
              <ul className="space-y-2 text-sm text-white/80">
                {t.toss.items.map((it) => (
                  <li key={it} className="flex items-start gap-2">
                    <Check className="mt-0.5 size-4 shrink-0 text-brand-200" />
                    {it}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* ── Features ───────────────────────────────────────── */}
      <section className="mx-auto w-full max-w-5xl px-4 py-20">
        <div className="mb-8 font-mono text-[11px] uppercase tracking-[0.2em] text-white/35">{t.sectionLabel}</div>
        <div className="grid gap-px overflow-hidden rounded-xl border border-white/10 bg-white/10 sm:grid-cols-2 lg:grid-cols-3">
          {t.features.map((f, i) => {
            const Icon = FEATURE_ICONS[i];
            return (
              <div key={f.title} className="bg-[#0f0f0f] p-6 transition-colors hover:bg-[#141414]">
                <div className="mb-4 grid size-9 place-items-center rounded-lg border border-white/10 bg-white/5">
                  <Icon className="size-4.5 text-brand-200" />
                </div>
                <div className="mb-1 font-mono text-[10px] uppercase tracking-widest text-white/30">{f.label}</div>
                <h3 className="mb-1.5 font-medium">{f.title}</h3>
                <p className="text-sm text-white/55">{f.desc}</p>
              </div>
            );
          })}
        </div>

        <div className="mt-12 flex flex-col items-start gap-3 rounded-xl border border-white/10 bg-[#0f0f0f] p-6 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="font-medium">{t.llmTitle}</h2>
            <p className="mt-1 text-sm text-white/55">{t.llmDesc}</p>
          </div>
          <Link
            href={`${p}/docs/guide/agents`}
            className="shrink-0 rounded-md border border-white/15 px-4 py-2 text-sm text-white/80 transition-colors hover:bg-white/5"
          >
            {t.llmCta}
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
          {t.disclaimer}
        </p>
      </footer>
    </main>
  );
}
