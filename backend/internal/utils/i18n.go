package utils

var translations = map[string]map[string]string{
	"en": {
		"vote":             "Vote",
		"login":            "Login",
		"logout":           "Logout",
		"dashboard":        "Dashboard",
		"policies":         "Policies",
		"submit":           "Submit",
		"submit_policy":    "Submit Policy",
		"admin":            "Admin",
		"superuser":        "Superuser",
		"all_policies":     "All Policies",
		"policy_title":     "Policy Title",
		"description":      "Description",
		"comments":         "Comments",
		"add_comment":      "Add Comment",
		"vote_up":          "Vote Up",
		"vote_down":        "Vote Down",
		"search":           "Search",
		"filter":           "Filter",
		"category":         "Category",
		"all_categories":   "All Categories",
		"status":           "Status",
		"pending":          "Pending",
		"approved":         "Approved",
		"rejected":         "Rejected",
		"uncertain":        "Uncertain",
		"in_progress":      "In Progress",
		"completed":        "Completed",
		"on_hold":          "On Hold",
		"cannot_implement": "Cannot Implement",
		"sort_by":          "Sort By",
		"newest":           "Newest",
		"oldest":           "Oldest",
		"most_voted":       "Most Voted",
		"trending":         "Trending",
		"export":           "Export",
		"analytics":        "Analytics",
		"dark_mode":        "Dark Mode",
		"light_mode":       "Light Mode",
	},
	"ro": {
		"vote":             "Votează",
		"login":            "Autentificare",
		"logout":           "Deconectare",
		"dashboard":        "Panou",
		"policies":         "Politici",
		"submit":           "Trimite",
		"submit_policy":    "Trimite Politică",
		"admin":            "Admin",
		"superuser":        "Superuser",
		"all_policies":     "Toate Politicile",
		"policy_title":     "Titlu Politică",
		"description":      "Descriere",
		"comments":         "Comentarii",
		"add_comment":      "Adaugă Comentariu",
		"vote_up":          "Votează Pro",
		"vote_down":        "Votează Contra",
		"search":           "Căutare",
		"filter":           "Filtrează",
		"category":         "Categorie",
		"all_categories":   "Toate Categoriile",
		"status":           "Status",
		"pending":          "În Așteptare",
		"approved":         "Aprobat",
		"rejected":         "Respins",
		"uncertain":        "Incert",
		"in_progress":      "În Progres",
		"completed":        "Finalizat",
		"on_hold":          "În Așteptare",
		"cannot_implement": "Nu Poate Fi Implementat",
		"sort_by":          "Sortează După",
		"newest":           "Cele Mai Noi",
		"oldest":           "Cele Mai Vechi",
		"most_voted":       "Cele Mai Votate",
		"trending":         "Trending",
		"export":           "Exportă",
		"analytics":        "Analize",
		"dark_mode":        "Mod Întunecat",
		"light_mode":       "Mod Luminos",
	},
}

func T(lang, key string) string {
	if lang == "" {
		lang = "en"
	}
	if langMap, ok := translations[lang]; ok {
		if val, ok := langMap[key]; ok {
			return val
		}
	}
	// Fallback to English
	if val, ok := translations["en"][key]; ok {
		return val
	}
	return key
}
