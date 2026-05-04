/**
 * IDEMPOTENCY TEST - Kiểm tra không tạo duplicate order
 * Gửi cùng 1 request nhiều lần với cùng Idempotency Key
 * Kỳ vọng: Chỉ 1 order được tạo, tất cả request trả về cùng order_id
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const firstRequestSuccess = new Rate('first_request_success');
const idempotencyCorrect = new Rate('idempotency_correct');

export const options = {
  vus: 10,
  iterations: 50,
  thresholds: {
    first_request_success: ['rate>0.95'],
    idempotency_correct: ['rate>0.99'],
  },
};

const BASE_URL = 'http://100.65.255.2';

export default function () {
  const idempotencyKey = `idem-vu${__VU}-group${Math.floor(__ITER / 5)}`;
  const userId = `user-idem-${__VU}`;

  const payload = JSON.stringify({
    user_id: userId,
    items: [{ product_id: 'product-001', quantity: 1 }],
    currency: 'VND',
    payment_method: 'credit_card',
    shipping_address: 'Idempotency Test Street, HCMC',
    note: `idempotency test key:${idempotencyKey}`,
  });

  const headers = {
    'Content-Type': 'application/json',
    'X-Idempotency-Key': idempotencyKey,
  };

  // Request lần 1
  const res1 = http.post(`${BASE_URL}/api/orders`, payload, { headers });
  const success1 = check(res1, {
    'first request 2xx': (r) => r.status >= 200 && r.status < 300,
  });
  firstRequestSuccess.add(success1);

  let orderId1 = null;
  try { orderId1 = JSON.parse(res1.body)?.data?.order?.id; } catch { }

  sleep(0.2);

  // Request lần 2 - cùng key
  const res2 = http.post(`${BASE_URL}/api/orders`, payload, { headers });
  let orderId2 = null;
  try { orderId2 = JSON.parse(res2.body)?.data?.order?.id; } catch { }

  const idem2 = check(res2, {
    'req2 returns same order_id': () => orderId1 !== null && orderId2 !== null && orderId1 === orderId2,
  });
  idempotencyCorrect.add(idem2);

  sleep(0.2);

  // Request lần 3 - cùng key
  const res3 = http.post(`${BASE_URL}/api/orders`, payload, { headers });
  let orderId3 = null;
  try { orderId3 = JSON.parse(res3.body)?.data?.order?.id; } catch { }

  const idem3 = check(res3, {
    'req3 returns same order_id': () => orderId1 !== null && orderId3 !== null && orderId1 === orderId3,
  });
  idempotencyCorrect.add(idem3);

  sleep(1);
}

export function handleSummary(data) {
  const m = data.metrics;
  const idemRate = m.idempotency_correct?.values?.rate;
  return {
    'stdout': JSON.stringify({
      test_type: 'IDEMPOTENCY TEST',
      results: {
        total_requests: m.http_reqs?.values?.count,
        first_request_success: (m.first_request_success?.values?.rate * 100)?.toFixed(2) + '%',
        idempotency_correct_rate: (idemRate * 100)?.toFixed(2) + '%',
        error_rate: (m.http_req_failed?.values?.rate * 100)?.toFixed(2) + '%',
      },
      conclusion: idemRate > 0.99
        ? '✅ PASS: Idempotency hoạt động đúng - không có duplicate order'
        : '❌ FAIL: Có duplicate order được tạo',
    }, null, 2),
  };
}