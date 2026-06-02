package observability

import (
	"context"
	"encoding/json"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func InjectTraceHeaders(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	if len(carrier) == 0 {
		return map[string]string{}
	}

	headers := make(map[string]string, len(carrier))
	for key, value := range carrier {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		headers[key] = value
	}

	return headers
}

func MarshalTraceHeaders(headers map[string]string) []byte {
	if len(headers) == 0 {
		return []byte("{}")
	}

	raw, err := json.Marshal(headers)
	if err != nil {
		return []byte("{}")
	}

	return raw
}

func UnmarshalTraceHeaders(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}
	}

	var headers map[string]string
	if err := json.Unmarshal([]byte(raw), &headers); err != nil {
		return map[string]string{}
	}

	cleaned := make(map[string]string, len(headers))
	for key, value := range headers {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		cleaned[key] = value
	}

	return cleaned
}
