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
              <th>📅 Дата</th>
              <th>🔧 Действия</th>
            </tr>
          </thead>
          <tbody id="history-tbody">
            <tr>
              <td colspan="6" class="loading">
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

  function loadHistory() {
    const tbody = document.getElementById('history-tbody');
    tbody.innerHTML = `
      <tr>
        <td colspan="6" class="loading">
          <div class="loading-spinner"></div>
          <span>Загрузка истории...</span>
        </td>
      </tr>
    `;

    // Добавляем анимацию к кнопке обновления
    refreshBtn.classList.add('refreshing');
    refreshBtn.disabled = true;

    fetch('/api/v1/messages/campaigns')
      .then(response => {
        if (!response.ok) {
          throw new Error('Ошибка загрузки истории');
        }
        return response.json();
      })
      .then(campaigns => {
        console.log('API response:', campaigns);
        
        // Проверяем, что campaigns существует и является массивом
        if (!campaigns || !Array.isArray(campaigns)) {
          console.log('Invalid response format:', typeof campaigns, campaigns);
          showEmptyState();
          return;
        }
        
        allCampaigns = campaigns;
        updateStats(campaigns);
        renderCampaigns(campaigns);
      })
      .catch(error => {
        console.error('Error loading history:', error);
        showErrorState();
        showToast('Ошибка загрузки истории', 'danger');
      })
      .finally(() => {
        refreshBtn.classList.remove('refreshing');
        refreshBtn.disabled = false;
      });
  }

  function updateStats(campaigns) {
    const total = campaigns.length;
    const completed = campaigns.filter(c => c.status === 'finished').length;
    const active = campaigns.filter(c => c.status === 'started').length;
    const failed = campaigns.filter(c => c.status === 'failed').length;

    document.getElementById('total-campaigns').textContent = total;
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
            ${campaign.media_filename ? '<div class="media-indicator"></div>' : ''}
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
              <span class="processed-number">${campaign.processed_count}</span>
              <span class="separator">/</span>
              <span class="total-number">${campaign.total}</span>
            </div>
            <div class="progress-bar">
              <div class="progress-fill" style="width: ${getProgressPercentage(campaign.processed_count, campaign.total)}%"></div>
            </div>
            <div class="progress-percentage">${getProgressPercentage(campaign.processed_count, campaign.total)}%</div>
          </div>
        </td>
        <td class="campaign-speed">
          <span class="speed-number">${campaign.messages_per_hour}</span>
          <span class="speed-label">/час</span>
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
        <td colspan="6" class="empty">
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
        <td colspan="6" class="error">
          <div class="error-state">
            <div class="error-icon">❌</div>
            <div class="error-text">Ошибка загрузки истории</div>
            <button class="retry-btn" onclick="loadHistory()">Повторить</button>
          </div>
        </td>
      </tr>
    `;
  }

  function showCampaignDetails(campaignId) {
    modal.style.display = 'block';
    modalTitle.textContent = 'Загрузка...';
    modalBody.innerHTML = '<div class="loading">Загрузка деталей...</div>';

    fetch(`/api/v1/messages/campaigns/${campaignId}`)
      .then(response => {
        if (!response.ok) {
          throw new Error('Рассылка не найдена');
        }
        return response.json();
      })
      .then(campaign => {
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
                  <span class="status status-${campaign.status}">${getStatusText(campaign.status)}</span>
                </div>
                <div class="detail-item">
                  <label>Дата:</label>
                  <span class="detail-value">${formatDate(campaign.created_at)}</span>
                </div>
              </div>
            </div>
            
            <div class="detail-section">
              <h4>📊 Статистика</h4>
              <div class="detail-grid">
                <div class="detail-item">
                  <label>Прогресс обработки:</label>
                  <div class="progress-detail">
                    <div class="progress-numbers">
                      <span class="processed-number">${campaign.processed_count}</span>
                      <span class="separator">/</span>
                      <span class="total-number">${campaign.total}</span>
                      <span class="progress-percentage">(${getProgressPercentage(campaign.processed_count, campaign.total)}%)</span>
                    </div>
                    <div class="progress-bar">
                      <div class="progress-fill" style="width: ${getProgressPercentage(campaign.processed_count, campaign.total)}%"></div>
                    </div>
                  </div>
                </div>
                <div class="detail-item">
                  <label>Сообщений в час:</label>
                  <span class="detail-value">${campaign.messages_per_hour}</span>
                </div>
                ${campaign.initiator ? `
                <div class="detail-item">
                  <label>Инициатор:</label>
                  <span class="detail-value">${campaign.initiator}</span>
                </div>
                ` : ''}
              </div>
            </div>
            
            <div class="detail-section">
              <h4>💬 Сообщение</h4>
              <div class="message-preview">${campaign.message}</div>
            </div>
            
            ${campaign.media_filename ? `
            <div class="detail-section" style="padding-bottom: 25px;">
              <h4>📎 Медиа файл</h4>
              <div class="media-info">
                <div class="media-item">
                  <span class="media-icon">📎</span>
                  <div class="media-details">
                    <div class="media-name">${campaign.media_filename}</div>
                    <div class="media-type">${campaign.media_type} (${campaign.media_mime})</div>
                  </div>
                </div>
              </div>
            </div>
            ` : ''}
          </div>
        `;
      })
      .catch(error => {
        console.error('Error loading campaign details:', error);
        modalTitle.textContent = 'Ошибка';
        modalBody.innerHTML = '<div class="error">Ошибка загрузки деталей рассылки</div>';
        showToast('Ошибка загрузки деталей', 'danger');
      });
  }

  function getStatusIcon(status) {
    const iconMap = {
      'started': '🔄',
      'finished': '✅',
      'failed': '❌',
      'pending': '⏳'
    };
    return iconMap[status] || '❓';
  }

  function getStatusText(status) {
    const statusMap = {
      'started': 'Запущена',
      'finished': 'Завершена',
      'failed': 'Ошибка',
      'pending': 'Ожидает'
    };
    return statusMap[status] || status;
  }

  function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function getRelativeTime(dateString) {
    const date = new Date(dateString);
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
  }

  function getProgressPercentage(processedCount, totalCount) {
    return Math.round((processedCount / totalCount) * 100);
  }

  // Делаем функцию loadHistory доступной глобально для кнопки повтора
  window.loadHistory = loadHistory;
} 