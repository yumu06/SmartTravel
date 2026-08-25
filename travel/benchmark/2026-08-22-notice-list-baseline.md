# Notice List Local Load-Test Baseline

Date: 2026-08-22  
Target: `GET /travel/notice/list?pageNum=1&pageSize=20`  
Environment: Windows local loopback, one Gin process, local MySQL, request logging enabled.

## Results

| Requests | Concurrency | Successful | Error rate | Successful QPS | P50 | P95 | P99 | Max |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 500 | 10 | 500 | 0% | 7,781 | 1.09 ms | 2.15 ms | 5.82 ms | 6.36 ms |
| 5,000 | 50 | 5,000 | 0% | 8,068 | 5.48 ms | 12.91 ms | 19.91 ms | 87.69 ms |
| 5,000 | 200 | 5,000 | 0% | 11,797 | 16.27 ms | 33.36 ms | 48.94 ms | 80.83 ms |
| 5,000 | 300 | 4,822 | 3.56% | ~9,922 | 25.02 ms | 77.56 ms | 166.70 ms | 259.02 ms |
| 5,000 | 400 | 3,141 | 37.18% | ~8,507 | 13.21 ms | 70.25 ms | 361.37 ms | 368.68 ms |
| 5,000 | 500 | 3,911 | 21.78% | ~8,724 | 24.64 ms | 174.45 ms | 442.31 ms | 448.31 ms |

## Findings

- Concurrency 200 is the highest tested zero-error point in this instantaneous-burst test.
- At concurrency 300 and above, failures are TCP connection refusals. They are not HTTP 5xx responses or request timeouts.
- The service remains alive and returns HTTP 200 after each burst, so the failures indicate local listener/backlog saturation rather than a process crash.
- The endpoint executes both a page query and `COUNT(*) FROM notices` per request. Under peak load, GORM recorded count queries between 220 ms and 396 ms.
- Gin request logging emitted several megabytes during the run and materially affects this local result.

## Interpretation Limits

This is a local instantaneous-burst baseline, not a production capacity claim. A production-style run should use staged ramp-up, release mode, controlled request logging, production-sized data, host resource metrics, MySQL metrics, and a load generator on another machine.

## Reproduction

```powershell
go run ./cmd/loadtest -n 5000 -c 200 -timeout 2s
```
