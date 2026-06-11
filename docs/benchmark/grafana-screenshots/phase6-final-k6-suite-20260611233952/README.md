# Grafana Screenshot Note - phase6-final-k6-suite-20260611233952

No Grafana screenshots are committed as primary benchmark evidence for this run.

Reason:

- The final k6 suite was executed as a continuous sequence of many tests.
- Functional, baseline, flash-sale, spike, stress, and soak scenarios ran close to each other.
- Grafana time-series panels therefore overlap multiple scenarios in the same time window.
- A static screenshot would be difficult to map accurately to one specific test phase.

Primary evidence for this benchmark is stored in:

- ../k6-final-artifacts/phase6-final-k6-suite-20260611233952/detailed-k6-metrics.md
- ../k6-final-artifacts/phase6-final-k6-suite-20260611233952/detailed-k6-metrics.tsv
- ../k6-final-artifacts/phase6-final-k6-suite-20260611233952/k6-output-scenario-and-threshold-extracts.md
- ../k6-final-artifacts/phase6-final-k6-suite-20260611233952/k6-result-interpretation.md
- ../k6-final-artifacts/phase6-final-k6-suite-20260611233952/postcheck/

Grafana was still useful for live observation during the run, but it is not used as the main report evidence for this final suite.

For a future benchmark with Grafana screenshots, each scenario should be run in a separate time window with an explicit start/end timestamp marker.
