package bulk_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"whatsapp-service/internal/bulk/domain"
	"whatsapp-service/internal/bulk/mocks"
)

func TestMockBulkCampaignStorage(t *testing.T) {
	storage := mocks.NewMockBulkCampaignStorage()

	// Тест создания кампании
	campaign := &domain.BulkCampaign{
		ID:              "test-campaign-1",
		Message:         "Test message",
		Total:           10,
		Status:          "started",
		MessagesPerHour: 5,
	}

	err := storage.Create(campaign)
	require.NoError(t, err)

	// Тест получения кампании
	retrieved, err := storage.GetByID("test-campaign-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "test-campaign-1", retrieved.ID)
	assert.Equal(t, "Test message", retrieved.Message)
	assert.Equal(t, 10, retrieved.Total)
	assert.Equal(t, "started", retrieved.Status)

	// Тест получения несуществующей кампании
	notFound, err := storage.GetByID("non-existent")
	require.NoError(t, err)
	assert.Nil(t, notFound)

	// Тест обновления статуса
	err = storage.UpdateStatus("test-campaign-1", "finished")
	require.NoError(t, err)

	updated, err := storage.GetByID("test-campaign-1")
	require.NoError(t, err)
	assert.Equal(t, "finished", updated.Status)

	// Тест конкурентного доступа
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			campaign := &domain.BulkCampaign{
				ID:              fmt.Sprintf("concurrent-campaign-%d", index),
				Message:         fmt.Sprintf("Message %d", index),
				Total:           index,
				Status:          "started",
				MessagesPerHour: 5,
			}
			err := storage.Create(campaign)
			require.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Проверяем, что все кампании созданы
	for i := 0; i < 10; i++ {
		campaign, err := storage.GetByID(fmt.Sprintf("concurrent-campaign-%d", i))
		require.NoError(t, err)
		assert.NotNil(t, campaign)
		assert.Equal(t, fmt.Sprintf("Message %d", i), campaign.Message)
	}
}

func TestMockBulkCampaignStatusStorage(t *testing.T) {
	storage := mocks.NewMockBulkCampaignStatusStorage()

	// Тест создания статуса
	status := &domain.BulkCampaignStatus{
		CampaignID:  "test-campaign-1",
		PhoneNumber: "71234567890",
		Status:      "pending",
	}

	err := storage.Create(status)
	require.NoError(t, err)

	// Проверяем, что ID был сгенерирован
	assert.NotEmpty(t, status.ID)

	// Тест создания второго статуса
	status2 := &domain.BulkCampaignStatus{
		CampaignID:  "test-campaign-1",
		PhoneNumber: "79876543210",
		Status:      "pending",
	}

	err = storage.Create(status2)
	require.NoError(t, err)

	// Тест получения статусов по ID кампании
	statuses, err := storage.ListByCampaignID("test-campaign-1")
	require.NoError(t, err)
	assert.Len(t, statuses, 2)

	// Проверяем, что номера телефонов корректны
	phoneNumbers := make([]string, 0, len(statuses))
	for _, s := range statuses {
		phoneNumbers = append(phoneNumbers, s.PhoneNumber)
	}
	assert.Contains(t, phoneNumbers, "71234567890")
	assert.Contains(t, phoneNumbers, "79876543210")

	// Тест обновления статуса
	now := time.Now().Format(time.RFC3339)
	errMsg := "test error"
	err = storage.Update(status.ID, "failed", &errMsg, &now)
	require.NoError(t, err)

	// Проверяем обновление
	updatedStatuses, err := storage.ListByCampaignID("test-campaign-1")
	require.NoError(t, err)
	for _, s := range updatedStatuses {
		if s.PhoneNumber == "71234567890" {
			assert.Equal(t, "failed", s.Status)
			assert.Equal(t, "test error", *s.Error)
			assert.Equal(t, now, *s.SentAt)
		}
	}

	// Тест получения статусов несуществующей кампании
	emptyStatuses, err := storage.ListByCampaignID("non-existent")
	require.NoError(t, err)
	assert.Len(t, emptyStatuses, 0)

	// Тест конкурентного доступа
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			status := &domain.BulkCampaignStatus{
				CampaignID:  "concurrent-campaign",
				PhoneNumber: fmt.Sprintf("7123456789%d", index),
				Status:      "pending",
			}
			err := storage.Create(status)
			require.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Проверяем, что все статусы созданы
	concurrentStatuses, err := storage.ListByCampaignID("concurrent-campaign")
	require.NoError(t, err)
	assert.Len(t, concurrentStatuses, 10)
}

func TestMockBulkCampaignStatusStorage_UpdateNonExistent(t *testing.T) {
	storage := mocks.NewMockBulkCampaignStatusStorage()

	// Тест обновления несуществующего статуса
	now := time.Now().Format(time.RFC3339)
	errMsg := "test error"
	err := storage.Update("non-existent-id", "failed", &errMsg, &now)
	require.NoError(t, err) // Мок не должен возвращать ошибку при обновлении несуществующего

	// Проверяем, что ничего не изменилось
	statuses, err := storage.ListByCampaignID("any-campaign")
	require.NoError(t, err)
	assert.Len(t, statuses, 0)
}
