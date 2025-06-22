package http

// setupRoutes настраивает маршруты сервера.
func (s *Server) setupRoutes() {
	// Health check
	s.engine.GET(HealthPath, s.healthHandler)

	// API v1
	v1 := s.engine.Group("/api/" + APIVersion)
	{
		v1.GET(StatusPath, s.statusHandler)
		// Здесь будут добавлены другие эндпоинты
	}
}
