package inventoryapp

import (
	"context"
	"log"
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
	return &Service{repo: repo, publisher: publisher}
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
		// Atomic UPDATE: trừ kho và kiểm tra đủ hàng trong 1 câu SQL
		// Không cần SELECT FOR UPDATE → giảm lock contention đáng kể
		err := s.repo.AtomicReserveInventory(ctx, tx, item.ProductID, item.Quantity)
		if err != nil {
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
			EventType:     "inventory.reserved",
			OrderID:       event.OrderID,
			UserID:        event.UserID,
			Status:        domaininventory.ReservationReserved,
			Currency:      event.Currency,
			PaymentMethod: event.PaymentMethod,
			TotalAmount:   event.TotalAmount,
			Items:         successItems,
		}
		if err := s.publisher.PublishReserved(ctx, reservedEvent); err != nil {
			return errs.WrapInternal(err, "failed to publish inventory.reserved")
		}
	}
	return nil
}

func (s *Service) ListAllInventories(ctx context.Context) ([]domaininventory.Inventory, error) {
	return s.repo.ListAllInventories(ctx)
}

func (s *Service) RollbackInventory(ctx context.Context, orderID string) error {
	log.Printf("🔄 Đang chuẩn bị Rollback cho đơn hàng: %s", orderID)
	var res *domaininventory.InventoryReservation
	var err error
	for i := 0; i < 3; i++ {
		res, err = s.repo.FindReservationByOrderID(ctx, orderID)
		if err == nil {
			break
		}
		log.Printf("⏳ Đợi DB commit (Lần %d)...", i+1)
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		log.Printf("❌ Không tìm thấy đơn hàng %s để Rollback", orderID)
		return nil
	}
	if res.Status == "CANCELLED" || res.Status == "ROLLBACKED" {
		log.Printf("ℹ️ Đơn hàng %s đã được Rollback trước đó rồi.", orderID)
		return nil
	}
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var productIDs []string
	for _, item := range res.Items {
		productIDs = append(productIDs, item.ProductID)
	}
	invs, err := s.repo.GetInventoriesForUpdate(ctx, tx, productIDs)
	if err != nil {
		return err
	}
	for _, item := range res.Items {
		inv, ok := invs[item.ProductID]
		if !ok {
			continue
		}
		newAvailable := inv.AvailableQuantity + item.ReservedQuantity
		newReserved := inv.ReservedQuantity - item.ReservedQuantity
		log.Printf("📦 Nhả hàng SP %s: %d -> %d", item.ProductID, inv.AvailableQuantity, newAvailable)
		err = s.repo.UpdateInventoryQuantities(ctx, tx, item.ProductID, inv.OnHandQuantity, newReserved, newAvailable)
		if err != nil {
			return err
		}
	}
	err = s.repo.UpdateReservationStatus(ctx, tx, res.ID, "ROLLBACKED")
	if err != nil {
		return err
	}
	log.Printf("✅ Đã hoàn tất Rollback cho đơn hàng: %s", orderID)
	return tx.Commit(ctx)
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