// Утилиты для работы с API

/**
 * Обрабатывает ошибки API и возвращает понятное сообщение об ошибке
 * @param {Response} response - Объект ответа fetch
 * @returns {Promise<string>} Сообщение об ошибке
 */
export async function handleApiError(response) {
  if (response.ok) return null;
  
  try {
    const errorData = await response.json();
    // Проверяем новую структуру ошибки AppErrorResponse
    if (errorData.type && errorData.message) {
      return errorData.message;
    }
    // Fallback для старой структуры
    return errorData.error || errorData.message || 'Неизвестная ошибка';
  } catch (e) {
    return `HTTP ${response.status}: ${response.statusText}`;
  }
}

/**
 * Выполняет fetch запрос с обработкой ошибок
 * @param {string} url - URL для запроса
 * @param {Object} options - Опции fetch
 * @param {Function} showToast - Функция для показа уведомлений
 * @returns {Promise<Object>} Результат запроса
 */
export async function apiRequest(url, options = {}, showToast = null) {
  try {
    const response = await fetch(url, options);
    
    if (response.ok) {
      return await response.json();
    } else {
      const errorMessage = await handleApiError(response);
      if (showToast) {
        showToast(errorMessage || 'Ошибка запроса', 'danger');
      }
      throw new Error(errorMessage || 'Ошибка запроса');
    }
  } catch (error) {
    console.error('API request error:', error);
    if (showToast && error.message !== 'Ошибка запроса') {
      showToast(error.message || 'Ошибка запроса', 'danger');
    }
    throw error;
  }
}

/**
 * Выполняет GET запрос
 * @param {string} url - URL для запроса
 * @param {Function} showToast - Функция для показа уведомлений
 * @returns {Promise<Object>} Результат запроса
 */
export async function apiGet(url, showToast = null) {
  return apiRequest(url, { method: 'GET' }, showToast);
}

/**
 * Выполняет POST запрос
 * @param {string} url - URL для запроса
 * @param {Object|FormData} data - Данные для отправки
 * @param {Function} showToast - Функция для показа уведомлений
 * @returns {Promise<Object>} Результат запроса
 */
export async function apiPost(url, data, showToast = null) {
  const options = { method: 'POST' };
  
  if (data instanceof FormData) {
    options.body = data;
  } else {
    options.headers = { 'Content-Type': 'application/json' };
    options.body = JSON.stringify(data);
  }
  
  return apiRequest(url, options, showToast);
}

/**
 * Выполняет PUT запрос
 * @param {string} url - URL для запроса
 * @param {Object} data - Данные для отправки
 * @param {Function} showToast - Функция для показа уведомлений
 * @returns {Promise<Object>} Результат запроса
 */
export async function apiPut(url, data, showToast = null) {
  return apiRequest(url, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  }, showToast);
}

/**
 * Выполняет DELETE запрос
 * @param {string} url - URL для запроса
 * @param {Function} showToast - Функция для показа уведомлений
 * @returns {Promise<Object>} Результат запроса
 */
export async function apiDelete(url, showToast = null) {
  return apiRequest(url, { method: 'DELETE' }, showToast);
} 