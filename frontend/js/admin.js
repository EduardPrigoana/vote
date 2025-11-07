requireAdmin();

if (isSuperuser()) {
  document.getElementById('superuser-link').style.display = 'block';
}

const policiesContainer = document.getElementById('policies-container');
const alertContainer = document.getElementById('alert-container');
const statusFilter = document.getElementById('status-filter');

let currentPolicyId = null;
let currentAction = null;
let policyToDelete = null;
let selectedPolicies = new Set();
let editingPolicyId = null;

function switchTab(tab) {
  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.classList.remove('active');
  });
  document.querySelector(`[data-tab="${tab}"]`).classList.add('active');

  document.querySelectorAll('.tab-content').forEach(content => {
    content.style.display = 'none';
  });

  if (tab === 'policies') {
    document.getElementById('policies-tab').style.display = 'block';
  } else if (tab === 'analytics') {
    document.getElementById('analytics-tab').style.display = 'block';
    loadAnalytics();
  } else if (tab === 'audit') {
    document.getElementById('audit-tab').style.display = 'block';
    loadAuditLog();
  }

  if (typeof plausible !== 'undefined') {
    plausible('Admin Tab Switch', { props: { tab: tab } });
  }
}

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
          <p>Adjust your filter.</p>
        </div>
      `;
      return;
    }

    policiesContainer.innerHTML = `
      <table>
        <thead>
          <tr>
            <th style="width: 40px;">
              <input type="checkbox" id="select-all" onchange="toggleSelectAll(this.checked)">
            </th>
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
                  <input type="checkbox" class="policy-checkbox" value="${policy.id}" onchange="togglePolicySelection('${policy.id}', this.checked)">
                </td>
                <td>
                  <strong>${escapeHtml(policy.title)}</strong><br>
                  <small style="color: var(--muted-foreground);">${escapeHtml(policy.description.substring(0, 100))}...</small>
                  ${policy.admin_comment ? `<br><small style="color: var(--foreground); opacity: 0.8;">${escapeHtml(policy.admin_comment)}</small>` : ''}
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
                    ${renderActionButtons(policy)}
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

function renderActionButtons(policy) {
  let buttons = '';
  
  buttons += `<button class="btn btn-secondary btn-sm" onclick="openEditModal('${policy.id}')">Edit</button>`;
  
  if (policy.status !== 'approved') {
    buttons += `<button class="btn btn-success btn-sm" onclick="updateStatus('${policy.id}', 'approved')">Approve</button>`;
  }
  if (policy.status !== 'in_progress' && policy.status === 'approved') {
    buttons += `<button class="btn btn-secondary btn-sm" onclick="updateStatus('${policy.id}', 'in_progress')">In Progress</button>`;
  }
  if (policy.status !== 'completed' && policy.status === 'in_progress') {
    buttons += `<button class="btn btn-success btn-sm" onclick="updateStatus('${policy.id}', 'completed')">Complete</button>`;
  }
  if (policy.status !== 'uncertain') {
    buttons += `<button class="btn btn-warning btn-sm" onclick="updateStatus('${policy.id}', 'uncertain')">Uncertain</button>`;
  }
  if (policy.status !== 'rejected') {
    buttons += `<button class="btn btn-danger btn-sm" onclick="updateStatus('${policy.id}', 'rejected')">Reject</button>`;
  }
  if (policy.status === 'approved' || policy.status === 'in_progress') {
    buttons += `<button class="btn btn-secondary btn-sm" onclick="updateStatus('${policy.id}', 'on_hold')">Hold</button>`;
  }
  if (policy.status === 'approved' || policy.status === 'uncertain') {
    buttons += `<button class="btn btn-secondary btn-sm" onclick="updateStatus('${policy.id}', 'cannot_implement')">Can't Do</button>`;
  }
  
  buttons += `<button class="btn btn-secondary btn-sm" onclick="addComment('${policy.id}')">Comment</button>`;
  buttons += `<button class="btn btn-danger btn-sm" onclick="confirmDelete('${policy.id}', '${escapeHtml(policy.title)}')">Delete</button>`;
  
  return buttons;
}

async function openEditModal(policyId) {
  try {
    const policy = await apiRequest(`/admin/policies/${policyId}`);
    const categories = await apiRequest('/categories');
    
    editingPolicyId = policyId;
    
    document.getElementById('edit-policy-title').value = policy.title;
    document.getElementById('edit-policy-description').value = policy.description;
    
    const categorySelect = document.getElementById('edit-policy-category');
    categorySelect.innerHTML = '<option value="">Select category...</option>';
    categories.forEach(cat => {
      const option = document.createElement('option');
      option.value = cat.id;
      option.textContent = cat.name;
      if (policy.category_id === cat.id) {
        option.selected = true;
      }
      categorySelect.appendChild(option);
    });
    
    document.getElementById('edit-modal').style.display = 'block';
    
  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  }
}

function closeEditModal() {
  document.getElementById('edit-modal').style.display = 'none';
  editingPolicyId = null;
}

document.getElementById('edit-policy-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  
  const title = document.getElementById('edit-policy-title').value.trim();
  const description = document.getElementById('edit-policy-description').value.trim();
  const category = document.getElementById('edit-policy-category').value || null;
  
  try {
    await apiRequest(`/admin/policies/${editingPolicyId}`, {
      method: 'PUT',
      body: JSON.stringify({
        title: title,
        description: description,
        category_id: category
      }),
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">Policy updated</div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    closeEditModal();
    loadPolicies();
    loadStats();

    if (typeof plausible !== 'undefined') {
      plausible('Edit Policy');
    }

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  }
});

function toggleSelectAll(checked) {
  document.querySelectorAll('.policy-checkbox').forEach(checkbox => {
    checkbox.checked = checked;
    togglePolicySelection(checkbox.value, checked);
  });
}

function togglePolicySelection(policyId, selected) {
  if (selected) {
    selectedPolicies.add(policyId);
  } else {
    selectedPolicies.delete(policyId);
  }
  
  document.getElementById('bulk-delete-btn').style.display = 
    selectedPolicies.size > 0 ? 'inline-flex' : 'none';
}

async function bulkDelete() {
  if (selectedPolicies.size === 0) return;
  
  if (!confirm(`Delete ${selectedPolicies.size} selected policies? This cannot be undone.`)) {
    return;
  }

  try {
    await apiRequest('/admin/policies/bulk', {
      method: 'POST',
      body: JSON.stringify({
        policy_ids: Array.from(selectedPolicies),
        action: 'delete'
      })
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">Successfully deleted ${selectedPolicies.size} policies</div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    selectedPolicies.clear();
    loadPolicies();
    loadStats();

    if (typeof plausible !== 'undefined') {
      plausible('Bulk Delete');
    }

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  }
}

async function exportCSV() {
  const ids = selectedPolicies.size > 0 ? Array.from(selectedPolicies).join(',') : '';
  const url = `/api/v1/admin/export/csv${ids ? `?ids=${ids}` : ''}`;
  
  try {
    const response = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${getAuthToken()}`
      }
    });
    
    if (!response.ok) {
      throw new Error('Export failed');
    }
    
    const blob = await response.blob();
    const downloadUrl = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = downloadUrl;
    a.download = `policies_${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(downloadUrl);
    document.body.removeChild(a);
    
    if (typeof plausible !== 'undefined') {
      plausible('Export CSV', { props: { selected: selectedPolicies.size } });
    }
  } catch (error) {
    alert('Export failed: ' + error.message);
  }
}

async function exportExcel() {
  const ids = selectedPolicies.size > 0 ? Array.from(selectedPolicies).join(',') : '';
  const url = `/api/v1/admin/export/xlsx${ids ? `?ids=${ids}` : ''}`;
  
  try {
    const response = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${getAuthToken()}`
      }
    });
    
    if (!response.ok) {
      throw new Error('Export failed');
    }
    
    const blob = await response.blob();
    const downloadUrl = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = downloadUrl;
    a.download = `policies_${new Date().toISOString().split('T')[0]}.xlsx`;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(downloadUrl);
    document.body.removeChild(a);
    
    if (typeof plausible !== 'undefined') {
      plausible('Export Excel', { props: { selected: selectedPolicies.size } });
    }
  } catch (error) {
    alert('Export failed: ' + error.message);
  }
}

async function loadAnalytics() {
  const container = document.getElementById('analytics-tab');
  container.innerHTML = '<div class="loading">Loading analytics...</div>';

  try {
    const analytics = await apiRequest('/admin/analytics');

    container.innerHTML = `
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-value">${analytics.total_policies}</div>
          <div class="stat-label">Total Policies</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">${analytics.total_votes}</div>
          <div class="stat-label">Total Votes</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">${analytics.participation_rate.toFixed(1)}%</div>
          <div class="stat-label">Participation Rate</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">${analytics.policy_success_rate.toFixed(1)}%</div>
          <div class="stat-label">Success Rate</div>
        </div>
      </div>

      <h2 style="margin-top: 2rem; margin-bottom: 1rem;">Top Classrooms</h2>
      <table>
        <thead>
          <tr>
            <th>Code</th>
            <th>Votes</th>
            <th>Policies</th>
            <th>Engagement</th>
          </tr>
        </thead>
        <tbody>
          ${analytics.top_classrooms.map(classroom => `
            <tr>
              <td><strong>${escapeHtml(classroom.login_code)}</strong></td>
              <td>${classroom.vote_count}</td>
              <td>${classroom.policy_count}</td>
              <td><strong>${classroom.engagement_score}</strong></td>
            </tr>
          `).join('')}
        </tbody>
      </table>

      <h2 style="margin-top: 2rem; margin-bottom: 1rem;">Categories</h2>
      <table>
        <thead>
          <tr>
            <th>Category</th>
            <th>Policies</th>
            <th>Votes</th>
          </tr>
        </thead>
        <tbody>
          ${analytics.category_distribution.map(cat => `
            <tr>
              <td><strong>${escapeHtml(cat.category_name)}</strong></td>
              <td>${cat.policy_count}</td>
              <td>${cat.vote_count}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    `;
  } catch (error) {
    container.innerHTML = `<div class="alert alert-error">${error.message}</div>`;
  }
}

async function loadAuditLog() {
  const container = document.getElementById('audit-container');
  container.innerHTML = '<div class="loading">Loading audit log...</div>';

  try {
    const logs = await apiRequest('/admin/audit-log?limit=100');

    if (logs.length === 0) {
      container.innerHTML = '<div class="empty-state"><p>No audit logs</p></div>';
      return;
    }

    container.innerHTML = `
      <table>
        <thead>
          <tr>
            <th>Time</th>
            <th>User</th>
            <th>Action</th>
            <th>Entity</th>
            <th>Details</th>
          </tr>
        </thead>
        <tbody>
          ${logs.map(log => `
            <tr>
              <td><small>${new Date(log.created_at).toLocaleString()}</small></td>
              <td>${log.user_code || 'System'}</td>
              <td><code>${log.action}</code></td>
              <td>${log.entity_type}</td>
              <td><small>${log.details ? JSON.stringify(log.details) : '-'}</small></td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    `;

  } catch (error) {
    container.innerHTML = `<div class="alert alert-error">${error.message}</div>`;
  }
}

function updateStatus(policyId, status) {
  currentPolicyId = policyId;
  currentAction = status;
  
  document.getElementById('modal-title').textContent = `${status.charAt(0).toUpperCase() + status.slice(1).replace('_', ' ')} Policy`;
  document.getElementById('admin-comment').value = '';
  document.getElementById('modal').style.display = 'block';
}

function addComment(policyId) {
  currentPolicyId = policyId;
  currentAction = 'comment';
  
  document.getElementById('modal-title').textContent = 'Add Comment';
  document.getElementById('admin-comment').value = '';
  document.getElementById('modal').style.display = 'block';
}

function confirmDelete(policyId, policyTitle) {
  policyToDelete = policyId;
  const tempDiv = document.createElement('div');
  tempDiv.innerHTML = policyTitle;
  const decodedTitle = tempDiv.textContent;
  
  document.getElementById('delete-policy-title').textContent = decodedTitle;
  document.getElementById('delete-modal').style.display = 'block';
}

function closeModal() {
  document.getElementById('modal').style.display = 'none';
  currentPolicyId = null;
  currentAction = null;
}

function closeDeleteModal() {
  document.getElementById('delete-modal').style.display = 'none';
  policyToDelete = null;
}

document.getElementById('modal-confirm').addEventListener('click', async () => {
  const comment = document.getElementById('admin-comment').value.trim();

  try {
    if (currentAction === 'comment') {
      await apiRequest(`/admin/policies/${currentPolicyId}/comment`, {
        method: 'POST',
        body: JSON.stringify({ comment: comment || null }),
      });
    } else {
      await apiRequest(`/admin/policies/${currentPolicyId}/status`, {
        method: 'POST',
        body: JSON.stringify({
          status: currentAction,
          comment: comment || null,
        }),
      });
    }

    alertContainer.innerHTML = `
      <div class="alert alert-success">Policy updated</div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    closeModal();
    loadPolicies();
    loadStats();

    if (typeof plausible !== 'undefined') {
      plausible('Admin Action', { props: { action: currentAction } });
    }

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  }
});

document.getElementById('confirm-delete-btn').addEventListener('click', async () => {
  if (!policyToDelete) return;

  try {
    await apiRequest(`/admin/policies/${policyToDelete}`, {
      method: 'DELETE',
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">Policy deleted</div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    closeDeleteModal();
    loadPolicies();
    loadStats();

    if (typeof plausible !== 'undefined') {
      plausible('Delete Policy');
    }

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
    closeDeleteModal();
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