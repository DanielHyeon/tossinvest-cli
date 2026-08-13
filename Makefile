BINARY := bin/tossctl
TOSSCTL_INSTALL_PATH ?= $(HOME)/.local/bin/tossctl
GOFMT ?= $(shell go env GOROOT)/bin/gofmt
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/JungHoonGhae/tossinvest-cli/internal/version.Version=$(VERSION) \
	-X github.com/JungHoonGhae/tossinvest-cli/internal/version.Commit=$(COMMIT) \
	-X github.com/JungHoonGhae/tossinvest-cli/internal/version.Date=$(DATE)

.PHONY: build stage-local-update run test vet cover validate gate lint fmt tidy clean \
	image \
	sdd-doctor sdd-sync sdd-sync-full sdd-test sdd-check sdd-hooks-install sdd-infra

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/tossctl

# Build a reviewable sibling candidate. This target never overwrites the
# installed executable and never restarts a process; the authenticated settings
# screen performs those two acts after showing the candidate's hash/build facts.
stage-local-update: build
	mkdir -p $(dir $(TOSSCTL_INSTALL_PATH))
	install -m755 $(BINARY) $(TOSSCTL_INSTALL_PATH).candidate

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/tossctl

# -timeout 30m: internal/journal opens a migrated SQLite database per test and
# there are ~670 of them, so the package sits in the several-minute range and each
# new migration step adds to every one of them. It crossed Go's 600s default at
# schema v30 (a084). The suite is slow, not hung — measured 480s at v29 and 795s
# at v30 under parallel load — and a per-package limit that a legitimate migration
# can trip is a limit that reports the wrong failure.
test:
	go test -timeout 30m ./...

# vet only — `make lint` 은 gofmt 검사까지 함께 돌린다. 포맷 검사 없이 정적 분석만
# 빠르게 돌리고 싶을 때 이 타겟을 쓴다.
vet:
	go vet ./...

# 커버리지 프로파일 생성 + 합산(total) 한 줄 출력. coverage.out 은 .gitignore 대상.
cover:
	go test -timeout 30m ./... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

# openspec 스펙/변경 검증 (SDD 게이트). openspec CLI 가 필요하다.
validate:
	openspec validate --all --strict --no-interactive

# StockOS-derived Full SDD toolchain.
sdd-doctor:
	python3 tools/sdd/sdd_doctor.py

sdd-sync:
	python3 tools/sdd/sdd_sync.py

sdd-sync-full:
	python3 tools/sdd/sdd_sync.py --full

sdd-test:
	python3 -m unittest discover -s scripts -p 'test_*.py'
	python3 -m unittest discover -s tools/logic-map -p 'test_*.py'
	python3 -m unittest discover -s tools/sdd -p 'test_*.py'
	python3 -m unittest discover -s tools/sdd-history -p 'test_*.py'
	python3 -m unittest discover -s tools/pm -p 'test_*.py'
	python3 -m unittest discover -s tools/deploy -p 'test_*.py'
	go test ./tools/logic-map

sdd-check: sdd-doctor
	python3 tools/sdd/check_agent_config_sync.py
	python3 scripts/memory_index.py check
	python3 tools/sdd/check_index_freshness.py
	python3 tools/pm/generate_master_tracker.py --check
	python3 -m compileall -q scripts tools/logic-map tools/sdd tools/sdd-history tools/pm tools/deploy
	$(MAKE) sdd-test

sdd-hooks-install:
	git config core.hooksPath .githooks
	git config sdd.captureCommits true

sdd-infra:
	uv venv --python 3.13 .sdd/.venv
	uv pip install --python .sdd/.venv/bin/python -r tools/sdd/requirements.txt

# change 완료 게이트: `make gate CHANGE=<change-id>`.
# CHANGE 미지정 시 gate.sh 가 usage 를 출력하고 exit 2 한다.
# NTFS 마운트라 실행 비트가 없으므로 bash 로 명시 호출한다(docs/baseline.md 참고).
gate:
	bash tools/gate.sh $(CHANGE)

# 이미지를 만들고 그 자리에서 되돌릴 이름을 박는다.
#
# `docker compose build` 는 tossos:local 태그를 새 이미지로 옮기고 직전 이미지는
# 태그를 잃는다. 태그 없는 이미지는 다음 prune 에 사라져 롤백 대상이 없어진다.
# 2026-08-11 · 08-12 · 08-13 세 번 그렇게 잃었고, 세 번 다 원인은 build 와 up 사이에
# 문서에만 있는 `docker tag` 한 줄이 끼어 있던 것이었다. 그래서 두 일은 한 명령이다 —
# 사이에 사람이 기억할 단계가 없어야 빠지지 않는다.
#
# 판단(직전 이미지가 이 빌드에 이름을 잃는가, 핀 이름이 쓸 수 있는 값인가)은
# tools/deploy/image_pin.py 에 있다. recipe 안에 쓴 판단은 어떤 테스트도 닿을 수
# 없는 판단이기 때문이다 — cmd/tossctl/soakautostart.go:78-81 이 같은 이유로 같은
# 선택을 한다.
# CHANGE 와 COMMIT 은 recipe 문자열이 아니라 **환경**으로 넘긴다. make 가 값을
# 명령줄에 끼워 넣으면 셸이 먼저 해석하므로 따옴표가 든 값은 image_pin 에 닿기
# 전에 이미 빠져나간다 — 그때 image_pin 의 검증은 아무것도 막지 못한다.
image: export TOSSOS_PIN_CHANGE = $(CHANGE)
image: export TOSSOS_PIN_COMMIT = $(COMMIT)
image:
	@set -e; \
	pin=$$(python3 tools/deploy/image_pin.py name); \
	if tags=$$(docker image inspect tossos:local --format '{{json .RepoTags}}' 2>&1); then \
		state=present; \
	elif printf '%s' "$$tags" | grep -qi 'no such image'; then \
		state=absent; tags=''; \
	else \
		state=unknown; tags=''; \
	fi; \
	if docker image inspect "$$pin" >/dev/null 2>&1; then taken=yes; else taken=no; fi; \
	python3 tools/deploy/image_pin.py guard --state "$$state" --tags "$$tags" --pin-taken "$$taken"; \
	echo "핀 이름: $$pin"
	docker compose build
	@set -e; \
	pin=$$(python3 tools/deploy/image_pin.py name); \
	id=$$(docker image inspect tossos:local --format '{{.Id}}'); \
	[ -n "$$id" ]; \
	docker tag "$$id" "$$pin"; \
	got=$$(docker image inspect "$$pin" --format '{{.Id}}'); \
	[ "$$got" = "$$id" ]; \
	echo "박았다: $$pin -> $$id"; \
	echo "다음: 이 핀의 schema 가 현재 저널을 읽는지 적고(docs/operations.md), 사람이 교체를 승인한다"

# lint is gofmt + vet only — no extra tooling to install. `gofmt -l` lists
# unformatted files without changing them, so the check fails loudly instead of
# silently reformatting; run `make fmt` to fix.
lint:
	@unformatted=$$($(GOFMT) -l ./cmd ./internal ./tools/logic-map); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt: 아래 파일이 포맷되지 않았습니다 — \`make fmt\` 를 실행하세요:"; \
		echo "$$unformatted" | sed 's/^/  /'; \
		exit 1; \
	fi
	go vet ./...

fmt:
	$(GOFMT) -w ./cmd ./internal ./tools/logic-map

tidy:
	go mod tidy

clean:
	rm -rf bin coverage.out
