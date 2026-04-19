package inventoryapp

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	domaininventory "github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/domain/inventory"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/infrastructure/persistence"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/shared/errs"
)

type EventPublisher interface {
	PublishReserved(ctx context.Context, event InventoryReservedEvent) error
	PublishFailed(ctx context.Context, event InventoryFailedEvent) error
}

type Service struct {
	repo      domaininventory.Repository
	publisher EventPublisher
}

func NewService(repo domaininventory.Repository, publisher EventPublisher) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *Service) GetInventory(ctx context.Context, productID string) (*domaininventory.Inventory, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, errs.BadRequest("product_id is required")
	}

	inv, err := s.repo.GetInventoryByProductID(ctx, productID)
	if err != nil {
		if persistence.IsNotFound(err) {
			return nil, errs.NotFound("inventory not found")
		}
		return nil, errs.WrapInternal(err, "failed to get inventory")
	}

	return inv, nil
}

func (s *Service) GetReservationByOrderID(ctx context.Context, orderID string) (*domaininventory.InventoryReservation, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, errs.BadRequest("order_id is required")
	}

	reservation, err := s.repo.FindReservationByOrderID(ctx, orderID)
	if err != nil {
		if persistence.IsNotFound(err) {
			return nil, errs.NotFound("reservation not found")
		}
		return nil, errs.WrapInternal(err, "failed to get reservation")
	}

	return reservation, nil
}

func (s *Service) HandleOrderCreated(ctx context.Context, event OrderCreatedEvent) error {
	if err := validateOrderCreatedEvent(event); err != nil {
		return err
	}

	existing, err := s.repo.FindReservationByOrderID(ctx, event.OrderID)
	if err == nil && existing != nil {
		return nil
	}
	if err != nil && !persistence.IsNotFound(err) {
		return errs.WrapInternal(err, "failed to check existing reservation")
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return errs.WrapInternal(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	productIDs := make([]string, 0, len(event.Items))
	for _, item := range event.Items {
		productIDs = append(productIDs, item.ProductID)
	}

	lockedInventories, err := s.repo.GetInventoriesForUpdate(ctx, tx, productIDs)
	if err != nil {
		return errs.WrapInternal(err, "failed to lock inventories")
	}

	now := time.Now()
	reservationID := uuid.NewString()
	reservation := &domaininventory.InventoryReservation{
		ID:        reservationID,
		OrderID:   event.OrderID,
		UserID:    event.UserID,
		Status:    domaininventory.ReservationPending,
		Reason:    "",
		Items:     []domaininventory.InventoryReservationItem{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	failedItems := make([]InventoryFailedEventItem, 0)
	reservationItems := make([]domaininventory.InventoryReservationItem, 0, len(event.Items))

	for _, item := range event.Items {
		inv, ok := lockedInventories[item.ProductID]
		if !ok {
			failedItems = append(failedItems, InventoryFailedEventItem{
				ProductID:         item.ProductID,
				RequestedQuantity: item.Quantity,
				AvailableQuantity: 0,
			})

			reservationItems = append(reservationItems, domaininventory.InventoryReservationItem{
				ID:                uuid.NewString(),
				ReservationID:     reservationID,
				ProductID:         item.ProductID,
				RequestedQuantity: item.Quantity,
				ReservedQuantity:  0,
				Status:            domaininventory.ItemFailed,
				CreatedAt:         now,
			})
			continue
		}

		if inv.AvailableQuantity < item.Quantity {
			failedItems = append(failedItems, InventoryFailedEventItem{
				ProductID:         item.ProductID,
				RequestedQuantity: item.Quantity,
				AvailableQuantity: inv.AvailableQuantity,
			})

			reservationItems = append(reservationItems, domaininventory.InventoryReservationItem{
				ID:                uuid.NewString(),
				ReservationID:     reservationID,
				ProductID:         item.ProductID,
				RequestedQuantity: item.Quantity,
				ReservedQuantity:  0,
				Status:            domaininventory.ItemFailed,
				CreatedAt:         now,
			})
			continue
		}

		reservationItems = append(reservationItems, domaininventory.InventoryReservationItem{
			ID:                uuid.NewString(),
			ReservationID:     reservationID,
			ProductID:         item.ProductID,
			RequestedQuantity: item.Quantity,
			ReservedQuantity:  item.Quantity,
			Status:            domaininventory.ItemReserved,
			CreatedAt:         now,
		})
	}

	if len(failedItems) > 0 {
		reservation.Status = domaininventory.ReservationFailed
		reservation.Reason = "insufficient stock or product not found"
		reservation.Items = reservationItems

		if err := s.repo.CreateReservation(ctx, tx, reservation); err != nil {
			return errs.WrapInternal(err, "failed to create failed reservation")
		}

		if err := s.repo.CreateReservationItems(ctx, tx, reservationItems); err != nil {
			return errs.WrapInternal(err, "failed to create failed reservation items")
		}

		if err := tx.Commit(ctx); err != nil {
			return errs.WrapInternal(err, "failed to commit failed reservation")
		}

		if s.publisher != nil {
			failEvent := InventoryFailedEvent{
				EventType: "inventory.failed",
				OrderID:   event.OrderID,
				UserID:    event.UserID,
				Status:    domaininventory.ReservationFailed,
				Reason:    reservation.Reason,
				Items:     failedItems,
			}
			if err := s.publisher.PublishFailed(ctx, failEvent); err != nil {
				return errs.WrapInternal(err, "failed to publish inventory.failed")
			}
		}

		return nil
	}

	for _, item := range event.Items {
		inv := lockedInventories[item.ProductID]

		newReserved := inv.ReservedQuantity + item.Quantity
		newAvailable := inv.AvailableQuantity - item.Quantity

		if err := s.repo.UpdateInventoryQuantities(
			ctx,
			tx,
			item.ProductID,
			inv.OnHandQuantity,
			newReserved,
			newAvailable,
		); err != nil {
			return errs.WrapInternal(err, "failed to update inventory quantities")
		}
	}

	reservation.Status = domaininventory.ReservationReserved
	reservation.Reason = ""
	reservation.Items = reservationItems

	if err := s.repo.CreateReservation(ctx, tx, reservation); err != nil {
		return errs.WrapInternal(err, "failed to create reservation")
	}

	if err := s.repo.CreateReservationItems(ctx, tx, reservationItems); err != nil {
		return errs.WrapInternal(err, "failed to create reservation items")
	}

	if err := tx.Commit(ctx); err != nil {
		return errs.WrapInternal(err, "failed to commit reservation")
	}

	if s.publisher != nil {
		successItems := make([]InventoryReservedEventItem, 0, len(event.Items))
		for _, item := range event.Items {
			successItems = append(successItems, InventoryReservedEventItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}

		reservedEvent := InventoryReservedEvent{
			EventType: "inventory.reserved",
			OrderID:   event.OrderID,
			UserID:    event.UserID,
			Status:    domaininventory.ReservationReserved,
			Items:     successItems,
		}

		if err := s.publisher.PublishReserved(ctx, reservedEvent); err != nil {
			return errs.WrapInternal(err, "failed to publish inventory.reserved")
		}
	}

	return nil
}

func validateOrderCreatedEvent(event OrderCreatedEvent) error {
	if strings.TrimSpace(event.OrderID) == "" {
		return errs.BadRequest("order_id is required")
	}

	if strings.TrimSpace(event.UserID) == "" {
		return errs.BadRequest("user_id is required")
	}

	if len(event.Items) == 0 {
		return errs.BadRequest("items must not be empty")
	}

	for _, item := range event.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return errs.BadRequest("product_id is required")
		}
		if item.Quantity <= 0 {
			return errs.BadRequest("quantity must be greater than 0")
		}
	}

	return nil
}
