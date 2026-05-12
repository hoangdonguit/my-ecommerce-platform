import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

const orderDuration = new Trend('order_duration_ms');
const totalOrders = new Counter('total_orders');

export const options = {
  stages: [
    { duration: '5m', target: 30 },
    { duration: '50m', target: 30 },
    { duration: '5m', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

const BASE_URL = 'http://100.65.255.2';
const HEADERS = { 'Content-Type': 'application/json', 'X-API-Key': 'UIT-DOAN-2026-SECRET' };

export default function () {
  const userId = `user-soak-${__VU}`;
  const rand = Math.random();

  if (rand < 0.30) { // 30% Tạo đơn
    const start = Date.now();
    const res = http.post(`${BASE_URL}/api/orders`, JSON.stringify({
      user_id: userId,
      items: [{ product_id: 'prod-123', quantity: 1 }],
      currency: 'VND',
      payment_method: 'COD',
      shipping_address: 'Soak Test'
    }), {
      headers: { ...HEADERS, 'X-Idempotency-Key': `soak-${__VU}-${Date.now()}` }
    });
    orderDuration.add(Date.now() - start);
    if (check(res, { 'create 2xx': (r) => r.status < 300 })) totalOrders.add(1);
    sleep(1);
  } else if (rand < 0.90) { // 60% Liệt kê đơn (Pagination)
    const res = http.get(`${BASE_URL}/api/orders?user_id=${userId}&page=1&limit=5`, { headers: HEADERS });
    check(res, { 'list 200': (r) => r.status === 200 });
    sleep(1);
  } else { // 10% Health
    const res = http.get(`${BASE_URL}/api/health`, { headers: HEADERS });
    check(res, { 'health 200': (r) => r.status === 200 });
    sleep(0.5);
  }
}