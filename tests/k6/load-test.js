/**
 * LOAD TEST - Mô phỏng lưu lượng người dùng bình thường và cao điểm
 * Cấu hình: Ramp up 0→50 VUs trong 2 phút, duy trì 10 phút, ramp down 2 phút
 * Ngưỡng pass: throughput > 100 req/s, p50 < 200ms, p95 < 500ms, error < 0.5%
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter, Trend } from 'k6/metrics';

const errorRate = new Rate('error_rate');
const orderSuccessRate = new Rate('order_success_rate');
const orderDuration = new Trend('order_duration_ms');
const totalOrders = new Counter('total_orders_created');

export const options = {
  stages: [
    { duration: '2m', target: 50 },   // Ramp up đến 50 VUs
    { duration: '10m', target: 50 },  // Duy trì 50 VUs
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_failed: ['rate<0.005'],       // error rate < 0.5%
    http_req_duration: ['p(50)<200', 'p(95)<500'],
    order_success_rate: ['rate>0.99'],     // order success > 99%
    http_reqs: ['rate>100'],              // throughput > 100 req/s
  },
};

const BASE_URL = 'http://100.65.255.2';

const PRODUCTS = ['product-001', 'product-002', 'product-003'];
const PAYMENT_METHODS = ['credit_card', 'debit_card', 'bank_transfer'];

function generateIdempotencyKey() {
  return `load-${__VU}-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

function randomItem(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

// Scenario A: Tạo order mới (60% traffic)
function createOrder(userId) {
  const idempotencyKey = generateIdempotencyKey();
  const start = Date.now();

  const res = http.post(
    `${BASE_URL}/api/orders`,
    JSON.stringify({
      user_id: userId,
      items: [{ product_id: randomItem(PRODUCTS), quantity: Math.floor(Math.random() * 3) + 1 }],
      currency: 'VND',
      payment_method: randomItem(PAYMENT_METHODS),
      shipping_address: `${Math.floor(Math.random() * 999)} Load Test Street, HCMC`,
      note: 'k6 load test',
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        'X-Idempotency-Key': idempotencyKey,
      },
      tags: { name: 'create_order', scenario: 'load' },
    }
  );

  orderDuration.add(Date.now() - start);
  const success = check(res, {
    'create order 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  orderSuccessRate.add(success);
  errorRate.add(!success);
  if (success) totalOrders.add(1);

  return res;
}

// Scenario B: Kiểm tra trạng thái order (30% traffic)
function checkOrderStatus(userId) {
  const res = http.get(
    `${BASE_URL}/api/orders?user_id=${userId}&page=1&limit=5`,
    { tags: { name: 'list_orders', scenario: 'load' } }
  );

  check(res, { 'list orders 200': (r) => r.status === 200 });
  errorRate.add(res.status !== 200);
  return res;
}

// Scenario C: Health check (10% traffic)
function healthCheck() {
  const res = http.get(`${BASE_URL}/api/health`, {
    tags: { name: 'health_check', scenario: 'load' },
  });
  check(res, { 'health 200': (r) => r.status === 200 });
  errorRate.add(res.status !== 200);
}

export default function () {
  const userId = `user-load-${__VU}-${Math.floor(__ITER / 10)}`;
  const rand = Math.random();

  if (rand < 0.60) {
    createOrder(userId);
    sleep(Math.random() * 2 + 1);
  } else if (rand < 0.90) {
    checkOrderStatus(userId);
    sleep(Math.random() * 1 + 0.5);
  } else {
    healthCheck();
    sleep(0.5);
  }
}

export function handleSummary(data) {
  const m = data.metrics;
  return {
    'stdout': JSON.stringify({
      test_type: 'LOAD TEST',
      duration: '14 minutes (2m ramp-up + 10m steady + 2m ramp-down)',
      results: {
        total_requests: m.http_reqs?.values?.count,
        throughput_rps: m.http_reqs?.values?.rate?.toFixed(2),
        error_rate: (m.http_req_failed?.values?.rate * 100)?.toFixed(2) + '%',
        p50_latency: m.http_req_duration?.values['p(50)']?.toFixed(2) + 'ms',
        p95_latency: m.http_req_duration?.values['p(95)']?.toFixed(2) + 'ms',
        p99_latency: m.http_req_duration?.values['p(99)']?.toFixed(2) + 'ms',
        orders_created: m.total_orders_created?.values?.count,
        order_success_rate: (m.order_success_rate?.values?.rate * 100)?.toFixed(2) + '%',
      },
      thresholds_passed: {
        error_rate_under_0_5pct: m.http_req_failed?.values?.rate < 0.005,
        p95_under_500ms: m.http_req_duration?.values['p(95)'] < 500,
        throughput_over_100rps: m.http_reqs?.values?.rate > 100,
      },
    }, null, 2),
  };
}
