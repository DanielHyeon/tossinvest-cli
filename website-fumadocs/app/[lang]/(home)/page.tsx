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
        토스증권 계좌·시세·거래내역을 조회하고, 주문을 넣습니다. 공식 Open API가 다루지 못하는
        범위까지 <code className="font-mono text-white/90">tossctl</code> 로 지금 바로.
      </>
    ),
    cta: '5분 만에 시작',
    thesis: {
      label: '왜 지금',
      headline: '공식 API, 기다릴 필요 없습니다',
      body: '토스증권 공식 Open API는 사전 신청자에게만, 그것도 조회·주문 같은 기본 기능부터 천천히 열리고 있습니다. tossctl은 이 기능을 모두 지원하고, 공식에 아직 없는 기능까지 제공합니다. 신청이나 승인은 필요 없습니다.',
      points: [
        { k: '단계적·승인제', v: '공식 API는 신청한 사람에게, 좁은 기능부터 조금씩 열립니다.' },
        { k: '기다릴 필요 없이', v: 'tossctl은 공식 출시를 기다릴 필요 없이 전 기능을 지원합니다.' },
        { k: '에이전트 연동', v: '모든 명령이 JSON으로 출력되어 AI 에이전트와 바로 연동됩니다.' },
      ],
    },
    sectionLabel: '왜 tossctl 인가',
    compareLabel: '공식 OPEN API 의 상위집합',
    compareLead: (
      <>
        공식 Open API(예정)의 조회·거래를 <span className="text-brand-200">100% 커버</span>하고,
        공식엔 없는 기능 <span className="text-brand-200">18개 이상</span>을 더합니다.
      </>
    ),
    stats: [
      { n: '약 4%', l: '공식 API가 다루는 토스 웹앱 비중' },
      { n: '18+', l: '공식에 없는 tossctl 고유 기능' },
      { n: '100%', l: '공식 조회·거래 커버리지' },
    ],
    coverage: {
      bright: '공식 API (~20개 · 약 4%)',
      dim: '토스 웹앱 의미있는 API ~430개',
      note: 'tossctl은 공식 전부와 고유 기능을 포함하고, 이 범위를 계속 넓혀갑니다.',
    },
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
      { label: 'DATA', title: '넓은 조회', desc: '계좌·시세·호가·체결·수급·지수·업종·배당·거래내역까지 하나의 CLI로 조회합니다.' },
      { label: 'SAFETY', title: '안전한 거래', desc: '기본 비활성에 dry-run preview, --execute·--confirm 2단계까지. 실수로 주문이 나가지 않습니다.' },
      { label: 'AGENTS', title: '에이전트 친화', desc: '모든 명령이 --output json 으로 출력돼 AI 에이전트와 바로 연동됩니다.' },
      { label: 'INTELLIGENCE', title: '토스 AI 기능', desc: '공식 API에는 없는 AI 시그널·뉴스 브리핑·조건검색·커뮤니티 랭킹을 제공합니다.' },
      { label: 'REALTIME', title: '실시간 푸시', desc: '주문·체결·보유 변동을 SSE로 실시간 스트리밍합니다.' },
      { label: 'AUTOMATION', title: '자동화 우선', desc: 'table·JSON·CSV·SSE로 출력해 스크립트·파이프라인에 바로 연결합니다.' },
    ],
  },
  en: {
    sub: 'connect your AI agents to Toss Securities',
    desc: (
      <>
        Read accounts, quotes, and transactions, and place orders. A wider surface than the official
        Open API — with <code className="font-mono text-white/90">tossctl</code>, right now.
      </>
    ),
    cta: 'Start in 5 minutes',
    thesis: {
      label: 'WHY NOW',
      headline: "You don't have to wait for the official API",
      body: "Toss Securities' official Open API opens only to pre-approved applicants, and so far only for basics like reads and orders. tossctl already does all of that, plus what the official API still doesn't have. No application, no approval.",
      points: [
        { k: 'Staged · gated', v: 'The official API opens slowly, to approved applicants, starting narrow.' },
        { k: 'No waiting', v: 'tossctl uses the full surface today, without waiting for the rollout.' },
        { k: 'Agents included', v: 'Every command answers in JSON, so people and agents use it the same way.' },
      ],
    },
    sectionLabel: 'WHY TOSSCTL',
    compareLabel: 'A SUPERSET OF THE OFFICIAL OPEN API',
    compareLead: (
      <>
        Covers <span className="text-brand-200">100%</span> of the official Open API's (upcoming)
        read &amp; trade scope, and adds <span className="text-brand-200">18+</span> features it lacks.
      </>
    ),
    stats: [
      { n: '~4%', l: 'of the Toss web app the official API covers' },
      { n: '18+', l: 'unique features the official API lacks' },
      { n: '100%', l: 'of official reads & trades covered' },
    ],
    coverage: {
      bright: 'Official API (~20 · ~4%)',
      dim: '~430 meaningful Toss web-app APIs',
      note: 'tossctl covers all of the official plus unique features, and keeps expanding.',
    },
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

// Scrolling "works with" strip.
const MARQUEE = [...AGENTS, ...INTEGRATIONS, { name: 'HTTP', logo: '/logos/http.svg', sub: 'api' }];
function Marquee() {
  const group = (key: string) => (
    <div key={key} className="flex shrink-0 items-center gap-12 pe-12" aria-hidden={key === 'b'}>
      {MARQUEE.map((m) => (
        <span key={m.name} className="inline-flex shrink-0 items-center gap-2 text-white/45">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={m.logo} alt="" className="size-5 object-contain" />
          <span className="font-mono text-xs">{m.name}</span>
        </span>
      ))}
    </div>
  );
  return (
    <div className="overflow-hidden border-b border-white/10 py-6 [mask-image:linear-gradient(90deg,transparent,#000_8%,#000_92%,transparent)]">
      <div className="flex w-max" style={{ animation: 'tossMarquee 28s linear infinite' }}>
        {group('a')}
        {group('b')}
      </div>
    </div>
  );
}

// Coverage dot-map: the official API is a sliver of the Toss web-app surface.
function CoverageGrid({ bright, dim, note }: { bright: string; dim: string; note: string }) {
  const total = 160;
  const official = 6; // ~4%
  return (
    <div className="rounded-xl border border-white/10 bg-[#0f0f0f] p-6">
      <div className="grid grid-cols-[repeat(20,minmax(0,1fr))] gap-1.5">
        {Array.from({ length: total }).map((_, i) => (
          <span
            key={i}
            className={
              'aspect-square rounded-[2px] ' + (i < official ? 'bg-brand-200' : 'bg-white/[0.06]')
            }
          />
        ))}
      </div>
      <div className="mt-5 flex flex-wrap items-center gap-x-6 gap-y-2 font-mono text-[11px] text-white/45">
        <span className="inline-flex items-center gap-2">
          <span className="size-2.5 rounded-[2px] bg-brand-200" /> {bright}
        </span>
        <span className="inline-flex items-center gap-2">
          <span className="size-2.5 rounded-[2px] bg-white/[0.12]" /> {dim}
        </span>
      </div>
      <p className="mt-3 text-[12px] leading-relaxed text-white/45">{note}</p>
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

      {/* ── Works-with marquee ─────────────────────────────── */}
      <Marquee />

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

          <div className="mb-8 grid gap-6 lg:grid-cols-[1.5fr_1fr] lg:items-stretch">
            <CoverageGrid bright={t.coverage.bright} dim={t.coverage.dim} note={t.coverage.note} />
            <div className="grid grid-cols-3 divide-x divide-white/10 overflow-hidden rounded-xl border border-white/10 bg-[#0f0f0f] lg:grid-cols-1 lg:divide-x-0 lg:divide-y">
              {t.stats.map((s) => (
                <div key={s.l} className="px-4 py-6 text-center lg:py-5">
                  <div className="font-sans text-3xl font-bold tracking-tight text-brand-200">{s.n}</div>
                  <div className="mx-auto mt-1.5 max-w-[18ch] text-[12px] leading-snug text-white/45">{s.l}</div>
                </div>
              ))}
            </div>
          </div>

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
