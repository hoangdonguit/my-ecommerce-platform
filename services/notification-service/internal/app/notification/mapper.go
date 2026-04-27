package notificationapp

import domainnotification "github.com/hoangdonguit/my-ecommerce-platform/notification-service/internal/domain/notification"

func ToNotificationResponse(n *domainnotification.Notification) NotificationResponse {
	return NotificationResponse{
		ID:            n.ID,
		UserID:        n.UserID,
		OrderID:       n.OrderID,
		EventType:     n.EventType,
		Channel:       n.Channel,
		Recipient:     n.Recipient,
		Title:         n.Title,
		Message:       n.Message,
		Status:        n.Status,
		FailureReason: n.FailureReason,
		SentAt:        n.SentAt,
		CreatedAt:     n.CreatedAt,
		UpdatedAt:     n.UpdatedAt,
	}
}

func ToNotificationResponses(items []domainnotification.Notification) []NotificationResponse {
	result := make([]NotificationResponse, 0, len(items))

	for i := range items {
		copyItem := items[i]
		result = append(result, ToNotificationResponse(&copyItem))
	}

	return result
}
