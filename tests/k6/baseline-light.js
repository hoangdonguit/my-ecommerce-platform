import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate } from "k6/metrics";

http.setResponseCallback(http.expectedStatuses(200, 201, 409));

export const accepted_orders = new Counter("accepted_orders");
export const rejected_orders = new Counter("rejected_orders");
export const unexpected_error_rate = new Rate("unexpected_error_rate");

export const options = {
  scenarios: {
    baseline_light: {
      executor: "constant-vus",
      vus: 10,
      duration: "60s",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    unexpected_error_rate: ["rate<0.01"],
    http_req_duration: ["p(95)<3000"],
  },
};

const BASE_URL = __ENV.GATEWAY_URL || "http://100.65.255.2:30517";
const API_KEY = __ENV.API_KEY || "";
const PRODUCT_ID = __ENV.PRODUCT_ID || "prod-123";

function buildBody() {
  return JSON.stringify({
    user_id: `baseline-user-${__VU}-${__ITER}`,
    items: [
      {
        product_id: PRODUCT_ID,
        quantity: 1,
      },
    ],
    currency: "VND",
    payment_method: "COD",
    shipping_address: "UIT Baseline Light Test",
    note: "baseline light k6 test",
  });
}

export default function () {
  const idemKey = `baseline-light-${Date.now()}-${__VU}-${__ITER}`;

  const res = http.post(`${BASE_URL}/api/orders`, buildBody(), {
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
    "accepted or rejected": () => accepted || rejected,
  });

  sleep(0.2);
}
