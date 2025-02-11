package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Course struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
}

type Order struct {
	ID        string    `json:"id"`
	Course    string    `json:"course"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	UserEmail string    `json:"user_email"`
	UserName  string    `json:"user_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	PaidAt    time.Time `json:"paid_at"`
}

func ListCourse(ctx context.Context) ([]Course, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/courses", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var courses []Course
	err = json.Unmarshal(body, &courses)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func GetCourse(ctx context.Context, course string) (*Course, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/courses/"+course, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var courses Course
	err = json.Unmarshal(body, &courses)
	if err != nil {
		return nil, err
	}
	return &courses, nil
}

func CreateOrder(ctx context.Context, course, userName, userEmail string) (*Order, error) {
	type payload struct {
		Course    string `json:"course"`
		UserName  string `json:"user_name"`
		UserEmail string `json:"user_email"`
	}
	p := payload{
		Course:    course,
		UserName:  userName,
		UserEmail: userEmail,
	}
	pb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8080/orders", bytes.NewReader(pb))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var order Order
	err = json.Unmarshal(body, &order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func GetOrder(ctx context.Context, orderNumber string) (*Order, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/orders/"+orderNumber, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var order Order
	err = json.Unmarshal(body, &order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}
