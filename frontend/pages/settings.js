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
  fetch('/api/v1/settings')
    .then(r => r.json())
    .then(data => {
      if (data && data.api_key) {
        form.apiKey.value = data.api_key;
        form.whatsappId.value = data.whatsapp_id;
        form.whatsgateUrl.value = data.base_url;
      }
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
    fetch('/api/v1/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    })
      .then(r => r.ok ? showToast('Настройки сохранены', 'success') : r.json().then(d => Promise.reject(d)))
      .catch(() => showToast('Ошибка сохранения', 'danger'))
      .finally(() => { btn.disabled = false; btn.textContent = 'Сохранить'; });
  };
  document.getElementById('reset-settings').onclick = () => {
    fetch('/api/v1/settings/reset', { method: 'DELETE' })
      .then(r => r.ok ? showToast('Настройки сброшены', 'success') : showToast('Ошибка сброса', 'danger'));
    form.reset();
  };
} 