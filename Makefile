# TurfGPS — the canonical gate and build runner.
#
# `local-gates` is the source of truth for WHICH gates are mandatory and when
# each becomes live. This file is how they are run. It exists so the command
# lines live in one place instead of being copied into agent prompts, where ten
# of them once went stale together; cite the skill, run the targets.
#
# ---------------------------------------------------------------------------
# THE ONE THING THIS FILE EXISTS TO GET RIGHT
#
# `Architecture.md § D8` puts the Go module in `service/` and leaves no module
# at the repository root. Every Go command therefore resolves against nothing
# when it is run from the root: it finds no Go files rather than no faults,
# exits zero, and prints precisely what a clean module prints. All five gates
# pass over an empty tree, character-for-character identical to a clean run.
#
# So every recipe below carries its own directory, on the same line as the
# command it guards. make runs each recipe line in a fresh shell, so a bare
# `cd` on a line of its own is undone before the next line starts. GO_DIR is
# defined once and used everywhere; a recipe that invokes the Go toolchain
# without it is a defect on sight, however green it reports.
#
# `make gates` prints the report line `local-gates` law 1 requires, including
# the directory, and prints it only after all five have actually passed. The
# line is produced by the run rather than typed afterwards, because a report
# naming no directory is not evidence that anything was compiled.
# ---------------------------------------------------------------------------
#
# CGO_ENABLED is not set anywhere in this file, and must not be added. See the
# header of `service/Dockerfile` for why: `NFR-003` is deliberately silent about
# linkage while `Architecture.md § D6` is open, and a build flag here would
# settle it.
#
# The frontend block of `local-gates` is dormant until `web/package.json`
# exists. Its targets belong to the commit that creates it — writing them now
# would mean shipping recipes nobody can run.

GO_DIR  := service
BIN_DIR := bin
IMAGE   ?= turfgps-service:dev

.DEFAULT_GOAL := help
.PHONY: help gates fmt vet lint test build image clean

help:
	@echo 'TurfGPS — see `local-gates` for which gates are mandatory.'
	@echo ''
	@echo '  make gates   fmt, vet, lint, test and build, all from $(GO_DIR)/'
	@echo '  make fmt     gofmt -l . — fails when it names a file'
	@echo '  make vet     go vet ./...'
	@echo '  make lint    golangci-lint run'
	@echo '  make test    go test -race -count=1 ./...  (needs a C compiler)'
	@echo '  make build   one executable, into ./$(BIN_DIR)/'
	@echo '  make image   container image $(IMAGE), context $(GO_DIR)/'
	@echo '  make clean   remove ./$(BIN_DIR)/'

# The five backend gates of `local-gates`, in its order. The report line runs
# only if every one of them succeeded.
gates: fmt vet lint test build
	@echo ''
	@echo 'dir: $(GO_DIR) | fmt: clean | vet: PASS | lint: 0 | test: PASS | build: SUCCESS'

# gofmt -l lists unformatted files and still exits 0, so a recipe that merely
# calls it passes whatever it finds. The output is the result and has to be
# tested, not just displayed.
fmt:
	@cd $(GO_DIR) && unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo 'fmt: NOT clean. gofmt -l names:'; \
		echo "$$unformatted"; \
		exit 1; \
	fi; \
	echo 'fmt: clean'

vet:
	cd $(GO_DIR) && go vet ./...

lint:
	cd $(GO_DIR) && golangci-lint run

# The race detector is not optional here: `Architecture.md § D1` chose Go for a
# long-lived stateful service holding many concurrent solve sessions, so
# concurrency is the thing most likely to break.
#
# -race requires cgo, and cgo requires a C compiler. On a host without one this
# target fails with `-race requires cgo`, which means the HOST is missing a
# capability — not that the code is faulty. Fix the host, or run this gate
# where a compiler exists.
#
# Do not "fix" it by dropping -race, which silently deletes the gate this
# codebase most needs, nor by setting CGO_ENABLED=0, which deletes the gate AND
# takes a position on linkage that `NFR-003` forbids this story to take.
test:
	cd $(GO_DIR) && go test -race -count=1 ./...

# -o writes to $(BIN_DIR)/ at the repository root, which .gitignore covers.
# `go build ./...` would instead drop the binary beside its package, where the
# extensionless Linux name is not ignored and rides along in the next commit.
build:
	cd $(GO_DIR) && go build -trimpath -o ../$(BIN_DIR)/turfgps ./cmd/turfgps

# The context is $(GO_DIR)/, not the root — the same directory discipline the
# Go recipes carry, expressed the way docker takes it.
image:
	docker build -t $(IMAGE) $(GO_DIR)

clean:
	rm -rf $(BIN_DIR)
