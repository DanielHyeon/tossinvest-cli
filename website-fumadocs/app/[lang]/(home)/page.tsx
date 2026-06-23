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
  Star,
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
    proof: {
      label: 'Trusted by builders from',
      note: 'GitHub ★ 400+ · 회사는 stargazer 공개 프로필 기준',
    },
    thesis: {
      label: '왜 지금',
      headline: '공식 API, 기다릴 필요 없습니다',
      body: '토스증권 공식 Open API가 지금 지원하는 범위는 토스 WTS 전체 기능의 약 4%(~20개)뿐입니다. 그마저도 사전 신청자에게, 조회·주문 같은 기본 기능부터 천천히 열리고 있습니다. tossctl은 이 4%를 100% 포함하고, 공식에 아직 없는 나머지 WTS 기능까지 제공합니다. 신청이나 승인은 필요 없습니다.',
      points: [
        { k: '공식은 약 4%', v: '공식 Open API는 토스 WTS 기능 ~430개 중 약 20개(4%)만 지원합니다.' },
        { k: '기다릴 필요 없이', v: '공식 출시를 기다릴 필요 없이, 공식 범위 100%와 그 너머 기능까지 지금 지원합니다.' },
        { k: '선택적 하이브리드', v: '공식 API 없이도 전부 동작합니다. 원하면 공식 API를 연결해 지원 기능은 공식 경로로 — 안정성을 선택적으로 더합니다.' },
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
      { n: '약 4%', l: '공식 Open API가 지원하는 WTS 기능 비중 (~20/430)' },
      { n: '18+', l: '공식에 없는 tossctl 고유 기능' },
      { n: '100%', l: '공식 지원 기능 커버리지' },
    ],
    coverage: {
      bright: '공식 Open API — WTS 기능의 약 4% (~20개)',
      dim: '토스 WTS 전체 기능 ~430개',
      note: 'tossctl은 공식이 다루는 약 4%를 100% 포함하고, 공식에 없는 나머지 WTS 기능까지 넓혀갑니다.',
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
      { label: 'DATA', title: '넓은 조회', desc: '계좌·시세·호가·체결·수급·지수·업종·배당·거래내역을 명령 한 줄로 조회합니다.' },
      { label: 'SAFETY', title: '안전한 거래', desc: '거래는 기본으로 꺼져 있고, 주문 전 미리보기와 두 번의 확인을 거칩니다. 실수로 주문이 나가지 않습니다.' },
      { label: 'AGENTS', title: '에이전트 친화', desc: '모든 결과를 AI가 그대로 읽는 형식으로 내보내 에이전트와 바로 연동됩니다.' },
      { label: 'INTELLIGENCE', title: '토스 AI 기능', desc: '공식 API에는 없는 AI 시그널·뉴스 브리핑·조건검색·커뮤니티 랭킹을 제공합니다.' },
      { label: 'REALTIME', title: '실시간 푸시', desc: '주문·체결·보유 변동을 실시간으로 받아봅니다.' },
      { label: 'AUTOMATION', title: '자동화 우선', desc: '표·파일·실시간 등 원하는 형식으로 내보내 스크립트와 자동화에 바로 연결합니다.' },
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
    proof: {
      label: 'Trusted by builders from',
      note: '400+ GitHub stars · companies from stargazers’ public profiles',
    },
    thesis: {
      label: 'WHY NOW',
      headline: "You don't have to wait for the official API",
      body: "Toss Securities' official Open API covers only about 4% (~20) of the full Toss WTS feature set today — and even that opens slowly, to pre-approved applicants, starting with basics like reads and orders. tossctl covers 100% of that 4% and adds the rest of the WTS features the official API still lacks. No application, no approval.",
      points: [
        { k: 'Official ≈ 4%', v: 'The official Open API supports ~20 of ~430 Toss WTS features — about 4%.' },
        { k: 'No waiting', v: 'No waiting for the rollout — 100% of the official scope and beyond, today.' },
        { k: 'Hybrid, optional', v: 'Works fully without the official API. Plug it in if you want — supported features route through it for extra stability.' },
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
      { n: '~4%', l: 'of Toss WTS features the official API supports (~20/430)' },
      { n: '18+', l: 'unique features the official API lacks' },
      { n: '100%', l: 'of officially-supported features covered' },
    ],
    coverage: {
      bright: 'Official Open API — ~4% of WTS (~20)',
      dim: '~430 total Toss WTS features',
      note: 'tossctl covers 100% of the official ~4% and keeps expanding into the rest of WTS.',
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
      { label: 'DATA', title: 'Broad reads', desc: 'Accounts, quotes, orderbook, ticks, flows, indices, sectors, dividends, ledger — in one command.' },
      { label: 'SAFETY', title: 'Safe trading', desc: 'Trading is off by default, with an order preview and two confirmations before anything runs. No accidental orders.' },
      { label: 'AGENTS', title: 'Agent-friendly', desc: 'Every result comes back in a format AI reads directly, so agents like Claude, Codex, and Cursor connect right away.' },
      { label: 'INTELLIGENCE', title: 'Toss AI features', desc: 'AI signals, news briefing, screener, community rankings — none of which the official API has.' },
      { label: 'REALTIME', title: 'Real-time push', desc: 'See order, fill, and holdings changes the moment they happen.' },
      { label: 'AUTOMATION', title: 'Automation-first', desc: 'Export as a table, file, or live feed — drop it straight into scripts and automation.' },
    ],
  },
} as const;

// MZ8Ua-style connector trails: three rows of fading dots run from the node
// edge outward toward the three agent / integration cards on each side.
// Rows are aligned to the card centers (node box is HUB_H tall, cards ~70px apart).
const HUB_H = 198;
function DotTrails() {
  const cy = HUB_H / 2; // 99 — vertical center of the node box
  const rows = [-70, 0, 70]; // aligns with the 3 cards in each side column
  const n = 8;
  return (
    <div className="pointer-events-none absolute inset-0 hidden lg:block">
      {rows.map((dy) =>
        Array.from({ length: n }).map((_, i) => {
          const dist = 78 + i * 19; // 78 → ~211px from center (reaches into the gap)
          const op = 0.5 * (1 - i / (n + 1)); // brightest near node, fades outward — neutral
          return (
            <span key={`${dy}-${i}`}>
              <span
                className="absolute rounded-full bg-white"
                style={{ width: 3, height: 3, left: 120 - dist, top: cy + dy, opacity: op }}
              />
              <span
                className="absolute rounded-full bg-white"
                style={{ width: 3, height: 3, left: 120 + dist, top: cy + dy, opacity: op }}
              />
            </span>
          );
        }),
      )}
    </div>
  );
}

// Scrolling "works with" strip — AI coding agents that drive tossctl.
// Logos keep their brand colors; only naturally-black marks ship as white SVGs.
const MARQUEE = [
  { name: 'Claude Code', logo: '/logos/mq/claude.svg' },
  { name: 'Codex CLI', logo: '/logos/codex.svg' },
  { name: 'Gemini CLI', logo: '/logos/mq/googlegemini.svg' },
  { name: 'Cursor', logo: '/logos/mq/cursor.svg' },
  { name: 'GitHub Copilot', logo: '/logos/mq/githubcopilot.svg' },
  { name: 'OpenCode', logo: '/logos/opencode.svg' },
  { name: 'Qwen Code', logo: '/logos/mq/qwen.svg' },
  { name: 'DeepSeek', logo: '/logos/mq/deepseek.svg' },
  { name: 'Mistral', logo: '/logos/mq/mistralai.svg' },
  { name: 'Kimi CLI', logo: '/logos/mq/moonshotai.svg' },
  { name: 'OpenClaw', logo: '/logos/openclaw.svg' },
];
function Marquee() {
  // Two identical groups translated by exactly -50% → seamless infinite loop.
  // The trailing gap (gap-12) equals the inner gap so the seam is invisible.
  const group = (key: string) => (
    <div key={key} className="flex shrink-0 items-center gap-12 pe-12" aria-hidden={key === 'b'}>
      {MARQUEE.map((m) => (
        <span key={m.name} className="inline-flex shrink-0 items-center gap-2.5 text-white/60">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={m.logo} alt="" className="size-6 object-contain md:size-8" />
          <span className="font-mono text-xs md:text-sm">{m.name}</span>
        </span>
      ))}
    </div>
  );
  return (
    <div className="relative border-y border-white/10 py-12">
      {/* "Works with" sits ON the top border, legend-style */}
      <span className="absolute left-1/2 top-0 -translate-x-1/2 -translate-y-1/2 bg-[#0a0a0a] px-4 font-mono text-[11px] uppercase tracking-[0.3em] text-white/40">
        Works with
      </span>
      <div className="overflow-hidden [mask-image:linear-gradient(90deg,transparent,#000_6%,#000_94%,transparent)]">
        <div className="flex w-max animate-[tossMarquee_36s_linear_infinite]">
          {group('a')}
          {group('b')}
        </div>
      </div>
    </div>
  );
}

// "Trusted by builders from" grid. Companies are derived from stargazers'
// public GitHub `company` field — framed as *builders from*, never an
// endorsement. Logos where simple-icons has them (white), wordmark text
// otherwise. Toss is intentionally excluded (this is an unofficial tool).
const COMPANIES: { name: string; logo?: string }[] = [
  { name: 'Naver', logo: '/logos/companies/naver.svg' },
  { name: 'Kakao', logo: '/logos/companies/kakao.svg' },
  { name: 'Daangn', logo: '/logos/companies/daangn.png' },
  { name: 'Upbit', logo: '/logos/companies/upbit.svg' },
  { name: 'TMAP', logo: '/logos/companies/tmap.png' },
  { name: 'SK telecom', logo: '/logos/companies/sktelecom.svg' },
  { name: 'TeamSparta', logo: '/logos/companies/teamsparta.png' },
  { name: 'TwelveLabs' },
];

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

function SpokeCard({
  logo,
  name,
  sub,
  mirror = false,
}: {
  logo: string;
  name: string;
  sub: string;
  mirror?: boolean;
}) {
  return (
    <div
      className={
        'flex items-center gap-3 rounded-lg border border-white/10 bg-[#111111] p-3 transition-colors hover:border-white/25 ' +
        (mirror ? 'flex-row-reverse text-right' : '')
      }
    >
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
              'linear-gradient(to right, rgba(255,255,255,0.045) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.045) 1px, transparent 1px)',
            backgroundSize: '208px 148px',
            backgroundPosition: 'center',
            maskImage: 'radial-gradient(95% 75% at 50% 42%, black, transparent)',
            WebkitMaskImage: 'radial-gradient(95% 75% at 50% 42%, black, transparent)',
          }}
        />
        <div
          className="pointer-events-none absolute left-1/2 top-0 h-[460px] w-[760px] -translate-x-1/2"
          style={{ background: 'radial-gradient(50% 50% at 50% 0%, rgba(52,211,153,0.16), transparent)' }}
        />

        <div className="relative mx-auto w-full max-w-6xl px-4 pb-16 pt-14">
          <div className="mb-12 flex items-center justify-center gap-3 font-mono text-xs text-white/45">
            <span className="font-sans text-base font-bold tracking-tight text-white">tossctl</span>
            <span className="text-white/25">×</span>
            <span className="font-sans text-sm font-bold tracking-tight text-white/40">AI Agents</span>
          </div>

          {/* heading block (above the hub, MZ8Ua-style) */}
          <div className="mx-auto mb-16 max-w-2xl text-center">
            <h1 className="font-sans text-4xl font-bold tracking-tight md:text-5xl">tossinvest-cli</h1>
            <p className="mt-3 font-mono text-xs text-white/45">{t.sub}</p>
            <p className="mx-auto mt-5 max-w-md break-keep text-sm text-white/65">{t.desc}</p>
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

          {/* hub & spoke — node vertically aligned with the cards on each side */}
          <div className="relative grid items-center gap-8 lg:grid-cols-[1fr_auto_1fr]">
            <div className="relative z-2 hidden flex-col gap-3 lg:flex">
              <div className="absolute -top-7 left-0 font-mono text-[10px] uppercase tracking-[0.2em] text-white/30">AI AGENTS</div>
              {AGENTS.map((a) => (
                <SpokeCard key={a.name} {...a} />
              ))}
            </div>

            <div
              className="relative z-1 grid place-items-center justify-self-center"
              style={{ width: 240, height: HUB_H }}
            >
              {/* soft circle glow behind the node (replaces the concentric strokes) */}
              <span
                className="pointer-events-none absolute rounded-full"
                style={{
                  width: 360,
                  height: 360,
                  background:
                    'radial-gradient(circle, rgba(255,255,255,0.10), rgba(255,255,255,0.03) 42%, transparent 70%)',
                }}
              />
              <DotTrails />
              <TossctlIcon className="relative size-[80px] drop-shadow-[0_0_40px_rgba(255,255,255,0.08)]" />
            </div>

            <div className="relative z-2 hidden flex-col gap-3 lg:flex">
              <div className="absolute -top-7 right-0 font-mono text-[10px] uppercase tracking-[0.2em] text-white/30">AGENTS · SHELL</div>
              {INTEGRATIONS.map((a) => (
                <SpokeCard key={a.name} {...a} mirror />
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

      {/* ── Social proof: "Trusted by builders from" (stargazers' employers) ── */}
      <section className="border-b border-white/10">
        <div className="mx-auto max-w-5xl px-4 py-14 text-center">
          <div className="mb-8 font-sans text-sm font-semibold tracking-wide text-white/80">
            {t.proof.label}
          </div>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {COMPANIES.map((c) => (
              <div
                key={c.name}
                className="flex h-20 items-center justify-center rounded-xl border border-white/10 bg-[#0f0f0f]"
              >
                {c.logo ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  // 브랜드 로고는 원래 색 유지 + 높이 고정(아이콘/워드마크 혼재).
                  <img
                    src={c.logo}
                    alt={c.name}
                    className="h-7 w-auto max-w-[130px] object-contain"
                  />
                ) : (
                  <span className="font-sans text-base font-semibold tracking-tight text-white/70">
                    {c.name}
                  </span>
                )}
              </div>
            ))}
          </div>
          <a
            href="https://github.com/JungHoonGhae/tossinvest-cli/stargazers"
            className="mt-6 inline-flex items-center gap-1.5 font-mono text-[11px] text-white/35 transition-colors hover:text-white/55"
          >
            <Star className="size-3" />
            {t.proof.note}
          </a>
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
          <p className="mt-4 max-w-2xl break-keep text-white/60">{t.thesis.body}</p>
          <div className="mt-10 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
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
