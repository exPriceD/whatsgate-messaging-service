package campaign

// CreateCampaignRequest представляет HTTP-запрос на создание кампании
type CreateCampaignRequest struct {
	Name                 string   `json:"name" form:"name" binding:"required"`
	Message              string   `json:"message" form:"message" binding:"required"`
	AdditionalPhones     []string `json:"additional_phones" form:"additional_phones"`
	ExcludePhones        []string `json:"exclude_phones" form:"exclude_phones"`
	MessagesPerHour      int      `json:"messages_per_hour" form:"messages_per_hour"`
	Initiator            string   `json:"initiator" form:"initiator"`
	SelectedCategoryName string   `json:"selected_category_name" form:"selected_category_name"`
}
