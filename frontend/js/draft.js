const DRAFT_KEY = 'policy_draft';

function saveDraft(title, description, categoryId) {
  const draft = {
    title: title || '',
    description: description || '',
    category_id: categoryId || '',
    saved_at: new Date().toISOString()
  };
  
  localStorage.setItem(DRAFT_KEY, JSON.stringify(draft));
}

function loadDraft() {
  const draftStr = localStorage.getItem(DRAFT_KEY);
  if (!draftStr) return null;
  
  try {
    return JSON.parse(draftStr);
  } catch (e) {
    return null;
  }
}

function clearDraft() {
  localStorage.removeItem(DRAFT_KEY);
}

function hasDraft() {
  return !!localStorage.getItem(DRAFT_KEY);
}