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
# Measured from this root, four of the five are loud: vet, lint and build each
# fail to resolve a module, and test fails on the absent-cgo host reason that is
# identical from either directory, so on this host it discriminates nothing.
# Only gofmt exits 0, and its output is identical to a clean run — not because
# it read nothing, but because it walks the filesystem instead of resolving a
# module, and every Go file in this repository sits inside the module
# directory. It reads the same tree by accident of layout, and reads more the
# moment a .go file lands outside it.
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
# The frontend block of `local-gates` is dormant until `web/package.json`
# exists. Its targets belong to the commit that creates it — writing them now
# would mean shipping recipes nobody can run.

GO_DIR  := service
BIN_DIR := bin
IMAGE   ?= turfgps-service:dev

.DEFAULT_GOAL := help
.PHONY: help gates d8-claims fmt vet lint test build image clean

help:
	@echo 'TurfGPS — see `local-gates` for which gates are mandatory.'
	@echo ''
	@echo '  make gates   fmt, vet, lint, test and build, all from $(GO_DIR)/'
	@echo '  make d8-claims  does anything restate the root-run model instead of citing it'
	@echo '  make fmt     gofmt -l . — fails when it names a file, or cannot read the tree'
	@echo '  make vet     go vet ./...'
	@echo '  make lint    golangci-lint run'
	@echo '  make test    go test -race -count=1 ./...  (needs a C compiler)'
	@echo '  make build   all of $(GO_DIR)/cmd/..., into ./$(BIN_DIR)/'
	@echo '  make image   container image $(IMAGE), context $(GO_DIR)/'
	@echo '  make clean   remove ./$(BIN_DIR)/'

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

# `local-gates § Documentation gates` gate 2, for the one fact this repository
# has already had restated thirteen times: what a Go command does when it is
# run from a root holding no module. It is NOT the documentation gates — those
# have no runner, as that section says — and this target's name says so rather
# than implying a coverage it does not have.
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
