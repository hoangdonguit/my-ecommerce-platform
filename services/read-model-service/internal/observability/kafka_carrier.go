package observability

import kafkago "github.com/segmentio/kafka-go"

type KafkaHeadersCarrier struct {
	Headers *[]kafkago.Header
}

func NewKafkaHeadersCarrier(headers *[]kafkago.Header) KafkaHeadersCarrier {
	return KafkaHeadersCarrier{Headers: headers}
}

func (c KafkaHeadersCarrier) Get(key string) string {
	if c.Headers == nil {
		return ""
	}

	for _, header := range *c.Headers {
		if header.Key == key {
			return string(header.Value)
		}
	}

	return ""
}

func (c KafkaHeadersCarrier) Set(key string, value string) {
	if c.Headers == nil {
		return
	}

	headers := *c.Headers

	for i := range headers {
		if headers[i].Key == key {
			headers[i].Value = []byte(value)
			*c.Headers = headers
			return
		}
	}

	*c.Headers = append(headers, kafkago.Header{
		Key:   key,
		Value: []byte(value),
	})
}

func (c KafkaHeadersCarrier) Keys() []string {
	if c.Headers == nil {
		return nil
	}

	keys := make([]string, 0, len(*c.Headers))
	for _, header := range *c.Headers {
		keys = append(keys, header.Key)
	}

	return keys
}
