const translations = {
  en: {
    vote: "Vote",
    login: "Login",
    logout: "Logout",
    dashboard: "Dashboard",
    policies: "Policies",
    submit: "Submit",
    submit_policy: "Submit Policy",
    admin: "Admin",
    superuser: "Superuser",
    all_policies: "All Policies",
    policy_title: "Policy Title",
    description: "Description",
    comments: "Comments",
    add_comment: "Add Comment",
    search: "Search",
    filter: "Filter",
    category: "Category",
    all_categories: "All Categories",
    status: "Status",
    pending: "Pending",
    approved: "Approved",
    rejected: "Rejected",
    uncertain: "Uncertain",
    in_progress: "In Progress",
    completed: "Completed",
    on_hold: "On Hold",
    cannot_implement: "Cannot Implement",
    sort_by: "Sort By",
    newest: "Newest",
    oldest: "Oldest",
    most_voted: "Most Voted",
    trending: "Trending",
    export: "Export",
    analytics: "Analytics",
    dark_mode: "Dark Mode",
    light_mode: "Light Mode",
    support: "Support",
    oppose: "Oppose",
    no_comments_yet: "No comments yet. Be the first to share your thoughts!",
    write_comment: "Write a comment...",
    submit_comment: "Submit Comment",
    loading: "Loading...",
    classroom_code: "Classroom Code",
    continue: "Continue",
  },
  ro: {
    vote: "Votează",
    login: "Autentificare",
    logout: "Deconectare",
    dashboard: "Panou",
    policies: "Politici",
    submit: "Trimite",
    submit_policy: "Trimite Politică",
    admin: "Admin",
    superuser: "Superuser",
    all_policies: "Toate Politicile",
    policy_title: "Titlu Politică",
    description: "Descriere",
    comments: "Comentarii",
    add_comment: "Adaugă Comentariu",
    search: "Căutare",
    filter: "Filtrează",
    category: "Categorie",
    all_categories: "Toate Categoriile",
    status: "Status",
    pending: "În Așteptare",
    approved: "Aprobat",
    rejected: "Respins",
    uncertain: "Incert",
    in_progress: "În Progres",
    completed: "Finalizat",
    on_hold: "În Așteptare",
    cannot_implement: "Nu Poate Fi Implementat",
    sort_by: "Sortează După",
    newest: "Cele Mai Noi",
    oldest: "Cele Mai Vechi",
    most_voted: "Cele Mai Votate",
    trending: "Trending",
    export: "Exportă",
    analytics: "Analize",
    dark_mode: "Mod Întunecat",
    light_mode: "Mod Luminos",
    support: "Susține",
    oppose: "Opune",
    no_comments_yet: "Niciun comentariu încă. Fii primul care își exprimă părerea!",
    write_comment: "Scrie un comentariu...",
    submit_comment: "Trimite Comentariu",
    loading: "Se încarcă...",
    classroom_code: "Cod Clasă",
    continue: "Continuă",
  }
};

const LANG_KEY = 'preferred_language';

function getCurrentLanguage() {
  return localStorage.getItem(LANG_KEY) || 'en';
}

function setLanguage(lang) {
  localStorage.setItem(LANG_KEY, lang);
  updatePageLanguage();
}

function t(key) {
  const lang = getCurrentLanguage();
  return translations[lang]?.[key] || translations.en[key] || key;
}

function updatePageLanguage() {
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    const translation = t(key);
    
    if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') {
      el.placeholder = translation;
    } else {
      el.textContent = translation;
    }
  });
}

// Initialize language on page load
document.addEventListener('DOMContentLoaded', updatePageLanguage);