// Страница массовой рассылки
const fileIcon = `<svg fill="none" viewBox="0 0 20 20"><rect width="16" height="18" x="2" y="1" fill="#fff" stroke="#2d8cff" stroke-width="1.5" rx="4"/><path stroke="#2d8cff" stroke-width="1.5" d="M6 6h8M6 10h8M6 14h5"/></svg>`;

export function renderBulkPage() {
  return `
    <h2>Массовая рассылка</h2>
    <form id="bulk-form" class="form" enctype="multipart/form-data">
      <label>Название рассылки <input name="name" required autocomplete="off" placeholder="Например: Летняя акция"></label>
      <label style="pointer-events: none;">
        Файл номеров (xlsx)
        <span class="file-input-wrapper">
          <span class="file-input-label">${fileIcon} <span>Выбрать файл</span>
            <input type="file" name="numbers_file" class="file-input" accept=".xlsx" required>
          </span>
          <span class="file-name" id="file-name-xlsx">Файл не выбран</span>
        </span>
      </label>
      <label>Сообщений в час <input type="number" name="messages_per_hour" min="1" value="20" required placeholder="Например: 25"></label>
      <label>Сообщение <textarea style="height: 210px;" name="message" required placeholder="Введите текст сообщения..."></textarea></label>
      <label style="pointer-events: none;">
        Медиа файл
        <span class="file-input-wrapper">
          <span class="file-input-label">${fileIcon} <span>Выбрать медиа</span>
            <input type="file" name="media_file" class="file-input" accept="image/*,video/*,audio/*">
          </span>
          <span class="file-name" id="file-name-media">Файл не выбран</span>
        </span>
      </label>
      <div class="form-actions">
        <input name="testPhone" placeholder="Номер для теста" style="flex:1;" autocomplete="off" disabled>
        <button type="button" id="send-test" disabled>Отправить тест</button>
        <button type="submit">Отправить</button>
      </div>
    </form>
  `;
}

export function initBulkForm(showToast) {
  const form = document.getElementById('bulk-form');
  // Кастомные file input'ы
  const fileInput = form.querySelector('input[name="numbers_file"]');
  const fileName = document.getElementById('file-name-xlsx');
  fileInput.onchange = () => {
    fileName.textContent = fileInput.files[0]?.name || 'Файл не выбран';
    if (fileInput.files[0]) fileName.style.color = '#2ecc40';
    else fileName.style.color = '#b0b0b0';
  };
  const mediaInput = form.querySelector('input[name="media_file"]');
  const mediaName = document.getElementById('file-name-media');
  mediaInput.onchange = () => {
    mediaName.textContent = mediaInput.files[0]?.name || 'Файл не выбран';
    if (mediaInput.files[0]) mediaName.style.color = '#2ecc40';
    else mediaName.style.color = '#b0b0b0';
  };

  const testBtn = document.getElementById('send-test');
  testBtn.disabled = false;
  form.testPhone.disabled = false;
  form.testPhone.placeholder = '7XXXXXXXXXX';

  // Вспомогательная функция для спиннера
  function setLoading(isLoading, btn) {
    if (isLoading) {
      btn.disabled = true;
      btn.innerHTML = 'Отправка... <span class="spinner"></span>';
    } else {
      btn.disabled = false;
      btn.textContent = btn.id === 'send-test' ? 'Отправить тест' : 'Отправить';
    }
  }

  testBtn.onclick = async () => {
    const testPhone = form.testPhone.value.trim();
    const message = form.message.value.trim();
    // Валидация номера
    if (!/^7\d{10}$/.test(testPhone)) {
      showToast('Введите корректный номер: 11 цифр, начинается с 7', 'danger');
      form.testPhone.classList.add('error');
      form.testPhone.focus();
      return;
    } else {
      form.testPhone.classList.remove('error');
    }
    // Валидация сообщения
    if (!message) {
      showToast('Введите текст сообщения', 'danger');
      form.message.classList.add('error');
      form.message.focus();
      return;
    } else {
      form.message.classList.remove('error');
    }
    // Валидация медиа (если есть)
    if (form.media_file.files[0]) {
      const file = form.media_file.files[0];
      if (file.size > 20 * 1024 * 1024) {
        showToast('Медиафайл не должен превышать 20 МБ', 'danger');
        return;
      }
    }
    setLoading(true, testBtn);
    const fd = new FormData();
    fd.append('phone', testPhone);
    fd.append('message', message);
    if (form.media_file.files[0]) fd.append('media_file', form.media_file.files[0]);
    fetch('/api/v1/messages/test-send', {
      method: 'POST',
      body: fd
    })
      .then(r => r.ok ? showToast('Тестовое сообщение отправлено', 'success') : r.json().then(d => Promise.reject(d)))
      .catch(() => showToast('Ошибка отправки теста', 'danger'))
      .finally(() => setLoading(false, testBtn));
  };

  form.onsubmit = e => {
    e.preventDefault();
    // Валидация массовой рассылки
    const message = form.message.value.trim();
    if (!message) {
      showToast('Введите текст сообщения', 'danger');
      form.message.classList.add('error');
      form.message.focus();
      return;
    } else {
      form.message.classList.remove('error');
    }
    if (!form.numbers_file.files[0]) {
      showToast('Выберите файл номеров', 'danger');
      form.numbers_file.classList.add('error');
      form.numbers_file.focus();
      return;
    } else {
      form.numbers_file.classList.remove('error');
    }
    if (form.media_file.files[0]) {
      const file = form.media_file.files[0];
      if (file.size > 20 * 1024 * 1024) {
        showToast('Медиафайл не должен превышать 20 МБ', 'danger');
        return;
      }
    }
    setLoading(true, form.querySelector('button[type="submit"]'));
    const fd = new FormData();
    fd.append('name', form.name.value.trim());
    fd.append('message', message);
    fd.append('messages_per_hour', form.messages_per_hour.value);
    fd.append('numbers_file', form.numbers_file.files[0]);
    if (form.media_file.files[0]) fd.append('media_file', form.media_file.files[0]);
    fd.append('async', 'false');
    fetch('/api/v1/messages/bulk-send', {
      method: 'POST',
      body: fd
    })
      .then(r => r.ok ? showToast('Рассылка запущена', 'success') : r.json().then(d => Promise.reject(d)))
      .catch(() => showToast('Ошибка запуска рассылки', 'danger'))
      .finally(() => setLoading(false, form.querySelector('button[type="submit"]')));
  };
} 