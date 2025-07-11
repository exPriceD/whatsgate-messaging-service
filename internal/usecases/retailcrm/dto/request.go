package dto

// GetAvailableCategoriesRequest представляет запрос на получение доступных категорий
type GetAvailableCategoriesRequest struct {
	// Пока пустой, но может быть расширен в будущем
}

// FilterCustomersByCategoryRequest представляет запрос на фильтрацию клиентов по категории
type FilterCustomersByCategoryRequest struct {
	PhoneNumbers         []string // Номера телефонов для фильтрации
	SelectedCategoryName string   // Название выбранной категории
}

// TestConnectionRequest представляет запрос на проверку соединения
type TestConnectionRequest struct {
	// Пока пустой, но может быть расширен в будущем
}
