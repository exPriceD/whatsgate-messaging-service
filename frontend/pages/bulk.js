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
  };
  const mediaInput = form.querySelector('input[name="media_file"]');
  const mediaName = document.getElementById('file-name-media');
  mediaInput.onchange = () => {
    mediaName.textContent = mediaInput.files[0]?.name || 'Файл не выбран';
  };

  const testBtn = document.getElementById('send-test');
  testBtn.disabled = false;
  form.testPhone.disabled = false;
  testBtn.onclick = async () => {
    const testPhone = form.testPhone.value.trim();
    if (!testPhone) return showToast('Введите номер для теста', 'danger');
    const fd = new FormData();
    fd.append('phone', testPhone);
    fd.append('message', form.message.value);
    if (form.media_file.files[0]) fd.append('media_file', form.media_file.files[0]);
    fetch('/api/v1/messages/test-send', {
      method: 'POST',
      body: fd
    })
      .then(r => r.ok ? showToast('Тестовое сообщение отправлено', 'success') : r.json().then(d => Promise.reject(d)))
      .catch(() => showToast('Ошибка отправки теста', 'danger'));
  };
  form.onsubmit = e => {
    e.preventDefault();
    const fd = new FormData();
    fd.append('message', form.message.value);
    fd.append('messages_per_hour', form.messages_per_hour.value);
    fd.append('numbers_file', form.numbers_file.files[0]);
    if (form.media_file.files[0]) fd.append('media_file', form.media_file.files[0]);
    fd.append('async', 'false');
    fetch('/api/v1/messages/bulk-send', {
      method: 'POST',
      body: fd
    })
      .then(r => r.ok ? showToast('Рассылка запущена', 'success') : r.json().then(d => Promise.reject(d)))
      .catch(() => showToast('Ошибка запуска рассылки', 'danger'));
  };
} 