/**
 * SOAK TEST - Phát hiện memory leak và performance degradation
 * Cấu hình: 30-50 VUs, chạy 1-2 giờ liên tục
 * Mục đích: Phát hiện memory leak, resource exhaustion theo thời gian
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

const errorRate = new Rate('error_rate');
const orderDuration = new Trend('order_duration_ms');
const totalOrders = new Counter('total_orders');
const failedOrders = new Counter('failed_orders');

export const options = {
  stages: [
    { duration: '5m', target: 30 },   // Ramp up chậm
    { duration: '50m', target: 30 },  // Duy trì 50 phút (đổi thành 110m cho 2h)
    { duration: '5m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],           // error rate < 1%
    http_req_duration: ['p(95)<500'],          // p95 < 500ms ổn định
    error_rate: ['rate<0.01'],
  },
};

const BASE_URL = 'http://100.65.255.2';
const PRODUCTS = ['product-001', 'product-002', 'product-003'];

function generateIdempotencyKey() {
  return `soak-${__VU}-${__ITER}-${Date.now()}`;
}

// Theo dõi latency theo thời gian để phát hiện drift
let iterationCount = 0;
let latencyBaseline = null;

export default function () {
  iterationCount++;
  const userId = `user-soak-${__VU}`;
  const rand = Math.random();

  if (rand < 0.6) {
    // 60% - Tạo order
    const idempotencyKey = generateIdempotencyKey();
    const start = Date.now();

    const res = http.post(
      `${BASE_URL}/api/orders`,
      JSON.stringify({
        user_id: userId,
        items: [{
          product_id: PRODUCTS[Math.floor(Math.random() * PRODUCTS.length)],
          quantity: 1,
        }],
        currency: 'VND',
        payment_method: 'credit_card',
        shipping_address: 'Soak Test Street, HCMC',
        note: `soak iter:${__ITER}`,
      }),
      {
        headers: {
          'Content-Type': 'application/json',
          'X-Idempotency-Key': idempotencyKey,
        },
        tags: { name: 'create_order' },
        timeout: '10s',
      }
    );

    const duration = Date.now() - start;
    orderDuration.add(duration);

    // Theo dõi latency drift
    if (iterationCount === 10) {
      latencyBaseline = duration;
    }
    if (latencyBaseline && iterationCount % 100 === 0) {
      const drift = ((duration - latencyBaseline) / latencyBaseline * 100).toFixed(1);
      if (Math.abs(parseFloat(drift)) > 10) {
        console.warn(`VU:${__VU} - Latency drift: ${drift}% (baseline:${latencyBaseline}ms current:${duration}ms)`);
      }
    }

    const success = check(res, {
      'soak order 2xx': (r) => r.status >= 200 && r.status < 300,
    });

    errorRate.add(!success);
    if (success) totalOrders.add(1);
    else failedOrders.add(1);

    sleep(Math.random() * 2 + 1);

  } else if (rand < 0.9) {
    // 30% - List orders
    const res = http.get(
      `${BASE_URL}/api/orders?user_id=${userId}&page=1&limit=5`,
      { tags: { name: 'list_orders' }, timeout: '5s' }
    );
    check(res, { 'list 200': (r) => r.status === 200 });
    errorRate.add(res.status !== 200);
    sleep(1);

  } else {
    // 10% - Health check
    const res = http.get(`${BASE_URL}/api/health`, {
      tags: { name: 'health_check' },
    });
    check(res, { 'health 200': (r) => r.status === 200 });
    sleep(0.5);
  }
}

export function handleSummary(data) {
  const m = data.metrics;
  return {
    'stdout': JSON.stringify({
      test_type: 'SOAK TEST',
      duration: '60 minutes',
      results: {
        total_requests: m.http_reqs?.values?.count,
        avg_rps: m.http_reqs?.values?.rate?.toFixed(2),
        error_rate: (m.http_req_failed?.values?.rate * 100)?.toFixed(2) + '%',
        p50_latency: m.http_req_duration?.values['p(50)']?.toFixed(2) + 'ms',
        p95_latency: m.http_req_duration?.values['p(95)']?.toFixed(2) + 'ms',
        p99_latency: m.http_req_duration?.values['p(99)']?.toFixed(2) + 'ms',
        min_latency: m.http_req_duration?.values?.min?.toFixed(2) + 'ms',
        max_latency: m.http_req_duration?.values?.max?.toFixed(2) + 'ms',
        orders_created: m.total_orders?.values?.count,
        failed_orders: m.failed_orders?.values?.count,
      },
      memory_leak_check: 'Xem Grafana: heap growth < 10% = không có leak',
      latency_drift_check: 'p95 cuối test so với đầu test, drift < 10% = ổn định',
    }, null, 2),
  };
}
