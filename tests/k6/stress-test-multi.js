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
    stress_multi_products: {
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
const PRODUCT_IDS = (__ENV.PRODUCT_IDS || "prod-123,prod-456,prod-789")
  .split(",")
  .map((x) => x.trim())
  .filter((x) => x.length > 0);

function pickProduct() {
  return PRODUCT_IDS[(__VU + __ITER) % PRODUCT_IDS.length];
}

export default function () {
  const productID = pickProduct();
  const idemKey = `stress-multi-${Date.now()}-${__VU}-${__ITER}-${productID}`;

  const body = JSON.stringify({
    user_id: `stress-multi-user-${__VU}-${__ITER}`,
    items: [{ product_id: productID, quantity: 1 }],
    currency: "VND",
    payment_method: "COD",
    shipping_address: "UIT Stress Test Lab",
    note: "stress test multi products",
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
    console.error(`unexpected_status=${res.status} product=${productID} body=${res.body}`);
  }

  check(res, {
    "accepted or rejected": () => accepted || rejected,
  });
}
