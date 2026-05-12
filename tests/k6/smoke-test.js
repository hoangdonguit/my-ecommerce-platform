import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

export const options = {
  vus: 1,
  duration: '2m',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

const BASE_URL = 'http://100.65.255.2'; 
const HEADERS = { 'Content-Type': 'application/json', 'X-API-Key': 'UIT-DOAN-2026-SECRET' };

function generateIdempotencyKey() {
  return `smoke-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

export default function () {
  const userId = `user-smoke-${__VU}`;

  // 1. Health Check
  const healthRes = http.get(`${BASE_URL}/api/health`, { headers: HEADERS });
  check(healthRes, { 'health 200': (r) => r.status === 200 });

  // 2. Services Health Check
  const servicesRes = http.get(`${BASE_URL}/api/health/services`, { headers: HEADERS });
  check(servicesRes, { 'services health 200': (r) => r.status === 200 });

  // 3. Create Order
  const orderRes = http.post(`${BASE_URL}/api/orders`, JSON.stringify({
    user_id: userId,
    items: [{ product_id: 'prod-123', quantity: 1 }],
    currency: 'VND',
    payment_method: 'COD',
    shipping_address: '123 Smoke Test'
  }), {
    headers: { ...HEADERS, 'X-Idempotency-Key': generateIdempotencyKey() }
  });
  check(orderRes, { 'order created 2xx': (r) => r.status >= 200 && r.status < 300 });

  sleep(1); 
}