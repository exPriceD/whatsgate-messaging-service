import { apiGet, apiPut, apiDelete } from '../ui/api.js';

// Страница настроек с двумя отдельными формами
export function renderSettingsPage() {
  return `
    <h2>Настройки</h2>
    
    <!-- WhatsGate настройки -->
    <div class="settings-section">
      <h3>WhatsGate</h3>
      <form id="whatsgate-settings-form" class="form">
        <label>WhatsApp ID <input name="whatsappId" required autocomplete="off" placeholder="Введите WhatsApp ID..."></label>
        <label>API ключ <input name="apiKey" required autocomplete="off" placeholder="Введите API ключ..."></label>
        <label>Base URL <input name="baseUrl" required autocomplete="off" placeholder="https://api.whatsgate.ru"></label>
        <div class="form-actions">
          <button type="submit">Сохранить</button>
          <button type="button" id="reset-whatsgate-settings">Сбросить</button>
        </div>
      </form>
    </div>

    <!-- RetailCRM настройки -->
    <div class="settings-section">
      <h3>RetailCRM</h3>
      <form id="retailcrm-settings-form" class="form">
        <label>API ключ <input name="apiKey" required autocomplete="off" placeholder="Введите RetailCRM API ключ..."></label>
        <label>Base URL <input name="baseUrl" required autocomplete="off" placeholder="https://example.retailcrm.ru"></label>
        <div class="form-actions">
          <button type="submit">Сохранить</button>
          <button type="button" id="reset-retailcrm-settings">Сбросить</button>
        </div>
      </form>
    </div>
  `;
}

export function initSettingsForm(showToast) {
  const whatsgateForm = document.getElementById('whatsgate-settings-form');
  const retailcrmForm = document.getElementById('retailcrm-settings-form');
  
  // Загрузка настроек WhatsGate
  loadWhatsgateSettings(whatsgateForm, showToast);
  
  // Загрузка настроек RetailCRM
  loadRetailCRMSettings(retailcrmForm, showToast);
  
  // Обработчики форм
  setupWhatsgateForm(whatsgateForm, showToast);
  setupRetailCRMForm(retailcrmForm, showToast);
}

// Загрузка настроек WhatsGate
function loadWhatsgateSettings(form, showToast) {
  apiGet('/api/v1/whatsgate-settings', showToast)
    .then(response => {
      const data = response.data || response;
      if (data && data.api_key) {
        form.whatsappId.value = data.whatsapp_id || '';
        form.apiKey.value = data.api_key || '';
        form.baseUrl.value = data.base_url || '';
      }
    })
    .catch(error => {
      console.error('Error loading WhatsGate settings:', error);
    });
}

// Загрузка настроек RetailCRM
function loadRetailCRMSettings(form, showToast) {
  apiGet('/api/v1/retailcrm-settings', showToast)
    .then(response => {
      const data = response.data || response;
      if (data && data.api_key) {
        form.apiKey.value = data.api_key || '';
        form.baseUrl.value = data.base_url || '';
      }
    })
    .catch(error => {
      console.error('Error loading RetailCRM settings:', error);
    });
}

// Настройка формы WhatsGate
function setupWhatsgateForm(form, showToast) {
  form.onsubmit = e => {
    e.preventDefault();
    
    // Валидация
    if (!validateWhatsgateForm(form, showToast)) {
      return;
    }
    
    const body = {
      whatsapp_id: form.whatsappId.value.trim(),
      api_key: form.apiKey.value.trim(),
      base_url: form.baseUrl.value.trim()
    };
    
    const btn = form.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Сохранение...';
    
    apiPut('/api/v1/whatsgate-settings', body, showToast)
      .then(() => {
        showToast('WhatsGate настройки сохранены', 'success');
      })
      .catch(error => {
        console.error('Error saving WhatsGate settings:', error);
      })
      .finally(() => { 
        btn.disabled = false; 
        btn.textContent = 'Сохранить';
      });
  };
  
  // Сброс настроек WhatsGate
  document.getElementById('reset-whatsgate-settings').onclick = () => {
    apiDelete('/api/v1/whatsgate-settings/reset', showToast)
      .then(() => {
        showToast('WhatsGate настройки сброшены', 'success');
        form.reset();
      })
      .catch(error => {
        console.error('Error resetting WhatsGate settings:', error);
      });
  };
}

// Настройка формы RetailCRM
function setupRetailCRMForm(form, showToast) {
  form.onsubmit = e => {
    e.preventDefault();
    
    // Валидация
    if (!validateRetailCRMForm(form, showToast)) {
      return;
    }
    
    const body = {
      api_key: form.apiKey.value.trim(),
      base_url: form.baseUrl.value.trim()
    };
    
    const btn = form.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Сохранение...';
    
    apiPut('/api/v1/retailcrm-settings', body, showToast)
      .then(() => {
        showToast('RetailCRM настройки сохранены', 'success');
      })
      .catch(error => {
        console.error('Error saving RetailCRM settings:', error);
      })
      .finally(() => { 
        btn.disabled = false; 
        btn.textContent = 'Сохранить';
      });
  };
  
  // Сброс настроек RetailCRM
  document.getElementById('reset-retailcrm-settings').onclick = () => {
    apiDelete('/api/v1/retailcrm-settings/reset', showToast)
      .then(() => {
        showToast('RetailCRM настройки сброшены', 'success');
        form.reset();
      })
      .catch(error => {
        console.error('Error resetting RetailCRM settings:', error);
      });
  };
}

// Валидация формы WhatsGate
function validateWhatsgateForm(form, showToast) {
  let isValid = true;
  
  // Проверка WhatsApp ID
  if (!form.whatsappId.value.trim()) {
    showToast('Введите WhatsApp ID', 'danger');
    form.whatsappId.classList.add('error');
    form.whatsappId.focus();
    isValid = false;
  } else {
    form.whatsappId.classList.remove('error');
  }
  
  // Проверка API ключа
  if (!form.apiKey.value.trim()) {
    showToast('Введите API ключ для WhatsGate', 'danger');
    form.apiKey.classList.add('error');
    if (isValid) form.apiKey.focus();
    isValid = false;
  } else {
    form.apiKey.classList.remove('error');
  }
  
  // Проверка Base URL
  if (!form.baseUrl.value.trim()) {
    showToast('Введите Base URL для WhatsGate', 'danger');
    form.baseUrl.classList.add('error');
    if (isValid) form.baseUrl.focus();
    isValid = false;
  } else {
    form.baseUrl.classList.remove('error');
  }
  
  return isValid;
}

// Валидация формы RetailCRM
function validateRetailCRMForm(form, showToast) {
  let isValid = true;
  
  // Проверка API ключа
  if (!form.apiKey.value.trim()) {
    showToast('Введите API ключ для RetailCRM', 'danger');
    form.apiKey.classList.add('error');
    form.apiKey.focus();
    isValid = false;
  } else {
    form.apiKey.classList.remove('error');
  }
  
  // Проверка Base URL
  if (!form.baseUrl.value.trim()) {
    showToast('Введите Base URL для RetailCRM', 'danger');
    form.baseUrl.classList.add('error');
    if (isValid) form.baseUrl.focus();
    isValid = false;
  } else {
    form.baseUrl.classList.remove('error');
  }
  
  return isValid;
} 