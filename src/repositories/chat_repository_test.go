package repositories

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/nvnamsss/chat/src/adapters"
	"github.com/nvnamsss/chat/src/configs"
	"github.com/nvnamsss/chat/src/dtos"
	"github.com/nvnamsss/chat/src/logger"
	"github.com/nvnamsss/chat/src/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB adapters.DBAdapter
)

func TestMain(m *testing.M) {
	var err error
	// Initialize logger for tests
	logger.Init("debug", "test")
	sport := getEnvOrDefault("TEST_DB_PORT", "5432")
	port, err := strconv.Atoi(sport)
	if err != nil {
		panic("Invalid port number: " + err.Error())
	}
	// Set up test database connection
	config := configs.Database{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     port,
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		Name:     getEnvOrDefault("TEST_DB_NAME", "chat_test"),
		SSLMode:  "disable",
	}

	testDB, err = adapters.NewDBAdapter(config)
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	// Run migrations for test database
	err = testDB.AutoMigrate(&models.Chat{}, &models.Message{})
	if err != nil {
		panic("Failed to run migrations: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	sqlDB, _ := testDB.GetDB().DB()
	sqlDB.Close()

	os.Exit(code)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupTest(t *testing.T) (ChatRepository, func()) {
	// Create a new repository instance
	repo := NewChatRepository(testDB)

	// Create cleanup function
	cleanup := func() {
		// Clean up test data
		testDB.GetDB().Exec("DELETE FROM messages")
		testDB.GetDB().Exec("DELETE FROM chats")
	}

	// Run cleanup before test
	cleanup()

	return repo, cleanup
}

func createTestChat(t *testing.T, repo ChatRepository, userID string, title string) *models.Chat {
	chat := &models.Chat{
		UserID: userID,
		Title:  title,
	}
	err := repo.Create(context.Background(), chat)
	require.NoError(t, err)
	require.NotZero(t, chat.ID)
	return chat
}

func TestChatRepository_Create(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	t.Run("successful creation", func(t *testing.T) {
		chat := &models.Chat{
			UserID: "user1",
			Title:  "Test Chat",
		}

		err := repo.Create(context.Background(), chat)
		require.NoError(t, err)
		assert.NotZero(t, chat.ID)
		assert.NotZero(t, chat.CreatedAt)
		assert.NotZero(t, chat.UpdatedAt)
	})

}

func TestChatRepository_Get(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	t.Run("successful retrieval", func(t *testing.T) {
		// Create test chat
		original := createTestChat(t, repo, "user1", "Test Chat")

		// Retrieve chat
		retrieved, err := repo.Get(context.Background(), original.ID)
		require.NoError(t, err)
		assert.Equal(t, original.ID, retrieved.ID)
		assert.Equal(t, original.UserID, retrieved.UserID)
		assert.Equal(t, original.Title, retrieved.Title)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.Get(context.Background(), 99999)
		require.Error(t, err)
	})
}

func TestChatRepository_GetByUserID(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	t.Run("successful retrieval with pagination", func(t *testing.T) {
		// Create test chats
		userID := "user1"
		for i := 1; i <= 5; i++ {
			createTestChat(t, repo, userID, "Test Chat")
		}

		// Test pagination
		chats, total, err := repo.GetByUserID(context.Background(), userID, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, chats, 2)
	})

	t.Run("empty result", func(t *testing.T) {
		chats, total, err := repo.GetByUserID(context.Background(), "nonexistent", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, chats)
	})
}

func TestChatRepository_Search(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	t.Run("successful search", func(t *testing.T) {
		// Create test chats
		userID := "user1"
		createTestChat(t, repo, userID, "AI Chat")
		createTestChat(t, repo, userID, "ML Discussion")
		createTestChat(t, repo, userID, "Another AI Chat")

		// Search for chats
		req := &dtos.SearchChatsRequest{
			Query:  "AI",
			Limit:  10,
			Offset: 0,
		}

		chats, total, err := repo.Search(context.Background(), req, userID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, chats, 2)
	})

	t.Run("no results", func(t *testing.T) {
		req := &dtos.SearchChatsRequest{
			Query:  "NonexistentTerm",
			Limit:  10,
			Offset: 0,
		}

		chats, total, err := repo.Search(context.Background(), req, "user1")
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, chats)
	})
}

func TestChatRepository_Update(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	t.Run("successful update", func(t *testing.T) {
		// Create test chat
		chat := createTestChat(t, repo, "user1", "Original Title")

		// Update the chat
		chat.Title = "Updated Title"
		time.Sleep(time.Millisecond) // Ensure updated_at will be different
		err := repo.Update(context.Background(), chat)
		require.NoError(t, err)

		// Verify update
		updated, err := repo.Get(context.Background(), chat.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", updated.Title)
		assert.True(t, updated.UpdatedAt.After(updated.CreatedAt))
	})

	t.Run("update nonexistent", func(t *testing.T) {
		chat := &models.Chat{
			ID:     99999,
			UserID: "user1",
			Title:  "Updated Title",
		}
		err := repo.Update(context.Background(), chat)
		require.Error(t, err)
	})
}

func TestChatRepository_Delete(t *testing.T) {
	repo, cleanup := setupTest(t)
	defer cleanup()

	t.Run("successful deletion", func(t *testing.T) {
		// Create test chat
		chat := createTestChat(t, repo, "user1", "Test Chat")

		// Delete the chat
		err := repo.Delete(context.Background(), chat.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.Get(context.Background(), chat.ID)
		require.Error(t, err)
	})

	t.Run("delete nonexistent", func(t *testing.T) {
		err := repo.Delete(context.Background(), 99999)
		require.Error(t, err)
	})
}
