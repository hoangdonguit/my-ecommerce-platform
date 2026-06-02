package observability

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	kafkago "github.com/segmentio/kafka-go"
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

func KafkaHeadersFromJSON(raw []byte) []kafkago.Header {
	rawText := strings.TrimSpace(string(raw))
	if rawText == "" {
		return nil
	}

	var headers map[string]string
	if err := json.Unmarshal([]byte(rawText), &headers); err != nil {
		return nil
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		key = strings.TrimSpace(key)
		if key != "" {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	result := make([]kafkago.Header, 0, len(keys))
	for _, key := range keys {
		value := strings.TrimSpace(headers[key])
		if value == "" {
			continue
		}

		result = append(result, kafkago.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	return result
}
