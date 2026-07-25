BINARY := bin/tossctl
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/JungHoonGhae/tossinvest-cli/internal/version.Version=$(VERSION) \
	-X github.com/JungHoonGhae/tossinvest-cli/internal/version.Commit=$(COMMIT) \
	-X github.com/JungHoonGhae/tossinvest-cli/internal/version.Date=$(DATE)

.PHONY: build run test vet cover validate gate lint fmt tidy clean

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/tossctl

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

# change 완료 게이트: `make gate CHANGE=<change-id>`.
# CHANGE 미지정 시 gate.sh 가 usage 를 출력하고 exit 2 한다.
# NTFS 마운트라 실행 비트가 없으므로 bash 로 명시 호출한다(docs/baseline.md 참고).
gate:
	bash tools/gate.sh $(CHANGE)

# lint is gofmt + vet only — no extra tooling to install. `gofmt -l` lists
# unformatted files without changing them, so the check fails loudly instead of
# silently reformatting; run `make fmt` to fix.
lint:
	@unformatted=$$(gofmt -l ./cmd ./internal); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt: 아래 파일이 포맷되지 않았습니다 — \`make fmt\` 를 실행하세요:"; \
		echo "$$unformatted" | sed 's/^/  /'; \
		exit 1; \
	fi
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

clean:
	rm -rf bin coverage.out
