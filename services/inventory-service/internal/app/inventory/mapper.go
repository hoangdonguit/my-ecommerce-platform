package inventoryapp

import domaininventory "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/domain/inventory"

func ToInventoryResponse(inv *domaininventory.Inventory) InventoryResponse {
	return InventoryResponse{
		ProductID:         inv.ProductID,
		SKU:               inv.SKU,
		OnHandQuantity:    inv.OnHandQuantity,
		ReservedQuantity:  inv.ReservedQuantity,
		AvailableQuantity: inv.AvailableQuantity,
		UpdatedAt:         inv.UpdatedAt,
	}
}

func ToReservationResponse(res *domaininventory.InventoryReservation) ReservationResponse {
	items := make([]ReservationItemResponse, 0, len(res.Items))

	for _, item := range res.Items {
		items = append(items, ReservationItemResponse{
			ProductID:         item.ProductID,
			RequestedQuantity: item.RequestedQuantity,
			ReservedQuantity:  item.ReservedQuantity,
			Status:            item.Status,
		})
	}

	return ReservationResponse{
		ID:        res.ID,
		OrderID:   res.OrderID,
		UserID:    res.UserID,
		Status:    res.Status,
		Reason:    res.Reason,
		Items:     items,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}
}
