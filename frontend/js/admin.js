requireAdmin();

const policiesContainer = document.getElementById('policies-container');
const alertContainer = document.getElementById('alert-container');
const statusFilter = document.getElementById('status-filter');

let currentPolicyId = null;
let currentAction = null;

// Load stats
async function loadStats() {
  try {
    const stats = await apiRequest('/admin/stats');
    document.getElementById('stat-total').textContent = stats.total_policies;
    document.getElementById('stat-pending').textContent = stats.pending_policies;
    document.getElementById('stat-votes').textContent = stats.total_votes;
    document.getElementById('stat-students').textContent = stats.active_students;
  } catch (error) {
    console.error('Failed to load stats:', error);
  }
}

async function loadPolicies() {
  const status = statusFilter.value;
  const queryString = status ? `?status=${status}` : '';

  try {
    const policies = await apiRequest(`/admin/policies${queryString}`);

    if (policies.length === 0) {
      policiesContainer.innerHTML = `
        <div class="empty-state">
          <h3>No policies found</h3>
          <p>Try adjusting your filter.</p>
        </div>
      `;
      return;
    }

    policiesContainer.innerHTML = `
      <table>
        <thead>
          <tr>
            <th>Title</th>
            <th>Status</th>
            <th>Votes</th>
            <th>Submitted</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          ${policies.map(policy => {
            const totalVotes = (policy.upvotes || 0) + (policy.downvotes || 0);
            const supportPercentage = totalVotes > 0 ? ((policy.upvotes || 0) / totalVotes * 100).toFixed(1) : 0;
            
            return `
              <tr>
                <td>
                  <strong>${escapeHtml(policy.title)}</strong><br>
                  <small style="color: var(--muted-foreground);">${escapeHtml(policy.description.substring(0, 100))}...</small>
                  ${policy.admin_comment ? `<br><small style="color: var(--foreground); opacity: 0.8;">💬 ${escapeHtml(policy.admin_comment)}</small>` : ''}
                </td>
                <td>
                  <span class="badge badge-${policy.status}">${policy.status}</span>
                </td>
                <td>
                  <div style="min-width: 150px;">
                    ${totalVotes > 0 ? `
                      <div style="margin-bottom: 0.5rem;">
                        <div style="display: flex; justify-content: space-between; font-size: 0.75rem; margin-bottom: 0.25rem;">
                          <span>${policy.upvotes || 0} up</span>
                          <span>${policy.downvotes || 0} down</span>
                        </div>
                        <div style="width: 100%; height: 6px; background: var(--muted); border-radius: 9999px; overflow: hidden; border: 1px solid var(--border);">
                          <div style="height: 100%; background: var(--foreground); width: ${supportPercentage}%; transition: width 0.3s;"></div>
                        </div>
                      </div>
                      <small style="color: var(--muted-foreground);">
                        ${supportPercentage}% support (${totalVotes} total)
                      </small>
                    ` : `
                      <small style="color: var(--muted-foreground);">No votes</small>
                    `}
                  </div>
                </td>
                <td>
                  <small>${new Date(policy.created_at).toLocaleDateString()}</small>
                </td>
                <td>
                  <div style="display: flex; gap: 0.5rem; flex-wrap: wrap;">
                    ${policy.status !== 'approved' ? `
                      <button class="btn btn-success btn-sm" onclick="updateStatus('${policy.id}', 'approved')">
                        Approve
                      </button>
                    ` : ''}
                    ${policy.status !== 'uncertain' ? `
                      <button class="btn btn-warning btn-sm" onclick="updateStatus('${policy.id}', 'uncertain')">
                        Uncertain
                      </button>
                    ` : ''}
                    ${policy.status !== 'rejected' ? `
                      <button class="btn btn-danger btn-sm" onclick="updateStatus('${policy.id}', 'rejected')">
                        Reject
                      </button>
                    ` : ''}
                    <button class="btn btn-secondary btn-sm" onclick="addComment('${policy.id}')">
                      Comment
                    </button>
                    <button class="btn btn-danger btn-sm" onclick="confirmDelete('${policy.id}', '${escapeHtml(policy.title)}')">
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            `;
          }).join('')}
        </tbody>
      </table>
    `;

  } catch (error) {
    policiesContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  }
}

function updateStatus(policyId, status) {
  currentPolicyId = policyId;
  currentAction = status;
  
  document.getElementById('modal-title').textContent = `${status.charAt(0).toUpperCase() + status.slice(1)} Policy`;
  document.getElementById('admin-comment').value = '';
  document.getElementById('modal').style.display = 'block';
}

function addComment(policyId) {
  currentPolicyId = policyId;
  currentAction = 'comment';
  
  document.getElementById('modal-title').textContent = 'Add/Update Comment';
  document.getElementById('admin-comment').value = '';
  document.getElementById('modal').style.display = 'block';
}

function confirmDelete(policyId, policyTitle) {
  // Decode HTML entities for display
  const tempDiv = document.createElement('div');
  tempDiv.innerHTML = policyTitle;
  const decodedTitle = tempDiv.textContent;
  
  if (confirm(`Are you sure you want to permanently delete this policy?\n\n"${decodedTitle}"\n\nThis action cannot be undone and will delete all associated votes.`)) {
    deletePolicy(policyId);
  }
}

async function deletePolicy(policyId) {
  try {
    await apiRequest(`/admin/policies/${policyId}`, {
      method: 'DELETE',
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">
        <strong>Success!</strong> Policy deleted successfully.
      </div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    loadPolicies();
    loadStats();

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">
        <strong>Error:</strong> ${error.message}
      </div>
    `;
    
    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 5000);
  }
}

function closeModal() {
  document.getElementById('modal').style.display = 'none';
  currentPolicyId = null;
  currentAction = null;
}

document.getElementById('modal-confirm').addEventListener('click', async () => {
  const comment = document.getElementById('admin-comment').value.trim();

  try {
    if (currentAction === 'comment') {
      // Just add/update comment without changing status
      await apiRequest(`/admin/policies/${currentPolicyId}/comment`, {
        method: 'POST',
        body: JSON.stringify({
          comment: comment || null,
        }),
      });
    } else {
      // Update status with optional comment
      await apiRequest(`/admin/policies/${currentPolicyId}/status`, {
        method: 'POST',
        body: JSON.stringify({
          status: currentAction,
          comment: comment || null,
        }),
      });
    }

    alertContainer.innerHTML = `
      <div class="alert alert-success">
        <strong>Success!</strong> Policy updated successfully.
      </div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    closeModal();
    loadPolicies();
    loadStats();

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">
        <strong>Error:</strong> ${error.message}
      </div>
    `;
  }
});

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

statusFilter.addEventListener('change', loadPolicies);

loadStats();
loadPolicies();