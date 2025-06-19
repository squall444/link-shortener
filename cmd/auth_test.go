package main

import (
	"bytes"
	"encoding/json"
	"goadv/internal/auth"
	"goadv/internal/user"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDb() *gorm.DB {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("DSN")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func initData(db *gorm.DB) {
	db.Create(&user.User{
		Email:    "a@a.ru",
		Password: "$2a$10$kikxIRYaI/OITTmN3sWa0OsGnTpMPImlg.Ir6Uf/3CKxzM9vML57q",
		Name:     "Vasya",
	})
}

func removeData(db *gorm.DB) {
	db.Unscoped().
		Where("email = ?", "a@a.ru").
		Delete(&user.User{})
}

func TestLoginSuccess(t *testing.T) {
	db := initDb()
	initData(db)

	app, _ := App()
	ts := httptest.NewServer(app)
	defer ts.Close()

	data, _ := json.Marshal(&auth.LoginRequest{
		Email:    "a@a.ru",
		Password: "12345",
	})
	res, err := http.Post(ts.URL+"/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d Got %d", http.StatusOK, res.StatusCode)
	}
	var resData auth.LoginResponse
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		t.Fatal(err)
	}
	if resData.Token == "" {
		t.Fatalf("Token empty")
	}
	removeData(db)
}

func TestLoginFail(t *testing.T) {
	db := initDb()
	initData(db)

	app, _ := App()
	ts := httptest.NewServer(app)
	defer ts.Close()

	data, _ := json.Marshal(&auth.LoginRequest{
		Email:    "a@a.ru",
		Password: "1",
	})
	res, err := http.Post(ts.URL+"/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Expected %d Got %d", http.StatusUnauthorized, res.StatusCode)
	}
	removeData(db)
}
