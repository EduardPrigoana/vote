requireAuth();

if (isAdmin()) {
  document.getElementById('admin-link').style.display = 'block';
}

const policiesContainer = document.getElementById('policies-container');
const alertContainer = document.getElementById('alert-container');
const voteStatusCache = {};

async function loadPolicies() {
  try {
    const policies = await apiRequest('/policies');

    if (policies.length === 0) {
      policiesContainer.innerHTML = `
        <div class="empty-state">
          <h3>No policies yet</h3>
          <p>Be the first to submit a policy idea!</p>
          <a href="/submit.html" class="btn btn-primary">Submit Policy</a>
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

  return `
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">${escapeHtml(policy.title)}</h3>
        <span class="badge badge-${policy.status}">${policy.status}</span>
      </div>
      <p>${escapeHtml(policy.description)}</p>
      ${policy.admin_comment ? `
        <div class="info-box">
          <strong>Admin:</strong> ${escapeHtml(policy.admin_comment)}
        </div>
      ` : ''}
      
      <div class="vote-container">
        <button 
          class="vote-btn upvote" 
          data-policy-id="${policy.id}" 
          data-vote-type="upvote"
          aria-label="Upvote this policy"
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
          aria-label="Downvote this policy"
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

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

loadPolicies();