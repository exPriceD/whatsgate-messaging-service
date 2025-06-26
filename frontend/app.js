const fileIcon = `<svg fill="none" viewBox="0 0 20 20"><rect width="16" height="18" x="2" y="1" fill="#fff" stroke="#2d8cff" stroke-width="1.5" rx="4"/><path stroke="#2d8cff" stroke-width="1.5" d="M6 6h8M6 10h8M6 14h5"/></svg>`;

const pages = {
  settings: `
    <h2>Настройки</h2>
    <form id="settings-form" class="form">
      <label>WhatsGate API ключ <input name="apiKey" required autocomplete="off" placeholder="Введите API ключ..."></label>
      <label>WhatsappID <input name="whatsappId" required autocomplete="off" placeholder="Введите WhatsappID..."></label>
      <label>WhatsGate URL <input name="whatsgateUrl" required autocomplete="off" value="https://whatsgate.ru/api/v1" placeholder="https://api.whatsgate.ru"></label>
      <label>RetailCRM API ключ <input name="retailCrmApiKey" required autocomplete="off" placeholder="Введите API ключ..."></label>
      <div class="form-actions">
        <button type="submit">Сохранить</button>
        <button type="button" id="reset-settings">Сбросить</button>
      </div>
    </form>
  `,
  bulk: `
    <h2>Массовая рассылка</h2>
    <form id="bulk-form" class="form" enctype="multipart/form-data">
      <label>Название рассылки <input name="name" required autocomplete="off" placeholder="Например: Летняя акция"></label>
      <label style="pointer-events: none;">
        Файл номеров (xlsx)
        <span class="file-input-wrapper">
          <span class="file-input-label">${fileIcon} <span>Выбрать файл</span>
            <input type="file" name="file" class="file-input" accept=".xlsx" required>
          </span>
          <span class="file-name" id="file-name-xlsx">Файл не выбран</span>
        </span>
      </label>
      <label>Сообщений в час <input type="number" name="rate" min="1" value="20" required placeholder="Например: 25"></label>
      <label>Сообщение <textarea style="height: 210px;" name="message" required placeholder="Введите текст сообщения..."></textarea></label>
      <label style="pointer-events: none;">
        Медиа файл
        <span class="file-input-wrapper">
          <span class="file-input-label">${fileIcon} <span>Выбрать медиа</span>
            <input type="file" name="media" class="file-input" accept="image/*,video/*,audio/*">
          </span>
          <span class="file-name" id="file-name-media">Файл не выбран</span>
        </span>
      </label>
      <div class="form-actions">
        <input name="testPhone" placeholder="Номер для теста" style="flex:1;" autocomplete="off">
        <button type="button" id="send-test">Отправить тест</button>
        <button type="submit">Отправить</button>
      </div>
    </form>
  `,
  subscribe: `<h2>Уведомление о подписке</h2><p>Здесь будет форма для уведомлений.</p>`,
  history: `<h2>История рассылок</h2><p>Здесь будет история рассылок.</p>`
};

function showToast(msg, type = "") {
  const toast = document.getElementById('toast');
  toast.textContent = msg;
  toast.className = 'show' + (type ? ' ' + type : '');
  setTimeout(() => { toast.className = ''; }, 3000);
}

function loadPage(page) {
  document.getElementById('main-content').innerHTML = pages[page];
  if (page === 'settings') initSettingsForm();
  if (page === 'bulk') initBulkForm();
}

document.querySelectorAll('.sidebar li').forEach(li => {
  li.onclick = () => {
    document.querySelectorAll('.sidebar li').forEach(l => l.classList.remove('active'));
    li.classList.add('active');
    loadPage(li.dataset.page);
  };
});

// SETTINGS FORM
function initSettingsForm() {
  const form = document.getElementById('settings-form');
  fetch('/api/settings').then(r => r.json()).then(data => {
    if (data && data.apiKey) {
      form.apiKey.value = data.apiKey;
      form.whatsappId.value = data.whatsappId;
      form.whatsgateUrl.value = data.whatsgateUrl;
      form.retailCrmApiKey.value = data.retailCrmApiKey || '';
    }
  });
  form.onsubmit = e => {
    e.preventDefault();
    const body = {
      apiKey: form.apiKey.value.trim(),
      whatsappId: form.whatsappId.value.trim(),
      whatsgateUrl: form.whatsgateUrl.value.trim(),
      retailCrmApiKey: form.retailCrmApiKey.value.trim()
    };
    fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    })
      .then(r => r.ok ? showToast('Настройки сохранены', 'success') : r.json().then(d => Promise.reject(d)))
      .catch(() => showToast('Ошибка сохранения', 'danger'));
  };
  document.getElementById('reset-settings').onclick = () => {
    form.reset();
    showToast('Поля сброшены');
  };
}

// BULK FORM
function isValidPhone(phone) {
  // Только 11 цифр, начинается с 7
  return /^7\d{10}$/.test(phone);
}

function renderStatsBlock(stats) {
  let html = `<div class="bulk-stats-block">
    <h3>Статистика номеров</h3>
    <ul>
      <li>Всего: <b>${stats.total}</b></li>
      <li>Валидных: <b style="color:var(--success)">${stats.valid}</b></li>
      <li>Невалидных: <b style="color:var(--danger)">${stats.invalid}</b></li>
      <li>Пустых: <b>${stats.empty}</b></li>
    </ul>
  </div>`;
  let statsDiv = document.getElementById('bulk-stats');
  if (!statsDiv) {
    statsDiv = document.createElement('div');
    statsDiv.id = 'bulk-stats';
    form.parentNode.insertBefore(statsDiv, form.nextSibling);
  }
  statsDiv.innerHTML = html;
}

function parsePhonesFromSheet(file, cb) {
  const reader = new FileReader();
  reader.onload = function(e) {
    const data = new Uint8Array(e.target.result);
    const workbook = XLSX.read(data, {type: 'array'});
    // Берём первый лист
    const sheet = workbook.Sheets[workbook.SheetNames[0]];
    const rows = XLSX.utils.sheet_to_json(sheet, {header:1});
    // Ищем все значения, похожие на номера
    let phones = [];
    rows.forEach(row => {
      row.forEach(cell => {
        if (typeof cell === 'string' || typeof cell === 'number') {
          let val = String(cell).replace(/\D/g, '');
          phones.push(val);
        }
      });
    });
    cb(phones);
  };
  reader.readAsArrayBuffer(file);
}

function updateBulkStats(file) {
  if (!file) {
    renderStatsBlock({total:0, valid:0, invalid:0, empty:0});
    return;
  }
  parsePhonesFromSheet(file, phones => {
    let stats = {total:0, valid:0, invalid:0, empty:0};
    phones.forEach(p => {
      if (!p || p.length === 0) stats.empty++;
      else if (isValidPhone(p)) stats.valid++;
      else stats.invalid++;
      stats.total++;
    });
    renderStatsBlock(stats);
  });
}

function initBulkForm() {
  const form = document.getElementById('bulk-form');
  // Кастомные file input'ы
  const fileInput = form.querySelector('input[name="file"]');
  const fileName = document.getElementById('file-name-xlsx');
  fileInput.onchange = () => {
    fileName.textContent = fileInput.files[0]?.name || 'Файл не выбран';
    updateBulkStats(fileInput.files[0]);
  };
  // Показываем статистику при повторном открытии формы, если файл уже выбран
  if (fileInput.files[0]) updateBulkStats(fileInput.files[0]);
  const mediaInput = form.querySelector('input[name="media"]');
  const mediaName = document.getElementById('file-name-media');
  mediaInput.onchange = () => {
    mediaName.textContent = mediaInput.files[0]?.name || 'Файл не выбран';
  };
  const testBtn = document.getElementById('send-test');
  testBtn.onclick = () => {
    const testPhone = form.testPhone.value.trim();
    if (!testPhone) return showToast('Введите номер для теста', 'danger');
    const fd = new FormData(form);
    fd.set('testPhone', testPhone);
    fetch('/api/messages/test', {
      method: 'POST',
      body: fd
    })
      .then(r => r.ok ? showToast('Тестовое сообщение отправлено', 'success') : r.json().then(d => Promise.reject(d)))
      .catch(() => showToast('Ошибка отправки теста', 'danger'));
  };
  form.onsubmit = e => {
    e.preventDefault();
    const fd = new FormData(form);
    fetch('/api/messages/bulk', {
      method: 'POST',
      body: fd
    })
      .then(r => r.ok ? showToast('Рассылка запущена', 'success') : r.json().then(d => Promise.reject(d)))
      .catch(() => showToast('Ошибка запуска рассылки', 'danger'));
  };
}

// Инициализация первой страницы
loadPage('settings'); 