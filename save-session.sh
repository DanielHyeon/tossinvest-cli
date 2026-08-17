#!/bin/bash
# ============================================================
# 세션 컨텍스트 자동 저장 스크립트
# "Prompt is too long" 전에 실행하여 작업 상태를 저장
# 새 세션에서 session-summary.md를 참조하여 이어서 작업
# ============================================================

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="$PROJECT_DIR/.ai-context"
OUTPUT_FILE="$OUTPUT_DIR/session-summary.md"
# Claude Code 메모리 경로: 프로젝트 절대경로에서 /→- 변환 (선행 - 포함)
MEMORY_DIR="$HOME/.claude/projects/$(echo "$PROJECT_DIR" | tr '/' '-')/memory"

mkdir -p "$OUTPUT_DIR"

# 기존 세션 파일이 있으면 타임스탬프 백업 (덮어쓰기 방지)
if [ -f "$OUTPUT_FILE" ]; then
  PREV_TS=$(date -r "$OUTPUT_FILE" '+%Y%m%d_%H%M%S')
  cp "$OUTPUT_FILE" "$OUTPUT_DIR/session-summary_${PREV_TS}.md"
  # 백업은 최근 5개만 유지 (오래된 것 자동 삭제)
  ls -t "$OUTPUT_DIR"/session-summary_*.md 2>/dev/null | tail -n +6 | xargs -r rm -f
fi

# 현재 시각
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

# Git 상태 수집
BRANCH=$(git -C "$PROJECT_DIR" branch --show-current 2>/dev/null)
GIT_STATUS=$(git -C "$PROJECT_DIR" status --short 2>/dev/null)
GIT_DIFF_STAT=$(git -C "$PROJECT_DIR" diff --stat 2>/dev/null)
GIT_DIFF_STAGED=$(git -C "$PROJECT_DIR" diff --cached --stat 2>/dev/null)
RECENT_COMMITS=$(git -C "$PROJECT_DIR" log --oneline -10 2>/dev/null)

# 최근 수정된 파일 (30분 이내)
RECENT_FILES=$(find "$PROJECT_DIR" \
  -not -path "*/.git/*" \
  -not -path "*/node_modules/*" \
  -not -path "*/target/*" \
  -not -path "*/build/*" \
  -not -path "*/__pycache__/*" \
  -not -path "*/.ai-context/*" \
  -type f \
  -mmin -30 \
  2>/dev/null | head -30 | sed "s|$PROJECT_DIR/||g")

# Claude Code 메모리 파일 내용 (있으면)
MEMORY_CONTENT=""
if [ -f "$MEMORY_DIR/MEMORY.md" ]; then
  MEMORY_CONTENT=$(cat "$MEMORY_DIR/MEMORY.md")
fi

# TodoWrite 내용 (있으면) — Claude Code의 할일 목록
TODO_FILE="$HOME/.claude/todos/current.json"
TODO_CONTENT=""
if [ -f "$TODO_FILE" ]; then
  TODO_CONTENT=$(cat "$TODO_FILE" 2>/dev/null)
fi

# Claude Code 최근 대화 정보 수집
# Claude Code JSONL 구조: { type: "user"|"assistant", message: { role, content } }
# content는 배열 — text 타입 항목에서 실제 사용자 입력 추출
CLAUDE_CODE_PROJECT_DIR="$HOME/.claude/projects/$(echo "$PROJECT_DIR" | tr '/' '-')"
CLAUDE_CODE_CHAT=""
if [ -d "$CLAUDE_CODE_PROJECT_DIR" ]; then
  # ls glob은 파일 수가 많으면 실패 → find 사용
  LATEST_JSONL=$(find "$CLAUDE_CODE_PROJECT_DIR" -maxdepth 1 -name "*.jsonl" -type f -printf "%T@ %p\n" 2>/dev/null | sort -rn | head -1 | cut -d' ' -f2-)
  if [ -n "$LATEST_JSONL" ]; then
    CLAUDE_CODE_CHAT=$(python3 -c "
import json, sys
try:
    conversations = []  # (role, text) 쌍 리스트
    with open('$LATEST_JSONL', 'r') as f:
        lines = f.readlines()

    # 시스템/노이즈 메시지 필터링 접두사
    skip_prefixes = (
        '<system-reminder>', '<task-notification>',
        '<ide_opened_file>', '<ide_selection>',
        '<available-deferred-tools>',
        'This session is being continued',
        '[Request interrupted',
        'Read the output file',
        'Note: /media/',
        'Called the Read tool',
        'Result of calling',
    )

    def extract_text(content):
        \"\"\"content에서 사용자가 읽을 수 있는 텍스트만 추출\"\"\"
        texts = []
        if isinstance(content, str):
            texts.append(content.strip())
        elif isinstance(content, list):
            for part in content:
                if isinstance(part, dict) and part.get('type') == 'text':
                    texts.append(part.get('text', '').strip())
        return ' '.join(t for t in texts if t)

    for line in lines:
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue

        msg_type = obj.get('type')
        msg = obj.get('message', {})
        role = msg.get('role', '')
        content = msg.get('content', [])
        text = extract_text(content)

        if not text or len(text) < 6:
            continue
        if any(text.startswith(p) for p in skip_prefixes):
            continue

        if msg_type == 'user' and role == 'user':
            conversations.append(('user', text[:300]))
        elif msg_type == 'assistant' and role == 'assistant':
            # Claude 응답: 도구 호출 제외, 텍스트 응답만 (최대 500자)
            conversations.append(('assistant', text[:500]))

    # 최근 15쌍 대화 출력 (user + assistant)
    recent = conversations[-30:]  # 최대 30개 항목 (약 15쌍)
    for role, text in recent:
        if role == 'user':
            print(f'👤 사용자: {text}')
        else:
            print(f'🤖 Claude: {text}')
        print()
except Exception as e:
    print(f'(Claude Code 대화 파싱 실패: {e})')
" 2>/dev/null)
  fi
fi

# 세션 요약 파일 생성
cat > "$OUTPUT_FILE" << HEREDOC
# 이전 세션 컨텍스트

> 저장 시각: $TIMESTAMP
> 브랜치: $BRANCH

## 현재 Git 상태 (변경된 파일)
\`\`\`
$GIT_STATUS
\`\`\`

## 변경 통계 (unstaged)
\`\`\`
$GIT_DIFF_STAT
\`\`\`

## 변경 통계 (staged)
\`\`\`
$GIT_DIFF_STAGED
\`\`\`

## 최근 커밋 (10개)
\`\`\`
$RECENT_COMMITS
\`\`\`

## 최근 30분 내 수정된 파일
\`\`\`
$RECENT_FILES
\`\`\`

## 수동 메모 (세션 종료 전 직접 기록)
<!-- 여기에 현재 하고 있던 작업, 다음에 할 작업을 적으세요 -->
<!-- 예: "Plan 47 Phase 3 구현 중, NoticeInboxService 작업 완료, 다음은 프론트엔드 연동" -->


## Claude Code 최근 대화 — 사용자 요청 + Claude 완료 작업 (자동 수집)
$CLAUDE_CODE_CHAT

## Claude Code 메모리 (자동 수집)
$MEMORY_CONTENT

HEREDOC

echo "✅ 세션 저장 완료: $OUTPUT_FILE"
echo "📝 필요하면 $OUTPUT_FILE 열어서 '수동 메모' 섹션에 현재 작업 내용을 적으세요."
echo ""
echo "🔄 새 세션에서 이어하려면:"
echo "   Claude Code에서 다음과 같이 입력:"
echo "   '@.ai-context/session-summary.md 이전 세션 이어서 작업해줘'"
