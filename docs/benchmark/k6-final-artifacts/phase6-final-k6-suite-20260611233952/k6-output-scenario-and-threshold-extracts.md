# K6 Output Scenario and Threshold Extracts


## 01-functional-smoke

Script: tests/k6/smoke-test.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/smoke-test.js
        output: -

     scenarios: (100.00%) 1 scenario, 1 max VUs, 2m30s max duration (incl. graceful stop):
              * default: 1 looping VUs for 2m0s (gracefulStop: 30s)


running (0m01.0s), 1/1 VUs, 0 complete and 0 interrupted iterations
default   [   1% ] 1 VUs  0m01.0s/2m0s

running (0m02.0s), 1/1 VUs, 0 complete and 0 interrupted iterations
default   [   2% ] 1 VUs  0m02.0s/2m0s

running (0m03.0s), 1/1 VUs, 1 complete and 0 interrupted iterations
default   [   2% ] 1 VUs  0m03.0s/2m0s

running (0m04.0s), 1/1 VUs, 2 complete and 0 interrupted iterations
default   [   3% ] 1 VUs  0m04.0s/2m0s

running (0m05.0s), 1/1 VUs, 3 complete and 0 interrupted iterations
default   [   4% ] 1 VUs  0m05.0s/2m0s

running (0m06.0s), 1/1 VUs, 4 complete and 0 interrupted iterations
default   [   5% ] 1 VUs  0m06.0s/2m0s

running (0m07.0s), 1/1 VUs, 5 complete and 0 interrupted iterations
default   [   6% ] 1 VUs  0m07.0s/2m0s

running (0m08.0s), 1/1 VUs, 6 complete and 0 interrupted iterations
default   [   7% ] 1 VUs  0m08.0s/2m0s

running (0m09.0s), 1/1 VUs, 7 complete and 0 interrupted iterations
default   [   7% ] 1 VUs  0m09.0s/2m0s

running (0m10.0s), 1/1 VUs, 8 complete and 0 interrupted iterations
default   [   8% ] 1 VUs  0m10.0s/2m0s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<500' p(95)=43.52ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 330     2.746915/s
    checks_succeeded...: 100.00% 330 out of 330
    checks_failed......: 0.00%   0 out of 330

    ✓ health 200
    ✓ services health 200
    ✓ order created 2xx

    HTTP
    http_req_duration..............: avg=30.35ms min=7.24ms med=27.62ms max=1.05s p(90)=39.4ms p(95)=43.52ms
      { expected_response:true }...: avg=30.35ms min=7.24ms med=27.62ms max=1.05s p(90)=39.4ms p(95)=43.52ms
    http_req_failed................: 0.00%  0 out of 330
    http_reqs......................: 330    2.746915/s

    EXECUTION
    iteration_duration.............: avg=1.09s   min=1.05s  med=1.08s   max=2.15s p(90)=1.1s   p(95)=1.11s  
    iterations.....................: 110    0.915638/s
    vus............................: 1      min=1        max=1
    vus_max........................: 1      min=1        max=1

    NETWORK
    data_received..................: 179 kB 1.5 kB/s
    data_sent......................: 75 kB  625 B/s




running (2m00.1s), 0/1 VUs, 110 complete and 0 interrupted iterations
default ✓ [ 100% ] 1 VUs  2m0s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 330     2.746915/s
    checks_succeeded...: 100.00% 330 out of 330
    checks_failed......: 0.00%   0 out of 330

    ✓ health 200
    ✓ services health 200
    ✓ order created 2xx

    HTTP
    http_req_duration..............: avg=30.35ms min=7.24ms med=27.62ms max=1.05s p(90)=39.4ms p(95)=43.52ms
      { expected_response:true }...: avg=30.35ms min=7.24ms med=27.62ms max=1.05s p(90)=39.4ms p(95)=43.52ms
    http_req_failed................: 0.00%  0 out of 330
    http_reqs......................: 330    2.746915/s

    EXECUTION
    iteration_duration.............: avg=1.09s   min=1.05s  med=1.08s   max=2.15s p(90)=1.1s   p(95)=1.11s  
    iterations.....................: 110    0.915638/s
    vus............................: 1      min=1        max=1
    vus_max........................: 1      min=1        max=1

    NETWORK
    data_received..................: 179 kB 1.5 kB/s
    data_sent......................: 75 kB  625 B/s




running (2m00.1s), 0/1 VUs, 110 complete and 0 interrupted iterations
default ✓ [ 100% ] 1 VUs  2m0s


## 02-idempotency

Script: tests/k6/idempotency-test.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/idempotency-test.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 10m30s max duration (incl. graceful stop):
              * default: 100 iterations shared among 20 VUs (maxDuration: 10m0s, gracefulStop: 30s)


running (00m01.0s), 20/20 VUs, 54 complete and 0 interrupted iterations
default   [  54% ] 20 VUs  00m01.0s/10m0s  054/100 shared iters


  █ THRESHOLDS 

    idempotency_correct
    ✓ 'rate>0.99' rate=100.00%


  █ TOTAL RESULTS 

    checks_total.......: 100     72.582164/s
    checks_succeeded...: 100.00% 100 out of 100
    checks_failed......: 0.00%   0 out of 100

    ✓ idem status 2xx

    CUSTOM
    idempotency_correct............: 100.00% 100 out of 100

    HTTP
    http_req_duration..............: avg=77.76ms  min=11.92ms  med=41.12ms  max=661.45ms p(90)=116.91ms p(95)=452.37ms
      { expected_response:true }...: avg=77.76ms  min=11.92ms  med=41.12ms  max=661.45ms p(90)=116.91ms p(95)=452.37ms
    http_req_failed................: 0.00%   0 out of 200
    http_reqs......................: 200     145.164327/s

    EXECUTION
    iteration_duration.............: avg=257.41ms min=144.18ms med=187.85ms max=790.57ms p(90)=585.13ms p(95)=685.36ms

### Threshold block

  █ THRESHOLDS 

    idempotency_correct
    ✓ 'rate>0.99' rate=100.00%


  █ TOTAL RESULTS 

    checks_total.......: 100     72.582164/s
    checks_succeeded...: 100.00% 100 out of 100
    checks_failed......: 0.00%   0 out of 100

    ✓ idem status 2xx

    CUSTOM
    idempotency_correct............: 100.00% 100 out of 100

    HTTP
    http_req_duration..............: avg=77.76ms  min=11.92ms  med=41.12ms  max=661.45ms p(90)=116.91ms p(95)=452.37ms
      { expected_response:true }...: avg=77.76ms  min=11.92ms  med=41.12ms  max=661.45ms p(90)=116.91ms p(95)=452.37ms
    http_req_failed................: 0.00%   0 out of 200
    http_reqs......................: 200     145.164327/s

    EXECUTION
    iteration_duration.............: avg=257.41ms min=144.18ms med=187.85ms max=790.57ms p(90)=585.13ms p(95)=685.36ms
    iterations.....................: 100     72.582164/s
    vus............................: 20      min=20         max=20
    vus_max........................: 20      min=20         max=20

    NETWORK
    data_received..................: 158 kB  115 kB/s
    data_sent......................: 72 kB   52 kB/s




running (00m01.4s), 00/20 VUs, 100 complete and 0 interrupted iterations
default ✓ [ 100% ] 20 VUs  00m01.4s/10m0s  100/100 shared iters

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 100     72.582164/s
    checks_succeeded...: 100.00% 100 out of 100
    checks_failed......: 0.00%   0 out of 100

    ✓ idem status 2xx

    CUSTOM
    idempotency_correct............: 100.00% 100 out of 100

    HTTP
    http_req_duration..............: avg=77.76ms  min=11.92ms  med=41.12ms  max=661.45ms p(90)=116.91ms p(95)=452.37ms
      { expected_response:true }...: avg=77.76ms  min=11.92ms  med=41.12ms  max=661.45ms p(90)=116.91ms p(95)=452.37ms
    http_req_failed................: 0.00%   0 out of 200
    http_reqs......................: 200     145.164327/s

    EXECUTION
    iteration_duration.............: avg=257.41ms min=144.18ms med=187.85ms max=790.57ms p(90)=585.13ms p(95)=685.36ms
    iterations.....................: 100     72.582164/s
    vus............................: 20      min=20         max=20
    vus_max........................: 20      min=20         max=20

    NETWORK
    data_received..................: 158 kB  115 kB/s
    data_sent......................: 72 kB   52 kB/s




running (00m01.4s), 00/20 VUs, 100 complete and 0 interrupted iterations
default ✓ [ 100% ] 20 VUs  00m01.4s/10m0s  100/100 shared iters


## 03-baseline-5rps

Script: tests/k6/baseline-e2e-5rps.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-e2e-5rps.js
        output: -

     scenarios: (100.00%) 1 scenario, 30 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_e2e_5rps: 5.00 iterations/s for 1m0s (maxVUs: 10-30, gracefulStop: 30s)


running (0m01.0s), 01/10 VUs, 4 complete and 0 interrupted iterations
baseline_e2e_5rps   [   2% ] 01/10 VUs  0m01.0s/1m0s  5.00 iters/s

running (0m02.0s), 01/10 VUs, 9 complete and 0 interrupted iterations
baseline_e2e_5rps   [   3% ] 01/10 VUs  0m02.0s/1m0s  5.00 iters/s

running (0m03.0s), 01/10 VUs, 14 complete and 0 interrupted iterations
baseline_e2e_5rps   [   5% ] 01/10 VUs  0m03.0s/1m0s  5.00 iters/s

running (0m04.0s), 01/10 VUs, 19 complete and 0 interrupted iterations
baseline_e2e_5rps   [   7% ] 01/10 VUs  0m04.0s/1m0s  5.00 iters/s

running (0m05.0s), 01/10 VUs, 24 complete and 0 interrupted iterations
baseline_e2e_5rps   [   8% ] 01/10 VUs  0m05.0s/1m0s  5.00 iters/s

running (0m06.0s), 01/10 VUs, 29 complete and 0 interrupted iterations
baseline_e2e_5rps   [  10% ] 01/10 VUs  0m06.0s/1m0s  5.00 iters/s

running (0m07.0s), 01/10 VUs, 34 complete and 0 interrupted iterations
baseline_e2e_5rps   [  12% ] 01/10 VUs  0m07.0s/1m0s  5.00 iters/s

running (0m08.0s), 01/10 VUs, 39 complete and 0 interrupted iterations
baseline_e2e_5rps   [  13% ] 01/10 VUs  0m08.0s/1m0s  5.00 iters/s

running (0m09.0s), 01/10 VUs, 44 complete and 0 interrupted iterations
baseline_e2e_5rps   [  15% ] 01/10 VUs  0m09.0s/1m0s  5.00 iters/s

running (0m10.0s), 01/10 VUs, 49 complete and 0 interrupted iterations
baseline_e2e_5rps   [  17% ] 01/10 VUs  0m10.0s/1m0s  5.00 iters/s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<1500' p(95)=62.42ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 301     4.995857/s
    checks_succeeded...: 100.00% 301 out of 301
    checks_failed......: 0.00%   0 out of 301

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 301    4.995857/s
    unexpected_error_rate..........: 0.00%  0 out of 301

    HTTP
    http_req_duration..............: avg=36.1ms   min=19.38ms  med=30.41ms  max=171.94ms p(90)=46.31ms  p(95)=62.42ms 
      { expected_response:true }...: avg=36.1ms   min=19.38ms  med=30.41ms  max=171.94ms p(90)=46.31ms  p(95)=62.42ms 
    http_req_failed................: 0.00%  0 out of 301
    http_reqs......................: 301    4.995857/s

    EXECUTION
    iteration_duration.............: avg=237.18ms min=220.56ms med=231.47ms max=372.18ms p(90)=246.76ms p(95)=263.95ms
    iterations.....................: 301    4.995857/s
    vus............................: 1      min=1        max=1 
    vus_max........................: 10     min=10       max=10

    NETWORK
    data_received..................: 291 kB 4.8 kB/s
    data_sent......................: 151 kB 2.5 kB/s




running (1m00.2s), 00/10 VUs, 301 complete and 0 interrupted iterations
baseline_e2e_5rps ✓ [ 100% ] 00/10 VUs  1m0s  5.00 iters/s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 301     4.995857/s
    checks_succeeded...: 100.00% 301 out of 301
    checks_failed......: 0.00%   0 out of 301

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 301    4.995857/s
    unexpected_error_rate..........: 0.00%  0 out of 301

    HTTP
    http_req_duration..............: avg=36.1ms   min=19.38ms  med=30.41ms  max=171.94ms p(90)=46.31ms  p(95)=62.42ms 
      { expected_response:true }...: avg=36.1ms   min=19.38ms  med=30.41ms  max=171.94ms p(90)=46.31ms  p(95)=62.42ms 
    http_req_failed................: 0.00%  0 out of 301
    http_reqs......................: 301    4.995857/s

    EXECUTION
    iteration_duration.............: avg=237.18ms min=220.56ms med=231.47ms max=372.18ms p(90)=246.76ms p(95)=263.95ms
    iterations.....................: 301    4.995857/s
    vus............................: 1      min=1        max=1 
    vus_max........................: 10     min=10       max=10

    NETWORK
    data_received..................: 291 kB 4.8 kB/s
    data_sent......................: 151 kB 2.5 kB/s




running (1m00.2s), 00/10 VUs, 301 complete and 0 interrupted iterations
baseline_e2e_5rps ✓ [ 100% ] 00/10 VUs  1m0s  5.00 iters/s


## 04-baseline-10rps

Script: tests/k6/baseline-e2e-10rps.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-e2e-10rps.js
        output: -

     scenarios: (100.00%) 1 scenario, 60 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_e2e_10rps: 10.00 iterations/s for 1m0s (maxVUs: 20-60, gracefulStop: 30s)


running (0m01.0s), 01/20 VUs, 9 complete and 0 interrupted iterations
baseline_e2e_10rps   [   2% ] 01/20 VUs  0m01.0s/1m0s  10.00 iters/s

running (0m02.0s), 01/20 VUs, 19 complete and 0 interrupted iterations
baseline_e2e_10rps   [   3% ] 01/20 VUs  0m02.0s/1m0s  10.00 iters/s

running (0m03.0s), 01/20 VUs, 29 complete and 0 interrupted iterations
baseline_e2e_10rps   [   5% ] 01/20 VUs  0m03.0s/1m0s  10.00 iters/s

running (0m04.0s), 01/20 VUs, 39 complete and 0 interrupted iterations
baseline_e2e_10rps   [   7% ] 01/20 VUs  0m04.0s/1m0s  10.00 iters/s

running (0m05.0s), 01/20 VUs, 49 complete and 0 interrupted iterations
baseline_e2e_10rps   [   8% ] 01/20 VUs  0m05.0s/1m0s  10.00 iters/s

running (0m06.0s), 01/20 VUs, 59 complete and 0 interrupted iterations
baseline_e2e_10rps   [  10% ] 01/20 VUs  0m06.0s/1m0s  10.00 iters/s

running (0m07.0s), 01/20 VUs, 69 complete and 0 interrupted iterations
baseline_e2e_10rps   [  12% ] 01/20 VUs  0m07.0s/1m0s  10.00 iters/s

running (0m08.0s), 01/20 VUs, 79 complete and 0 interrupted iterations
baseline_e2e_10rps   [  13% ] 01/20 VUs  0m08.0s/1m0s  10.00 iters/s

running (0m09.0s), 01/20 VUs, 89 complete and 0 interrupted iterations
baseline_e2e_10rps   [  15% ] 01/20 VUs  0m09.0s/1m0s  10.00 iters/s

running (0m10.0s), 01/20 VUs, 99 complete and 0 interrupted iterations
baseline_e2e_10rps   [  17% ] 01/20 VUs  0m10.0s/1m0s  10.00 iters/s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<1500' p(95)=47.33ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 601     9.995467/s
    checks_succeeded...: 100.00% 601 out of 601
    checks_failed......: 0.00%   0 out of 601

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 601    9.995467/s
    unexpected_error_rate..........: 0.00%  0 out of 601

    HTTP
    http_req_duration..............: avg=32.15ms  min=18.63ms  med=30.67ms  max=75.05ms  p(90)=41.11ms  p(95)=47.33ms 
      { expected_response:true }...: avg=32.15ms  min=18.63ms  med=30.67ms  max=75.05ms  p(90)=41.11ms  p(95)=47.33ms 
    http_req_failed................: 0.00%  0 out of 601
    http_reqs......................: 601    9.995467/s

    EXECUTION
    iteration_duration.............: avg=133.04ms min=119.53ms med=131.61ms max=175.54ms p(90)=142.85ms p(95)=148.06ms
    iterations.....................: 601    9.995467/s
    vus............................: 1      min=1        max=1 
    vus_max........................: 20     min=20       max=20

    NETWORK
    data_received..................: 583 kB 9.7 kB/s
    data_sent......................: 305 kB 5.1 kB/s




running (1m00.1s), 00/20 VUs, 601 complete and 0 interrupted iterations
baseline_e2e_10rps ✓ [ 100% ] 00/20 VUs  1m0s  10.00 iters/s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 601     9.995467/s
    checks_succeeded...: 100.00% 601 out of 601
    checks_failed......: 0.00%   0 out of 601

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 601    9.995467/s
    unexpected_error_rate..........: 0.00%  0 out of 601

    HTTP
    http_req_duration..............: avg=32.15ms  min=18.63ms  med=30.67ms  max=75.05ms  p(90)=41.11ms  p(95)=47.33ms 
      { expected_response:true }...: avg=32.15ms  min=18.63ms  med=30.67ms  max=75.05ms  p(90)=41.11ms  p(95)=47.33ms 
    http_req_failed................: 0.00%  0 out of 601
    http_reqs......................: 601    9.995467/s

    EXECUTION
    iteration_duration.............: avg=133.04ms min=119.53ms med=131.61ms max=175.54ms p(90)=142.85ms p(95)=148.06ms
    iterations.....................: 601    9.995467/s
    vus............................: 1      min=1        max=1 
    vus_max........................: 20     min=20       max=20

    NETWORK
    data_received..................: 583 kB 9.7 kB/s
    data_sent......................: 305 kB 5.1 kB/s




running (1m00.1s), 00/20 VUs, 601 complete and 0 interrupted iterations
baseline_e2e_10rps ✓ [ 100% ] 00/20 VUs  1m0s  10.00 iters/s


## 05-baseline-20rps

Script: tests/k6/baseline-e2e-20rps.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-e2e-20rps.js
        output: -

     scenarios: (100.00%) 1 scenario, 120 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_e2e_20rps: 20.00 iterations/s for 1m0s (maxVUs: 40-120, gracefulStop: 30s)


running (0m01.0s), 002/040 VUs, 18 complete and 0 interrupted iterations
baseline_e2e_20rps   [   2% ] 002/040 VUs  0m01.0s/1m0s  20.00 iters/s

running (0m02.0s), 002/040 VUs, 38 complete and 0 interrupted iterations
baseline_e2e_20rps   [   3% ] 002/040 VUs  0m02.0s/1m0s  20.00 iters/s

running (0m03.0s), 002/040 VUs, 58 complete and 0 interrupted iterations
baseline_e2e_20rps   [   5% ] 002/040 VUs  0m03.0s/1m0s  20.00 iters/s

running (0m04.0s), 002/040 VUs, 78 complete and 0 interrupted iterations
baseline_e2e_20rps   [   7% ] 002/040 VUs  0m04.0s/1m0s  20.00 iters/s

running (0m05.0s), 003/040 VUs, 97 complete and 0 interrupted iterations
baseline_e2e_20rps   [   8% ] 003/040 VUs  0m05.0s/1m0s  20.00 iters/s

running (0m06.0s), 002/040 VUs, 118 complete and 0 interrupted iterations
baseline_e2e_20rps   [  10% ] 002/040 VUs  0m06.0s/1m0s  20.00 iters/s

running (0m07.0s), 003/040 VUs, 137 complete and 0 interrupted iterations
baseline_e2e_20rps   [  12% ] 003/040 VUs  0m07.0s/1m0s  20.00 iters/s

running (0m08.0s), 002/040 VUs, 158 complete and 0 interrupted iterations
baseline_e2e_20rps   [  13% ] 002/040 VUs  0m08.0s/1m0s  20.00 iters/s

running (0m09.0s), 002/040 VUs, 178 complete and 0 interrupted iterations
baseline_e2e_20rps   [  15% ] 002/040 VUs  0m09.0s/1m0s  20.00 iters/s

running (0m10.0s), 002/040 VUs, 198 complete and 0 interrupted iterations
baseline_e2e_20rps   [  17% ] 002/040 VUs  0m10.0s/1m0s  20.00 iters/s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<1500' p(95)=62.98ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 1201    19.970039/s
    checks_succeeded...: 100.00% 1201 out of 1201
    checks_failed......: 0.00%   0 out of 1201

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 1201   19.970039/s
    unexpected_error_rate..........: 0.00%  0 out of 1201

    HTTP
    http_req_duration..............: avg=37.23ms min=20.34ms  med=33.89ms max=151.95ms p(90)=52.16ms  p(95)=62.98ms 
      { expected_response:true }...: avg=37.23ms min=20.34ms  med=33.89ms max=151.95ms p(90)=52.16ms  p(95)=62.98ms 
    http_req_failed................: 0.00%  0 out of 1201
    http_reqs......................: 1201   19.970039/s

    EXECUTION
    iteration_duration.............: avg=138.1ms min=120.65ms med=134.8ms max=252.25ms p(90)=152.89ms p(95)=164.09ms
    iterations.....................: 1201   19.970039/s
    vus............................: 3      min=2         max=3 
    vus_max........................: 40     min=40        max=40

    NETWORK
    data_received..................: 1.2 MB 19 kB/s
    data_sent......................: 610 kB 10 kB/s




running (1m00.1s), 000/040 VUs, 1201 complete and 0 interrupted iterations
baseline_e2e_20rps ✓ [ 100% ] 000/040 VUs  1m0s  20.00 iters/s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 1201    19.970039/s
    checks_succeeded...: 100.00% 1201 out of 1201
    checks_failed......: 0.00%   0 out of 1201

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 1201   19.970039/s
    unexpected_error_rate..........: 0.00%  0 out of 1201

    HTTP
    http_req_duration..............: avg=37.23ms min=20.34ms  med=33.89ms max=151.95ms p(90)=52.16ms  p(95)=62.98ms 
      { expected_response:true }...: avg=37.23ms min=20.34ms  med=33.89ms max=151.95ms p(90)=52.16ms  p(95)=62.98ms 
    http_req_failed................: 0.00%  0 out of 1201
    http_reqs......................: 1201   19.970039/s

    EXECUTION
    iteration_duration.............: avg=138.1ms min=120.65ms med=134.8ms max=252.25ms p(90)=152.89ms p(95)=164.09ms
    iterations.....................: 1201   19.970039/s
    vus............................: 3      min=2         max=3 
    vus_max........................: 40     min=40        max=40

    NETWORK
    data_received..................: 1.2 MB 19 kB/s
    data_sent......................: 610 kB 10 kB/s




running (1m00.1s), 000/040 VUs, 1201 complete and 0 interrupted iterations
baseline_e2e_20rps ✓ [ 100% ] 000/040 VUs  1m0s  20.00 iters/s


## 06-baseline-30rps

Script: tests/k6/baseline-e2e-30rps.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-e2e-30rps.js
        output: -

     scenarios: (100.00%) 1 scenario, 180 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_e2e_30rps: 30.00 iterations/s for 1m0s (maxVUs: 60-180, gracefulStop: 30s)


running (0m01.0s), 006/060 VUs, 24 complete and 0 interrupted iterations
baseline_e2e_30rps   [   2% ] 006/060 VUs  0m01.0s/1m0s  30.00 iters/s

running (0m02.0s), 004/060 VUs, 56 complete and 0 interrupted iterations
baseline_e2e_30rps   [   3% ] 004/060 VUs  0m02.0s/1m0s  30.00 iters/s

running (0m03.0s), 004/060 VUs, 86 complete and 0 interrupted iterations
baseline_e2e_30rps   [   5% ] 004/060 VUs  0m03.0s/1m0s  30.00 iters/s

running (0m04.0s), 003/060 VUs, 117 complete and 0 interrupted iterations
baseline_e2e_30rps   [   7% ] 003/060 VUs  0m04.0s/1m0s  30.00 iters/s

running (0m05.0s), 004/060 VUs, 146 complete and 0 interrupted iterations
baseline_e2e_30rps   [   8% ] 004/060 VUs  0m05.0s/1m0s  30.00 iters/s

running (0m06.0s), 004/060 VUs, 176 complete and 0 interrupted iterations
baseline_e2e_30rps   [  10% ] 004/060 VUs  0m06.0s/1m0s  30.00 iters/s

running (0m07.0s), 004/060 VUs, 206 complete and 0 interrupted iterations
baseline_e2e_30rps   [  12% ] 004/060 VUs  0m07.0s/1m0s  30.00 iters/s

running (0m08.0s), 004/060 VUs, 236 complete and 0 interrupted iterations
baseline_e2e_30rps   [  13% ] 004/060 VUs  0m08.0s/1m0s  30.00 iters/s

running (0m09.0s), 004/060 VUs, 266 complete and 0 interrupted iterations
baseline_e2e_30rps   [  15% ] 004/060 VUs  0m09.0s/1m0s  30.00 iters/s

running (0m10.0s), 004/060 VUs, 296 complete and 0 interrupted iterations
baseline_e2e_30rps   [  17% ] 004/060 VUs  0m10.0s/1m0s  30.00 iters/s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<1500' p(95)=70.7ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 1800    29.952542/s
    checks_succeeded...: 100.00% 1800 out of 1800
    checks_failed......: 0.00%   0 out of 1800

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 1800   29.952542/s
    unexpected_error_rate..........: 0.00%  0 out of 1800

    HTTP
    http_req_duration..............: avg=39.94ms  min=17.33ms  med=35.25ms max=255.44ms p(90)=57.38ms p(95)=70.7ms  
      { expected_response:true }...: avg=39.94ms  min=17.33ms  med=35.25ms max=255.44ms p(90)=57.38ms p(95)=70.7ms  
    http_req_failed................: 0.00%  0 out of 1800
    http_reqs......................: 1800   29.952542/s

    EXECUTION
    iteration_duration.............: avg=141.88ms min=118.19ms med=136.1ms max=411.2ms  p(90)=161ms   p(95)=175.28ms
    iterations.....................: 1800   29.952542/s
    vus............................: 4      min=3         max=6 
    vus_max........................: 60     min=60        max=60

    NETWORK
    data_received..................: 1.7 MB 29 kB/s
    data_sent......................: 915 kB 15 kB/s




running (1m00.1s), 000/060 VUs, 1800 complete and 0 interrupted iterations
baseline_e2e_30rps ✓ [ 100% ] 000/060 VUs  1m0s  30.00 iters/s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 1800    29.952542/s
    checks_succeeded...: 100.00% 1800 out of 1800
    checks_failed......: 0.00%   0 out of 1800

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 1800   29.952542/s
    unexpected_error_rate..........: 0.00%  0 out of 1800

    HTTP
    http_req_duration..............: avg=39.94ms  min=17.33ms  med=35.25ms max=255.44ms p(90)=57.38ms p(95)=70.7ms  
      { expected_response:true }...: avg=39.94ms  min=17.33ms  med=35.25ms max=255.44ms p(90)=57.38ms p(95)=70.7ms  
    http_req_failed................: 0.00%  0 out of 1800
    http_reqs......................: 1800   29.952542/s

    EXECUTION
    iteration_duration.............: avg=141.88ms min=118.19ms med=136.1ms max=411.2ms  p(90)=161ms   p(95)=175.28ms
    iterations.....................: 1800   29.952542/s
    vus............................: 4      min=3         max=6 
    vus_max........................: 60     min=60        max=60

    NETWORK
    data_received..................: 1.7 MB 29 kB/s
    data_sent......................: 915 kB 15 kB/s




running (1m00.1s), 000/060 VUs, 1800 complete and 0 interrupted iterations
baseline_e2e_30rps ✓ [ 100% ] 000/060 VUs  1m0s  30.00 iters/s


## 07-baseline-40rps

Script: tests/k6/baseline-e2e-40rps.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-e2e-40rps.js
        output: -

     scenarios: (100.00%) 1 scenario, 240 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_e2e_40rps: 40.00 iterations/s for 1m0s (maxVUs: 80-240, gracefulStop: 30s)


running (0m01.0s), 005/080 VUs, 35 complete and 0 interrupted iterations
baseline_e2e_40rps   [   2% ] 005/080 VUs  0m01.0s/1m0s  40.00 iters/s

running (0m02.0s), 005/080 VUs, 75 complete and 0 interrupted iterations
baseline_e2e_40rps   [   3% ] 005/080 VUs  0m02.0s/1m0s  40.00 iters/s

running (0m03.0s), 006/080 VUs, 114 complete and 0 interrupted iterations
baseline_e2e_40rps   [   5% ] 006/080 VUs  0m03.0s/1m0s  40.00 iters/s

running (0m04.0s), 006/080 VUs, 154 complete and 0 interrupted iterations
baseline_e2e_40rps   [   7% ] 006/080 VUs  0m04.0s/1m0s  40.00 iters/s

running (0m05.0s), 010/080 VUs, 190 complete and 0 interrupted iterations
baseline_e2e_40rps   [   8% ] 010/080 VUs  0m05.0s/1m0s  40.00 iters/s

running (0m06.0s), 006/080 VUs, 234 complete and 0 interrupted iterations
baseline_e2e_40rps   [  10% ] 006/080 VUs  0m06.0s/1m0s  40.00 iters/s

running (0m07.0s), 008/080 VUs, 272 complete and 0 interrupted iterations
baseline_e2e_40rps   [  12% ] 008/080 VUs  0m07.0s/1m0s  40.00 iters/s

running (0m08.0s), 006/080 VUs, 314 complete and 0 interrupted iterations
baseline_e2e_40rps   [  13% ] 006/080 VUs  0m08.0s/1m0s  40.00 iters/s

running (0m09.0s), 007/080 VUs, 353 complete and 0 interrupted iterations
baseline_e2e_40rps   [  15% ] 007/080 VUs  0m09.0s/1m0s  40.00 iters/s

running (0m10.0s), 005/080 VUs, 395 complete and 0 interrupted iterations
baseline_e2e_40rps   [  17% ] 005/080 VUs  0m10.0s/1m0s  40.00 iters/s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<1500' p(95)=98.23ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 2401    39.929437/s
    checks_succeeded...: 100.00% 2401 out of 2401
    checks_failed......: 0.00%   0 out of 2401

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 2401   39.929437/s
    unexpected_error_rate..........: 0.00%  0 out of 2401

    HTTP
    http_req_duration..............: avg=48.4ms   min=19.4ms  med=41.2ms   max=248.67ms p(90)=76.63ms  p(95)=98.23ms 
      { expected_response:true }...: avg=48.4ms   min=19.4ms  med=41.2ms   max=248.67ms p(90)=76.63ms  p(95)=98.23ms 
    http_req_failed................: 0.00%  0 out of 2401
    http_reqs......................: 2401   39.929437/s

    EXECUTION
    iteration_duration.............: avg=149.23ms min=120.5ms med=141.99ms max=349.25ms p(90)=177.05ms p(95)=198.67ms
    iterations.....................: 2401   39.929437/s
    vus............................: 6      min=5         max=10
    vus_max........................: 80     min=80        max=80

    NETWORK
    data_received..................: 2.3 MB 39 kB/s
    data_sent......................: 1.2 MB 20 kB/s




running (1m00.1s), 000/080 VUs, 2401 complete and 0 interrupted iterations
baseline_e2e_40rps ✓ [ 100% ] 000/080 VUs  1m0s  40.00 iters/s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 2401    39.929437/s
    checks_succeeded...: 100.00% 2401 out of 2401
    checks_failed......: 0.00%   0 out of 2401

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 2401   39.929437/s
    unexpected_error_rate..........: 0.00%  0 out of 2401

    HTTP
    http_req_duration..............: avg=48.4ms   min=19.4ms  med=41.2ms   max=248.67ms p(90)=76.63ms  p(95)=98.23ms 
      { expected_response:true }...: avg=48.4ms   min=19.4ms  med=41.2ms   max=248.67ms p(90)=76.63ms  p(95)=98.23ms 
    http_req_failed................: 0.00%  0 out of 2401
    http_reqs......................: 2401   39.929437/s

    EXECUTION
    iteration_duration.............: avg=149.23ms min=120.5ms med=141.99ms max=349.25ms p(90)=177.05ms p(95)=198.67ms
    iterations.....................: 2401   39.929437/s
    vus............................: 6      min=5         max=10
    vus_max........................: 80     min=80        max=80

    NETWORK
    data_received..................: 2.3 MB 39 kB/s
    data_sent......................: 1.2 MB 20 kB/s




running (1m00.1s), 000/080 VUs, 2401 complete and 0 interrupted iterations
baseline_e2e_40rps ✓ [ 100% ] 000/080 VUs  1m0s  40.00 iters/s


## 08-baseline-50rps

Script: tests/k6/baseline-e2e-50rps.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-e2e-50rps.js
        output: -

     scenarios: (100.00%) 1 scenario, 300 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_e2e_50rps: 50.00 iterations/s for 1m0s (maxVUs: 100-300, gracefulStop: 30s)


running (0m01.0s), 007/100 VUs, 43 complete and 0 interrupted iterations
baseline_e2e_50rps   [   2% ] 007/100 VUs  0m01.0s/1m0s  50.00 iters/s

running (0m02.0s), 007/100 VUs, 93 complete and 0 interrupted iterations
baseline_e2e_50rps   [   3% ] 007/100 VUs  0m02.0s/1m0s  50.00 iters/s

running (0m03.0s), 008/100 VUs, 142 complete and 0 interrupted iterations
baseline_e2e_50rps   [   5% ] 008/100 VUs  0m03.0s/1m0s  50.00 iters/s

running (0m04.0s), 007/100 VUs, 193 complete and 0 interrupted iterations
baseline_e2e_50rps   [   7% ] 007/100 VUs  0m04.0s/1m0s  50.00 iters/s

running (0m05.0s), 007/100 VUs, 243 complete and 0 interrupted iterations
baseline_e2e_50rps   [   8% ] 007/100 VUs  0m05.0s/1m0s  50.00 iters/s

running (0m06.0s), 007/100 VUs, 293 complete and 0 interrupted iterations
baseline_e2e_50rps   [  10% ] 007/100 VUs  0m06.0s/1m0s  50.00 iters/s

running (0m07.0s), 009/100 VUs, 341 complete and 0 interrupted iterations
baseline_e2e_50rps   [  12% ] 009/100 VUs  0m07.0s/1m0s  50.00 iters/s

running (0m08.0s), 007/100 VUs, 393 complete and 0 interrupted iterations
baseline_e2e_50rps   [  13% ] 007/100 VUs  0m08.0s/1m0s  50.00 iters/s

running (0m09.0s), 008/100 VUs, 442 complete and 0 interrupted iterations
baseline_e2e_50rps   [  15% ] 008/100 VUs  0m09.0s/1m0s  50.00 iters/s

running (0m10.0s), 010/100 VUs, 490 complete and 0 interrupted iterations
baseline_e2e_50rps   [  17% ] 009/100 VUs  0m10.0s/1m0s  50.00 iters/s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<1500' p(95)=119.8ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 3001    49.897631/s
    checks_succeeded...: 100.00% 3001 out of 3001
    checks_failed......: 0.00%   0 out of 3001

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 3001   49.897631/s
    unexpected_error_rate..........: 0.00%  0 out of 3001

    HTTP
    http_req_duration..............: avg=59.01ms  min=20.13ms  med=49.67ms max=428.33ms p(90)=98.92ms  p(95)=119.8ms 
      { expected_response:true }...: avg=59.01ms  min=20.13ms  med=49.67ms max=428.33ms p(90)=98.92ms  p(95)=119.8ms 
    http_req_failed................: 0.00%  0 out of 3001
    http_reqs......................: 3001   49.897631/s

    EXECUTION
    iteration_duration.............: avg=159.84ms min=120.44ms med=150.5ms max=529.24ms p(90)=199.55ms p(95)=220.46ms
    iterations.....................: 3001   49.897631/s
    vus............................: 8      min=6         max=14 
    vus_max........................: 100    min=100       max=100

    NETWORK
    data_received..................: 2.9 MB 49 kB/s
    data_sent......................: 1.5 MB 25 kB/s




running (1m00.1s), 000/100 VUs, 3001 complete and 0 interrupted iterations
baseline_e2e_50rps ✓ [ 100% ] 000/100 VUs  1m0s  50.00 iters/s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 3001    49.897631/s
    checks_succeeded...: 100.00% 3001 out of 3001
    checks_failed......: 0.00%   0 out of 3001

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 3001   49.897631/s
    unexpected_error_rate..........: 0.00%  0 out of 3001

    HTTP
    http_req_duration..............: avg=59.01ms  min=20.13ms  med=49.67ms max=428.33ms p(90)=98.92ms  p(95)=119.8ms 
      { expected_response:true }...: avg=59.01ms  min=20.13ms  med=49.67ms max=428.33ms p(90)=98.92ms  p(95)=119.8ms 
    http_req_failed................: 0.00%  0 out of 3001
    http_reqs......................: 3001   49.897631/s

    EXECUTION
    iteration_duration.............: avg=159.84ms min=120.44ms med=150.5ms max=529.24ms p(90)=199.55ms p(95)=220.46ms
    iterations.....................: 3001   49.897631/s
    vus............................: 8      min=6         max=14 
    vus_max........................: 100    min=100       max=100

    NETWORK
    data_received..................: 2.9 MB 49 kB/s
    data_sent......................: 1.5 MB 25 kB/s




running (1m00.1s), 000/100 VUs, 3001 complete and 0 interrupted iterations
baseline_e2e_50rps ✓ [ 100% ] 000/100 VUs  1m0s  50.00 iters/s


## 09-baseline-60rps

Script: tests/k6/baseline-e2e-60rps.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-e2e-60rps.js
        output: -

     scenarios: (100.00%) 1 scenario, 360 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_e2e_60rps: 60.00 iterations/s for 1m0s (maxVUs: 120-360, gracefulStop: 30s)


running (0m01.0s), 009/120 VUs, 51 complete and 0 interrupted iterations
baseline_e2e_60rps   [   2% ] 009/120 VUs  0m01.0s/1m0s  60.00 iters/s

running (0m02.0s), 010/120 VUs, 110 complete and 0 interrupted iterations
baseline_e2e_60rps   [   3% ] 010/120 VUs  0m02.0s/1m0s  60.00 iters/s

running (0m03.0s), 009/120 VUs, 171 complete and 0 interrupted iterations
baseline_e2e_60rps   [   5% ] 009/120 VUs  0m03.0s/1m0s  60.00 iters/s

running (0m04.0s), 015/120 VUs, 225 complete and 0 interrupted iterations
baseline_e2e_60rps   [   7% ] 015/120 VUs  0m04.0s/1m0s  60.00 iters/s

running (0m05.0s), 014/120 VUs, 286 complete and 0 interrupted iterations
baseline_e2e_60rps   [   8% ] 014/120 VUs  0m05.0s/1m0s  60.00 iters/s

running (0m06.0s), 015/120 VUs, 345 complete and 0 interrupted iterations
baseline_e2e_60rps   [  10% ] 015/120 VUs  0m06.0s/1m0s  60.00 iters/s

running (0m07.0s), 010/120 VUs, 410 complete and 0 interrupted iterations
baseline_e2e_60rps   [  12% ] 010/120 VUs  0m07.0s/1m0s  60.00 iters/s

running (0m08.0s), 008/120 VUs, 472 complete and 0 interrupted iterations
baseline_e2e_60rps   [  13% ] 008/120 VUs  0m08.0s/1m0s  60.00 iters/s

running (0m09.0s), 008/120 VUs, 532 complete and 0 interrupted iterations
baseline_e2e_60rps   [  15% ] 008/120 VUs  0m09.0s/1m0s  60.00 iters/s

running (0m10.0s), 008/120 VUs, 592 complete and 0 interrupted iterations
baseline_e2e_60rps   [  17% ] 008/120 VUs  0m10.0s/1m0s  60.00 iters/s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<1500' p(95)=149.17ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 3600    59.86452/s
    checks_succeeded...: 100.00% 3600 out of 3600
    checks_failed......: 0.00%   0 out of 3600

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 3600   59.86452/s
    unexpected_error_rate..........: 0.00%  0 out of 3600

    HTTP
    http_req_duration..............: avg=70.94ms  min=20.71ms  med=59.66ms  max=366.87ms p(90)=120.4ms  p(95)=149.17ms
      { expected_response:true }...: avg=70.94ms  min=20.71ms  med=59.66ms  max=366.87ms p(90)=120.4ms  p(95)=149.17ms
    http_req_failed................: 0.00%  0 out of 3600
    http_reqs......................: 3600   59.86452/s

    EXECUTION
    iteration_duration.............: avg=171.74ms min=121.39ms med=160.29ms max=467.97ms p(90)=221.27ms p(95)=249.38ms
    iterations.....................: 3600   59.86452/s
    vus............................: 12     min=8         max=17 
    vus_max........................: 120    min=120       max=120

    NETWORK
    data_received..................: 3.5 MB 58 kB/s
    data_sent......................: 1.8 MB 30 kB/s




running (1m00.1s), 000/120 VUs, 3600 complete and 0 interrupted iterations
baseline_e2e_60rps ✓ [ 100% ] 000/120 VUs  1m0s  60.00 iters/s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 3600    59.86452/s
    checks_succeeded...: 100.00% 3600 out of 3600
    checks_failed......: 0.00%   0 out of 3600

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 3600   59.86452/s
    unexpected_error_rate..........: 0.00%  0 out of 3600

    HTTP
    http_req_duration..............: avg=70.94ms  min=20.71ms  med=59.66ms  max=366.87ms p(90)=120.4ms  p(95)=149.17ms
      { expected_response:true }...: avg=70.94ms  min=20.71ms  med=59.66ms  max=366.87ms p(90)=120.4ms  p(95)=149.17ms
    http_req_failed................: 0.00%  0 out of 3600
    http_reqs......................: 3600   59.86452/s

    EXECUTION
    iteration_duration.............: avg=171.74ms min=121.39ms med=160.29ms max=467.97ms p(90)=221.27ms p(95)=249.38ms
    iterations.....................: 3600   59.86452/s
    vus............................: 12     min=8         max=17 
    vus_max........................: 120    min=120       max=120

    NETWORK
    data_received..................: 3.5 MB 58 kB/s
    data_sent......................: 1.8 MB 30 kB/s




running (1m00.1s), 000/120 VUs, 3600 complete and 0 interrupted iterations
baseline_e2e_60rps ✓ [ 100% ] 000/120 VUs  1m0s  60.00 iters/s


## 10-baseline-70rps

Script: tests/k6/baseline-e2e-70rps.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-e2e-70rps.js
        output: -

     scenarios: (100.00%) 1 scenario, 420 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_e2e_70rps: 70.00 iterations/s for 1m0s (maxVUs: 140-420, gracefulStop: 30s)


running (0m01.0s), 010/140 VUs, 60 complete and 0 interrupted iterations
baseline_e2e_70rps   [   2% ] 010/140 VUs  0m01.0s/1m0s  70.00 iters/s

running (0m02.0s), 012/140 VUs, 128 complete and 0 interrupted iterations
baseline_e2e_70rps   [   3% ] 012/140 VUs  0m02.0s/1m0s  70.00 iters/s

running (0m03.0s), 012/140 VUs, 198 complete and 0 interrupted iterations
baseline_e2e_70rps   [   5% ] 012/140 VUs  0m03.0s/1m0s  70.00 iters/s

running (0m04.0s), 014/140 VUs, 266 complete and 0 interrupted iterations
baseline_e2e_70rps   [   7% ] 014/140 VUs  0m04.0s/1m0s  70.00 iters/s

running (0m05.0s), 014/140 VUs, 336 complete and 0 interrupted iterations
baseline_e2e_70rps   [   8% ] 014/140 VUs  0m05.0s/1m0s  70.00 iters/s

running (0m06.0s), 012/140 VUs, 408 complete and 0 interrupted iterations
baseline_e2e_70rps   [  10% ] 012/140 VUs  0m06.0s/1m0s  70.00 iters/s

running (0m07.0s), 013/140 VUs, 477 complete and 0 interrupted iterations
baseline_e2e_70rps   [  12% ] 013/140 VUs  0m07.0s/1m0s  70.00 iters/s

running (0m08.0s), 011/140 VUs, 549 complete and 0 interrupted iterations
baseline_e2e_70rps   [  13% ] 011/140 VUs  0m08.0s/1m0s  70.00 iters/s

running (0m09.0s), 010/140 VUs, 620 complete and 0 interrupted iterations
baseline_e2e_70rps   [  15% ] 010/140 VUs  0m09.0s/1m0s  70.00 iters/s

running (0m10.0s), 013/140 VUs, 687 complete and 0 interrupted iterations
baseline_e2e_70rps   [  17% ] 013/140 VUs  0m10.0s/1m0s  70.00 iters/s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<1500' p(95)=225.47ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 4200    69.753328/s
    checks_succeeded...: 100.00% 4200 out of 4200
    checks_failed......: 0.00%   0 out of 4200

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 4200   69.753328/s
    unexpected_error_rate..........: 0.00%  0 out of 4200

    HTTP
    http_req_duration..............: avg=109.77ms min=21.56ms  med=95.78ms  max=561.2ms  p(90)=193.51ms p(95)=225.47ms
      { expected_response:true }...: avg=109.77ms min=21.56ms  med=95.78ms  max=561.2ms  p(90)=193.51ms p(95)=225.47ms
    http_req_failed................: 0.00%  0 out of 4200
    http_reqs......................: 4200   69.753328/s

    EXECUTION
    iteration_duration.............: avg=210.58ms min=121.72ms med=196.46ms max=661.35ms p(90)=293.94ms p(95)=326.04ms
    iterations.....................: 4200   69.753328/s
    vus............................: 14     min=9         max=30 
    vus_max........................: 140    min=140       max=140

    NETWORK
    data_received..................: 4.1 MB 68 kB/s
    data_sent......................: 2.1 MB 36 kB/s




running (1m00.2s), 000/140 VUs, 4200 complete and 0 interrupted iterations
baseline_e2e_70rps ✓ [ 100% ] 000/140 VUs  1m0s  70.00 iters/s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 4200    69.753328/s
    checks_succeeded...: 100.00% 4200 out of 4200
    checks_failed......: 0.00%   0 out of 4200

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 4200   69.753328/s
    unexpected_error_rate..........: 0.00%  0 out of 4200

    HTTP
    http_req_duration..............: avg=109.77ms min=21.56ms  med=95.78ms  max=561.2ms  p(90)=193.51ms p(95)=225.47ms
      { expected_response:true }...: avg=109.77ms min=21.56ms  med=95.78ms  max=561.2ms  p(90)=193.51ms p(95)=225.47ms
    http_req_failed................: 0.00%  0 out of 4200
    http_reqs......................: 4200   69.753328/s

    EXECUTION
    iteration_duration.............: avg=210.58ms min=121.72ms med=196.46ms max=661.35ms p(90)=293.94ms p(95)=326.04ms
    iterations.....................: 4200   69.753328/s
    vus............................: 14     min=9         max=30 
    vus_max........................: 140    min=140       max=140

    NETWORK
    data_received..................: 4.1 MB 68 kB/s
    data_sent......................: 2.1 MB 36 kB/s




running (1m00.2s), 000/140 VUs, 4200 complete and 0 interrupted iterations
baseline_e2e_70rps ✓ [ 100% ] 000/140 VUs  1m0s  70.00 iters/s


## 11-baseline-80rps

Script: tests/k6/baseline-e2e-80rps.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-e2e-80rps.js
        output: -

     scenarios: (100.00%) 1 scenario, 480 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_e2e_80rps: 80.00 iterations/s for 1m0s (maxVUs: 160-480, gracefulStop: 30s)


running (0m01.0s), 012/160 VUs, 67 complete and 0 interrupted iterations
baseline_e2e_80rps   [   2% ] 012/160 VUs  0m01.0s/1m0s  80.00 iters/s

running (0m02.0s), 028/160 VUs, 131 complete and 0 interrupted iterations
baseline_e2e_80rps   [   3% ] 028/160 VUs  0m02.0s/1m0s  80.00 iters/s

running (0m03.0s), 028/160 VUs, 211 complete and 0 interrupted iterations
baseline_e2e_80rps   [   5% ] 028/160 VUs  0m03.0s/1m0s  80.00 iters/s

running (0m04.0s), 017/160 VUs, 302 complete and 0 interrupted iterations
baseline_e2e_80rps   [   7% ] 017/160 VUs  0m04.0s/1m0s  80.00 iters/s

running (0m05.0s), 017/160 VUs, 382 complete and 0 interrupted iterations
baseline_e2e_80rps   [   8% ] 017/160 VUs  0m05.0s/1m0s  80.00 iters/s

running (0m06.0s), 022/160 VUs, 457 complete and 0 interrupted iterations
baseline_e2e_80rps   [  10% ] 022/160 VUs  0m06.0s/1m0s  80.00 iters/s

running (0m07.0s), 014/160 VUs, 545 complete and 0 interrupted iterations
baseline_e2e_80rps   [  12% ] 014/160 VUs  0m07.0s/1m0s  80.00 iters/s

running (0m08.0s), 013/160 VUs, 626 complete and 0 interrupted iterations
baseline_e2e_80rps   [  13% ] 013/160 VUs  0m08.0s/1m0s  80.00 iters/s

running (0m09.0s), 015/160 VUs, 704 complete and 0 interrupted iterations
baseline_e2e_80rps   [  15% ] 015/160 VUs  0m09.0s/1m0s  80.00 iters/s

running (0m10.0s), 017/160 VUs, 783 complete and 0 interrupted iterations
baseline_e2e_80rps   [  17% ] 017/160 VUs  0m10.0s/1m0s  80.00 iters/s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<1500' p(95)=239.32ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 4800    79.74199/s
    checks_succeeded...: 100.00% 4800 out of 4800
    checks_failed......: 0.00%   0 out of 4800

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 4800   79.74199/s
    unexpected_error_rate..........: 0.00%  0 out of 4800

    HTTP
    http_req_duration..............: avg=121.91ms min=22.4ms   med=108.45ms max=519.77ms p(90)=211.61ms p(95)=239.32ms
      { expected_response:true }...: avg=121.91ms min=22.4ms   med=108.45ms max=519.77ms p(90)=211.61ms p(95)=239.32ms
    http_req_failed................: 0.00%  0 out of 4800
    http_reqs......................: 4800   79.74199/s

    EXECUTION
    iteration_duration.............: avg=222.7ms  min=123.24ms med=209.12ms max=620.88ms p(90)=312.4ms  p(95)=340.12ms
    iterations.....................: 4800   79.74199/s
    vus............................: 16     min=11        max=34 
    vus_max........................: 160    min=160       max=160

    NETWORK
    data_received..................: 4.7 MB 78 kB/s
    data_sent......................: 2.4 MB 41 kB/s




running (1m00.2s), 000/160 VUs, 4800 complete and 0 interrupted iterations
baseline_e2e_80rps ✓ [ 100% ] 000/160 VUs  1m0s  80.00 iters/s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 4800    79.74199/s
    checks_succeeded...: 100.00% 4800 out of 4800
    checks_failed......: 0.00%   0 out of 4800

    ✓ order request accepted or rejected

    CUSTOM
    accepted_orders................: 4800   79.74199/s
    unexpected_error_rate..........: 0.00%  0 out of 4800

    HTTP
    http_req_duration..............: avg=121.91ms min=22.4ms   med=108.45ms max=519.77ms p(90)=211.61ms p(95)=239.32ms
      { expected_response:true }...: avg=121.91ms min=22.4ms   med=108.45ms max=519.77ms p(90)=211.61ms p(95)=239.32ms
    http_req_failed................: 0.00%  0 out of 4800
    http_reqs......................: 4800   79.74199/s

    EXECUTION
    iteration_duration.............: avg=222.7ms  min=123.24ms med=209.12ms max=620.88ms p(90)=312.4ms  p(95)=340.12ms
    iterations.....................: 4800   79.74199/s
    vus............................: 16     min=11        max=34 
    vus_max........................: 160    min=160       max=160

    NETWORK
    data_received..................: 4.7 MB 78 kB/s
    data_sent......................: 2.4 MB 41 kB/s




running (1m00.2s), 000/160 VUs, 4800 complete and 0 interrupted iterations
baseline_e2e_80rps ✓ [ 100% ] 000/160 VUs  1m0s  80.00 iters/s


## 12-baseline-light

Script: tests/k6/baseline-light.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/baseline-light.js
        output: -

     scenarios: (100.00%) 1 scenario, 10 max VUs, 1m30s max duration (incl. graceful stop):
              * baseline_light: 10 looping VUs for 1m0s (gracefulStop: 30s)


running (0m01.0s), 10/10 VUs, 28 complete and 0 interrupted iterations
baseline_light   [   2% ] 10 VUs  0m01.0s/1m0s

running (0m02.0s), 10/10 VUs, 63 complete and 0 interrupted iterations
baseline_light   [   3% ] 10 VUs  0m02.0s/1m0s

running (0m03.0s), 10/10 VUs, 96 complete and 0 interrupted iterations
baseline_light   [   5% ] 10 VUs  0m03.0s/1m0s

running (0m04.0s), 10/10 VUs, 129 complete and 0 interrupted iterations
baseline_light   [   7% ] 10 VUs  0m04.0s/1m0s

running (0m05.0s), 10/10 VUs, 167 complete and 0 interrupted iterations
baseline_light   [   8% ] 10 VUs  0m05.0s/1m0s

running (0m06.0s), 10/10 VUs, 202 complete and 0 interrupted iterations
baseline_light   [  10% ] 10 VUs  0m06.0s/1m0s

running (0m07.0s), 10/10 VUs, 236 complete and 0 interrupted iterations
baseline_light   [  12% ] 10 VUs  0m07.0s/1m0s

running (0m08.0s), 10/10 VUs, 268 complete and 0 interrupted iterations
baseline_light   [  13% ] 10 VUs  0m08.0s/1m0s

running (0m09.0s), 10/10 VUs, 306 complete and 0 interrupted iterations
baseline_light   [  15% ] 10 VUs  0m09.0s/1m0s

running (0m10.0s), 10/10 VUs, 343 complete and 0 interrupted iterations
baseline_light   [  17% ] 10 VUs  0m10.0s/1m0s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<3000' p(95)=202.35ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 2044    33.944183/s
    checks_succeeded...: 100.00% 2044 out of 2044
    checks_failed......: 0.00%   0 out of 2044

    ✓ accepted or rejected

    CUSTOM
    accepted_orders................: 2044   33.944183/s
    unexpected_error_rate..........: 0.00%  0 out of 2044

    HTTP
    http_req_duration..............: avg=93.38ms  min=20.56ms  med=85.62ms  max=448.91ms p(90)=184.84ms p(95)=202.35ms
      { expected_response:true }...: avg=93.38ms  min=20.56ms  med=85.62ms  max=448.91ms p(90)=184.84ms p(95)=202.35ms
    http_req_failed................: 0.00%  0 out of 2044
    http_reqs......................: 2044   33.944183/s

    EXECUTION
    iteration_duration.............: avg=294.08ms min=220.75ms med=286.28ms max=649.4ms  p(90)=385.35ms p(95)=402.56ms
    iterations.....................: 2044   33.944183/s
    vus............................: 10     min=10        max=10
    vus_max........................: 10     min=10        max=10

    NETWORK
    data_received..................: 1.7 MB 29 kB/s
    data_sent......................: 800 kB 13 kB/s




running (1m00.2s), 00/10 VUs, 2044 complete and 0 interrupted iterations
baseline_light ✓ [ 100% ] 10 VUs  1m0s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 2044    33.944183/s
    checks_succeeded...: 100.00% 2044 out of 2044
    checks_failed......: 0.00%   0 out of 2044

    ✓ accepted or rejected

    CUSTOM
    accepted_orders................: 2044   33.944183/s
    unexpected_error_rate..........: 0.00%  0 out of 2044

    HTTP
    http_req_duration..............: avg=93.38ms  min=20.56ms  med=85.62ms  max=448.91ms p(90)=184.84ms p(95)=202.35ms
      { expected_response:true }...: avg=93.38ms  min=20.56ms  med=85.62ms  max=448.91ms p(90)=184.84ms p(95)=202.35ms
    http_req_failed................: 0.00%  0 out of 2044
    http_reqs......................: 2044   33.944183/s

    EXECUTION
    iteration_duration.............: avg=294.08ms min=220.75ms med=286.28ms max=649.4ms  p(90)=385.35ms p(95)=402.56ms
    iterations.....................: 2044   33.944183/s
    vus............................: 10     min=10        max=10
    vus_max........................: 10     min=10        max=10

    NETWORK
    data_received..................: 1.7 MB 29 kB/s
    data_sent......................: 800 kB 13 kB/s




running (1m00.2s), 00/10 VUs, 2044 complete and 0 interrupted iterations
baseline_light ✓ [ 100% ] 10 VUs  1m0s


## 13-load-test

Script: tests/k6/load-test.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/load-test.js
        output: -

     scenarios: (100.00%) 1 scenario, 50 max VUs, 14m30s max duration (incl. graceful stop):
              * default: Up to 50 looping VUs for 14m0s over 3 stages (gracefulRampDown: 30s, gracefulStop: 30s)


running (00m01.0s), 01/50 VUs, 0 complete and 0 interrupted iterations
default   [   0% ] 01/50 VUs  00m01.0s/14m00.0s

running (00m02.0s), 01/50 VUs, 1 complete and 0 interrupted iterations
default   [   0% ] 01/50 VUs  00m02.0s/14m00.0s

running (00m03.0s), 02/50 VUs, 2 complete and 0 interrupted iterations
default   [   0% ] 02/50 VUs  00m03.0s/14m00.0s

running (00m04.0s), 02/50 VUs, 4 complete and 0 interrupted iterations
default   [   0% ] 02/50 VUs  00m04.0s/14m00.0s

running (00m05.0s), 03/50 VUs, 6 complete and 0 interrupted iterations
default   [   1% ] 03/50 VUs  00m05.0s/14m00.0s

running (00m06.0s), 03/50 VUs, 9 complete and 0 interrupted iterations
default   [   1% ] 03/50 VUs  00m06.0s/14m00.0s

running (00m07.0s), 03/50 VUs, 12 complete and 0 interrupted iterations
default   [   1% ] 03/50 VUs  00m07.0s/14m00.0s

running (00m08.0s), 04/50 VUs, 15 complete and 0 interrupted iterations
default   [   1% ] 04/50 VUs  00m08.0s/14m00.0s

running (00m09.0s), 04/50 VUs, 18 complete and 0 interrupted iterations
default   [   1% ] 04/50 VUs  00m09.0s/14m00.0s

running (00m10.0s), 05/50 VUs, 22 complete and 0 interrupted iterations
default   [   1% ] 05/50 VUs  00m10.0s/14m00.0s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<500' p(95)=172.04ms

    http_req_failed
    ✓ 'rate<0.005' rate=0.00%

    order_success_rate
    ✓ 'rate>0.99' rate=100.00%


  █ TOTAL RESULTS 

    checks_total.......: 20842   24.811798/s
    checks_succeeded...: 100.00% 20842 out of 20842
    checks_failed......: 0.00%   0 out of 20842

    ✓ create order 2xx

    CUSTOM
    order_duration_ms..............: avg=66.872373 min=17      med=49      max=418      p(90)=124      p(95)=172     
    order_success_rate.............: 100.00% 20842 out of 20842
    total_orders_created...........: 20842   24.811798/s

    HTTP
    http_req_duration..............: avg=66.76ms   min=16.82ms med=48.67ms max=418.16ms p(90)=124.03ms p(95)=172.04ms
      { expected_response:true }...: avg=66.76ms   min=16.82ms med=48.67ms max=418.16ms p(90)=124.03ms p(95)=172.04ms
    http_req_failed................: 0.00%   0 out of 20842
    http_reqs......................: 20842   24.811798/s

    EXECUTION
    iteration_duration.............: avg=1.04s     min=1s      med=1.03s   max=1.41s    p(90)=1.1s     p(95)=1.13s   
    iterations.....................: 34672   41.276013/s
    vus............................: 1       min=1              max=50
    vus_max........................: 50      min=50             max=50

    NETWORK
    data_received..................: 17 MB   20 kB/s
    data_sent......................: 7.7 MB  9.2 kB/s




running (14m00.0s), 00/50 VUs, 34672 complete and 0 interrupted iterations
default ✓ [ 100% ] 00/50 VUs  14m0s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 20842   24.811798/s
    checks_succeeded...: 100.00% 20842 out of 20842
    checks_failed......: 0.00%   0 out of 20842

    ✓ create order 2xx

    CUSTOM
    order_duration_ms..............: avg=66.872373 min=17      med=49      max=418      p(90)=124      p(95)=172     
    order_success_rate.............: 100.00% 20842 out of 20842
    total_orders_created...........: 20842   24.811798/s

    HTTP
    http_req_duration..............: avg=66.76ms   min=16.82ms med=48.67ms max=418.16ms p(90)=124.03ms p(95)=172.04ms
      { expected_response:true }...: avg=66.76ms   min=16.82ms med=48.67ms max=418.16ms p(90)=124.03ms p(95)=172.04ms
    http_req_failed................: 0.00%   0 out of 20842
    http_reqs......................: 20842   24.811798/s

    EXECUTION
    iteration_duration.............: avg=1.04s     min=1s      med=1.03s   max=1.41s    p(90)=1.1s     p(95)=1.13s   
    iterations.....................: 34672   41.276013/s
    vus............................: 1       min=1              max=50
    vus_max........................: 50      min=50             max=50

    NETWORK
    data_received..................: 17 MB   20 kB/s
    data_sent......................: 7.7 MB  9.2 kB/s




running (14m00.0s), 00/50 VUs, 34672 complete and 0 interrupted iterations
default ✓ [ 100% ] 00/50 VUs  14m0s


## 14-flash-sale

Script: tests/k6/flash-sale-test.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/flash-sale-test.js
        output: -

     scenarios: (100.00%) 1 scenario, 200 max VUs, 1m20s max duration (incl. graceful stop):
              * flash_sale: Up to 200 looping VUs for 1m10s over 3 stages (gracefulRampDown: 10s, gracefulStop: 30s)


running (0m01.0s), 010/200 VUs, 18 complete and 0 interrupted iterations
flash_sale   [   1% ] 010/200 VUs  0m01.0s/1m10.0s

running (0m02.0s), 020/200 VUs, 69 complete and 0 interrupted iterations
flash_sale   [   3% ] 020/200 VUs  0m02.0s/1m10.0s

running (0m03.0s), 030/200 VUs, 141 complete and 0 interrupted iterations
flash_sale   [   4% ] 030/200 VUs  0m03.0s/1m10.0s

running (0m04.0s), 040/200 VUs, 222 complete and 0 interrupted iterations
flash_sale   [   6% ] 040/200 VUs  0m04.0s/1m10.0s

running (0m05.0s), 050/200 VUs, 321 complete and 0 interrupted iterations
flash_sale   [   7% ] 050/200 VUs  0m05.0s/1m10.0s

running (0m06.0s), 060/200 VUs, 470 complete and 0 interrupted iterations
flash_sale   [   9% ] 060/200 VUs  0m06.0s/1m10.0s

running (0m07.0s), 070/200 VUs, 600 complete and 0 interrupted iterations
flash_sale   [  10% ] 070/200 VUs  0m07.0s/1m10.0s

running (0m08.0s), 080/200 VUs, 763 complete and 0 interrupted iterations
flash_sale   [  11% ] 080/200 VUs  0m08.0s/1m10.0s

running (0m09.0s), 090/200 VUs, 932 complete and 0 interrupted iterations
flash_sale   [  13% ] 090/200 VUs  0m09.0s/1m10.0s

running (0m10.0s), 100/200 VUs, 1106 complete and 0 interrupted iterations
flash_sale   [  14% ] 100/200 VUs  0m10.0s/1m10.0s

### Threshold block

  █ THRESHOLDS 

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 12993   185.193257/s
    checks_succeeded...: 100.00% 12993 out of 12993
    checks_failed......: 0.00%   0 out of 12993

    ✓ accepted or rejected by flash sale gate

    CUSTOM
    accepted_orders................: 12993   185.193257/s
    status_2xx_rate................: 100.00% 12993 out of 12993
    status_409_rate................: 0.00%   0 out of 12993
    unexpected_error_rate..........: 0.00%   0 out of 12993

    HTTP
    http_req_duration..............: avg=651ms    min=22.57ms  med=660.98ms max=1.96s p(90)=1.01s p(95)=1.1s
      { expected_response:true }...: avg=651ms    min=22.57ms  med=660.98ms max=1.96s p(90)=1.01s p(95)=1.1s
    http_req_failed................: 0.00%   0 out of 12993
    http_reqs......................: 12993   185.193257/s

    EXECUTION
    iteration_duration.............: avg=851.71ms min=223.83ms med=861.75ms max=2.16s p(90)=1.21s p(95)=1.3s
    iterations.....................: 12993   185.193257/s
    vus............................: 4       min=4              max=200
    vus_max........................: 200     min=200            max=200

    NETWORK
    data_received..................: 11 MB   156 kB/s
    data_sent......................: 5.3 MB  76 kB/s




running (1m10.2s), 000/200 VUs, 12993 complete and 0 interrupted iterations
flash_sale ✓ [ 100% ] 000/200 VUs  1m10s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 12993   185.193257/s
    checks_succeeded...: 100.00% 12993 out of 12993
    checks_failed......: 0.00%   0 out of 12993

    ✓ accepted or rejected by flash sale gate

    CUSTOM
    accepted_orders................: 12993   185.193257/s
    status_2xx_rate................: 100.00% 12993 out of 12993
    status_409_rate................: 0.00%   0 out of 12993
    unexpected_error_rate..........: 0.00%   0 out of 12993

    HTTP
    http_req_duration..............: avg=651ms    min=22.57ms  med=660.98ms max=1.96s p(90)=1.01s p(95)=1.1s
      { expected_response:true }...: avg=651ms    min=22.57ms  med=660.98ms max=1.96s p(90)=1.01s p(95)=1.1s
    http_req_failed................: 0.00%   0 out of 12993
    http_reqs......................: 12993   185.193257/s

    EXECUTION
    iteration_duration.............: avg=851.71ms min=223.83ms med=861.75ms max=2.16s p(90)=1.21s p(95)=1.3s
    iterations.....................: 12993   185.193257/s
    vus............................: 4       min=4              max=200
    vus_max........................: 200     min=200            max=200

    NETWORK
    data_received..................: 11 MB   156 kB/s
    data_sent......................: 5.3 MB  76 kB/s




running (1m10.2s), 000/200 VUs, 12993 complete and 0 interrupted iterations
flash_sale ✓ [ 100% ] 000/200 VUs  1m10s


## 15-flash-sale-spike

Script: tests/k6/flash-sale-spike-test.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/flash-sale-spike-test.js
        output: -

     scenarios: (100.00%) 1 scenario, 500 max VUs, 1m0s max duration (incl. graceful stop):
              * flash_sale_spike: Up to 500 looping VUs for 50s over 3 stages (gracefulRampDown: 10s, gracefulStop: 30s)


running (0m01.0s), 049/500 VUs, 30 complete and 0 interrupted iterations
flash_sale_spike   [   2% ] 049/500 VUs  01.0s/50.0s

running (0m02.0s), 098/500 VUs, 146 complete and 0 interrupted iterations
flash_sale_spike   [   4% ] 098/500 VUs  02.0s/50.0s

running (0m03.0s), 148/500 VUs, 280 complete and 0 interrupted iterations
flash_sale_spike   [   6% ] 148/500 VUs  03.0s/50.0s

running (0m04.0s), 198/500 VUs, 429 complete and 0 interrupted iterations
flash_sale_spike   [   8% ] 198/500 VUs  04.0s/50.0s

running (0m05.0s), 248/500 VUs, 604 complete and 0 interrupted iterations
flash_sale_spike   [  10% ] 248/500 VUs  05.0s/50.0s

running (0m06.0s), 298/500 VUs, 737 complete and 0 interrupted iterations
flash_sale_spike   [  12% ] 298/500 VUs  06.0s/50.0s

running (0m07.0s), 348/500 VUs, 921 complete and 0 interrupted iterations
flash_sale_spike   [  14% ] 348/500 VUs  07.0s/50.0s

running (0m08.0s), 398/500 VUs, 1118 complete and 0 interrupted iterations
flash_sale_spike   [  16% ] 398/500 VUs  08.0s/50.0s

running (0m09.0s), 448/500 VUs, 1275 complete and 0 interrupted iterations
flash_sale_spike   [  18% ] 448/500 VUs  09.0s/50.0s

running (0m10.0s), 498/500 VUs, 1460 complete and 0 interrupted iterations
flash_sale_spike   [  20% ] 498/500 VUs  10.0s/50.0s

### Threshold block

  █ THRESHOLDS 

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%

    unexpected_error_rate
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 9116    181.556211/s
    checks_succeeded...: 100.00% 9116 out of 9116
    checks_failed......: 0.00%   0 out of 9116

    ✓ accepted or rejected by flash sale gate

    CUSTOM
    accepted_orders................: 9116    181.556211/s
    status_2xx_rate................: 100.00% 9116 out of 9116
    status_409_rate................: 0.00%   0 out of 9116
    unexpected_error_rate..........: 0.00%   0 out of 9116

    HTTP
    http_req_duration..............: avg=2.13s min=167.03ms med=2.33s max=4.23s p(90)=2.85s p(95)=2.99s
      { expected_response:true }...: avg=2.13s min=167.03ms med=2.33s max=4.23s p(90)=2.85s p(95)=2.99s
    http_req_failed................: 0.00%   0 out of 9116
    http_reqs......................: 9116    181.556211/s

    EXECUTION
    iteration_duration.............: avg=2.23s min=272.8ms  med=2.43s max=4.33s p(90)=2.95s p(95)=3.09s
    iterations.....................: 9116    181.556211/s
    vus............................: 9       min=9            max=500
    vus_max........................: 500     min=500          max=500

    NETWORK
    data_received..................: 7.9 MB  157 kB/s
    data_sent......................: 4.0 MB  79 kB/s




running (0m50.2s), 000/500 VUs, 9116 complete and 0 interrupted iterations
flash_sale_spike ✓ [ 100% ] 000/500 VUs  50s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 9116    181.556211/s
    checks_succeeded...: 100.00% 9116 out of 9116
    checks_failed......: 0.00%   0 out of 9116

    ✓ accepted or rejected by flash sale gate

    CUSTOM
    accepted_orders................: 9116    181.556211/s
    status_2xx_rate................: 100.00% 9116 out of 9116
    status_409_rate................: 0.00%   0 out of 9116
    unexpected_error_rate..........: 0.00%   0 out of 9116

    HTTP
    http_req_duration..............: avg=2.13s min=167.03ms med=2.33s max=4.23s p(90)=2.85s p(95)=2.99s
      { expected_response:true }...: avg=2.13s min=167.03ms med=2.33s max=4.23s p(90)=2.85s p(95)=2.99s
    http_req_failed................: 0.00%   0 out of 9116
    http_reqs......................: 9116    181.556211/s

    EXECUTION
    iteration_duration.............: avg=2.23s min=272.8ms  med=2.43s max=4.33s p(90)=2.95s p(95)=3.09s
    iterations.....................: 9116    181.556211/s
    vus............................: 9       min=9            max=500
    vus_max........................: 500     min=500          max=500

    NETWORK
    data_received..................: 7.9 MB  157 kB/s
    data_sent......................: 4.0 MB  79 kB/s




running (0m50.2s), 000/500 VUs, 9116 complete and 0 interrupted iterations
flash_sale_spike ✓ [ 100% ] 000/500 VUs  50s


## 16-spike-test

Script: tests/k6/spike-test.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/spike-test.js
        output: -

     scenarios: (100.00%) 1 scenario, 1000 max VUs, 2m30s max duration (incl. graceful stop):
              * default: Up to 1000 looping VUs for 2m0s over 4 stages (gracefulRampDown: 30s, gracefulStop: 30s)


running (0m00.9s), 0005/1000 VUs, 24 complete and 0 interrupted iterations
default   [   1% ] 0005/1000 VUs  0m00.9s/2m00.0s

running (0m01.9s), 0010/1000 VUs, 92 complete and 0 interrupted iterations
default   [   2% ] 0010/1000 VUs  0m01.9s/2m00.0s

running (0m02.9s), 0015/1000 VUs, 183 complete and 0 interrupted iterations
default   [   2% ] 0015/1000 VUs  0m02.9s/2m00.0s

running (0m03.9s), 0020/1000 VUs, 263 complete and 0 interrupted iterations
default   [   3% ] 0020/1000 VUs  0m03.9s/2m00.0s

running (0m04.9s), 0025/1000 VUs, 394 complete and 0 interrupted iterations
default   [   4% ] 0025/1000 VUs  0m04.9s/2m00.0s

running (0m05.9s), 0030/1000 VUs, 536 complete and 0 interrupted iterations
default   [   5% ] 0030/1000 VUs  0m05.9s/2m00.0s

running (0m06.9s), 0035/1000 VUs, 651 complete and 0 interrupted iterations
default   [   6% ] 0035/1000 VUs  0m06.9s/2m00.0s

running (0m07.9s), 0039/1000 VUs, 791 complete and 0 interrupted iterations
default   [   7% ] 0039/1000 VUs  0m07.9s/2m00.0s

running (0m08.9s), 0044/1000 VUs, 949 complete and 0 interrupted iterations
default   [   7% ] 0044/1000 VUs  0m08.9s/2m00.0s

running (0m09.9s), 0049/1000 VUs, 1119 complete and 0 interrupted iterations
default   [   8% ] 0049/1000 VUs  0m09.9s/2m00.0s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 25548   212.497582/s
    checks_succeeded...: 100.00% 25548 out of 25548
    checks_failed......: 0.00%   0 out of 25548

    ✓ status is 2xx

    HTTP
    http_req_duration..............: avg=3.4s min=34.16ms med=3.77s max=7.49s p(90)=5.01s p(95)=5.32s
      { expected_response:true }...: avg=3.4s min=34.16ms med=3.77s max=7.49s p(90)=5.01s p(95)=5.32s
    http_req_failed................: 0.00%  0 out of 25548
    http_reqs......................: 25548  212.497582/s

    EXECUTION
    iteration_duration.............: avg=3.4s min=34.25ms med=3.77s max=7.5s  p(90)=5.01s p(95)=5.32s
    iterations.....................: 25548  212.497582/s
    vus............................: 8      min=5          max=1000
    vus_max........................: 1000   min=1000       max=1000

    NETWORK
    data_received..................: 21 MB  178 kB/s
    data_sent......................: 8.9 MB 74 kB/s




running (2m00.2s), 0000/1000 VUs, 25548 complete and 0 interrupted iterations
default ✓ [ 100% ] 0000/1000 VUs  2m0s


## 17-stress-test-multi

Script: tests/k6/stress-test-multi.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/stress-test-multi.js
        output: -

     scenarios: (100.00%) 1 scenario, 1000 max VUs, 9m30s max duration (incl. graceful stop):
              * default: Up to 1000 looping VUs for 9m0s over 4 stages (gracefulRampDown: 30s, gracefulStop: 30s)


running (0m00.9s), 0002/1000 VUs, 0 complete and 0 interrupted iterations
default   [   0% ] 0002/1000 VUs  0m00.9s/9m00.0s

running (0m01.9s), 0004/1000 VUs, 2 complete and 0 interrupted iterations
default   [   0% ] 0004/1000 VUs  0m01.9s/9m00.0s

running (0m02.9s), 0005/1000 VUs, 6 complete and 0 interrupted iterations
default   [   1% ] 0005/1000 VUs  0m02.9s/9m00.0s

running (0m03.9s), 0007/1000 VUs, 10 complete and 0 interrupted iterations
default   [   1% ] 0007/1000 VUs  0m03.9s/9m00.0s

running (0m04.9s), 0009/1000 VUs, 15 complete and 0 interrupted iterations
default   [   1% ] 0009/1000 VUs  0m04.9s/9m00.0s

running (0m05.9s), 0010/1000 VUs, 20 complete and 0 interrupted iterations
default   [   1% ] 0010/1000 VUs  0m05.9s/9m00.0s

running (0m06.9s), 0012/1000 VUs, 30 complete and 0 interrupted iterations
default   [   1% ] 0012/1000 VUs  0m06.9s/9m00.0s

running (0m07.9s), 0014/1000 VUs, 38 complete and 0 interrupted iterations
default   [   1% ] 0014/1000 VUs  0m07.9s/9m00.0s

running (0m08.9s), 0015/1000 VUs, 50 complete and 0 interrupted iterations
default   [   2% ] 0015/1000 VUs  0m08.9s/9m00.0s

running (0m09.9s), 0017/1000 VUs, 65 complete and 0 interrupted iterations
default   [   2% ] 0017/1000 VUs  0m09.9s/9m00.0s

### Threshold block

  █ THRESHOLDS 

    error_rate
    ✓ 'rate<0.1' rate=0.00%

    http_req_duration
    ✓ 'p(95)<5000' p(95)=4.75s

    http_req_failed
    ✓ 'rate<0.1' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 82831   153.147383/s
    checks_succeeded...: 100.00% 82831 out of 82831
    checks_failed......: 0.00%   0 out of 82831

    ✓ status is 2xx

    CUSTOM
    error_rate.....................: 0.00% 0 out of 82831
    order_duration_ms..............: avg=2024.144825 min=37     med=1749  max=8272  p(90)=3999  p(95)=4760 
    success_orders.................: 82831 153.147383/s

    HTTP
    http_req_duration..............: avg=2.02s       min=32.1ms med=1.74s max=8.26s p(90)=3.99s p(95)=4.75s
      { expected_response:true }...: avg=2.02s       min=32.1ms med=1.74s max=8.26s p(90)=3.99s p(95)=4.75s
    http_req_failed................: 0.00% 0 out of 82831
    http_reqs......................: 82831 153.147383/s

    EXECUTION
    iteration_duration.............: avg=3.02s       min=1.03s  med=2.74s max=9.27s p(90)=4.99s p(95)=5.76s
    iterations.....................: 82831 153.147383/s
    vus............................: 6     min=2          max=1000
    vus_max........................: 1000  min=1000       max=1000

    NETWORK
    data_received..................: 75 MB 139 kB/s
    data_sent......................: 38 MB 70 kB/s




running (9m00.9s), 0000/1000 VUs, 82831 complete and 0 interrupted iterations
default ✓ [ 100% ] 0000/1000 VUs  9m0s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 82831   153.147383/s
    checks_succeeded...: 100.00% 82831 out of 82831
    checks_failed......: 0.00%   0 out of 82831

    ✓ status is 2xx

    CUSTOM
    error_rate.....................: 0.00% 0 out of 82831
    order_duration_ms..............: avg=2024.144825 min=37     med=1749  max=8272  p(90)=3999  p(95)=4760 
    success_orders.................: 82831 153.147383/s

    HTTP
    http_req_duration..............: avg=2.02s       min=32.1ms med=1.74s max=8.26s p(90)=3.99s p(95)=4.75s
      { expected_response:true }...: avg=2.02s       min=32.1ms med=1.74s max=8.26s p(90)=3.99s p(95)=4.75s
    http_req_failed................: 0.00% 0 out of 82831
    http_reqs......................: 82831 153.147383/s

    EXECUTION
    iteration_duration.............: avg=3.02s       min=1.03s  med=2.74s max=9.27s p(90)=4.99s p(95)=5.76s
    iterations.....................: 82831 153.147383/s
    vus............................: 6     min=2          max=1000
    vus_max........................: 1000  min=1000       max=1000

    NETWORK
    data_received..................: 75 MB 139 kB/s
    data_sent......................: 38 MB 70 kB/s




running (9m00.9s), 0000/1000 VUs, 82831 complete and 0 interrupted iterations
default ✓ [ 100% ] 0000/1000 VUs  9m0s


## 18-stress-test

Script: tests/k6/stress-test.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/stress-test.js
        output: -

     scenarios: (100.00%) 1 scenario, 1000 max VUs, 8m30s max duration (incl. graceful stop):
              * default: Up to 1000 looping VUs for 8m0s over 5 stages (gracefulRampDown: 30s, gracefulStop: 30s)


running (0m00.9s), 0002/1000 VUs, 0 complete and 0 interrupted iterations
default   [   0% ] 0002/1000 VUs  0m00.9s/8m00.0s

running (0m01.9s), 0004/1000 VUs, 2 complete and 0 interrupted iterations
default   [   0% ] 0004/1000 VUs  0m01.9s/8m00.0s

running (0m02.9s), 0005/1000 VUs, 4 complete and 0 interrupted iterations
default   [   1% ] 0005/1000 VUs  0m02.9s/8m00.0s

running (0m03.9s), 0007/1000 VUs, 9 complete and 0 interrupted iterations
default   [   1% ] 0007/1000 VUs  0m03.9s/8m00.0s

running (0m04.9s), 0009/1000 VUs, 13 complete and 0 interrupted iterations
default   [   1% ] 0009/1000 VUs  0m04.9s/8m00.0s

running (0m05.9s), 0010/1000 VUs, 20 complete and 0 interrupted iterations
default   [   1% ] 0010/1000 VUs  0m05.9s/8m00.0s

running (0m06.9s), 0012/1000 VUs, 30 complete and 0 interrupted iterations
default   [   1% ] 0012/1000 VUs  0m06.9s/8m00.0s

running (0m07.9s), 0014/1000 VUs, 38 complete and 0 interrupted iterations
default   [   2% ] 0014/1000 VUs  0m07.9s/8m00.0s

running (0m08.9s), 0015/1000 VUs, 49 complete and 0 interrupted iterations
default   [   2% ] 0015/1000 VUs  0m08.9s/8m00.0s

running (0m09.9s), 0017/1000 VUs, 62 complete and 0 interrupted iterations
default   [   2% ] 0017/1000 VUs  0m09.9s/8m00.0s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 75706   157.486645/s
    checks_succeeded...: 100.00% 75706 out of 75706
    checks_failed......: 0.00%   0 out of 75706

    ✓ status is 2xx

    CUSTOM
    error_rate.....................: 0.00% 0 out of 75706
    order_duration_ms..............: avg=2227.211159 min=59     med=1954.5 max=6432  p(90)=4311  p(95)=4653.75
    success_orders.................: 75706 157.486645/s

    HTTP
    http_req_duration..............: avg=2.22s       min=58.5ms med=1.95s  max=6.43s p(90)=4.31s p(95)=4.65s  
      { expected_response:true }...: avg=2.22s       min=58.5ms med=1.95s  max=6.43s p(90)=4.31s p(95)=4.65s  
    http_req_failed................: 0.00% 0 out of 75706
    http_reqs......................: 75706 157.486645/s

    EXECUTION
    iteration_duration.............: avg=3.22s       min=1.05s  med=2.95s  max=7.43s p(90)=5.31s p(95)=5.65s  
    iterations.....................: 75706 157.486645/s
    vus............................: 15    min=2          max=999 
    vus_max........................: 1000  min=1000       max=1000

    NETWORK
    data_received..................: 64 MB 132 kB/s
    data_sent......................: 31 MB 65 kB/s




running (8m00.7s), 0000/1000 VUs, 75706 complete and 0 interrupted iterations
default ✓ [ 100% ] 0000/1000 VUs  8m0s


## 19-soak-test

Script: tests/k6/soak-test.js

### Scenario header


         /\      Grafana   /‾‾/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ‾‾\ 
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: tests/k6/soak-test.js
        output: -

     scenarios: (100.00%) 1 scenario, 30 max VUs, 1h0m30s max duration (incl. graceful stop):
              * default: Up to 30 looping VUs for 1h0m0s over 3 stages (gracefulRampDown: 30s, gracefulStop: 30s)


running (0h00m01.0s), 01/30 VUs, 0 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m01.0s/1h00m00.0s

running (0h00m02.0s), 01/30 VUs, 1 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m02.0s/1h00m00.0s

running (0h00m03.0s), 01/30 VUs, 2 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m03.0s/1h00m00.0s

running (0h00m04.0s), 01/30 VUs, 3 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m04.0s/1h00m00.0s

running (0h00m05.0s), 01/30 VUs, 4 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m05.0s/1h00m00.0s

running (0h00m06.0s), 01/30 VUs, 5 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m06.0s/1h00m00.0s

running (0h00m07.0s), 01/30 VUs, 6 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m07.0s/1h00m00.0s

running (0h00m08.0s), 01/30 VUs, 7 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m08.0s/1h00m00.0s

running (0h00m09.0s), 01/30 VUs, 8 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m09.0s/1h00m00.0s

running (0h00m10.0s), 01/30 VUs, 9 complete and 0 interrupted iterations
default   [   0% ] 01/30 VUs  0h00m10.0s/1h00m00.0s

### Threshold block

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<500' p(95)=293.24ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 94555  26.263933/s
    checks_succeeded...: 99.99% 94553 out of 94555
    checks_failed......: 0.00%  2 out of 94555

    ✓ list 200
    ✗ create 2xx
      ↳  99% — ✓ 28359 / ✗ 2
    ✓ health 200

    CUSTOM
    order_duration_ms..............: avg=229.509679 min=34       med=214     max=2641  p(90)=311      p(95)=348     
    total_orders...................: 28359  7.877097/s

    HTTP
    http_req_duration..............: avg=97.7ms     min=6.65ms   med=43.32ms max=2.64s p(90)=253.82ms p(95)=293.24ms
      { expected_response:true }...: avg=97.7ms     min=6.65ms   med=43.32ms max=2.64s p(90)=253.8ms  p(95)=293.23ms
    http_req_failed................: 0.00%  2 out of 94555
    http_reqs......................: 94555  26.263933/s

    EXECUTION
    iteration_duration.............: avg=1.04s      min=506.81ms med=1.04s   max=3.64s p(90)=1.25s    p(95)=1.29s   
    iterations.....................: 94555  26.263933/s
    vus............................: 1      min=1          max=30
    vus_max........................: 30     min=30         max=30

    NETWORK
    data_received..................: 126 MB 35 kB/s
    data_sent......................: 22 MB  6.2 kB/s




running (1h00m00.2s), 00/30 VUs, 94555 complete and 0 interrupted iterations
default ✓ [ 100% ] 00/30 VUs  1h0m0s

### Total result block

  █ TOTAL RESULTS 

    checks_total.......: 94555  26.263933/s
    checks_succeeded...: 99.99% 94553 out of 94555
    checks_failed......: 0.00%  2 out of 94555

    ✓ list 200
    ✗ create 2xx
      ↳  99% — ✓ 28359 / ✗ 2
    ✓ health 200

    CUSTOM
    order_duration_ms..............: avg=229.509679 min=34       med=214     max=2641  p(90)=311      p(95)=348     
    total_orders...................: 28359  7.877097/s

    HTTP
    http_req_duration..............: avg=97.7ms     min=6.65ms   med=43.32ms max=2.64s p(90)=253.82ms p(95)=293.24ms
      { expected_response:true }...: avg=97.7ms     min=6.65ms   med=43.32ms max=2.64s p(90)=253.8ms  p(95)=293.23ms
    http_req_failed................: 0.00%  2 out of 94555
    http_reqs......................: 94555  26.263933/s

    EXECUTION
    iteration_duration.............: avg=1.04s      min=506.81ms med=1.04s   max=3.64s p(90)=1.25s    p(95)=1.29s   
    iterations.....................: 94555  26.263933/s
    vus............................: 1      min=1          max=30
    vus_max........................: 30     min=30         max=30

    NETWORK
    data_received..................: 126 MB 35 kB/s
    data_sent......................: 22 MB  6.2 kB/s




running (1h00m00.2s), 00/30 VUs, 94555 complete and 0 interrupted iterations
default ✓ [ 100% ] 00/30 VUs  1h0m0s

