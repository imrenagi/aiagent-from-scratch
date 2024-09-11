package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

var coursesDB = map[string]Course{
	"software-security": {
		Name:        "software-security",
		DisplayName: "Software Security",
		Description: "Learn how to secure your software",
		Price:       100.0,
		Currency:    "USD",
	},
}

var ordersDB = map[string]Order{}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, code int, message string) {
	errorResponse := Error{
		Message: message,
		Code:    code,
	}
	jsonError, _ := json.Marshal(errorResponse)
	w.WriteHeader(code)
	w.Write(jsonError)
}

func ListCoursesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	courses := make([]Course, 0, len(coursesDB))
	for _, course := range coursesDB {
		courses = append(courses, course)
	}

	jsonResponse, err := json.Marshal(courses)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error encoding JSON")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func GetCourseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	courseName := vars["course"]

	course, ok := coursesDB[courseName]
	if !ok {
		writeError(w, http.StatusNotFound, "Course not found")
		return
	}

	jsonResponse, err := json.Marshal(course)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error encoding JSON")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Define Order struct
	type CreateOrderRequest struct {
		Course    string `json:"course"`
		UserName  string `json:"user_name"`
		UserEmail string `json:"user_email"`
	}

	// Parse request body
	var newOrder CreateOrderRequest
	err := json.NewDecoder(r.Body).Decode(&newOrder)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if the course is valid
	course, ok := coursesDB[newOrder.Course]
	if !ok {
		writeError(w, http.StatusNotFound, "Course not found")
		return
	}

	order := Order{
		ID:        uuid.New().String(),
		Course:    newOrder.Course,
		UserName:  newOrder.UserName,
		UserEmail: newOrder.UserEmail,
		Status:    "pending",
		CreatedAt: time.Now(),
		Price:     course.Price,
		Currency:  course.Currency,
	}

	// Add new order to the map (assuming ordersDB is defined elsewhere)
	ordersDB[order.ID] = order

	// Create response with payment page URL
	type CreateOrderResponse struct {
		OrderID string `json:"order_id"`
	}

	response := CreateOrderResponse{
		OrderID: order.ID,
	}

	// Send response
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error encoding JSON")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orderID := vars["order"]

	// Get order from database
	order, ok := ordersDB[orderID]
	if !ok {
		writeError(w, http.StatusNotFound, "Order not found")
		return
	}

	// Prepare JSON response
	jsonResponse, err := json.Marshal(order)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error encoding JSON")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func PayOrderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orderID := vars["order"]

	// Get order from database
	order, ok := ordersDB[orderID]
	if !ok {
		writeError(w, http.StatusNotFound, "Order not found")
		return
	}

	// Update order status
	order.Status = "paid"
	order.PaidAt = time.Now()

	// Update order in database
	ordersDB[orderID] = order

	// Prepare JSON response
	jsonResponse, err := json.Marshal(order)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error encoding JSON")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func OrderPaymentPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order"]

	// Get order from database
	order, ok := ordersDB[orderID]
	if !ok {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Create HTML content
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pay for Order %s</title>
</head>
<body>
    <h1>Pay for Order %s</h1>
    <p>Total Amount: $%.2f</p>
    <form action="/orders/%s:pay" method="POST">
        <input type="submit" value="Pay Now">
    </form>
</body>
</html>
`, orderID, orderID, order.Price, orderID)

	// Set content type and write HTML
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/courses", ListCoursesHandler).Methods("GET")
	r.HandleFunc("/courses/{course}", GetCourseHandler).Methods("GET")
	r.HandleFunc("/orders", CreateOrderHandler).Methods("POST")
	r.HandleFunc("/orders/{order}", GetOrderHandler).Methods("GET")
	r.HandleFunc("/orders/{order}/payment", OrderPaymentPageHandler).Methods("GET")
	r.HandleFunc("/orders/{order}:pay", PayOrderHandler).Methods("POST")

	http.Handle("/", r)

	fmt.Println("Server is starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
