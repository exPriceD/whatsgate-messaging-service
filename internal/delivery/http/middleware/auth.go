package middleware

import (
	"net/http"
)

// Auth добавляет базовую аутентификацию (заготовка для будущего)
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Добавить логику аутентификации
		// Пока что пропускаем все запросы
		next.ServeHTTP(w, r)
	})
}

// RequireAuth проверяет авторизацию для защищенных эндпоинтов
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Проверка токена авторизации
		// Пример проверки заголовка Authorization

		// authHeader := r.Header.Get("Authorization")
		// if authHeader == "" {
		//     response.WriteError(w, http.StatusUnauthorized, "Authorization required")
		//     return
		// }

		// Пока что разрешаем все запросы
		next.ServeHTTP(w, r)
	})
}

// AdminOnly проверяет права администратора (заготовка)
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Проверка прав администратора

		// Пример отклонения без прав админа
		// response.WriteError(w, http.StatusForbidden, "Admin access required")
		// return

		next.ServeHTTP(w, r)
	})
}
