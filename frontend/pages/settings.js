import { apiGet, apiPut, apiDelete } from '../ui/api.js';

// Страница настроек
export function renderSettingsPage() {
  return `
    <h2>Настройки</h2>
    <form id="settings-form" class="form">
      <label>WhatsGate API ключ <input name="apiKey" required autocomplete="off" placeholder="Введите API ключ..."></label>
      <label>WhatsappID <input name="whatsappId" required autocomplete="off" placeholder="Введите WhatsappID..."></label>
      <label>WhatsGate URL <input name="whatsgateUrl" required autocomplete="off" placeholder="https://api.whatsgate.ru"></label>
      <label>RetailCRM API ключ <input name="retailCrmApiKey" autocomplete="off" placeholder="Введите API ключ..."></label>
      <div class="form-actions">
        <button type="submit">Сохранить</button>
        <button type="button" id="reset-settings">Сбросить</button>
      </div>
    </form>
  `;
}

export function initSettingsForm(showToast) {
  const form = document.getElementById('settings-form');
  
  // Загрузка настроек
  apiGet('/api/v1/settings', showToast)
    .then(response => {
      // Бэкенд возвращает данные в формате {data: {...}}
      const data = response.data || response;
      if (data && data.api_key) {
        form.apiKey.value = data.api_key;
        form.whatsappId.value = data.whatsapp_id;
        form.whatsgateUrl.value = data.base_url;
      }
    })
    .catch(error => {
      console.error('Error loading settings:', error);
      // Ошибка уже обработана в apiGet
    });

  form.onsubmit = e => {
    e.preventDefault();
    // Валидация
    if (!form.apiKey.value.trim()) {
      showToast('Введите API ключ', 'danger');
      form.apiKey.classList.add('error');
      form.apiKey.focus();
      return;
    } else {
      form.apiKey.classList.remove('error');
    }
    if (!form.whatsappId.value.trim()) {
      showToast('Введите WhatsappID', 'danger');
      form.whatsappId.classList.add('error');
      form.whatsappId.focus();
      return;
    } else {
      form.whatsappId.classList.remove('error');
    }
    if (!form.whatsgateUrl.value.trim()) {
      showToast('Введите WhatsGate URL', 'danger');
      form.whatsgateUrl.classList.add('error');
      form.whatsgateUrl.focus();
      return;
    } else {
      form.whatsgateUrl.classList.remove('error');
    }
    
    const body = {
      api_key: form.apiKey.value.trim(),
      whatsapp_id: form.whatsappId.value.trim(),
      base_url: form.whatsgateUrl.value.trim()
    };
    
    const btn = form.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Сохранение...';
    
    apiPut('/api/v1/settings', body, showToast)
      .then(() => {
        showToast('Настройки сохранены', 'success');
      })
      .catch(error => {
        console.error('Error saving settings:', error);
        // Ошибка уже обработана в apiPut
      })
      .finally(() => { 
        btn.disabled = false; 
        btn.textContent = 'Сохранить'; 
      });
  };
  
  document.getElementById('reset-settings').onclick = () => {
    apiDelete('/api/v1/settings/reset', showToast)
      .then(() => {
        showToast('Настройки сброшены', 'success');
        form.reset();
      })
      .catch(error => {
        console.error('Error resetting settings:', error);
        // Ошибка уже обработана в apiDelete
      });
  };
} 