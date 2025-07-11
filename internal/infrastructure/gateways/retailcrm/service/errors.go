package service

import "errors"

// Ошибки сервисов RetailCRM
var (
	ErrNoProductsFound    = errors.New("no products found")
	ErrNoProductGroups    = errors.New("no product groups found")
	ErrNoOrdersFound      = errors.New("no orders found for phone")
	ErrInvalidOrderStatus = errors.New("invalid order status")
	ErrInvalidProductData = errors.New("invalid product data")
	ErrInvalidOrderData   = errors.New("invalid order data")
	ErrAPIResponseError   = errors.New("api response error")
	ErrPaginationError    = errors.New("pagination error")
)
