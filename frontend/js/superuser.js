requireSuperuser();

const usersContainer = document.getElementById('users-container');
const alertContainer = document.getElementById('alert-container');
const createForm = document.getElementById('create-user-form');
const editForm = document.getElementById('edit-user-form');

let userToDelete = null;

async function loadUsers() {
  try {
    const users = await apiRequest('/superuser/users');

    if (users.length === 0) {
      usersContainer.innerHTML = `
        <div class="empty-state">
          <h3>No users</h3>
          <p>Create the first user above.</p>
        </div>
      `;
      return;
    }

    usersContainer.innerHTML = `
      <table>
        <thead>
          <tr>
            <th>Code</th>
            <th>Role</th>
            <th>Status</th>
            <th>Created</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          ${users.map(user => `
            <tr>
              <td><strong>${escapeHtml(user.login_code || 'N/A')}</strong></td>
              <td>
                <span class="badge badge-${user.role === 'superuser' ? 'approved' : user.role === 'admin' ? 'uncertain' : 'pending'}">
                  ${user.role}
                </span>
              </td>
              <td>
                <span style="color: ${user.is_active ? 'var(--success)' : 'var(--danger)'};">
                  ${user.is_active ? '● Active' : '○ Inactive'}
                </span>
              </td>
              <td>
                <small>${new Date(user.created_at).toLocaleDateString()}</small>
              </td>
              <td>
                <div style="display: flex; gap: 0.5rem; flex-wrap: wrap;">
                  <button class="btn btn-secondary btn-sm" onclick="toggleUserStatus('${user.id}')">
                    ${user.is_active ? 'Deactivate' : 'Activate'}
                  </button>
                  <button class="btn btn-secondary btn-sm" onclick="openEditModal('${user.id}', '${escapeHtml(user.login_code || '')}', '${user.role}', ${user.is_active})">
                    Edit
                  </button>
                  <button class="btn btn-danger btn-sm" onclick="confirmDelete('${user.id}', '${escapeHtml(user.login_code || '')}')">
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    `;

  } catch (error) {
    usersContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  }
}

createForm.addEventListener('submit', async (e) => {
  e.preventDefault();

  const loginCode = document.getElementById('login-code').value.trim();
  const role = document.getElementById('role').value;
  const isActive = document.getElementById('is-active').value === 'true';

  const submitBtn = createForm.querySelector('button[type="submit"]');
  submitBtn.disabled = true;
  submitBtn.textContent = 'Creating...';

  try {
    await apiRequest('/superuser/users', {
      method: 'POST',
      body: JSON.stringify({
        login_code: loginCode,
        role: role,
        is_active: isActive,
      }),
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">User created</div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    createForm.reset();
    loadUsers();

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = 'Create User';
  }
});

async function toggleUserStatus(userId) {
  try {
    await apiRequest(`/superuser/users/${userId}/toggle`, {
      method: 'POST',
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">Status updated</div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    loadUsers();

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  }
}

function openEditModal(userId, loginCode, role, isActive) {
  document.getElementById('edit-user-id').value = userId;
  document.getElementById('edit-login-code').value = loginCode;
  document.getElementById('edit-role').value = role;
  document.getElementById('edit-is-active').value = isActive.toString();
  document.getElementById('edit-modal').style.display = 'block';
}

function closeEditModal() {
  document.getElementById('edit-modal').style.display = 'none';
}

editForm.addEventListener('submit', async (e) => {
  e.preventDefault();

  const userId = document.getElementById('edit-user-id').value;
  const loginCode = document.getElementById('edit-login-code').value.trim();
  const role = document.getElementById('edit-role').value;
  const isActive = document.getElementById('edit-is-active').value === 'true';

  const submitBtn = editForm.querySelector('button[type="submit"]');
  submitBtn.disabled = true;
  submitBtn.textContent = 'Saving...';

  try {
    await apiRequest(`/superuser/users/${userId}`, {
      method: 'PUT',
      body: JSON.stringify({
        login_code: loginCode,
        role: role,
        is_active: isActive,
      }),
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">User updated</div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    closeEditModal();
    loadUsers();

  } catch (error) {
    alertContainer.innerHTML = `
      <div class="alert alert-error">${error.message}</div>
    `;
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = 'Save';
  }
});

function confirmDelete(userId, loginCode) {
  userToDelete = userId;
  document.getElementById('delete-user-code').textContent = loginCode;
  document.getElementById('delete-modal').style.display = 'block';
}

function closeDeleteModal() {
  document.getElementById('delete-modal').style.display = 'none';
  userToDelete = null;
}

document.getElementById('confirm-delete-btn').addEventListener('click', async () => {
  if (!userToDelete) return;

  try {
    await apiRequest(`/superuser/users/${userToDelete}`, {
      method: 'DELETE',
    });

    alertContainer.innerHTML = `
      <div class="alert alert-success">User deleted</div>
    `;

    setTimeout(() => {
      alertContainer.innerHTML = '';
    }, 3000);

    closeDeleteModal();
    loadUsers();

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

loadUsers();