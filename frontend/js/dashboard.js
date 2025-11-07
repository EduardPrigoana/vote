requireAuth();

if (isAdmin()) {
  document.getElementById('admin-link').style.display = 'block';
}

if (isSuperuser()) {
  document.getElementById('superuser-link').style.display = 'block';
}

const policiesContainer = document.getElementById('policies-container');
const alertContainer = document.getElementById('alert-container');
let searchTimeout;

async function loadCategories() {
  try {
    const categories = await apiRequest('/categories');
    const categoryFilter = document.getElementById('category-filter');
    categoryFilter.innerHTML = '<option value="">All Categories</option>';
    categories.forEach(cat => {
      const option = document.createElement('option');
      option.value = cat.id;
      option.textContent = cat.name;
      categoryFilter.appendChild(option);
    });
  } catch (error) {
    console.error('Failed to load categories:', error);
  }
}

async function loadPolicies() {
  const search = document.getElementById('search-input')?.value || '';
  const category = document.getElementById('category-filter')?.value || '';
  const status = document.getElementById('status-filter')?.value || '';
  const sort = document.getElementById('sort-filter')?.value || 'newest';

  const params = new URLSearchParams({ sort });
  if (search) params.append('search', search);
  if (category) params.append('category', category);
  if (status) params.append('status', status);

  try {
    const policies = await apiRequest(`/policies?${params.toString()}`);

    if (policies.length === 0) {
      policiesContainer.innerHTML = `
        <div class="empty-state">
          <h3>No policies found</h3>
          <p>Adjust filters or submit the first policy.</p>
          <a href="/submit" class="btn btn-primary">Submit Policy</a>
        </div>
      `;
      return;
    }

    policiesContainer.innerHTML = policies.map(renderPolicyCard).join('');
    attachVoteHandlers();

  } catch (error) {
    policiesContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  }
}

function renderPolicyCard(policy) {
  const deviceHasVoted = !!policy.current_user_vote;
  const totalVotes = (policy.upvotes || 0) + (policy.downvotes || 0);
  const supportPercentage = totalVotes > 0 ? ((policy.upvotes || 0) / totalVotes * 100).toFixed(1) : 0;
  
  return `
    <div class="card" id="policy-card-${policy.id}" data-policy-id="${policy.id}">
      <div class="card-header">
        <div style="flex: 1;">
          <h3 class="card-title">${escapeHtml(policy.title)}</h3>
          ${policy.category_name ? `<small style="color: var(--muted-foreground);">${escapeHtml(policy.category_name)}</small>` : ''}
        </div>
        <div style="display: flex; gap: 0.5rem; align-items: center;">
          <span class="badge badge-${policy.status}">${policy.status}</span>
          <button class="btn-icon" onclick="sharePolicy('${policy.id}', \`${escapeHtml(policy.title).replace(/`/g, '\\`')}\`)" title="Share policy">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"/><polyline points="16 6 12 2 8 6"/><line x1="12" y1="2" x2="12" y2="15"/></svg>
          </button>
          <a href="/policy/${policy.id}" class="btn-icon" title="View details">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
          </a>
        </div>
      </div>
      
      <p>${escapeHtml(policy.description)}</p>
      
      ${policy.admin_comment ? `<div class="info-box"><strong>Admin:</strong> ${escapeHtml(policy.admin_comment)}</div>` : ''}
      
      <div class="vote-progress">
        ${totalVotes > 0 ? `
            <div class="vote-progress-label">
              <span>Support: ${supportPercentage}%</span>
              <span>Oppose: ${(100 - supportPercentage).toFixed(1)}%</span>
            </div>
            <div class="vote-progress-bar-container">
              <div class="vote-progress-bar" style="width: ${supportPercentage}%"></div>
            </div>
        ` : `
            <div style="margin-bottom: 0.75rem; text-align:center;"><small>No votes yet</small></div>
        `}
        <div class="vote-stats">
          <span class="vote-stat">
            <span class="vote-stat-number" data-role="upvote-count">${policy.upvotes || 0}</span> support
          </span>
          <span class="vote-stat">
            <span class="vote-stat-number" data-role="downvote-count">${policy.downvotes || 0}</span> oppose
          </span>
        </div>
      </div>
      
      <div class="vote-actions">
          <div class="vote-container">
              <button class="vote-btn upvote" data-vote-type="upvote" ${deviceHasVoted ? 'disabled' : ''}>
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" /></svg>
                  <span class="vote-count" data-role="upvote-btn-count">${policy.upvotes || 0}</span>
              </button>
              <button class="vote-btn downvote" data-vote-type="downvote" ${deviceHasVoted ? 'disabled' : ''}>
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" /></svg>
                  <span class="vote-count" data-role="downvote-btn-count">${policy.downvotes || 0}</span>
              </button>
          </div>
          ${deviceHasVoted ? `<div class="vote-feedback"><small>Voted</small></div>` : '<div class="vote-feedback"></div>'}
      </div>
    </div>
  `;
}

function attachVoteHandlers() {
  policiesContainer.addEventListener('click', async (e) => {
    const button = e.target.closest('.vote-btn');
    if (!button || button.disabled) return;

    const card = button.closest('.card');
    const policyId = card.dataset.policyId;
    const voteType = button.dataset.voteType;

    const upvoteBtn = card.querySelector('.vote-btn.upvote');
    const downvoteBtn = card.querySelector('.vote-btn.downvote');

    upvoteBtn.disabled = true;
    downvoteBtn.disabled = true;

    try {
      await apiRequest('/votes', {
        method: 'POST',
        body: JSON.stringify({ policy_id: policyId, vote_type: voteType }),
      });

      card.querySelector('.vote-feedback').innerHTML = `<small>Vote recorded</small>`;

      showTempAlert('Vote recorded', 'success');

      if (typeof plausible !== 'undefined') {
        plausible('Vote', { props: { type: voteType } });
      }

    } catch (error) {
      showTempAlert(error.message, 'error', 5000);
      
      upvoteBtn.disabled = false;
      downvoteBtn.disabled = false;
    }
  });
}

function setupFilterHandlers() {
  document.getElementById('search-input')?.addEventListener('input', (e) => {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      loadPolicies();
      if (typeof plausible !== 'undefined' && e.target.value) {
        plausible('Search', { props: { query: e.target.value } });
      }
    }, 500);
  });

  document.getElementById('category-filter')?.addEventListener('change', () => {
    loadPolicies();
    if (typeof plausible !== 'undefined') plausible('Filter Category');
  });

  document.getElementById('status-filter')?.addEventListener('change', () => {
    loadPolicies();
    if (typeof plausible !== 'undefined') plausible('Filter Status');
  });

  document.getElementById('sort-filter')?.addEventListener('change', () => {
    loadPolicies();
    if (typeof plausible !== 'undefined') plausible('Change Sort');
  });
}

function showTempAlert(message, type, duration = 3000) {
  alertContainer.innerHTML = `<div class="alert alert-${type}">${message}</div>`;
  setTimeout(() => {
    alertContainer.innerHTML = '';
  }, duration);
}

function escapeHtml(text) {
  if (text === null || typeof text === 'undefined') return '';
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Share functionality
window.sharePolicy = function(policyId, title) {
  const url = `${window.location.origin}/policy/${policyId}`;
  
  if (navigator.share) {
    navigator.share({
      title: title,
      text: `Check out this policy: ${title}`,
      url: url
    }).then(() => {
      if (typeof plausible !== 'undefined') {
        plausible('Share Policy', { props: { method: 'native' } });
      }
    }).catch((error) => {
      if (error.name !== 'AbortError') {
        console.log('Share failed:', error);
        copyToClipboard(url);
        showTempAlert('Link copied to clipboard', 'success');
      }
    });
  } else {
    copyToClipboard(url);
    showTempAlert('Link copied to clipboard', 'success');
    
    if (typeof plausible !== 'undefined') {
      plausible('Share Policy', { props: { method: 'clipboard' } });
    }
  }
}

function copyToClipboard(text) {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text).catch(() => {
      fallbackCopyToClipboard(text);
    });
  } else {
    fallbackCopyToClipboard(text);
  }
}

function fallbackCopyToClipboard(text) {
  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.position = 'fixed';
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.select();
  try {
    document.execCommand('copy');
  } catch (err) {
    console.error('Failed to copy:', err);
  }
  document.body.removeChild(textarea);
}

loadCategories();
loadPolicies();
setupFilterHandlers();