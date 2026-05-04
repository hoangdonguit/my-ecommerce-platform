/**
 * SMOKE TEST - Kiểm tra hệ thống hoạt động đúng cơ bản
 * Mục đích: Xác minh hệ thống hoạt động ở tải tối thiểu
 * Cấu hình: 1-2 VUs, 2 phút
 * Ngưỡng pass: error rate < 1%, p95 < 500ms
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('error_rate');
const orderDuration = new Trend('order_duration');

// Cấu hình test
export const options = {
  vus: 1,
  duration: '2m',
  thresholds: {
    http_req_failed: ['rate<0.01'],        // error rate < 1%
    http_req_duration: ['p(95)<2000'],       // p95 < 500ms
    error_rate: ['rate<0.01'],
  },
};

// URL của API Gateway qua Istio
const BASE_URL = 'http://100.65.255.2';

// Tạo idempotency key ngẫu nhiên
function generateIdempotencyKey() {
  return `smoke-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

// Tạo order payload
function createOrderPayload(userId) {
  return JSON.stringify({
    user_id: userId,
    items: [
      {
        product_id: 'product-001',
        quantity: 1,
      },
    ],
    currency: 'VND',
    payment_method: 'credit_card',
    shipping_address: '123 Test Street, Ho Chi Minh City',
    note: 'k6 smoke test order',
  });
}

export default function () {
  const userId = `user-smoke-${__VU}`;

  // === TEST 1: Health Check ===
  const healthRes = http.get(`${BASE_URL}/api/health`, {
    tags: { name: 'health_check' },
  });

  check(healthRes, {
    'health check status 200': (r) => r.status === 200,
    'health check success true': (r) => {
      try {
        return JSON.parse(r.body).success === true;
      } catch { return false; }
    },
  });
  errorRate.add(healthRes.status !== 200);

  sleep(1);

  // === TEST 2: Services Health Check ===
  const servicesRes = http.get(`${BASE_URL}/api/health/services`, {
    tags: { name: 'services_health' },
    timeout: '3s',
  });

  check(servicesRes, {
    'services health status 200': (r) => r.status === 200,
    'all services ok': (r) => {
      try {
        const body = JSON.parse(r.body);
        const d = body.data;
        return d.order_service?.ok && d.inventory_service?.ok &&
               d.payment_service?.ok && d.notification_service?.ok;
      } catch { return false; }
    },
  });
  errorRate.add(servicesRes.status !== 200);

  sleep(1);

  // === TEST 3: Tạo Order ===
  const idempotencyKey = generateIdempotencyKey();
  const startTime = Date.now();

  const orderRes = http.post(
    `${BASE_URL}/api/orders`,
    createOrderPayload(userId),
    {
      headers: {
        'Content-Type': 'application/json',
        'X-Idempotency-Key': idempotencyKey,
      },
      tags: { name: 'create_order' },
    }
  );

  orderDuration.add(Date.now() - startTime);

  const orderCreated = check(orderRes, {
    'create order status 2xx': (r) => r.status >= 200 && r.status < 300,
    'create order has order_id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data?.order_id !== undefined || body.data?.id !== undefined;
      } catch { return false; }
    },
  });
  errorRate.add(!orderCreated);

  sleep(1);

  // === TEST 4: List Orders ===
  const listRes = http.get(
    `${BASE_URL}/api/orders?user_id=${userId}&page=1&limit=10`,
    { tags: { name: 'list_orders' } }
  );

  check(listRes, {
    'list orders status 200': (r) => r.status === 200,
  });
  errorRate.add(listRes.status !== 200);

  sleep(1);
}

export function handleSummary(data) {
  return {
    'stdout': JSON.stringify({
      test_type: 'SMOKE TEST',
      passed: data.metrics.http_req_failed.values.rate < 0.01,
      error_rate: (data.metrics.http_req_failed.values.rate * 100).toFixed(2) + '%',
      p95_latency: data.metrics.http_req_duration.values['p(95)'].toFixed(2) + 'ms',
      total_requests: data.metrics.http_reqs.values.count,
    }, null, 2),
  };
}
