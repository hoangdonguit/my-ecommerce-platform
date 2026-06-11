# Detailed K6 Metrics

| Test | Script | Exit | HTTP reqs | Iterations | p90 ms | p95 ms | p99 ms | Max ms | HTTP fail | Checks | Accepted | Success | Failed | Error rate |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 01-functional-smoke | tests/k6/smoke-test.js | 0 | 330 | 110 | 39.406 | 43.5205 | NA | 1052.6614 | NA | NA | NA | NA | NA | NA |
| 02-idempotency | tests/k6/idempotency-test.js | 0 | 200 | 100 | 116.9173 | 452.3701 | NA | 661.4566 | NA | NA | NA | NA | NA | NA |
| 03-baseline-5rps | tests/k6/baseline-e2e-5rps.js | 0 | 301 | 301 | 46.3109 | 62.428 | NA | 171.9479 | NA | NA | 301 | NA | NA | NA |
| 04-baseline-10rps | tests/k6/baseline-e2e-10rps.js | 0 | 601 | 601 | 41.1147 | 47.3328 | NA | 75.0542 | NA | NA | 601 | NA | NA | NA |
| 05-baseline-20rps | tests/k6/baseline-e2e-20rps.js | 0 | 1201 | 1201 | 52.1623 | 62.9861 | NA | 151.9577 | NA | NA | 1201 | NA | NA | NA |
| 06-baseline-30rps | tests/k6/baseline-e2e-30rps.js | 0 | 1800 | 1800 | 57.3825 | 70.7008 | NA | 255.441 | NA | NA | 1800 | NA | NA | NA |
| 07-baseline-40rps | tests/k6/baseline-e2e-40rps.js | 0 | 2401 | 2401 | 76.6336 | 98.2367 | NA | 248.6709 | NA | NA | 2401 | NA | NA | NA |
| 08-baseline-50rps | tests/k6/baseline-e2e-50rps.js | 0 | 3001 | 3001 | 98.9225 | 119.8076 | NA | 428.3398 | NA | NA | 3001 | NA | NA | NA |
| 09-baseline-60rps | tests/k6/baseline-e2e-60rps.js | 0 | 3600 | 3600 | 120.4074 | 149.1708 | NA | 366.873 | NA | NA | 3600 | NA | NA | NA |
| 10-baseline-70rps | tests/k6/baseline-e2e-70rps.js | 0 | 4200 | 4200 | 193.516 | 225.4762 | NA | 561.2039 | NA | NA | 4200 | NA | NA | NA |
| 11-baseline-80rps | tests/k6/baseline-e2e-80rps.js | 0 | 4800 | 4800 | 211.6152 | 239.3214 | NA | 519.7739 | NA | NA | 4800 | NA | NA | NA |
| 12-baseline-light | tests/k6/baseline-light.js | 0 | 2044 | 2044 | 184.8401 | 202.3542 | NA | 448.9183 | NA | NA | 2044 | NA | NA | NA |
| 13-load-test | tests/k6/load-test.js | 0 | 20842 | 34672 | 124.0311 | 172.0402 | NA | 418.1634 | NA | NA | NA | NA | NA | NA |
| 14-flash-sale | tests/k6/flash-sale-test.js | 0 | 12993 | 12993 | 1010.0035 | 1101.7314 | NA | 1968.1477 | NA | NA | 12993 | NA | NA | NA |
| 15-flash-sale-spike | tests/k6/flash-sale-spike-test.js | 0 | 9116 | 9116 | 2858.3411 | 2992.9919 | NA | 4231.9742 | NA | NA | 9116 | NA | NA | NA |
| 16-spike-test | tests/k6/spike-test.js | 0 | 25548 | 25548 | 5015.0205 | 5322.1511 | NA | 7490.7449 | NA | NA | NA | NA | NA | NA |
| 17-stress-test-multi | tests/k6/stress-test-multi.js | 0 | 82831 | 82831 | 3998.9069 | 4759.9712 | NA | 8266.7118 | NA | NA | NA | 82831 | NA | NA |
| 18-stress-test | tests/k6/stress-test.js | 0 | 75706 | 75706 | 4310.9889 | 4653.6564 | NA | 6431.9952 | NA | NA | NA | 75706 | NA | NA |
| 19-soak-test | tests/k6/soak-test.js | 0 | 94555 | 94555 | 253.8265 | 293.2423 | NA | 2640.7286 | NA | NA | NA | NA | NA | NA |
