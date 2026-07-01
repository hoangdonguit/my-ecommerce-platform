import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';

const orderDuration = new Trend('order_duration_ms');
const listDuration = new Trend('list_orders_duration_ms');
const healthDuration = new Trend('health_duration_ms');

const totalOrders = new Counter('total_orders');
const totalListRequests = new Counter('total_list_requests');
const totalHealthRequests = new Counter('total_health_requests');

const TARGET_VUS = Number(__ENV.TARGET_VUS || 30);
const RAMP_UP = __ENV.RAMP_UP || '5m';
const HOLD = __ENV.HOLD || '50m';
const RAMP_DOWN = __ENV.RAMP_DOWN || '5m';

export const options = {
  stages: [
    { duration: RAMP_UP, target: TARGET_VUS },
    { duration: HOLD, target: TARGET_VUS },
    { duration: RAMP_DOWN, target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<1500'],
    checks: ['rate>0.99'],
    order_duration_ms: ['p(95)<1500'],
  },
};

const BASE_URL = __ENV.GATEWAY_URL || 'http://100.65.255.2:30517';
const RUN_ID = __ENV.RUN_ID || `soak-${Date.now()}`;
const PRODUCT_ID = __ENV.PRODUCT_ID || 'prod-123';

const HEADERS = {
  'Content-Type': 'application/json',
  'X-API-Key': (__ENV.API_KEY || ''),
  'X-Gateway-API-Key': (__ENV.API_KEY || ''),
};

export default function () {
  const userId = `user-${RUN_ID}-${__VU}`;
  const rand = Math.random();

  if (rand < 0.30) {
    const start = Date.now();

    const res = http.post(`${BASE_URL}/api/orders`, JSON.stringify({
      user_id: userId,
      customer_id: userId,
      items: [
        { product_id: PRODUCT_ID, quantity: 1 },
      ],
      currency: 'VND',
      payment_method: 'COD',
      shipping_address: 'Soak Test',
    }), {
      headers: {
        ...HEADERS,
        'X-Idempotency-Key': `${RUN_ID}-${__VU}-${__ITER}-${Date.now()}`,
      },
    });

    orderDuration.add(Date.now() - start);

    if (check(res, {
      'create order returns 2xx': (r) => r.status >= 200 && r.status < 300,
      'create order not 5xx': (r) => r.status < 500,
    })) {
      totalOrders.add(1);
    }

    sleep(1);
  } else if (rand < 0.90) {
    const start = Date.now();

    const res = http.get(`${BASE_URL}/api/orders?user_id=${userId}&page=1&limit=5`, {
      headers: HEADERS,
    });

    listDuration.add(Date.now() - start);

    if (check(res, {
      'list orders returns 200': (r) => r.status === 200,
      'list orders not 5xx': (r) => r.status < 500,
    })) {
      totalListRequests.add(1);
    }

    sleep(1);
  } else {
    const start = Date.now();

    const res = http.get(`${BASE_URL}/api/health`, {
      headers: HEADERS,
    });

    healthDuration.add(Date.now() - start);

    if (check(res, {
      'health returns 200': (r) => r.status === 200,
      'health not 5xx': (r) => r.status < 500,
    })) {
      totalHealthRequests.add(1);
    }

    sleep(0.5);
  }
}
