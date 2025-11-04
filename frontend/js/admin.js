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
          ${policies.map(policy => `
            <tr>
              <td>
                <strong>${escapeHtml(policy.title)}</strong><br>
                <small style="color: var(--text-muted);">${escapeHtml(policy.description.substring(0, 100))}...</small>
                ${policy.admin_comment ? `<br><small style="color: var(--primary);">💬 ${escapeHtml(policy.admin_comment)}</small>` : ''}
              </td>
              <td>
                <span class="badge badge-${policy.status}">${policy.status}</span>
              </td>
              <td>
                👍 ${policy.upvotes || 0} / 👎 ${policy.downvotes || 0}
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
                </div>
              </td>
            </tr>
          `).join('')}
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
      <div class="alert alert-success">Policy updated successfully!</div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    closeModal();
    loadPolicies();
    loadStats();

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
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