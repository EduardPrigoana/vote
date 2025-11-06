requireAuth();

if (isAdmin()) {
  document.getElementById('admin-link').style.display = 'block';
}

if (isSuperuser()) {
  document.getElementById('superuser-link').style.display = 'block';
}

const policiesContainer = document.getElementById('policies-container');
const alertContainer = document.getElementById('alert-container');
const voteStatusCache = {};
let categories = [];

// Load categories
async function loadCategories() {
  try {
    categories = await apiRequest('/categories?lang=en');
    
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

// Load policies with filters
async function loadPolicies() {
  const search = document.getElementById('search-input')?.value || '';
  const category = document.getElementById('category-filter')?.value || '';
  const status = document.getElementById('status-filter')?.value || '';
  const sort = document.getElementById('sort-filter')?.value || 'newest';

  const params = new URLSearchParams();
  if (search) params.append('search', search);
  if (category) params.append('category', category);
  if (status) params.append('status', status);
  if (sort) params.append('sort', sort);
  params.append('lang', 'en');

  try {
    const policies = await apiRequest(`/policies?${params.toString()}`);

    if (policies.length === 0) {
      policiesContainer.innerHTML = `
        <div class="empty-state">
          <h3>No policies found</h3>
          <p>Try adjusting your filters or be the first to submit a policy!</p>
          <a href="/submit" class="btn btn-primary">Submit Policy</a>
        </div>
      `;
      return;
    }

    // Load vote status for each policy
    for (const policy of policies) {
      try {
        const status = await apiRequest(`/votes/status/${policy.id}`);
        voteStatusCache[policy.id] = status;
      } catch (error) {
        console.error(`Failed to load vote status for ${policy.id}`, error);
      }
    }

    policiesContainer.innerHTML = policies.map(policy => renderPolicyCard(policy)).join('');
    attachVoteHandlers();
  } catch (error) {
    policiesContainer.innerHTML = `
      <div class="alert alert-error">
        <strong>Error:</strong> ${error.message}
      </div>
    `;
  }
}

function renderPolicyCard(policy) {
  const voteStatus = voteStatusCache[policy.id] || { device_has_voted: false };
  const deviceHasVoted = voteStatus.device_has_voted;
  const totalVotes = (policy.upvotes || 0) + (policy.downvotes || 0);
  const supportPercentage = totalVotes > 0 ? ((policy.upvotes || 0) / totalVotes * 100).toFixed(1) : 0;

  return `
    <div class="card">
      <div class="card-header">
        <div style="flex: 1;">
          <h3 class="card-title">${escapeHtml(policy.title)}</h3>
          ${policy.category_name ? `<small style="color: var(--muted-foreground);">${escapeHtml(policy.category_name)}</small>` : ''}
        </div>
        <span class="badge badge-${policy.status}">${policy.status}</span>
      </div>
      
      <p>${escapeHtml(policy.description)}</p>
      
      ${policy.admin_comment ? `
        <div class="info-box">
          <strong>Admin:</strong> ${escapeHtml(policy.admin_comment)}
        </div>
      ` : ''}
      
      ${totalVotes > 0 ? `
        <div class="vote-progress">
          <div class="vote-progress-label">
            <span>Support: ${supportPercentage}%</span>
            <span>Oppose: ${(100 - supportPercentage).toFixed(1)}%</span>
          </div>
          <div class="vote-progress-bar-container">
            <div class="vote-progress-bar" style="width: ${supportPercentage}%"></div>
          </div>
          <div class="vote-stats">
            <span class="vote-stat">
              <span class="vote-stat-number">${policy.upvotes || 0}</span> support
            </span>
            <span class="vote-stat">
              <span class="vote-stat-number">${policy.downvotes || 0}</span> oppose
            </span>
          </div>
        </div>
      ` : `
        <div style="margin-top: 1rem; padding: 0.75rem; background: var(--muted); border-radius: calc(var(--radius) - 2px); text-align: center;">
          <small style="color: var(--muted-foreground);">No votes yet — be the first!</small>
        </div>
      `}
      
      <div class="vote-container">
        <button 
          class="vote-btn upvote" 
          data-policy-id="${policy.id}" 
          data-vote-type="upvote"
          ${deviceHasVoted ? 'disabled' : ''}
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
          </svg>
          <span class="vote-count">${policy.upvotes || 0}</span>
        </button>
        
        <button 
          class="vote-btn downvote" 
          data-policy-id="${policy.id}" 
          data-vote-type="downvote"
          ${deviceHasVoted ? 'disabled' : ''}
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
          </svg>
          <span class="vote-count">${policy.downvotes || 0}</span>
        </button>
      </div>
      
      ${deviceHasVoted ? `
        <div style="margin-top: 0.75rem;">
          <small>✓ You've voted from this device</small>
        </div>
      ` : ''}
    </div>
  `;
}

function attachVoteHandlers() {
  document.querySelectorAll('.vote-btn').forEach(btn => {
    btn.addEventListener('click', async (e) => {
      const button = e.currentTarget;
      const policyId = button.dataset.policyId;
      const voteType = button.dataset.voteType;

      if (button.disabled) return;

      button.disabled = true;

      try {
        await apiRequest('/votes', {
          method: 'POST',
          body: JSON.stringify({
            policy_id: policyId,
            vote_type: voteType,
          }),
        });

        alertContainer.innerHTML = `
          <div class="alert alert-success">
            <strong>Success!</strong> Your vote has been recorded.
          </div>
        `;

        setTimeout(() => {
          alertContainer.innerHTML = '';
        }, 3000);

        if (typeof plausible !== 'undefined') {
          plausible('Vote', { props: { type: voteType } });
        }

        loadPolicies();

      } catch (error) {
        alertContainer.innerHTML = `
          <div class="alert alert-error">
            <strong>Error:</strong> ${error.message}
          </div>
        `;
        
        setTimeout(() => {
          alertContainer.innerHTML = '';
        }, 5000);

        button.disabled = false;
      }
    });
  });
}

// Search & filter handlers
let searchTimeout;
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
  
  if (typeof plausible !== 'undefined') {
    plausible('Filter Category');
  }
});

document.getElementById('status-filter')?.addEventListener('change', () => {
  loadPolicies();
  
  if (typeof plausible !== 'undefined') {
    plausible('Filter Status');
  }
});

document.getElementById('sort-filter')?.addEventListener('change', () => {
  loadPolicies();
  
  if (typeof plausible !== 'undefined') {
    plausible('Change Sort');
  }
});

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Initialize
loadCategories();
loadPolicies();