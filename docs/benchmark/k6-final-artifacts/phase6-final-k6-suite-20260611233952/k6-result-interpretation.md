# K6 Result Interpretation

## Stable capacity baseline

Baseline tests from 5 RPS to 80 RPS completed with exit code 0.

Highest baseline p95 observed: 239.3214 ms at 11-baseline-80rps.

Interpretation: current stable capacity rating is at least 80 RPS under this lab setup.

## Heavy/degradation scenarios

- 13-load-test: exit=0, p95=172.0402 ms, p99=NA ms, max=418.1634 ms.
- 14-flash-sale: exit=0, p95=1101.7314 ms, p99=NA ms, max=1968.1477 ms.
- 15-flash-sale-spike: exit=0, p95=2992.9919 ms, p99=NA ms, max=4231.9742 ms.
- 16-spike-test: exit=0, p95=5322.1511 ms, p99=NA ms, max=7490.7449 ms.
- 17-stress-test-multi: exit=0, p95=4759.9712 ms, p99=NA ms, max=8266.7118 ms.
- 18-stress-test: exit=0, p95=4653.6564 ms, p99=NA ms, max=6431.9952 ms.
- 19-soak-test: exit=0, p95=293.2423 ms, p99=NA ms, max=2640.7286 ms.

Interpretation: heavy tests are used to identify degradation and bottlenecks, not only pass/fail.
