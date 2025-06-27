import { renderSettingsPage, initSettingsForm } from './pages/settings.js';
import { renderBulkPage, initBulkForm } from './pages/bulk.js';
import { showToast } from './ui/toast.js';

const fileIcon = `<svg fill="none" viewBox="0 0 20 20"><rect width="16" height="18" x="2" y="1" fill="#fff" stroke="#2d8cff" stroke-width="1.5" rx="4"/><path stroke="#2d8cff" stroke-width="1.5" d="M6 6h8M6 10h8M6 14h5"/></svg>`;

const pages = {
  settings: renderSettingsPage(),
  bulk: renderBulkPage(),
  subscribe: `<h2>Уведомление о подписке</h2><p>Здесь будет форма для уведомлений.</p>`,
  history: `<h2>История рассылок</h2><p>Здесь будет история рассылок.</p>`
};

function loadPage(page) {
  document.getElementById('main-content').innerHTML = pages[page];
  if (page === 'settings') initSettingsForm(showToast);
  if (page === 'bulk') initBulkForm(showToast);
}

document.querySelectorAll('.sidebar li').forEach(li => {
  li.onclick = () => {
    document.querySelectorAll('.sidebar li').forEach(l => l.classList.remove('active'));
    li.classList.add('active');
    loadPage(li.dataset.page);
  };
});

// Инициализация первой страницы
loadPage('settings'); 