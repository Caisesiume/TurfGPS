# TurfGPS service image — one self-contained executable, per `NFR-003`.
#
# The build context is `service/`, not the repository root: `Architecture.md § D8`
# gives each deployable its own directory, and this image needs nothing from
# outside the Go module.
#
#     docker build -t turfgps-service:dev service/      (or: make image)
#
# ---------------------------------------------------------------------------
# WHAT THIS FILE MUST NOT DO, AND WHY. Read before editing.
#
# `NFR-003` is deliberately silent about linkage, because `Architecture.md § D6`
# proposes a raster sampler carrying a cgo dependency and that proposal is still
# open. Nothing here may demand a statically linked or cgo-free binary: that
# would settle an open architectural question from the wrong side, and it would
# be settled by a build flag rather than by the measurement D6 asks for.
#
# Two consequences below look like omissions and are neither:
#
#   * CGO_ENABLED is never set — not here, not in the Makefile, not in any
#     script. Whatever the toolchain defaults to is what this image builds, so
#     the artefact follows the module's needs instead of dictating them.
#
#   * The runtime base is not `scratch`. Scratch carries no libc, so a
#     dynamically linked binary cannot start there — which would force the
#     build to link statically, settling D6 from the wrong side by a longer
#     route. AC1 excludes exactly three things, listed at the runtime stage
#     below, and a base image's own libc is none of them.
# ---------------------------------------------------------------------------

# The builder's Debian release is paired with the runtime's and they move
# together. With cgo enabled the binary links against the builder's glibc, and
# a newer builder yields a binary the older runtime cannot load — a failure
# that appears only at start-up, in the image, after a clean build. bookworm
# here; base-debian12 is bookworm.
#
# The digest is what makes that pairing hold. Both tags move: a rebuild months
# apart can pick up a newer builder while the runtime stands still, and the
# mismatch above is exactly what that produces. The tag is kept beside the
# digest because it is the readable half, but the digest is what resolves. When
# either FROM is re-pinned, the other is checked for the same Debian release.
FROM golang:1.26-bookworm@sha256:116d58cbd88c1297624acc6e967a060012422bacf9930927e23fb719189c6f36 AS build

WORKDIR /src

# The module has no dependencies yet, so the whole source is one COPY. The
# usual manifest-first split, which exists to keep a dependency layer cached
# across source-only edits, belongs to the commit that adds the first
# dependency and buys nothing before it.
COPY . .

# -o writes outside the module tree on purpose. `go build ./...` drops the
# binary into the package directory, under the /src tree copied in above. /out
# holds the executable alone, at the one fixed path the runtime stage below
# copies from.
#
# -trimpath keeps absolute build paths out of the artefact, so the same source
# yields the same binary regardless of which directory it was checked out into.
RUN go build -trimpath -o /out/turfgps ./cmd/turfgps

# ---------------------------------------------------------------------------
# Runtime stage. `NFR-003` AC1 excludes exactly three things, and this base
# carries none of them:
#
#   no language runtime        — no Go toolchain, no interpreter of any kind
#   no application server      — the executable serves HTTP itself
#   no project-installed deps  — nothing is installed at this stage; there is
#                                no package manager and no shell to install with
#
# What the base does carry is glibc, ca-certificates and tzdata. Those belong
# to the base image rather than to this project, so they meet none of the three
# exclusions. The certificates are why the service can reach the Turf API over
# TLS without a second file being copied beside the executable.
#
# If D6 is settled toward godal, this base is where that lands: a cgo binary
# needing GDAL would need a base carrying it. Choosing a base with a libc keeps
# both of D6's outcomes reachable; choosing scratch would have closed one now.
#
# Pinned by digest, and paired with the builder's release — read the note above
# the builder stage before moving either.
FROM gcr.io/distroless/base-debian12:nonroot@sha256:b12529fbbd0bb15eea8905f69d83148679e0b4d7d434c8808100792029b1caae

# Exactly one file enters this stage. That is AC1's threshold and it is
# checkable by reading this file: one COPY, one path, no second artefact and
# no side files.
COPY --from=build /out/turfgps /usr/local/bin/turfgps

# Stated explicitly rather than left to the :nonroot tag, so the property
# survives someone retagging the base to debug something. 65532 is distroless's
# nonroot uid; the listen port is above 1024 and needs no privilege.
USER 65532:65532

EXPOSE 8080

# Exec form, so the executable is PID 1 and receives SIGTERM directly — which
# is what the entry point's signal handling expects. Nothing wraps it: there is
# no shell in this image to wrap it with, and that is the point.
ENTRYPOINT ["/usr/local/bin/turfgps"]
