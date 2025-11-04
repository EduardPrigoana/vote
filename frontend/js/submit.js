requireAuth();

if (isAdmin()) {
  document.getElementById('admin-link').style.display = 'block';
}

const form = document.getElementById('policy-form');
const alertContainer = document.getElementById('alert-container');

form.addEventListener('submit', async (e) => {
  e.preventDefault();

  const title = document.getElementById('title').value.trim();
  const description = document.getElementById('description').value.trim();
  const submitBtn = form.querySelector('button[type="submit"]');

  alertContainer.innerHTML = '';
  submitBtn.disabled = true;
  submitBtn.textContent = 'Submitting...';

  try {
    const data = await apiRequest('/policies', {
      method: 'POST',
      body: JSON.stringify({ title, description }),
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">
        ${data.message} Your policy is now pending admin review.
      </div>
    `;

    form.reset();

    setTimeout(() => {
      window.location.href = '/dashboard';
    }, 2000);

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = 'Submit Policy';
  }
});