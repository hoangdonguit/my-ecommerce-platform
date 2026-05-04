/**
 * STRESS TEST - Tìm breaking point của hệ thống
 * Cấu hình: Tăng tải theo bậc thang 100→200→300→400 VUs
 * Mục đích: Xác định điểm bắt đầu lỗi và hành vi recovery
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

const errorRate = new Rate('error_rate');
const orderDuration = new Trend('order_duration_ms');
const failedOrders = new Counter('failed_orders');
const successOrders = new Counter('success_orders');

export const options = {
  stages: [
    { duration: '2m', target: 50 },    // Warm up
    { duration: '3m', target: 100 },   // Bước 1: 100 VUs
    { duration: '3m', target: 200 },   // Bước 2: 200 VUs
    { duration: '3m', target: 300 },   // Bước 3: 300 VUs
    { duration: '3m', target: 400 },   // Bước 4: 400 VUs (breaking point?)
    { duration: '3m', target: 200 },   // Recovery: giảm xuống
    { duration: '2m', target: 0 },     // Ramp down
  ],
  thresholds: {
    // Stress test không có ngưỡng cứng - quan sát hành vi
    http_req_duration: ['p(99)<3000'],  // Chỉ cần p99 < 3s
  },
};

const BASE_URL = 'http://100.65.255.2';

function generateIdempotencyKey() {
  return `stress-${__VU}-${__ITER}-${Date.now()}`;
}

export default function () {
  const userId = `user-stress-${__VU}`;
  const idempotencyKey = generateIdempotencyKey();
  const start = Date.now();

  const res = http.post(
    `${BASE_URL}/api/orders`,
    JSON.stringify({
      user_id: userId,
      items: [{ product_id: 'product-001', quantity: 1 }],
      currency: 'VND',
      payment_method: 'credit_card',
      shipping_address: 'Stress Test Address, HCMC',
      note: `k6 stress test VU:${__VU} ITER:${__ITER}`,
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        'X-Idempotency-Key': idempotencyKey,
      },
      tags: { name: 'stress_order' },
      timeout: '10s',
    }
  );

  orderDuration.add(Date.now() - start);

  const success = check(res, {
    'status 2xx': (r) => r.status >= 200 && r.status < 300,
    'response time < 2s': (r) => r.timings.duration < 2000,
  });

  errorRate.add(!success);
  if (success) {
    successOrders.add(1);
  } else {
    failedOrders.add(1);
    // Log lỗi khi status không phải 2xx
    if (res.status >= 400) {
      console.error(`VU:${__VU} ITER:${__ITER} - Status:${res.status} Body:${res.body?.substring(0, 100)}`);
    }
  }

  // Giảm sleep khi stress để tạo tải cao hơn
  sleep(Math.random() * 0.5);
}

export function handleSummary(data) {
  const m = data.metrics;
  return {
    'stdout': JSON.stringify({
      test_type: 'STRESS TEST',
      stages: '50→100→200→300→400→200→0 VUs',
      results: {
        total_requests: m.http_reqs?.values?.count,
        peak_rps: m.http_reqs?.values?.rate?.toFixed(2),
        error_rate: (m.http_req_failed?.values?.rate * 100)?.toFixed(2) + '%',
        p50_latency: m.http_req_duration?.values['p(50)']?.toFixed(2) + 'ms',
        p95_latency: m.http_req_duration?.values['p(95)']?.toFixed(2) + 'ms',
        p99_latency: m.http_req_duration?.values['p(99)']?.toFixed(2) + 'ms',
        max_latency: m.http_req_duration?.values?.max?.toFixed(2) + 'ms',
        success_orders: m.success_orders?.values?.count,
        failed_orders: m.failed_orders?.values?.count,
      },
      note: 'Quan sát breaking point tại bước nào error rate tăng đột biến',
    }, null, 2),
  };
}
