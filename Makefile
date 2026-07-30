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

test:
	go test ./...

# vet only — `make lint` 은 gofmt 검사까지 함께 돌린다. 포맷 검사 없이 정적 분석만
# 빠르게 돌리고 싶을 때 이 타겟을 쓴다.
vet:
	go vet ./...

# 커버리지 프로파일 생성 + 합산(total) 한 줄 출력. coverage.out 은 .gitignore 대상.
cover:
	go test ./... -count=1 -coverprofile=coverage.out
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
	go test ./tools/logic-map

sdd-check: sdd-doctor
	python3 tools/sdd/check_agent_config_sync.py
	python3 scripts/memory_index.py check
	python3 tools/sdd/check_index_freshness.py
	python3 tools/pm/generate_master_tracker.py --check
	python3 -m compileall -q scripts tools/logic-map tools/sdd tools/sdd-history tools/pm
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
