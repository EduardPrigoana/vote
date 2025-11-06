requireAuth();

if (isAdmin()) {
  document.getElementById('admin-link').style.display = 'block';
}

if (isSuperuser()) {
  document.getElementById('superuser-link').style.display = 'block';
}

const form = document.getElementById('policy-form');
const alertContainer = document.getElementById('alert-container');
const titleInput = document.getElementById('title');
const descriptionInput = document.getElementById('description');
const categorySelect = document.getElementById('category');

// Load categories
async function loadCategories() {
  try {
    const categories = await apiRequest('/categories?lang=en');
    
    categorySelect.innerHTML = '<option value="">Select a category...</option>';
    
    categories.forEach(cat => {
      const option = document.createElement('option');
      option.value = cat.id;
      option.textContent = cat.name;
      categorySelect.appendChild(option);
    });
  } catch (error) {
    console.error('Failed to load categories:', error);
  }
}

// Character counters
titleInput.addEventListener('input', () => {
  const count = titleInput.value.length;
  document.getElementById('title-counter').textContent = `${count}/200`;
  
  if (count < 10) {
    document.getElementById('title-counter').style.color = 'var(--danger)';
  } else if (count > 180) {
    document.getElementById('title-counter').style.color = 'var(--warning)';
  } else {
    document.getElementById('title-counter').style.color = 'var(--muted-foreground)';
  }
});

descriptionInput.addEventListener('input', () => {
  const count = descriptionInput.value.length;
  document.getElementById('description-counter').textContent = `${count}/2000`;
  
  if (count < 50) {
    document.getElementById('description-counter').style.color = 'var(--danger)';
  } else if (count > 1900) {
    document.getElementById('description-counter').style.color = 'var(--warning)';
  } else {
    document.getElementById('description-counter').style.color = 'var(--muted-foreground)';
  }
});

// Submit form
form.addEventListener('submit', async (e) => {
  e.preventDefault();

  const title = titleInput.value.trim();
  const description = descriptionInput.value.trim();
  const category = categorySelect.value || null;
  const submitBtn = form.querySelector('button[type="submit"]');

  alertContainer.innerHTML = '';
  submitBtn.disabled = true;
  submitBtn.textContent = 'Submitting...';

  try {
    const data = await apiRequest('/policies', {
      method: 'POST',
      body: JSON.stringify({ 
        title, 
        description,
        category_id: category
      }),
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">
        ${data.message} Your policy is now pending admin review.
      </div>
    `;

    form.reset();
    document.getElementById('title-counter').textContent = '0/200';
    document.getElementById('description-counter').textContent = '0/2000';

    if (typeof plausible !== 'undefined') {
      plausible('Submit Policy', { 
        props: { 
          category: category || 'uncategorized',
          title_length: title.length,
          description_length: description.length
        } 
      });
    }

    setTimeout(() => {
      window.location.href = '/dashboard';
    }, 2000);

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
    
    if (typeof plausible !== 'undefined') {
      plausible('Submit Policy Error', { props: { error: error.message } });
    }
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = 'Submit Policy';
  }
});

// Initialize
loadCategories();