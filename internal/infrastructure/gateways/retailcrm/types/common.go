package types

// ProductShort содержит только id и name товара
type ProductShort struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ProductGroup содержит информацию о группе товаров
type ProductGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// OrderShort содержит только основные поля заказа для парсинга
type OrderShort struct {
	Status string           `json:"status"`
	Items  []OrderItemShort `json:"items"`
}

// OrderItemShort содержит только основные поля товара в заказе
type OrderItemShort struct {
	Offer struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"offer"`
}
