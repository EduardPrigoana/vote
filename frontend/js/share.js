function sharePolicy(policyId, title) {
  const url = `${window.location.origin}/policy/${policyId}`;
  
  if (navigator.share) {
    navigator.share({
      title: title,
      text: `Check out this policy: ${title}`,
      url: url
    }).catch((error) => console.log('Share failed:', error));
  } else {
    copyToClipboard(url);
    showTempAlert('Link copied to clipboard', 'success');
  }
}

function copyToClipboard(text) {
  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.position = 'fixed';
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand('copy');
  document.body.removeChild(textarea);
}

function openShareModal(policyId, title) {
  const url = `${window.location.origin}/policy/${policyId}`;
  const encodedUrl = encodeURIComponent(url);
  const encodedTitle = encodeURIComponent(title);
  
  const modal = document.getElementById('share-modal');
  if (!modal) return;
  
  const content = `
    <div class="modal-content">
      <h3>Share Policy</h3>
      <p style="margin-bottom: 1.5rem;">${escapeHtml(title)}</p>
      
      <div style="display: flex; gap: 0.75rem; flex-wrap: wrap; margin-bottom: 1.5rem;">
        <a href="https://twitter.com/intent/tweet?text=${encodedTitle}&url=${encodedUrl}" 
           target="_blank" 
           class="btn btn-secondary btn-sm">
          Twitter
        </a>
        <a href="https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}" 
           target="_blank" 
           class="btn btn-secondary btn-sm">
          Facebook
        </a>
        <a href="https://www.linkedin.com/sharing/share-offsite/?url=${encodedUrl}" 
           target="_blank" 
           class="btn btn-secondary btn-sm">
          LinkedIn
        </a>
        <button class="btn btn-secondary btn-sm" onclick="copyToClipboard('${url}'); showTempAlert('Link copied', 'success')">
          Copy Link
        </button>
      </div>
      
      <div style="display: flex; justify-content: flex-end;">
        <button class="btn btn-secondary" onclick="closeShareModal()">Close</button>
      </div>
    </div>
  `;
  
  modal.innerHTML = content;
  modal.style.display = 'block';
}

function closeShareModal() {
  const modal = document.getElementById('share-modal');
  if (modal) {
    modal.style.display = 'none';
  }
}