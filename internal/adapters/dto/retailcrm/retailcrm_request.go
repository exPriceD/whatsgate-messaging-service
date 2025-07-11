package retailcrm

// GetAvailableCategoriesRequest представляет HTTP-запрос на получение доступных категорий
type GetAvailableCategoriesRequest struct {
	// Пока пустой, но может быть расширен в будущем
}

// FilterCustomersByCategoryRequest представляет HTTP-запрос на фильтрацию клиентов по категории
type FilterCustomersByCategoryRequest struct {
	PhoneNumbers         []string `json:"phone_numbers" binding:"required"`
	SelectedCategoryName string   `json:"selected_category_name" binding:"required"`
}

// TestConnectionRequest представляет HTTP-запрос на проверку соединения
type TestConnectionRequest struct {
	// Пока пустой, но может быть расширен в будущем
}
