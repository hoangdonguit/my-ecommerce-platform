import http from "k6/http";
import { check } from "k6";
import { Counter, Rate } from "k6/metrics";

http.setResponseCallback(http.expectedStatuses(200, 201, 409));

export const accepted_orders = new Counter("accepted_orders");
export const rejected_orders = new Counter("rejected_orders");
export const unexpected_error_rate = new Rate("unexpected_error_rate");
export const status_2xx_rate = new Rate("status_2xx_rate");
export const status_409_rate = new Rate("status_409_rate");

const RPS_1 = Number(__ENV.RPS_1 || 220);
const RPS_2 = Number(__ENV.RPS_2 || 320);
const RPS_3 = Number(__ENV.RPS_3 || 450);
const RPS_4 = Number(__ENV.RPS_4 || 600);

const PRE_ALLOCATED_VUS = Number(__ENV.PRE_ALLOCATED_VUS || 500);
const MAX_VUS = Number(__ENV.MAX_VUS || 1200);

export const options = {
  scenarios: {
    stress_hot_product: {
      executor: "ramping-arrival-rate",
      timeUnit: "1s",
      preAllocatedVUs: PRE_ALLOCATED_VUS,
      maxVUs: MAX_VUS,
      stages: [
        { duration: "30s", target: RPS_1 },
        { duration: "45s", target: RPS_2 },
        { duration: "45s", target: RPS_3 },
        { duration: "45s", target: RPS_4 },
        { duration: "15s", target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.50"],
    unexpected_error_rate: ["rate<0.50"],
  },
};

const BASE_URL = __ENV.GATEWAY_URL || __ENV.BASE_URL || "http://100.65.255.2:30517";
const API_KEY = __ENV.API_KEY || "";
const PRODUCT_ID = __ENV.PRODUCT_ID || "prod-123";

export default function () {
  const idemKey = `stress-hot-${Date.now()}-${__VU}-${__ITER}-${PRODUCT_ID}`;

  const body = JSON.stringify({
    user_id: `stress-hot-user-${__VU}-${__ITER}`,
    items: [{ product_id: PRODUCT_ID, quantity: 1 }],
    currency: "VND",
    payment_method: "COD",
    shipping_address: "UIT Stress Test Lab",
    note: "stress test hot product",
  });

  const res = http.post(`${BASE_URL}/api/orders`, body, {
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": API_KEY,
      "X-Gateway-API-Key": API_KEY,
      "X-Idempotency-Key": idemKey,
    },
    timeout: "20s",
  });

  const accepted = res.status === 200 || res.status === 201;
  const rejected = res.status === 409;

  if (accepted) accepted_orders.add(1);
  if (rejected) rejected_orders.add(1);

  status_2xx_rate.add(accepted);
  status_409_rate.add(rejected);
  unexpected_error_rate.add(!(accepted || rejected));

  if (!(accepted || rejected)) {
    console.error(`unexpected_status=${res.status} product=${PRODUCT_ID} body=${res.body}`);
  }

  check(res, {
    "accepted or rejected": () => accepted || rejected,
  });
}
