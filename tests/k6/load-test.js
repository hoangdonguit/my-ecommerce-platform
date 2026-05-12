import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter, Trend } from 'k6/metrics';

const orderSuccessRate = new Rate('order_success_rate');
const orderDuration = new Trend('order_duration_ms');
const totalOrders = new Counter('total_orders_created');

export const options = {
  stages: [
    { duration: '2m', target: 50 },
    { duration: '10m', target: 50 },
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.005'],
    http_req_duration: ['p(95)<500'],
    order_success_rate: ['rate>0.99'],
  },
};

const BASE_URL = 'http://100.65.255.2';
const PRODUCTS = ['prod-123', 'prod-456', 'prod-789'];
const HEADERS = { 'Content-Type': 'application/json', 'X-API-Key': 'UIT-DOAN-2026-SECRET' };

export default function () {
  const userId = `user-load-${__VU}-${Math.floor(__ITER / 10)}`;
  const rand = Math.random();

  if (rand < 0.60) { // Giữ nguyên logic 60% của ông
    const start = Date.now();
    const res = http.post(`${BASE_URL}/api/orders`, JSON.stringify({
      user_id: userId,
      items: [{
        product_id: PRODUCTS[Math.floor(Math.random() * PRODUCTS.length)],
        quantity: Math.floor(Math.random() * 2) + 1
      }],
      currency: 'VND',
      payment_method: 'COD',
      shipping_address: 'Load Test Street'
    }), {
      headers: { ...HEADERS, 'X-Idempotency-Key': `load-${__VU}-${Date.now()}` }
    });

    orderDuration.add(Date.now() - start);
    const success = check(res, { 'create order 2xx': (r) => r.status >= 200 && r.status < 300 });
    orderSuccessRate.add(success);
    if (success) totalOrders.add(1);
  }
  sleep(1);
}