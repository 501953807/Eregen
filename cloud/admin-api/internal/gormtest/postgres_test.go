package gormtest

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestGORM_Postgres 需要 PostgreSQL 数据库连接
// 设置 DATABASE_URL 环境变量后运行
func TestGORM_Postgres(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set, skipping PostgreSQL test")
	}

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open postgres: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}, &TestOrder{}); err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}

	// 清理测试数据
	db.Exec("DELETE FROM test_orders")
	db.Exec("DELETE FROM test_users")

	// 测试 Create
	user := &TestUser{Name: "Postgres测试", Email: "pg@test.com", Age: 65, IsActive: true, Score: 90.0}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	if user.ID == 0 {
		t.Error("user ID should be auto-generated")
	}
	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should be auto-filled")
	}

	// 测试 Query
	var found TestUser
	if err := db.Where("id = ?", user.ID).First(&found).Error; err != nil {
		t.Fatalf("failed to find user: %v", err)
	}
	if found.Name != "Postgres测试" {
		t.Errorf("expected name 'Postgres测试', got '%s'", found.Name)
	}

	// 测试 Preload
	order := &TestOrder{UserID: uint(user.ID), Product: "测试商品", Amount: 99.99, Status: "pending"}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	var orderWithUser TestOrder
	if err := db.Preload("User").First(&orderWithUser, order.ID).Error; err != nil {
		t.Fatalf("failed to find order with preload: %v", err)
	}
	if orderWithUser.User.Name != "Postgres测试" {
		t.Errorf("expected user name 'Postgres测试', got '%s'", orderWithUser.User.Name)
	}

	// 清理
	db.Exec("DELETE FROM test_orders")
	db.Exec("DELETE FROM test_users")

	t.Logf("PostgreSQL POC 验证通过: ID=%d, Name=%s", user.ID, user.Name)
}
