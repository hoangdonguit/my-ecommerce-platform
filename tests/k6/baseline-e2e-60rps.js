import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate } from "k6/metrics";

http.setResponseCallback(http.expectedStatuses(200, 201, 409));

export const accepted_orders = new Counter("accepted_orders");
export const rejected_orders = new Counter("rejected_orders");
export const unexpected_error_rate = new Rate("unexpected_error_rate");

export const options = {
  scenarios: {
    baseline_e2e_60rps: {
      executor: "constant-arrival-rate",
      rate: 60,
      timeUnit: "1s",
      duration: "60s",
      preAllocatedVUs: 120,
      maxVUs: 360,
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    unexpected_error_rate: ["rate<0.01"],
    http_req_duration: ["p(95)<1500"],
  },
};

const BASE_URL = __ENV.GATEWAY_URL || "http://100.65.255.2:30517";
const PRODUCT_ID = __ENV.PRODUCT_ID || "prod-123";
const RUN_ID = __ENV.RUN_ID || `baseline-e2e-60rps-${Date.now()}`;
const API_KEY = __ENV.API_KEY || "";

export default function () {
  const idemKey = `${RUN_ID}-${__VU}-${__ITER}-${Date.now()}`;

  const payload = JSON.stringify({
    user_id: `${RUN_ID}-user-${__VU}-${__ITER}`,
    items: [
      {
        product_id: PRODUCT_ID,
        quantity: 1,
      },
    ],
    currency: "VND",
    payment_method: "COD",
    shipping_address: "UIT Baseline E2E 60RPS",
    note: RUN_ID,
  });

  const res = http.post(`${BASE_URL}/api/orders`, payload, {
    headers: {
      "Content-Type": "application/json",
      "X-Idempotency-Key": idemKey,
      ...(API_KEY ? { "X-API-Key": API_KEY, "X-Gateway-API-Key": API_KEY } : {}),
    },
    timeout: "20s",
  });

  const accepted = res.status === 201 || res.status === 200;
  const rejected = res.status === 409;

  if (accepted) accepted_orders.add(1);
  if (rejected) rejected_orders.add(1);

  unexpected_error_rate.add(!(accepted || rejected));

  check(res, {
    "order request accepted or rejected": () => accepted || rejected,
  });

  sleep(0.1);
}
