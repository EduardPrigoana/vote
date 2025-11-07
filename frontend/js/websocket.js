let ws = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;

function connectWebSocket() {
  if (!isAuthenticated()) {
    return;
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/ws`;

  try {
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('WebSocket connected');
      reconnectAttempts = 0;
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        handleWebSocketMessage(data);
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      
      if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
        reconnectAttempts++;
        setTimeout(() => {
          console.log(`Reconnecting... (attempt ${reconnectAttempts})`);
          connectWebSocket();
        }, 3000 * reconnectAttempts);
      }
    };
  } catch (error) {
    console.error('WebSocket connection failed:', error);
  }
}

function handleWebSocketMessage(data) {
  if (data.type === 'vote_update') {
    updatePolicyVotes(data.policy_id, data.data.upvotes, data.data.downvotes);
  } else if (data.type === 'policy_update') {
    updatePolicyStatus(data.policy_id, data.data.status);
  }
}

function updatePolicyVotes(policyId, upvotes, downvotes) {
  const card = document.querySelector(`[data-policy-id="${policyId}"]`);
  if (!card) return;

  const upvoteCount = card.querySelector('[data-role="upvote-count"]');
  const downvoteCount = card.querySelector('[data-role="downvote-count"]');
  const upvoteBtnCount = card.querySelector('[data-role="upvote-btn-count"]');
  const downvoteBtnCount = card.querySelector('[data-role="downvote-btn-count"]');

  if (upvoteCount) upvoteCount.textContent = upvotes;
  if (downvoteCount) downvoteCount.textContent = downvotes;
  if (upvoteBtnCount) upvoteBtnCount.textContent = upvotes;
  if (downvoteBtnCount) downvoteBtnCount.textContent = downvotes;

  const totalVotes = upvotes + downvotes;
  if (totalVotes > 0) {
    const supportPercentage = ((upvotes / totalVotes) * 100).toFixed(1);
    const progressBar = card.querySelector('.vote-progress-bar');
    if (progressBar) {
      progressBar.style.width = `${supportPercentage}%`;
    }

    const labels = card.querySelectorAll('.vote-progress-label span');
    if (labels.length >= 2) {
      labels[0].textContent = `Support: ${supportPercentage}%`;
      labels[1].textContent = `Oppose: ${(100 - supportPercentage).toFixed(1)}%`;
    }
  }
}

function updatePolicyStatus(policyId, status) {
  const card = document.querySelector(`[data-policy-id="${policyId}"]`);
  if (!card) return;

  const statusBadge = card.querySelector('.badge');
  if (statusBadge) {
    statusBadge.className = `badge badge-${status}`;
    statusBadge.textContent = status;
  }
}

if (typeof window !== 'undefined') {
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      setTimeout(connectWebSocket, 500);
    });
  } else {
    setTimeout(connectWebSocket, 500);
  }
}