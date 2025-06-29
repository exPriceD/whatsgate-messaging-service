import { apiPost } from '../ui/api.js';

// Страница массовой рассылки
const fileIcon = `<svg fill="none" viewBox="0 0 20 20"><rect width="16" height="18" x="2" y="1" fill="#fff" stroke="#2d8cff" stroke-width="1.5" rx="4"/><path stroke="#2d8cff" stroke-width="1.5" d="M6 6h8M6 10h8M6 14h5"/></svg>`;

export function renderBulkPage() {
  return `
    <h2>Массовая рассылка</h2>
    <div class="bulk-form-container">
      <form id="bulk-form" class="form" enctype="multipart/form-data">
        <label>Название рассылки <input name="name" required autocomplete="off" placeholder="Например: Летняя акция"></label>
        <label class="file-label">
          Файл номеров (xlsx)
          <span class="file-input-wrapper">
            <span class="file-input-label">${fileIcon} <span>Выбрать файл</span>
              <input type="file" name="numbers_file" class="file-input" accept=".xlsx" required>
            </span>
            <span class="file-name" id="file-name-xlsx">Файл не выбран</span>
          </span>
        </label>
        <label>Сообщений в час <input type="number" name="messages_per_hour" min="1" value="20" required placeholder="Например: 25"></label>
        <label>Сообщение <textarea name="message" required placeholder="Введите текст сообщения..."></textarea></label>
        <label class="file-label">
          Медиа файл
          <span class="file-input-wrapper">
            <span class="file-input-label">${fileIcon} <span>Выбрать медиа</span>
              <input type="file" name="media_file" class="file-input" accept="image/*,video/*,audio/*">
            </span>
            <span class="file-name" id="file-name-media">Файл не выбран</span>
          </span>
        </label>
        <div class="form-actions">
          <input name="testPhone" placeholder="Номер для теста" autocomplete="off" disabled>
          <button type="button" id="send-test" disabled>Отправить тест</button>
          <button type="submit">Отправить</button>
        </div>
      </form>
      
      <div class="bulk-form-sidebar">
        <div class="additional-numbers-section">
          <h4>➕ Добавить номера</h4>
          <p class="section-description">Дополнительные номера к файлу (по одному на строку)</p>
          <textarea 
            name="additional_numbers" 
            class="numbers-textarea" 
            placeholder="71234567890&#10;79876543210&#10;75551234567"
            rows="6"
          ></textarea>
        </div>
        
        <div class="exclude-numbers-section">
          <h4>🚫 Исключить номера</h4>
          <p class="section-description">Номера для исключения из файла (по одному на строку)</p>
          <textarea 
            name="exclude_numbers" 
            class="numbers-textarea" 
            placeholder="71234567890&#10;79876543210&#10;75551234567"
            rows="6"
          ></textarea>
          <div class="exclude-hint">
            💡 Скопируйте номера из деталей рассылки и вставьте сюда
          </div>
        </div>
        
        <div class="numbers-summary">
          <h4>📊 Сводка номеров</h4>
          <div class="summary-item">
            <span class="summary-label">Из файла:</span>
            <span class="summary-value" id="file-count">0</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">Добавить:</span>
            <span class="summary-value" id="add-count">0</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">Исключить:</span>
            <span class="summary-value" id="exclude-count">0</span>
          </div>
          <div class="summary-item total">
            <span class="summary-label">Итого:</span>
            <span class="summary-value" id="total-count">0</span>
          </div>
        </div>
      </div>
    </div>
  `;
}

export function initBulkForm(showToast) {
  const form = document.getElementById('bulk-form');
  
  // Кастомные file input'ы
  const fileInput = form.querySelector('input[name="numbers_file"]');
  const fileName = document.getElementById('file-name-xlsx');
  fileInput.onchange = async () => {
    fileName.textContent = fileInput.files[0]?.name || 'Файл не выбран';
    if (fileInput.files[0]) {
      fileName.style.color = '#2ecc40';
      // Подсчитываем количество строк в файле
      await countRowsInFile(fileInput.files[0]);
    } else {
      fileName.style.color = '#b0b0b0';
      document.getElementById('file-count').textContent = '0';
    }
    updateNumbersSummary();
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
    
    try {
      await apiPost('/api/v1/messages/test-send', fd, showToast);
      showToast('Тестовое сообщение отправлено', 'success');
    } catch (error) {
      console.error('Error sending test message:', error);
      // Ошибка уже обработана в apiPost
    } finally {
      setLoading(false, testBtn);
    }
  };

  form.onsubmit = async e => {
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
    
    // Добавляем дополнительные и исключаемые номера
    const additionalTextarea = document.querySelector('textarea[name="additional_numbers"]');
    const excludeTextarea = document.querySelector('textarea[name="exclude_numbers"]');
    
    const additionalNumbers = additionalTextarea ? additionalTextarea.value.trim() : '';
    const excludeNumbers = excludeTextarea ? excludeTextarea.value.trim() : '';
    
    if (additionalNumbers) {
      fd.append('additional_numbers', additionalNumbers);
    }
    if (excludeNumbers) {
      fd.append('exclude_numbers', excludeNumbers);
    }
    
    try {
      await apiPost('/api/v1/messages/bulk-send', fd, showToast);
      showToast('Рассылка запущена', 'success');
    } catch (error) {
      console.error('Error starting bulk campaign:', error);
      // Ошибка уже обработана в apiPost
    } finally {
      setLoading(false, form.querySelector('button[type="submit"]'));
    }
  };

  // Функция для подсчета строк в Excel файле
  async function countRowsInFile(file) {
    try {
      // Показываем индикатор загрузки
      const fileCountElement = document.getElementById('file-count');
      fileCountElement.textContent = '...';
      
      const formData = new FormData();
      formData.append('file', file);
      
      const response = await fetch('/api/v1/messages/count-file-rows', {
        method: 'POST',
        body: formData
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.success && data.rows > 0) {
          fileCountElement.textContent = data.rows;
          // Добавляем подсказку, что это приблизительное значение
          fileCountElement.title = 'Приблизительное количество строк в файле';
        } else {
          fileCountElement.textContent = '~';
          fileCountElement.title = 'Не удалось точно подсчитать строки';
        }
      } else {
        fileCountElement.textContent = '~';
        fileCountElement.title = 'Ошибка при подсчете строк';
      }
    } catch (error) {
      console.error('Error counting rows in file:', error);
      const fileCountElement = document.getElementById('file-count');
      fileCountElement.textContent = '~';
      fileCountElement.title = 'Ошибка при подсчете строк';
    }
  }

  // Функция для подсчета номеров в текстовом поле
  function countNumbers(text) {
    if (!text.trim()) return 0;
    return text.trim().split('\n').filter(line => line.trim()).length;
  }

  // Функция для обновления сводки номеров
  function updateNumbersSummary() {
    const additionalTextarea = document.querySelector('textarea[name="additional_numbers"]');
    const excludeTextarea = document.querySelector('textarea[name="exclude_numbers"]');
    
    const additionalText = additionalTextarea ? additionalTextarea.value : '';
    const excludeText = excludeTextarea ? excludeTextarea.value : '';
    
    const additionalCount = countNumbers(additionalText);
    const excludeCount = countNumbers(excludeText);
    
    // Получаем количество из файла
    const fileCountElement = document.getElementById('file-count');
    const fileCountText = fileCountElement ? fileCountElement.textContent : '0';
    const fileCount = fileCountText === '~' ? 0 : parseInt(fileCountText) || 0;
    
    const totalCount = Math.max(0, fileCount + additionalCount - excludeCount);
    
    document.getElementById('add-count').textContent = additionalCount;
    document.getElementById('exclude-count').textContent = excludeCount;
    document.getElementById('total-count').textContent = totalCount;
  }

  // Обработчики для текстовых полей
  const additionalTextarea = document.querySelector('textarea[name="additional_numbers"]');
  const excludeTextarea = document.querySelector('textarea[name="exclude_numbers"]');
  
  if (additionalTextarea) {
    additionalTextarea.addEventListener('input', updateNumbersSummary);
  }
  if (excludeTextarea) {
    excludeTextarea.addEventListener('input', updateNumbersSummary);
  }
} 