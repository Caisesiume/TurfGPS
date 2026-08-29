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
# at the repository root. Every Go command therefore resolves against the
# directory it is run from, and that directory — not the tree — is what decides
# what a gate measured.
#
# What each of the five reports from that wrong directory is NOT restated here.
# `local-gates § Backend (Go)` is its one home and carries it per gate, with
# the host and the date the measurement was taken on — including which gates
# are loud without being discriminating, which is the half a recipe can get
# wrong. A paragraph of it here would be one more statement of the model #118
# exists to collapse, in this file's own words, going stale on the next
# measurement with no diff here to notice.
#
# So every recipe below carries its own directory, on the same line as the
# command it guards. make runs each recipe line in a fresh shell, so a bare
# `cd` on a line of its own is undone before the next line starts. GO_DIR is
# defined once and used everywhere; a recipe that invokes the Go toolchain
# without it is a defect on sight, however green it reports.
#
# `make gates` prints the report line `local-gates` law 1 requires. It carries
# the directory because a report naming no directory is not evidence that
# anything was compiled — the subject of everything above.
#
# Every other field is derived from what its gate actually did: all five run,
# each one's exit status decides its own field, and the target exits nonzero
# when any of them failed. A failure renders, and so does a partial result.
#
# That derivation is stated here because the first cut of this recipe was two
# unconditional echoes of a fixed literal, and review measured what it cost:
# exactly one renderable value, ever — the line could not print `test: FAIL`
# on a run that was red, nor a nonzero lint count, nor anything partial. This
# repository has twice recorded an instrument that reported success having
# measured nothing, and a typed report line is that same shape.
#
# `lint` is the honest exception. golangci-lint's issue count is not parsed,
# so a failure renders `lint: FAIL` rather than a number nothing counted; the
# `0` on the passing side is its exit status, which it earns by reporting no
# issues. A field this recipe cannot derive is one it must not invent.
# ---------------------------------------------------------------------------
#
# CGO_ENABLED is not set anywhere in this file, and must not be added. See the
# header of `service/Dockerfile` for why: `NFR-003` is deliberately silent about
# linkage while `Architecture.md § D6` is open, and a build flag here would
# settle it.
#
# The frontend block of `local-gates` is live once `web/package.json` exists,
# which it now does. `web-gates` runs it, from $(WEB_DIR), and prints a report
# line of its own; that target below carries why the line is separate from the
# one above rather than three more fields on it.

GO_DIR  := service
WEB_DIR := web
BIN_DIR := bin
IMAGE   ?= turfgps-service:dev

.DEFAULT_GOAL := help
.PHONY: help gates web-gates d8-claims fmt vet lint test build image clean \
        web-install web-build web-lint web-test

help:
	@echo 'TurfGPS — see `local-gates` for which gates are mandatory.'
	@echo ''
	@echo '  make gates   fmt, vet, lint, test and build, all from $(GO_DIR)/'
	@echo '  make web-gates  build, lint and test, all from $(WEB_DIR)/'
	@echo '  make d8-claims  does anything restate the root-run model instead of citing it'
	@echo '  make fmt     gofmt -l . — fails when it names a file, or cannot read the tree'
	@echo '  make vet     go vet ./...'
	@echo '  make lint    golangci-lint run'
	@echo '  make test    go test -race -count=1 ./...  (needs a C compiler)'
	@echo '  make build   all of $(GO_DIR)/cmd/..., into ./$(BIN_DIR)/'
	@echo '  make image   container image $(IMAGE), context $(GO_DIR)/'
	@echo '  make clean   remove ./$(BIN_DIR)/'
	@echo ''
	@echo '  make web-install  npm ci — the lockfile exactly, never re-resolved'
	@echo '  make web-build  npm run build — tsc --noEmit, then vite build'
	@echo '  make web-lint   npm run lint, warnings failing it too'
	@echo '  make web-test   npm run test — vitest, one run, no watcher'

# The five backend gates of `local-gates`, in its order.
#
# They are invoked here rather than declared as prerequisites of this target.
# A prerequisite that fails aborts make before the recipe body ever runs, so
# the report line would be reachable only on the one path where every field
# reads as a pass — which is exactly how it came to have no way of describing
# a failure. Invoking them keeps each command in its own target, where the
# working directory is, and still lets this recipe see what each one returned.
#
# All five run even after one fails, so the line describes the whole run and
# not just its first casualty. The exit status is the run's, not the echo's.
gates:
	@rc=0; \
	if $(MAKE) --no-print-directory fmt;   then r_fmt='clean';     else r_fmt='FAIL';   rc=1; fi; \
	if $(MAKE) --no-print-directory vet;   then r_vet='PASS';      else r_vet='FAIL';   rc=1; fi; \
	if $(MAKE) --no-print-directory lint;  then r_lint='0';        else r_lint='FAIL';  rc=1; fi; \
	if $(MAKE) --no-print-directory test;  then r_test='PASS';     else r_test='FAIL';  rc=1; fi; \
	if $(MAKE) --no-print-directory build; then r_build='SUCCESS'; else r_build='FAIL'; rc=1; fi; \
	echo ''; \
	echo "dir: $(GO_DIR) | fmt: $$r_fmt | vet: $$r_vet | lint: $$r_lint | test: $$r_test | build: $$r_build"; \
	exit $$rc

# The three frontend gates of `local-gates`, in its order.
#
# Invoked rather than declared as prerequisites, and all three run after one
# fails, for the reasons `gates` above already gives — same shape, same
# argument, not restated.
#
# This prints a SECOND report line instead of three more fields on the first,
# and that is the one design decision in this block. `local-gates` law 1's code
# line names the single directory that decided which tree was measured, and a
# run entering both $(GO_DIR) and $(WEB_DIR) has no such directory to name — so
# whichever one a merged line printed would be false about half of it. There
# is no frontend `fmt` or `vet` to derive those two fields from, and a single
# merged `test:` could not say which stack was red. All three are the defect
# the header describes — a field nothing measured — and `d8-claims` below is
# kept out of that same line on exactly this argument.
web-gates:
	@rc=0; \
	if $(MAKE) --no-print-directory web-build; then r_build='SUCCESS'; else r_build='FAIL'; rc=1; fi; \
	if $(MAKE) --no-print-directory web-lint;  then r_lint='0';        else r_lint='FAIL';  rc=1; fi; \
	if $(MAKE) --no-print-directory web-test;  then r_test='PASS';     else r_test='FAIL';  rc=1; fi; \
	echo ''; \
	echo "dir: $(WEB_DIR) | build: $$r_build | lint: $$r_lint | test: $$r_test"; \
	exit $$rc

web-build:
	cd $(WEB_DIR) && npm run build

# --max-warnings 0 is what earns the `0` field above, and without it that field
# would be the header's forbidden one: a count nothing counted.
#
# The backend's `0` is golangci-lint's exit status, which it earns by exiting
# nonzero on any issue at all. eslint does not, and the difference is not
# theoretical here: `web/eslint.config.js` sets its own rules to error, but the
# shared configs it extends bring warn-level rules with them, so this tree can
# report issues on a run that exits 0. Measured 28 August 2026, eslint 10.9.1,
# one warning and no errors — exit 0 bare, exit 1 under --max-warnings 0, the
# same run and the same output either way.
#
# So a bare `npm run lint` is both a wrong field and a wrong gate. `local-gates`
# puts this gate's threshold at zero ISSUES rather than zero errors, and the
# flag is what makes the exit status test the threshold that was documented.
web-lint:
	cd $(WEB_DIR) && npm run lint -- --max-warnings 0

# vitest, one run and no watcher, which is what makes it usable as a gate; the
# script is `vitest run` and the recipe adds nothing. It exits nonzero when it
# finds no test files, so this gate cannot pass having run nothing.
web-test:
	cd $(WEB_DIR) && npm run test

# The lock-honoring install, and the only one this repository sanctions.
#
# `npm install` re-resolves the ^ ranges in package.json and may write a new
# package-lock.json, replacing the reviewed dependency set — integrity hashes
# and all — with one nothing has looked at, in a file the run itself edits.
# `npm ci` installs exactly what the lock pins and fails when the lock and the
# manifest disagree, so the tree the gates above measure is the tree review saw.
#
# Not a prerequisite of `web-gates`: that would reinstall on every gate run.
# This is the step for a fresh checkout, and for after the lock changes.
web-install:
	cd $(WEB_DIR) && npm ci

# `local-gates § Documentation gates` gate 2, for the one fact this repository
# has already had restated thirteen times: what a Go command does when it is
# run from a root holding no module. It is ONE duplication class inside that
# gate and not the documentation gates, whose every other part is still run by
# hand, as that section says — and this target's name says so rather than
# implying a coverage it does not have.
#
# It is deliberately NOT part of `gates` above. That target's whole output is
# the backend report line `local-gates` law 1 prescribes, whose every field is
# derived from one of the five Go gates; a documentation result inside it would
# be a field nothing in that line can describe.
#
# The recall corpus runs FIRST and make stops if it fails, so this target
# cannot report a verdict from an instrument whose recall was not just
# demonstrated. That ordering is the point of the target, not a convenience:
# the checker's phrase lists are what decide its verdict, and a broken list
# used to be indistinguishable from a clean tree.
#
# Neither line carries a directory, and that is not the omission the header
# above condemns: neither invokes the Go toolchain, and each script resolves
# its own root — the checker from `git rev-parse --show-toplevel`, the corpus
# from its own path — so where this recipe stands cannot change what either
# one measured.
d8-claims:
	@bash scripts/gates/tests/d8-root-run-claims-recall.sh
	@bash scripts/gates/d8-root-run-claims.sh

# gofmt -l names the files it would reformat and exits 0 whether or not it
# names any, so the list is the result and has to be tested, not merely
# displayed. That is true only of a tree gofmt could READ, and the earlier
# wording of this comment stated it unconditionally, which is the premise the
# recipe below was built on and the defect it shipped.
#
# A file that does not parse is reported on stderr with exit 2 and puts
# nothing on stdout. Testing the list alone therefore reads a tree gofmt
# refused as a tree gofmt approved: empty list, `fmt: clean`, exit 0. Both
# halves are checked here, and the status first, because it decides whether
# the list means anything — an unreadable tree is not a formatting verdict at
# all. The status is taken on the same line as the capture, where no command
# can be inserted between the two to reset it.
#
# The directory change is a checked step of its own, ahead of the capture.
# Chained as `cd $(GO_DIR) && unformatted=$$(gofmt -l .);` it short-circuited:
# a failed cd skipped the assignment, the variable stayed empty, the emptiness
# test read that as nothing to report, and the recipe printed `fmt: clean` and
# exited 0 — clean, having examined nothing. Note this is still one shell and
# so still holds the directory: what make discards between recipe lines is a
# bare `cd` on a line of its own, and every line below is a continuation.
#
# Three ways to report clean having measured nothing, then — an unenterable
# directory, an unreadable tree, and an absent gofmt, which is 127 and so is
# caught by the same status test. One shape, and the one this file exists to
# refuse. The cd was fixed while its twin two lines away was left standing;
# checking a status is cheap, and finding out which ones were dropped is not.
fmt:
	@cd $(GO_DIR) || { echo 'fmt: FAIL — cannot enter $(GO_DIR)/'; exit 1; }; \
	unformatted=$$(gofmt -l .); gofmt_status=$$?; \
	if [ $$gofmt_status -ne 0 ]; then \
		echo "fmt: FAIL — gofmt exited $$gofmt_status without reading the tree"; \
		exit $$gofmt_status; \
	fi; \
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
#
# Without it, go build puts the executable in the WORKING directory — $(GO_DIR)/
# here, where the recipe runs — and never in the package directory; where the
# pattern matches more than one command, it compiles them and writes nothing at
# all. Both measured. So without -o the destination is whatever directory the
# recipe stands in, or nowhere; -o is what makes it a decision.
#
# That is the reason: where the file lands. The ignore coverage above is true
# but is not the argument — this comment argued from it until this story added
# the very rule it called missing, in the same diff, and it went false with it.
#
# The pattern is ./cmd/... and not ./cmd/turfgps, matching the build this
# story's own AC1 test runs in `build_test.go`. Named singly, this gate compiles
# the one command it was told about and exits 0 over a second one it never
# looked at — the case `build_test.go` exists to catch, which a gate blind to
# it would let through with a green line. -o names the directory rather than
# the file so the pattern can resolve to more than one command; go then takes
# each executable's name from its own package directory.
build:
	cd $(GO_DIR) && go build -trimpath -o ../$(BIN_DIR)/ ./cmd/...

# The context is $(GO_DIR)/, not the root — the same directory discipline the
# Go recipes carry, expressed the way docker takes it.
image:
	docker build -t $(IMAGE) $(GO_DIR)

clean:
	rm -rf $(BIN_DIR)
