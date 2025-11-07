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
const DRAFT_KEY = 'policy_draft';
const AUTO_SAVE_INTERVAL = 30000;
let autoSaveTimer = null;
let draftLoaded = false;

async function loadCategories() {
  try {
    const categories = await apiRequest('/categories');
    
    categorySelect.innerHTML = '<option value="">Select category...</option>';
    
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

function saveDraft() {
  const draft = {
    title: titleInput.value.trim(),
    description: descriptionInput.value.trim(),
    category_id: categorySelect.value || null,
    saved_at: new Date().toISOString()
  };
  
  if (draft.title || draft.description) {
    localStorage.setItem(DRAFT_KEY, JSON.stringify(draft));
    showDraftIndicator('Draft saved');
  }
}

function loadDraft() {
  const draftStr = localStorage.getItem(DRAFT_KEY);
  if (!draftStr || draftLoaded) return;
  
  try {
    const draft = JSON.parse(draftStr);
    const savedDate = new Date(draft.saved_at);
    const now = new Date();
    const hoursSince = (now - savedDate) / (1000 * 60 * 60);
    
    if (hoursSince > 24) {
      clearDraft();
      return;
    }
    
    if (draft.title || draft.description) {
      if (confirm('You have an unsaved draft. Would you like to restore it?')) {
        titleInput.value = draft.title || '';
        descriptionInput.value = draft.description || '';
        if (draft.category_id) {
          categorySelect.value = draft.category_id;
        }
        updateCharacterCounters();
        draftLoaded = true;
        showDraftIndicator('Draft restored');
      } else {
        clearDraft();
      }
    }
  } catch (e) {
    clearDraft();
  }
}

function clearDraft() {
  localStorage.removeItem(DRAFT_KEY);
  showDraftIndicator('Draft cleared');
}

function showDraftIndicator(message) {
  const indicator = document.getElementById('draft-indicator');
  if (indicator) {
    indicator.textContent = message;
    indicator.style.opacity = '1';
    setTimeout(() => {
      indicator.style.opacity = '0';
    }, 2000);
  }
}

function startAutoSave() {
  stopAutoSave();
  autoSaveTimer = setInterval(() => {
    saveDraft();
  }, AUTO_SAVE_INTERVAL);
}

function stopAutoSave() {
  if (autoSaveTimer) {
    clearInterval(autoSaveTimer);
    autoSaveTimer = null;
  }
}

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

function updateCharacterCounters() {
  titleInput.dispatchEvent(new Event('input'));
  descriptionInput.dispatchEvent(new Event('input'));
}

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
        ${data.message} Pending admin review.
      </div>
    `;

    form.reset();
    document.getElementById('title-counter').textContent = '0/200';
    document.getElementById('description-counter').textContent = '0/2000';
    clearDraft();
    stopAutoSave();

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

loadCategories();
loadDraft();
startAutoSave();

window.addEventListener('beforeunload', () => {
  stopAutoSave();
  saveDraft();
});