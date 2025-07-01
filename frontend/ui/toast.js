export function showToast(msg, type = "") {
  const toast = document.getElementById('toast');
  let icon = '';
  if (type === 'success') icon = '✓ ';
  if (type === 'danger') icon = '✗ ';
  toast.textContent = icon + msg;
  toast.className = '';
  void toast.offsetWidth; // force reflow for animation
  toast.classList.add('show');
  if (type) toast.classList.add(type);
  toast.style.opacity = 1;
  setTimeout(() => { toast.style.opacity = 0; toast.className = ''; }, 5000);
} 