document.getElementById('login-form').addEventListener('submit', async (e) => {
  e.preventDefault();

  const code = document.getElementById('code').value.trim();
  const alertContainer = document.getElementById('alert-container');
  const submitBtn = e.target.querySelector('button[type="submit"]');

  alertContainer.innerHTML = '';
  submitBtn.disabled = true;
  submitBtn.textContent = 'Logging in...';

  try {
    const data = await apiRequest('/auth/code', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });

    saveAuth(data.token, data.role, data.user_id);

    alertContainer.innerHTML = `
      <div class="alert alert-success">Login successful! Redirecting...</div>
    `;

    setTimeout(() => {
      if (data.role === 'admin') {
        window.location.href = '/admin';
      } else {
        window.location.href = '/dashboard';
      }
    }, 1000);

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
    submitBtn.disabled = false;
    submitBtn.textContent = 'Login';
  }
});