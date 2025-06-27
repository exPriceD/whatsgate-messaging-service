export function showToast(msg, type = "") {
  const toast = document.getElementById('toast');
  toast.textContent = msg;
  toast.className = 'show' + (type ? ' ' + type : '');
  setTimeout(() => { toast.className = ''; }, 3000);
} 