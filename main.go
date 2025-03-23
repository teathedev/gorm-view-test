package main

import (
	"fmt"
	"time"

	"golang.org/x/exp/rand"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Define the User model
type User struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255"`
	Age  int
}

// Define the Product model
type Product struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"size:255"`
	Price  float64
	UserID uint // Foreign key to User
}

// Define the UserProductView view (read-only)
type RandomStruct struct {
	UserName    string  `gorm:"column:user_name"`
	ProductID   uint    `gorm:"column:product_id"`
	ProductName string  `gorm:"column:product_name"`
	Price       float64 `gorm:"column:price"`
}

func (RandomStruct) TableName() string {
	return "user_product_views"
}

func main() {
	// Database connection string
	dsn := "host=localhost user=gorm_user password=gorm_password dbname=gorm_view_test_db port=5432 sslmode=disable TimeZone=UTC"

	// Connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	// Auto migrate the tables
	err = db.AutoMigrate(&User{}, &Product{})
	if err != nil {
		panic("failed to migrate tables")
	}

	// Define the query for the view using GORM's query builder
	query := db.Model(&User{}).
		Select("users.name as user_name, products.id as product_id, products.name as product_name, products.price").
		Joins("JOIN products ON users.id = products.user_id")

	// Create the view using Migrator().CreateView
	err = db.Migrator().CreateView("user_product_views", gorm.ViewOption{
		Replace: true, // Replace the view if it already exists
		Query:   query,
	})
	if err != nil {
		panic("failed to create view")
	}

	// Insert 10 random users
	rand.Seed(uint64(time.Now().UnixNano()))
	for i := 1; i <= 10; i++ {
		user := User{
			Name: fmt.Sprintf("User%d", i),
			Age:  rand.Intn(50) + 18, // Random age between 18 and 67
		}
		db.Create(&user)

		// Insert 10 random products for each user
		for j := 1; j <= 10; j++ {
			product := Product{
				Name:   fmt.Sprintf("Product%d-%d", i, j),
				Price:  rand.Float64() * 100, // Random price between 0 and 100
				UserID: user.ID,
			}
			db.Create(&product)
		}
	}

	fmt.Println("Database setup complete! 10 users and 10 products per user inserted.")

	row := RandomStruct{}
	if err := db.First(&row).Error; err != nil {
		fmt.Println("Failed to fetch view!")
		fmt.Println(err)
		return
	}

	fmt.Printf("Fetched Row: %s", row.ProductName)
}
