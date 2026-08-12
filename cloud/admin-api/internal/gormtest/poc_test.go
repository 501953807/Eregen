package gormtest

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUser 测试用 User 模型
type TestUser struct {
	gorm.Model
	Name     string  `gorm:"size:100;not null"`
	Email    string  `gorm:"uniqueIndex;size:200"`
	Age      int
	IsActive bool
	Score    float64
}

// TestOrder 测试用 Order 模型（关联查询）
type TestOrder struct {
	gorm.Model
	UserID  uint    `gorm:"index"`
	Product string  `gorm:"size:200"`
	Amount  float64
	Status  string  `gorm:"size:50"`

	User TestUser `gorm:"foreignKey:UserID"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}, &TestOrder{}); err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}

	return db
}

func TestGORM_Sqlite_CreateAndQuery(t *testing.T) {
	db := setupTestDB(t)

	user := &TestUser{Name: "张三", Email: "zhangsan@example.com", Age: 65, IsActive: true, Score: 95.5}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	var found TestUser
	if err := db.Where("id = ?", user.ID).First(&found).Error; err != nil {
		t.Fatalf("failed to find user: %v", err)
	}

	if found.Name != "张三" {
		t.Errorf("expected name '张三', got '%s'", found.Name)
	}
	if found.Score != 95.5 {
		t.Errorf("expected score 95.5, got %f", found.Score)
	}
}

func TestGORM_Sqlite_AutoTimestamps(t *testing.T) {
	db := setupTestDB(t)

	user := &TestUser{Name: "李四", Email: "lisi@example.com", Age: 70}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should be auto-filled")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be auto-filled")
	}
}

func TestGORM_Sqlite_Update(t *testing.T) {
	db := setupTestDB(t)

	user := &TestUser{Name: "王五", Email: "wangwu@example.com", Age: 60}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	oldUpdatedAt := user.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	if err := db.Model(user).Update("score", 88.0).Error; err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	var updated TestUser
	db.First(&updated, user.ID)

	if updated.Score != 88.0 {
		t.Errorf("expected score 88.0, got %f", updated.Score)
	}
	if updated.UpdatedAt.Equal(oldUpdatedAt) {
		t.Error("UpdatedAt should change after update")
	}
}

func TestGORM_Sqlite_Preload(t *testing.T) {
	db := setupTestDB(t)

	user := &TestUser{Name: "赵六", Email: "zhaoliu@example.com", Age: 55}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	order := &TestOrder{UserID: uint(user.ID), Product: "降压药", Amount: 299.99, Status: "pending"}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	var orderWithUser TestOrder
	if err := db.Preload("User").First(&orderWithUser, order.ID).Error; err != nil {
		t.Fatalf("failed to find order with preload: %v", err)
	}

	if orderWithUser.User.Name != "赵六" {
		t.Errorf("expected user name '赵六', got '%s'", orderWithUser.User.Name)
	}
}

func TestGORM_Sqlite_BatchCreate(t *testing.T) {
	db := setupTestDB(t)

	orders := []*TestOrder{
		{UserID: 1, Product: "血糖仪", Amount: 199.0, Status: "completed"},
		{UserID: 1, Product: "血压计", Amount: 259.0, Status: "pending"},
		{UserID: 1, Product: "体温计", Amount: 39.9, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("failed to batch create orders: %v", err)
	}

	var count int64
	db.Model(&TestOrder{}).Count(&count)
	if count != 3 {
		t.Errorf("expected 3 orders, got %d", count)
	}
}

func TestGORM_Sqlite_SoftDelete(t *testing.T) {
	db := setupTestDB(t)

	user := &TestUser{Name: "孙七", Email: "sunqi@example.com", Age: 72}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := db.Delete(user).Error; err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	var deleted TestUser
	if err := db.First(&deleted, user.ID).Error; err == nil {
		t.Error("soft delete should hide record from normal query")
	}

	// Unscoped 应该能看到
	var unscoped TestUser
	if err := db.Unscoped().First(&unscoped, user.ID).Error; err != nil {
		t.Errorf("unscoped query should find soft-deleted record: %v", err)
	}
}

func TestGORM_Sqlite_DateFilter(t *testing.T) {
	db := setupTestDB(t)

	user := &TestUser{Name: "周八", Email: "zhouba@example.com", Age: 68}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var recentUsers []TestUser
	if err := db.Where("created_at > ?", sevenDaysAgo).Find(&recentUsers).Error; err != nil {
		t.Fatalf("failed to query with date filter: %v", err)
	}

	if len(recentUsers) != 1 {
		t.Errorf("expected 1 recent user, got %d", len(recentUsers))
	}
}

func TestGORM_Sqlite_MaxOpenConns1(t *testing.T) {
	// 验证 GORM 能正确处理 SQLite 文件锁
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}

	user := &TestUser{Name: "吴九", Email: "wujiu@example.com", Age: 63}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user with MaxOpenConns=1: %v", err)
	}
}
