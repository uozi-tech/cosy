# WeakDecode benchmark

The benchmark in `benchmark_test.go` is the stable comparison point for the
map-to-struct decoder. Run it with:

```sh
go test ./map2struct -run '^$' -bench '^BenchmarkWeakDecode$' -benchmem -count=5
```

## P0 baseline: mapstructure v1.5.0

Recorded on 2026-07-18 with Go 1.26.5, darwin/arm64, Apple M5 Pro:

| run | ns/op | B/op | allocs/op |
|---:|---:|---:|---:|
| 1 | 29,771 | 8,160 | 102 |
| 2 | 29,582 | 8,160 | 102 |
| 3 | 29,810 | 8,160 | 102 |
| 4 | 28,799 | 8,160 | 102 |
| 5 | 29,078 | 8,160 | 102 |

Median: **29,582 ns/op, 8,160 B/op, 102 allocs/op**.

## P1: compiled structcodec

Recorded on the same machine and Go toolchain:

| run | ns/op | B/op | allocs/op |
|---:|---:|---:|---:|
| 1 | 2,205 | 3,194 | 60 |
| 2 | 2,205 | 3,194 | 60 |
| 3 | 2,270 | 3,194 | 60 |
| 4 | 2,393 | 3,194 | 60 |
| 5 | 2,470 | 3,194 | 60 |

Median: **2,270 ns/op, 3,194 B/op, 60 allocs/op**. Compared with the P0
median this is **13.0x faster**, uses **61% fewer bytes**, and performs **42
fewer allocations** per decode.

## Review + hot-path round: plan-driven lookup, specialised decoders

Recorded on 2026-08-21 with Go 1.27.0, darwin/arm64, Apple M5 Pro (the
module now requires Go 1.27):

| run | ns/op | B/op | allocs/op |
|---:|---:|---:|---:|
| 1 | 803 | 536 | 13 |
| 2 | 819 | 536 | 13 |
| 3 | 892 | 536 | 13 |
| 4 | 852 | 536 | 13 |
| 5 | 817 | 536 | 13 |

Median: **819 ns/op, 536 B/op, 13 allocs/op**. Compared with the P0
median this is **36.1x faster** with **89 fewer allocations** per
decode; compared with P1 it is **2.8x faster** with **47 fewer
allocations**, after the decoder stopped copying the input map per decode,
resolved converters and nested plans at compile time and started copying
slices/maps instead of aliasing the payload.
