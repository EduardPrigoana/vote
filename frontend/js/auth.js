// Authentication utilities
const AUTH_TOKEN_KEY = 'auth_token';
const USER_ROLE_KEY = 'user_role';
const USER_ID_KEY = 'user_id';
const DEVICE_ID_KEY = 'device_fingerprint';

function saveAuth(token, role, userId) {
  localStorage.setItem(AUTH_TOKEN_KEY, token);
  localStorage.setItem(USER_ROLE_KEY, role);
  localStorage.setItem(USER_ID_KEY, userId);
}

function getAuthToken() {
  return localStorage.getItem(AUTH_TOKEN_KEY);
}

function getUserRole() {
  return localStorage.getItem(USER_ROLE_KEY);
}

function getUserId() {
  return localStorage.getItem(USER_ID_KEY);
}

function clearAuth() {
  localStorage.removeItem(AUTH_TOKEN_KEY);
  localStorage.removeItem(USER_ROLE_KEY);
  localStorage.removeItem(USER_ID_KEY);
}

function isAuthenticated() {
  return !!getAuthToken();
}

function isAdmin() {
  return getUserRole() === 'admin';
}

function logout() {
  clearAuth();
  window.location.href = '/login';
}

function requireAuth() {
  if (!isAuthenticated()) {
    window.location.href = '/login';
    return false;
  }
  return true;
}

function requireAdmin() {
  if (!isAuthenticated() || !isAdmin()) {
    window.location.href = '/dashboard';
    return false;
  }
  return true;
}

// Generate or get device fingerprint
function getDeviceFingerprint() {
  let fingerprint = localStorage.getItem(DEVICE_ID_KEY);
  
  if (!fingerprint) {
    fingerprint = 'dev_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    ctx.textBaseline = 'top';
    ctx.font = '14px Arial';
    ctx.fillText('fingerprint', 2, 2);
    
    fingerprint += '_' + canvas.toDataURL().slice(-50).replace(/[^a-zA-Z0-9]/g, '');
    
    localStorage.setItem(DEVICE_ID_KEY, fingerprint);
  }
  
  return fingerprint;
}

// API helper
async function apiRequest(endpoint, options = {}) {
  const token = getAuthToken();
  const headers = {
    'Content-Type': 'application/json',
    'X-Device-Fingerprint': getDeviceFingerprint(),
    ...options.headers,
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`/api/v1${endpoint}`, {
    ...options,
    headers,
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || 'Request failed');
  }

  return data;
}