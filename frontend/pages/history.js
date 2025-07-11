import { apiGet, apiPost, apiGetCampaignErrors } from '../ui/api.js';

// Страница истории рассылок
export function renderHistoryPage() {
  return `
    <div class="history-header">
      <h2>📊 История рассылок</h2>
      <p class="history-subtitle">Управление и мониторинг массовых рассылок</p>
    </div>
    
    <div class="history-stats">
      <div class="stat-card">
        <div class="stat-icon">📈</div>
        <div class="stat-content">
          <div class="stat-number" id="total-campaigns">-</div>
          <div class="stat-label">Всего рассылок</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">✅</div>
        <div class="stat-content">
          <div class="stat-number" id="completed-campaigns">-</div>
          <div class="stat-label">Завершено</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">🔄</div>
        <div class="stat-content">
          <div class="stat-number" id="active-campaigns">-</div>
          <div class="stat-label">Активные</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">❌</div>
        <div class="stat-content">
          <div class="stat-number" id="failed-campaigns">-</div>
          <div class="stat-label">Ошибки</div>
        </div>
      </div>
    </div>

    <div class="history-container">
      <div class="history-controls">
        <div class="search-box">
          <input type="text" id="search-campaigns" placeholder="🔍 Поиск по названию..." />
        </div>
        <div class="filter-controls">
          <select id="status-filter">
            <option value="">Все статусы</option>
            <option value="started">Запущена</option>
            <option value="finished">Завершена</option>
            <option value="failed">Ошибка</option>
            <option value="pending">Ожидает</option>
          </select>
          <button id="refresh-history" class="refresh-btn">
            <span class="refresh-icon">🔄</span>
            Обновить
          </button>
        </div>
      </div>
      
      <div class="history-table-container">
        <table id="history-table" class="history-table">
          <thead>
            <tr>
              <th>📝 Название</th>
              <th>📊 Статус</th>
              <th>📈 Прогресс</th>
              <th>⏱️ Сообщ./час</th>
              <th>🏷️ Категория</th>
              <th>❗ Ошибки</th>
              <th>📅 Дата</th>
              <th>🔧 Действия</th>
            </tr>
          </thead>
          <tbody id="history-tbody">
            <tr>
              <td colspan="8" class="loading">
                <div class="loading-spinner"></div>
                <span>Загрузка истории...</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    
    <!-- Модальное окно с деталями -->
    <div id="campaign-modal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h3 id="modal-title">📋 Детали рассылки</h3>
          <span class="close">&times;</span>
        </div>
        <div id="modal-body">
          <!-- Контент будет загружен динамически -->
        </div>
      </div>
    </div>
  `;
}

export function initHistoryPage(showToast) {
  const refreshBtn = document.getElementById('refresh-history');
  const searchInput = document.getElementById('search-campaigns');
  const statusFilter = document.getElementById('status-filter');
  const modal = document.getElementById('campaign-modal');
  const closeBtn = document.querySelector('.close');
  const modalTitle = document.getElementById('modal-title');
  const modalBody = document.getElementById('modal-body');

  let allCampaigns = []; // Храним все кампании для фильтрации

  // Загрузка истории при инициализации
  loadHistory();

  // Обработчики событий
  refreshBtn.onclick = loadHistory;
  searchInput.oninput = filterCampaigns;
  statusFilter.onchange = filterCampaigns;
  closeBtn.onclick = () => modal.style.display = 'none';
  window.onclick = (event) => {
    if (event.target === modal) {
      modal.style.display = 'none';
    }
  };

  async function loadHistory() {
    const tbody = document.getElementById('history-tbody');
    tbody.innerHTML = `
      <tr>
        <td colspan="8" class="loading">
          <div class="loading-spinner"></div>
          <span>Загрузка истории...</span>
        </td>
      </tr>
    `;

    // Добавляем анимацию к кнопке обновления
    refreshBtn.classList.add('refreshing');
    refreshBtn.disabled = true;

    try {
      // Используем наш новый List API с пагинацией
      const params = new URLSearchParams({
        limit: '500',
        offset: '0'
      });
      
      // Добавляем фильтр по статусу если выбран
      const statusFilterValue = statusFilter.value;
      if (statusFilterValue) {
        params.append('status', statusFilterValue);
      }
      
      const response = await apiGet(`/api/v1/campaigns?${params}`, showToast);
      console.log('API response:', response);
      
      // Наш новый API возвращает данные в формате {campaigns: [...], total: N, limit: N, offset: N}
      const campaigns = response.campaigns || [];
      const total = response.total || 0;
      
      console.log(`Loaded ${campaigns.length} campaigns of ${total} total`);
      
      allCampaigns = campaigns;
      updateStats(campaigns, total);
      renderCampaigns(campaigns);
      
    } catch (error) {
      console.error('Error loading history:', error);
      showErrorState();
      // Ошибка уже обработана в apiGet
    } finally {
      refreshBtn.classList.remove('refreshing');
      refreshBtn.disabled = false;
    }
  }

  function updateStats(campaigns, total = null) {
    const displayTotal = total !== null ? total : campaigns.length;
    const completed = campaigns.filter(c => c.status === 'finished').length;
    const active = campaigns.filter(c => c.status === 'started').length;
    const failed = campaigns.filter(c => c.status === 'failed').length;

    document.getElementById('total-campaigns').textContent = displayTotal;
    document.getElementById('completed-campaigns').textContent = completed;
    document.getElementById('active-campaigns').textContent = active;
    document.getElementById('failed-campaigns').textContent = failed;
  }

  function filterCampaigns() {
    const searchTerm = searchInput.value.toLowerCase();
    const statusFilterValue = statusFilter.value;

    const filtered = allCampaigns.filter(campaign => {
      const matchesSearch = !searchTerm || 
        (campaign.name && campaign.name.toLowerCase().includes(searchTerm)) ||
        campaign.message.toLowerCase().includes(searchTerm);
      
      const matchesStatus = !statusFilterValue || campaign.status === statusFilterValue;
      
      return matchesSearch && matchesStatus;
    });

    renderCampaigns(filtered);
  }

  function renderCampaigns(campaigns) {
    const tbody = document.getElementById('history-tbody');
    
    if (campaigns.length === 0) {
      showEmptyState();
      return;
    }

    tbody.innerHTML = campaigns.map(campaign => `
      <tr data-id="${campaign.id}" class="campaign-row">
        <td class="campaign-name">
          <div class="name-content">
            <div class="name-text">${campaign.name || 'Без названия'}</div>
          </div>
        </td>
        <td>
          <span class="status status-${campaign.status}">
            ${getStatusIcon(campaign.status)} ${getStatusText(campaign.status)}
          </span>
        </td>
        <td class="campaign-progress">
          <div class="progress-info">
            <div class="progress-numbers">
              <span class="processed-number">${campaign.processed_count || 0}</span>
              <span class="separator">/</span>
              <span class="total-number">${campaign.total_count || 0}</span>
            </div>
            <div class="progress-bar">
              <div class="progress-fill" style="width: ${getProgressPercentage(campaign.processed_count || 0, campaign.total_count || 0)}%"></div>
            </div>
            <div class="progress-percentage">${getProgressPercentage(campaign.processed_count || 0, campaign.total_count || 0)}%</div>
          </div>
        </td>
        <td class="campaign-speed">
          <span class="speed-number">${campaign.messages_per_hour || 0}</span>
          <span class="speed-label">/час</span>
        </td>
        <td class="campaign-category">
          ${campaign.category_name ? `<span class="category-tag">${campaign.category_name}</span>` : '<span class="no-category">—</span>'}
        </td>
        <td class="campaign-errors">
          <span class="error-count">${campaign.error_count || 0}</span>
        </td>
        <td class="campaign-date">
          <div class="date-content">
            <div class="date-main">${formatDate(campaign.created_at)}</div>
            <div class="date-relative">${getRelativeTime(campaign.created_at)}</div>
          </div>
        </td>
        <td class="campaign-actions">
          <button class="details-btn" data-id="${campaign.id}" title="Просмотреть детали">
            👁️ Детали
          </button>
        </td>
      </tr>
    `).join('');

    // Добавляем обработчики для кнопок деталей
    document.querySelectorAll('.details-btn').forEach(btn => {
      btn.onclick = (e) => {
        e.preventDefault();
        const campaignId = btn.dataset.id;
        showCampaignDetails(campaignId);
      };
    });
  }

  function showEmptyState() {
    const tbody = document.getElementById('history-tbody');
    tbody.innerHTML = `
      <tr>
        <td colspan="8" class="empty">
          <div class="empty-state">
            <div class="empty-icon">📭</div>
            <div class="empty-text">История рассылок пуста</div>
            <div class="empty-subtext">Создайте первую рассылку на странице "Массовая рассылка"</div>
          </div>
        </td>
      </tr>
    `;
  }

  function showErrorState() {
    const tbody = document.getElementById('history-tbody');
    tbody.innerHTML = `
      <tr>
        <td colspan="8" class="error">
          <div class="error-state">
            <div class="error-icon">❌</div>
            <div class="error-text">Ошибка загрузки истории</div>
            <button class="retry-btn" onclick="loadHistory()">Повторить</button>
          </div>
        </td>
      </tr>
    `;
  }

  async function showCampaignDetails(campaignId) {
    modal.style.display = 'block';
    modalTitle.textContent = 'Загрузка...';
    modalBody.innerHTML = '<div class="loading">Загрузка деталей...</div>';

    try {
      // Используем наш новый GetByID API
      const campaign = await apiGet(`/api/v1/campaigns/${campaignId}`, showToast);
      
      modalTitle.textContent = campaign.name || 'Детали рассылки';
      modalBody.innerHTML = `
        <div class="campaign-details">
          <div class="detail-section">
            <h4>📋 Основная информация</h4>
            <div class="detail-grid">
              <div class="detail-item">
                <label>ID Рассылки:</label>
                <span class="detail-value">${campaign.id}</span>
              </div>
              <div class="detail-item">
                <label>Название:</label>
                <span class="detail-value">${campaign.name || 'Не указано'}</span>
              </div>
              <div class="detail-item">
                <label>Статус:</label>
                <span class="status status-${campaign.status}">${getStatusIcon(campaign.status)} ${getStatusText(campaign.status)}</span>
              </div>
              <div class="detail-item">
                <label>Дата создания:</label>
                <span class="detail-value">${formatDate(campaign.created_at)}</span>
              </div>
              <div class="detail-item">
                <label>Категория:</label>
                <span class="detail-value">${campaign.category_name ? `<span class="category-tag">${campaign.category_name}</span>` : 'Без фильтрации'}</span>
              </div>
            </div>
          </div>
          
          <div class="detail-section">
            <h4>📊 Статистика отправки</h4>
            <div class="detail-grid">
              <div class="detail-item">
                <label>Корректных отправок:</label>
                <div class="progress-detail">
                  <div class="progress-numbers">
                    <span class="processed-number">${campaign.processed_count || 0}</span>
                    <span class="separator">/</span>
                    <span class="total-number">${campaign.total_count || 0}</span>
                    <span class="progress-percentage">(${getProgressPercentage(campaign.processed_count || 0, campaign.total_count || 0)}%)</span>
                  </div>
                  <div class="progress-bar">
                    <div class="progress-fill" style="width: ${getProgressPercentage(campaign.processed_count || 0, campaign.total_count || 0)}%"></div>
                  </div>
                </div>
              </div>
              <div class="detail-item">
                <label>Отправлено успешно:</label>
                <span class="detail-value success">${campaign.sent_numbers ? campaign.sent_numbers.filter(n => n.status === 'sent').length : 0}</span>
              </div>
              <div class="detail-item">
                <label>Ошибки отправки:</label>
                <span class="detail-value numbers-error">${campaign.failed_numbers ? campaign.failed_numbers.filter(n => n.status === 'failed').length : 0}</span>
              </div>
              <div class="detail-item">
                <label>Скорость отправки:</label>
                <span class="detail-value">${campaign.messages_per_hour || 0} сообщ./час</span>
              </div>
            </div>
          </div>
          
          <div class="detail-section">
            <h4>💬 Сообщение</h4>
            <div class="message-preview">${campaign.message}</div>
          </div>
          
          ${campaign.media ? `
          <div class="detail-section">
            <h4>📎 Медиа файл</h4>
            <div class="media-info">
              <div class="media-item">
                <span class="media-icon">📎</span>
                <div class="media-details">
                  <div class="media-name">${campaign.media.filename}</div>
                  <div class="media-type">${campaign.media.message_type} • ${campaign.media.mime_type}</div>
                </div>
              </div>
            </div>
          </div>
          ` : ''}
          
          ${campaign.sent_numbers && campaign.sent_numbers.length > 0 ? `
          <div class="detail-section">
            <h4>✅ Успешно отправлено (${campaign.sent_numbers.filter(n => n.status === 'sent').length})</h4>
            <div class="phone-numbers-container">
              <div class="phone-numbers-header">
                <span class="phone-numbers-label">Номера с успешной отправкой:</span>
              </div>
              <div class="phone-numbers-list">
                ${campaign.sent_numbers.filter(n => n.status === 'sent').slice(0, 50).map(number => `
                  <div class="phone-number-item success">
                    <span class="phone-number">${number.phone_number}</span>
                    <span class="phone-time">${formatDate(number.sent_at)}</span>
                  </div>
                `).join('')}
                ${campaign.sent_numbers.filter(n => n.status === 'sent').length > 50 ? `
                  <div class="phone-numbers-more">
                    ... и еще ${campaign.sent_numbers.filter(n => n.status === 'sent').length - 50} номеров
                  </div>
                ` : ''}
              </div>
              ${campaign.sent_numbers.filter(n => n.status === 'sent').length > 0 ? `
              <div class="phone-numbers-textarea-container">
                <div class="textarea-header">
                  <label class="phone-numbers-textarea-label">Все успешно отправленные номера (${campaign.sent_numbers.filter(n => n.status === 'sent').length}):</label>
                  <button class="copy-textarea-btn" onclick="copySuccessfulNumbers('${campaign.id}')" title="Копировать все успешно отправленные номера">
                    <span class="copy-btn-text">📋 Копировать номера</span>
                  </button>
                </div>
                <textarea id="successful-numbers-${campaign.id}" class="phone-numbers-textarea" readonly title="Выделите нужные номера для копирования">${campaign.sent_numbers.filter(n => n.status === 'sent').map(n => n.phone_number).join('\n')}</textarea>
              </div>
              ` : ''}
            </div>
          </div>
          ` : ''}
          
          ${campaign.failed_numbers && campaign.failed_numbers.filter(n => n.status === 'failed').length > 0 ? `
          <div class="detail-section">
            <h4>❌ Ошибки отправки (${campaign.failed_numbers.filter(n => n.status === 'failed').length})</h4>
            <div class="phone-numbers-container">
              <div class="phone-numbers-header">
                <span class="phone-numbers-label">Номера с ошибками отправки:</span>
              </div>
              <div class="phone-numbers-list">
                ${campaign.failed_numbers.filter(n => n.status === 'failed').slice(0, 50).map(number => `
                  <div class="phone-number-item error">
                    <span class="phone-number">${number.phone_number}</span>
                    <span class="phone-error">${number.error || 'Неизвестная ошибка'}</span>
                  </div>
                `).join('')}
                ${campaign.failed_numbers.filter(n => n.status === 'failed').length > 50 ? `
                  <div class="phone-numbers-more">
                    ... и еще ${campaign.failed_numbers.filter(n => n.status === 'failed').length - 50} номеров с ошибками
                  </div>
                ` : ''}
              </div>
              <div class="phone-numbers-textarea-container">
                <div class="textarea-header">
                  <label class="phone-numbers-textarea-label">Все номера с ошибками отправки (${campaign.failed_numbers.filter(n => n.status === 'failed').length}):</label>
                  <button class="copy-textarea-btn" onclick="copyFailedNumbers('${campaign.id}')" title="Копировать все номера с ошибками">
                    <span class="copy-btn-text">📋 Копировать ошибки</span>
                  </button>
                </div>
                <textarea id="failed-numbers-${campaign.id}" class="phone-numbers-textarea" readonly title="Выделите нужные номера для копирования">${campaign.failed_numbers.filter(n => n.status === 'failed').map(n => n.phone_number).join('\n')}</textarea>
              </div>
            </div>
          </div>
          ` : ''}
          
          ${campaign.status === 'started' || campaign.status === 'pending' ? `
          <div class="detail-section">
            <div class="cancel-campaign-container">
              <button class="cancel-campaign-btn" onclick="cancelCampaign('${campaign.id}', '${campaign.name.replace(/'/g, "\\'")}')">
                🚫 Отменить рассылку
              </button>
            </div>
          </div>
          ` : ''}
        </div>
      `;
      
    } catch (error) {
      console.error('Error loading campaign details:', error);
      modalBody.innerHTML = `
        <div class="error-state">
          <div class="error-icon">❌</div>
          <div class="error-text">Ошибка загрузки деталей</div>
          <div class="error-details">${error.message}</div>
        </div>
      `;
      // Ошибка уже обработана в apiGet
    }
  }

  function getStatusIcon(status) {
    const iconMap = {
      'started': '🔄',
      'finished': '✅',
      'failed': '❌',
      'pending': '⏳',
      'cancelled': '🚫'
    };
    return iconMap[status] || '❓';
  }

  function getStatusText(status) {
    const statusMap = {
      'started': 'Запущена',
      'finished': 'Завершена',
      'failed': 'Ошибка',
      'pending': 'Ожидает',
      'cancelled': 'Отменена'
    };
    return statusMap[status] || status;
  }

  function formatDate(dateString) {
    if (!dateString) return 'Не указано';
    try {
      const date = new Date(dateString);
      if (isNaN(date.getTime())) return 'Неверная дата';
      return date.toLocaleString('ru-RU', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch (error) {
      return 'Ошибка даты';
    }
  }

  function getRelativeTime(dateString) {
    if (!dateString) return 'Не указано';
    try {
      const date = new Date(dateString);
      if (isNaN(date.getTime())) return 'Неверная дата';
      
      const now = new Date();
      const diffMs = now - date;
      const diffMins = Math.floor(diffMs / 60000);
      const diffHours = Math.floor(diffMs / 3600000);
      const diffDays = Math.floor(diffMs / 86400000);

      if (diffMins < 1) return 'только что';
      if (diffMins < 60) return `${diffMins} мин назад`;
      if (diffHours < 24) return `${diffHours} ч назад`;
      if (diffDays < 7) return `${diffDays} дн назад`;
      return `${Math.floor(diffDays / 7)} нед назад`;
    } catch (error) {
      return 'Ошибка даты';
    }
  }

  function getProgressPercentage(processedCount, totalCount) {
    if (!totalCount || totalCount === 0) return 0;
    return Math.round((processedCount / totalCount) * 100);
  }

  // Вспомогательные функции
  function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  // Глобальная функция для копирования номеров телефонов
  window.copyPhoneNumbers = function(phoneNumbers) {
    const text = phoneNumbers.join('\n');
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(text).then(() => {
        showToast(`Скопировано ${phoneNumbers.length} номеров`, 'success');
      }).catch(() => {
        fallbackCopyToClipboard(text);
      });
    } else {
      fallbackCopyToClipboard(text);
    }
  };

  function fallbackCopyToClipboard(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'absolute';
    textArea.style.left = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    try {
      document.execCommand('copy');
      showToast(`Номера скопированы в буфер обмена`, 'success');
    } catch (err) {
      showToast('Ошибка копирования', 'danger');
    }
    document.body.removeChild(textArea);
  }

  // Глобальная функция для копирования успешно отправленных номеров
  window.copySuccessfulNumbers = function(campaignId) {
    const textarea = document.getElementById(`successful-numbers-${campaignId}`);
    const button = event.target.closest('.copy-textarea-btn');
    const buttonText = button.querySelector('.copy-btn-text');
    
    if (!textarea || !textarea.value.trim()) {
      showToast('Нет номеров для копирования', 'danger');
      return;
    }

    const originalText = buttonText.innerHTML;
    
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(textarea.value).then(() => {
        // Успешное копирование
        buttonText.innerHTML = '✅ Скопировано!';
        button.style.background = '#52c41a';
        const phoneCount = textarea.value.split('\n').filter(n => n.trim()).length;
        showToast(`Скопировано ${phoneCount} успешно отправленных номеров`, 'success');
        
        // Возвращаем исходный вид через 2 секунды
        setTimeout(() => {
          buttonText.innerHTML = originalText;
          button.style.background = '';
        }, 2000);
      }).catch(() => {
        fallbackCopySuccessfulNumbers(textarea.value, buttonText, button, originalText);
      });
    } else {
      fallbackCopySuccessfulNumbers(textarea.value, buttonText, button, originalText);
    }
  };

  function fallbackCopySuccessfulNumbers(text, buttonText, button, originalText) {
    const tempTextArea = document.createElement('textarea');
    tempTextArea.value = text;
    tempTextArea.style.position = 'absolute';
    tempTextArea.style.left = '-999999px';
    document.body.appendChild(tempTextArea);
    tempTextArea.focus();
    tempTextArea.select();
    
    try {
      document.execCommand('copy');
      buttonText.innerHTML = '✅ Скопировано!';
      button.style.background = '#52c41a';
      const phoneCount = text.split('\n').filter(n => n.trim()).length;
      showToast(`Скопировано ${phoneCount} успешно отправленных номеров`, 'success');
      
      setTimeout(() => {
        buttonText.innerHTML = originalText;
        button.style.background = '';
      }, 2000);
    } catch (err) {
      buttonText.innerHTML = '❌ Ошибка';
      button.style.background = '#ff4d4f';
      showToast('Ошибка копирования', 'danger');
      
      setTimeout(() => {
        buttonText.innerHTML = originalText;
        button.style.background = '';
      }, 2000);
    }
    
    document.body.removeChild(tempTextArea);
  }

  // Глобальная функция для копирования номеров с ошибками
  window.copyFailedNumbers = function(campaignId) {
    const textarea = document.getElementById(`failed-numbers-${campaignId}`);
    const button = event.target.closest('.copy-textarea-btn');
    const buttonText = button.querySelector('.copy-btn-text');
    
    if (!textarea || !textarea.value.trim()) {
      showToast('Нет номеров с ошибками для копирования', 'danger');
      return;
    }

    const originalText = buttonText.innerHTML;
    
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(textarea.value).then(() => {
        // Успешное копирование
        buttonText.innerHTML = '✅ Скопировано!';
        button.style.background = '#52c41a';
        const phoneCount = textarea.value.split('\n').filter(n => n.trim()).length;
        showToast(`Скопировано ${phoneCount} номеров с ошибками`, 'success');
        
        // Возвращаем исходный вид через 2 секунды
        setTimeout(() => {
          buttonText.innerHTML = originalText;
          button.style.background = '';
        }, 2000);
      }).catch(() => {
        fallbackCopyFailedNumbers(textarea.value, buttonText, button, originalText);
      });
    } else {
      fallbackCopyFailedNumbers(textarea.value, buttonText, button, originalText);
    }
  };

  function fallbackCopyFailedNumbers(text, buttonText, button, originalText) {
    const tempTextArea = document.createElement('textarea');
    tempTextArea.value = text;
    tempTextArea.style.position = 'absolute';
    tempTextArea.style.left = '-999999px';
    document.body.appendChild(tempTextArea);
    tempTextArea.focus();
    tempTextArea.select();
    
    try {
      document.execCommand('copy');
      buttonText.innerHTML = '✅ Скопировано!';
      button.style.background = '#52c41a';
      const phoneCount = text.split('\n').filter(n => n.trim()).length;
      showToast(`Скопировано ${phoneCount} номеров с ошибками`, 'success');
      
      setTimeout(() => {
        buttonText.innerHTML = originalText;
        button.style.background = '';
      }, 2000);
    } catch (err) {
      buttonText.innerHTML = '❌ Ошибка';
      button.style.background = '#ff4d4f';
      showToast('Ошибка копирования', 'danger');
      
      setTimeout(() => {
        buttonText.innerHTML = originalText;
        button.style.background = '';
      }, 2000);
    }
    
    document.body.removeChild(tempTextArea);
  }

  // Делаем функцию loadHistory доступной глобально для кнопки повтора
  window.loadHistory = loadHistory;
  
  // Функция для отмены рассылки (глобальная для доступа из onclick)
  window.cancelCampaign = async function(campaignId, campaignName) {
    if (!confirm(`Вы уверены, что хотите отменить рассылку "${campaignName}"?\n\nЭто действие нельзя отменить.`)) {
      return;
    }

    try {
      await apiPost(`/api/v1/campaigns/${campaignId}/cancel`, {}, showToast);
      showToast('Рассылка отменена', 'success');
      // Закрываем модальное окно
      modal.style.display = 'none';
      // Перезагружаем историю
      loadHistory();
    } catch (error) {
      console.error('Error cancelling campaign:', error);
      // Ошибка уже обработана в apiPost
    }
  };
} 