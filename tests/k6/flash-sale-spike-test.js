import http from "k6/http";

http.setResponseCallback(http.expectedStatuses(200, 201, 409));

import { check, sleep } from "k6";
import { Counter, Rate } from "k6/metrics";

export const accepted_orders = new Counter("accepted_orders");
export const rejected_orders = new Counter("rejected_orders");
export const unexpected_error_rate = new Rate("unexpected_error_rate");
export const status_2xx_rate = new Rate("status_2xx_rate");
export const status_409_rate = new Rate("status_409_rate");

export const options = {
  scenarios: {
    flash_sale_spike: {
      executor: "ramping-vus",
      stages: [
        { duration: "10s", target: 500 },
        { duration: "30s", target: 500 },
        { duration: "10s", target: 0 },
      ],
      gracefulRampDown: "10s",
    },
  },
  thresholds: {
    unexpected_error_rate: ["rate<0.01"],
    http_req_failed: ["rate<0.01"],
  },
};

const BASE_URL = __ENV.GATEWAY_URL || "http://localhost:8090";
const API_KEY = __ENV.API_KEY || "";
const PRODUCT_ID = __ENV.PRODUCT_ID || "prod-123";

function buildBody() {
  return JSON.stringify({
    user_id: `flash-spike-user-${__VU}-${__ITER}`,
    items: [
      {
        product_id: PRODUCT_ID,
        quantity: 1,
      },
    ],
    currency: "VND",
    payment_method: "COD",
    shipping_address: "UIT Flash Sale Spike Lab",
    note: "flash sale spike k6 test",
  });
}

export default function () {
  const idemKey = `flash-sale-spike-${Date.now()}-${__VU}-${__ITER}`;

  const res = http.post(`${BASE_URL}/api/orders`, buildBody(), {
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": API_KEY,
      "X-Idempotency-Key": idemKey,
    },
    timeout: "20s",
  });

  const accepted = res.status === 201 || res.status === 200;
  const rejected = res.status === 409;

  if (accepted) accepted_orders.add(1);
  if (rejected) rejected_orders.add(1);

  status_2xx_rate.add(accepted);
  status_409_rate.add(rejected);
  unexpected_error_rate.add(!(accepted || rejected));

  if (!(accepted || rejected)) {
    console.error(`unexpected_status=${res.status} body=${res.body}`);
  }

  check(res, {
    "accepted or rejected by flash sale gate": () => accepted || rejected,
  });

  sleep(0.1);
}
