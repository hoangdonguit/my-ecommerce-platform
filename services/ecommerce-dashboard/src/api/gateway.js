const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8090";

function createFallbackIdempotencyKey() {
  if (window.crypto && window.crypto.randomUUID) {
    return `web-${window.crypto.randomUUID()}`;
  }

  return `web-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

async function request(path, options = {}) {
  const url = `${API_BASE_URL}${path}`;

  const headers = {
    Accept: "application/json",
    ...(options.body ? { "Content-Type": "application/json" } : {}),
    ...(options.headers || {}),
  };

  let response;
  let text;

  try {
    response = await fetch(url, {
      ...options,
      headers,
    });

    text = await response.text();
  } catch (error) {
    throw new Error(`Không kết nối được Web Gateway tại ${API_BASE_URL}. Kiểm tra gateway đã chạy chưa.`);
  }

  let json = null;

  if (text) {
    try {
      json = JSON.parse(text);
    } catch {
      throw new Error(`Response không phải JSON hợp lệ từ ${url}`);
    }
  }

  if (!response.ok || json?.success === false) {
    const message =
      json?.message ||
      json?.error?.message ||
      `Request thất bại với HTTP ${response.status}`;

    const error = new Error(message);
    error.status = response.status;
    error.payload = json;
    throw error;
  }

  return json;
}

export async function getGatewayHealth() {
  return request("/api/health");
}

export async function getServicesHealth() {
  return request("/api/health/services");
}

export async function createOrder(input) {
  const idempotencyKey = input.idempotency_key || createFallbackIdempotencyKey();

  const body = {
    user_id: input.user_id,
    items: [
      {
        product_id: input.product_id,
        quantity: Number(input.quantity),
      },
    ],
    currency: input.currency,
    payment_method: input.payment_method,
    shipping_address: input.shipping_address,
    note: input.note || "",
  };

  return request("/api/orders", {
    method: "POST",
    headers: {
      "X-Idempotency-Key": idempotencyKey,
    },
    body: JSON.stringify(body),
  });
}

export async function listOrders(userId, page = 1, limit = 10) {
  const params = new URLSearchParams({
    user_id: userId,
    page: String(page),
    limit: String(limit),
  });

  return request(`/api/orders?${params.toString()}`);
}

export async function getOrder(orderId) {
  return request(`/api/orders/${encodeURIComponent(orderId)}`);
}

export async function getOrderSaga(orderId) {
  return request(`/api/orders/${encodeURIComponent(orderId)}/saga`);
}