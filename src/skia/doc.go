// Package skia — minimal compilable Skia backend (ponytail).
//
// Design: two build modes
//   - default (no tag): pure-Go stub using image.RGBA — compiles on any host, no cgo, no libskia.
//   - -tags skia: cgo path that imports github.com/go-skia/skia and links libskia.a
//     (see Dockerfile.skia stage 2). Stub types are replaced by real Skia handles there.
//
// This split lets CI vet/build stay green without a 100MB Skia checkout,
// while the Docker builder produces the real binary. The public API is identical
// in both modes (Canvas/Font/Image), so callers never branch on build tags.
//
// Non-goals (YAGNI): no interface with one impl, no factory, no config struct.
// Ponytail ceilings flagged inline.
package skia
