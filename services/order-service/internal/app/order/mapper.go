package orderapp

import domainorder "github.com/hoangdonguit/my-ecommerce-platform/order-service/internal/domain/order"

func ToOrderResponse(ord *domainorder.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(ord.Items))

	for _, item := range ord.Items {
		items = append(items, OrderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	return OrderResponse{
		ID:              ord.ID,
		UserID:          ord.UserID,
		Status:          ord.Status,
		Currency:        ord.Currency,
		PaymentMethod:   ord.PaymentMethod,
		ShippingAddress: ord.ShippingAddress,
		Note:            ord.Note,
		Items:           items,
		TotalAmount:     ord.TotalAmount,
		CreatedAt:       ord.CreatedAt,
		UpdatedAt:       ord.UpdatedAt,
	}
}

func ToOrderResponses(orders []domainorder.Order) []OrderResponse {
	result := make([]OrderResponse, 0, len(orders))

	for i := range orders {
		orderCopy := orders[i]
		result = append(result, ToOrderResponse(&orderCopy))
	}

	return result
}
