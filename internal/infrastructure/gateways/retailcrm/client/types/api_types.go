package types

import (
	"fmt"
	"time"
)

// RetailCRMError представляет ошибку от RetailCRM API
type RetailCRMError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// RetailCRMAPIError представляет ошибку API с дополнительной информацией
type RetailCRMAPIError struct {
	StatusCode int
	Code       int
	Message    string
	Details    string
}

// Error возвращает строковое представление ошибки
func (e *RetailCRMAPIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("retailcrm api error (status: %d, code: %d): %s - %s",
			e.StatusCode, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("retailcrm api error (status: %d, code: %d): %s",
		e.StatusCode, e.Code, e.Message)
}

// RetailCRMResponse представляет базовую структуру ответа RetailCRM
type RetailCRMResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ProductGroupResponse представляет ответ с группами товаров
type ProductGroupResponse struct {
	Success bool           `json:"success"`
	Data    []ProductGroup `json:"productGroup,omitempty"`
}

// ProductResponse представляет ответ с товарами
type ProductResponse struct {
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
	Data    []Product `json:"data,omitempty"`
}

// CustomerResponse представляет ответ с клиентом
type CustomerResponse struct {
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Data    Customer `json:"data,omitempty"`
}

// CustomersResponse представляет ответ со списком клиентов
type CustomersResponse struct {
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
	Data    []Customer `json:"data,omitempty"`
}

// OrderResponse представляет ответ с заказом
type OrderResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    Order  `json:"data,omitempty"`
}

// OrdersResponse представляет ответ со списком заказов
type OrdersResponse struct {
	Success bool    `json:"success"`
	Error   string  `json:"error,omitempty"`
	Data    []Order `json:"data,omitempty"`
}

// ProductGroup представляет группу товаров в RetailCRM
type ProductGroup struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Site       string `json:"site"`
	Lvl        int    `json:"lvl"`
	Active     bool   `json:"active"`
	ParentID   int    `json:"parentId,omitempty"`
	ExternalId string `json:"externalId"`
}

// Product представляет товар в RetailCRM
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	GroupID     int       `json:"groupId,omitempty"`
	GroupName   string    `json:"groupName,omitempty"`
	Price       float64   `json:"price"`
	Active      bool      `json:"active"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// Customer представляет клиента в RetailCRM
type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName,omitempty"`
	LastName  string    `json:"lastName,omitempty"`
	Email     string    `json:"email,omitempty"`
	Phones    []Phone   `json:"phones,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// Phone представляет телефон клиента
type Phone struct {
	Number string `json:"number"`
	Type   string `json:"type,omitempty"`
}

// Order представляет заказ в RetailCRM
type Order struct {
	ID         int         `json:"id"`
	Number     string      `json:"number"`
	CustomerID int         `json:"customerId"`
	Status     string      `json:"status"`
	TotalSumm  float64     `json:"totalSumm"`
	Items      []OrderItem `json:"items,omitempty"`
	CreatedAt  time.Time   `json:"createdAt,omitempty"`
	UpdatedAt  time.Time   `json:"updatedAt,omitempty"`
}

// OrderItem представляет товар в заказе
type OrderItem struct {
	ID        int     `json:"id"`
	ProductID int     `json:"productId"`
	Product   Product `json:"product,omitempty"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Sum       float64 `json:"sum"`
}
