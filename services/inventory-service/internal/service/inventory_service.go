package service

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/model"
	"github.com/hoangdonguit/my-ecommerce-platform/inventory-service/internal/repository"
)

type EventPublisher interface {
	PublishReserved(ctx context.Context, key string, event any) error
	PublishFailed(ctx context.Context, key string, event any) error
}

type InventoryService struct {
	repo      *repository.InventoryRepository
	publisher EventPublisher
}

func NewInventoryService(repo *repository.InventoryRepository, publisher EventPublisher) *InventoryService {
	return &InventoryService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *InventoryService) GetInventory(ctx context.Context, productID string) (*model.Inventory, error) {
	return s.repo.GetByProductID(ctx, productID)
}

func (s *InventoryService) HandleOrderCreated(ctx context.Context, event model.OrderCreatedEvent) error {
	if event.OrderID == "" || event.ProductID == "" || event.Quantity <= 0 {
		return errors.New("invalid order.created event")
	}

	existing, err := s.repo.FindReservationByOrderID(ctx, event.OrderID)
	if err == nil && existing != nil {
		log.Printf("order %s already processed, skip", event.OrderID)
		return nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	inv, err := s.repo.GetInventoryForUpdate(ctx, tx, event.ProductID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			failEvent := model.InventoryFailedEvent{
				EventType: "inventory.failed",
				OrderID:   event.OrderID,
				ProductID: event.ProductID,
				Quantity:  event.Quantity,
				Reason:    "product not found",
				Status:    "FAILED",
			}
			return s.publisher.PublishFailed(ctx, event.OrderID, failEvent)
		}
		return err
	}

	if inv.AvailableQuantity < event.Quantity {
		reservation := model.InventoryReservation{
			ID:               uuid.NewString(),
			OrderID:          event.OrderID,
			ProductID:        event.ProductID,
			ReservedQuantity: event.Quantity,
			Status:           "FAILED",
		}

		if err := s.repo.CreateReservation(ctx, tx, reservation); err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		failEvent := model.InventoryFailedEvent{
			EventType: "inventory.failed",
			OrderID:   event.OrderID,
			ProductID: event.ProductID,
			Quantity:  event.Quantity,
			Reason:    "insufficient stock",
			Status:    "FAILED",
		}

		return s.publisher.PublishFailed(ctx, event.OrderID, failEvent)
	}

	newQty := inv.AvailableQuantity - event.Quantity

	if err := s.repo.UpdateAvailableQuantity(ctx, tx, event.ProductID, newQty); err != nil {
		return err
	}

	reservation := model.InventoryReservation{
		ID:               uuid.NewString(),
		OrderID:          event.OrderID,
		ProductID:        event.ProductID,
		ReservedQuantity: event.Quantity,
		Status:           "RESERVED",
	}

	if err := s.repo.CreateReservation(ctx, tx, reservation); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	successEvent := model.InventoryReservedEvent{
		EventType: "inventory.reserved",
		OrderID:   event.OrderID,
		ProductID: event.ProductID,
		Quantity:  event.Quantity,
		Status:    "RESERVED",
	}

	return s.publisher.PublishReserved(ctx, event.OrderID, successEvent)
}
