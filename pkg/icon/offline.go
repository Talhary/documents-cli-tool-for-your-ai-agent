package icon

import (
	"strings"
)

// OfflineIcon contains metadata and SVG markup for an embedded icon.
type OfflineIcon struct {
	ID         string
	Collection string
	Name       string
	Category   string
	Keywords   []string
	SVG        string
}

// offlineLibrary holds the curated offline icon catalog.
var offlineLibrary = []OfflineIcon{
	// UI Icons (Lucide / Feather style 24x24 viewBox="0 0 24 24")
	{
		ID:         "lucide:search",
		Collection: "lucide",
		Name:       "search",
		Category:   "UI",
		Keywords:   []string{"search", "find", "magnifier", "lookup", "explore"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>`,
	},
	{
		ID:         "lucide:menu",
		Collection: "lucide",
		Name:       "menu",
		Category:   "UI",
		Keywords:   []string{"menu", "hamburger", "nav", "navigation", "bars"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="18" y2="18"/></svg>`,
	},
	{
		ID:         "lucide:x",
		Collection: "lucide",
		Name:       "x",
		Category:   "UI",
		Keywords:   []string{"x", "close", "cross", "cancel", "remove", "dismiss"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>`,
	},
	{
		ID:         "lucide:check",
		Collection: "lucide",
		Name:       "check",
		Category:   "UI",
		Keywords:   []string{"check", "checkmark", "tick", "confirm", "success", "done", "ok"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>`,
	},
	{
		ID:         "lucide:check-circle",
		Collection: "lucide",
		Name:       "check-circle",
		Category:   "UI",
		Keywords:   []string{"check-circle", "checked", "success", "verified", "passed"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><path d="m9 11 3 3L22 4"/></svg>`,
	},
	{
		ID:         "lucide:plus",
		Collection: "lucide",
		Name:       "plus",
		Category:   "UI",
		Keywords:   []string{"plus", "add", "create", "new", "insert"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="M12 5v14"/></svg>`,
	},
	{
		ID:         "lucide:minus",
		Collection: "lucide",
		Name:       "minus",
		Category:   "UI",
		Keywords:   []string{"minus", "remove", "subtract", "decrease"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/></svg>`,
	},
	{
		ID:         "lucide:trash",
		Collection: "lucide",
		Name:       "trash",
		Category:   "UI",
		Keywords:   []string{"trash", "delete", "remove", "bin", "garbage", "discard"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/></svg>`,
	},
	{
		ID:         "lucide:edit",
		Collection: "lucide",
		Name:       "edit",
		Category:   "UI",
		Keywords:   []string{"edit", "pencil", "write", "modify", "rename", "update"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.174 6.812a1 1 0 0 0-3.986-3.987L3.842 16.174a2 2 0 0 0-.5.83l-1.321 4.352a.5.5 0 0 0 .623.622l4.353-1.32a2 2 0 0 0 .83-.497z"/><path d="m15 5 4 4"/></svg>`,
	},
	{
		ID:         "lucide:copy",
		Collection: "lucide",
		Name:       "copy",
		Category:   "UI",
		Keywords:   []string{"copy", "duplicate", "clipboard", "clone"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>`,
	},
	{
		ID:         "lucide:download",
		Collection: "lucide",
		Name:       "download",
		Category:   "UI",
		Keywords:   []string{"download", "export", "save", "fetch", "archive"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>`,
	},
	{
		ID:         "lucide:upload",
		Collection: "lucide",
		Name:       "upload",
		Category:   "UI",
		Keywords:   []string{"upload", "import", "send", "publish"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" x2="12" y1="3" y2="15"/></svg>`,
	},
	{
		ID:         "lucide:refresh-cw",
		Collection: "lucide",
		Name:       "refresh",
		Category:   "UI",
		Keywords:   []string{"refresh", "reload", "sync", "cycle", "cw", "spin"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"/><path d="M21 3v5h-5"/><path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"/><path d="M8 16H3v5"/></svg>`,
	},
	{
		ID:         "lucide:filter",
		Collection: "lucide",
		Name:       "filter",
		Category:   "UI",
		Keywords:   []string{"filter", "funnel", "sort", "refine"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/></svg>`,
	},
	{
		ID:         "lucide:settings",
		Collection: "lucide",
		Name:       "settings",
		Category:   "UI",
		Keywords:   []string{"settings", "gear", "cog", "preferences", "config", "options"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>`,
	},
	{
		ID:         "lucide:user",
		Collection: "lucide",
		Name:       "user",
		Category:   "UI",
		Keywords:   []string{"user", "profile", "account", "avatar", "person", "member"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>`,
	},
	{
		ID:         "lucide:home",
		Collection: "lucide",
		Name:       "home",
		Category:   "UI",
		Keywords:   []string{"home", "house", "main", "dashboard", "homepage"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>`,
	},
	{
		ID:         "lucide:mail",
		Collection: "lucide",
		Name:       "mail",
		Category:   "UI",
		Keywords:   []string{"mail", "email", "envelope", "inbox", "message", "contact"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="16" x="2" y="4" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>`,
	},
	{
		ID:         "lucide:bell",
		Collection: "lucide",
		Name:       "bell",
		Category:   "UI",
		Keywords:   []string{"bell", "notification", "alert", "notice", "alarm"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"/><path d="M10.3 21a1.94 1.94 0 0 0 3.4 0"/></svg>`,
	},
	{
		ID:         "lucide:calendar",
		Collection: "lucide",
		Name:       "calendar",
		Category:   "UI",
		Keywords:   []string{"calendar", "date", "schedule", "event", "month"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="18" x="3" y="4" rx="2" ry="2"/><line x1="16" x2="16" y1="2" y2="6"/><line x1="8" x2="8" y1="2" y2="6"/><line x1="3" x2="21" y1="10" y2="10"/></svg>`,
	},
	{
		ID:         "lucide:clock",
		Collection: "lucide",
		Name:       "clock",
		Category:   "UI",
		Keywords:   []string{"clock", "time", "hour", "minute", "timer", "duration"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>`,
	},
	{
		ID:         "lucide:heart",
		Collection: "lucide",
		Name:       "heart",
		Category:   "UI",
		Keywords:   []string{"heart", "love", "like", "favorite", "wishlist"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/></svg>`,
	},
	{
		ID:         "lucide:star",
		Collection: "lucide",
		Name:       "star",
		Category:   "UI",
		Keywords:   []string{"star", "favorite", "rating", "bookmark", "featured"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>`,
	},
	{
		ID:         "lucide:lock",
		Collection: "lucide",
		Name:       "lock",
		Category:   "UI",
		Keywords:   []string{"lock", "security", "padlock", "private", "protected"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>`,
	},
	{
		ID:         "lucide:shield",
		Collection: "lucide",
		Name:       "shield",
		Category:   "UI",
		Keywords:   []string{"shield", "security", "protection", "safe", "guard", "defense"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10"/></svg>`,
	},
	{
		ID:         "lucide:file",
		Collection: "lucide",
		Name:       "file",
		Category:   "UI",
		Keywords:   []string{"file", "document", "page", "paper", "doc"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/></svg>`,
	},
	{
		ID:         "lucide:folder",
		Collection: "lucide",
		Name:       "folder",
		Category:   "UI",
		Keywords:   []string{"folder", "directory", "folder-open", "storage"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 8 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/></svg>`,
	},
	{
		ID:         "lucide:shopping-cart",
		Collection: "lucide",
		Name:       "shopping-cart",
		Category:   "Things",
		Keywords:   []string{"shopping-cart", "cart", "shop", "store", "ecommerce", "buy", "checkout"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="8" cy="21" r="1"/><circle cx="19" cy="21" r="1"/><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12"/></svg>`,
	},
	{
		ID:         "lucide:credit-card",
		Collection: "lucide",
		Name:       "credit-card",
		Category:   "Things",
		Keywords:   []string{"credit-card", "card", "payment", "pay", "bank", "checkout"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/></svg>`,
	},
	{
		ID:         "lucide:arrow-right",
		Collection: "lucide",
		Name:       "arrow-right",
		Category:   "UI",
		Keywords:   []string{"arrow-right", "arrow", "forward", "next", "continue", "right"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>`,
	},
	{
		ID:         "lucide:arrow-left",
		Collection: "lucide",
		Name:       "arrow-left",
		Category:   "UI",
		Keywords:   []string{"arrow-left", "arrow", "back", "previous", "left"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m12 19-7-7 7-7"/><path d="M19 12H5"/></svg>`,
	},
	{
		ID:         "lucide:chevron-down",
		Collection: "lucide",
		Name:       "chevron-down",
		Category:   "UI",
		Keywords:   []string{"chevron-down", "dropdown", "arrow-down", "expand", "collapse"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>`,
	},
	{
		ID:         "lucide:chevron-right",
		Collection: "lucide",
		Name:       "chevron-right",
		Category:   "UI",
		Keywords:   []string{"chevron-right", "breadcrumb", "arrow", "forward"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m9 18 6-6-6-6"/></svg>`,
	},
	{
		ID:         "lucide:alert-triangle",
		Collection: "lucide",
		Name:       "alert-triangle",
		Category:   "UI",
		Keywords:   []string{"alert-triangle", "warning", "caution", "alert", "danger"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" x2="12" y1="9" y2="13"/><line x1="12" x2="12.01" y1="17" y2="17"/></svg>`,
	},
	{
		ID:         "lucide:info",
		Collection: "lucide",
		Name:       "info",
		Category:   "UI",
		Keywords:   []string{"info", "information", "help", "hint", "about"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>`,
	},
	{
		ID:         "lucide:globe",
		Collection: "lucide",
		Name:       "globe",
		Category:   "Things",
		Keywords:   []string{"globe", "world", "earth", "web", "internet", "language", "international"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>`,
	},
	{
		ID:         "lucide:database",
		Collection: "lucide",
		Name:       "database",
		Category:   "Things",
		Keywords:   []string{"database", "sql", "storage", "db", "records", "data"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/><path d="M3 12c0 1.66 4 3 9 3s9-1.34 9-3"/></svg>`,
	},
	{
		ID:         "lucide:code",
		Collection: "lucide",
		Name:       "code",
		Category:   "UI",
		Keywords:   []string{"code", "programming", "developer", "html", "brackets"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>`,
	},

	// Brand / Product Logos
	{
		ID:         "logos:github",
		Collection: "logos",
		Name:       "github",
		Category:   "Logos",
		Keywords:   []string{"github", "git", "repo", "octocat", "vcs", "source"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="currentColor" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33s1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2"/></svg>`,
	},
	{
		ID:         "logos:react",
		Collection: "logos",
		Name:       "react",
		Category:   "Logos",
		Keywords:   []string{"react", "reactjs", "frontend", "javascript", "framework", "meta"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><ellipse cx="12" cy="12" rx="10" ry="4.2" stroke="#00d8ff" stroke-width="1.5" transform="rotate(0 12 12)"/><ellipse cx="12" cy="12" rx="10" ry="4.2" stroke="#00d8ff" stroke-width="1.5" transform="rotate(60 12 12)"/><ellipse cx="12" cy="12" rx="10" ry="4.2" stroke="#00d8ff" stroke-width="1.5" transform="rotate(120 12 12)"/><circle cx="12" cy="12" r="1.8" fill="#00d8ff"/></svg>`,
	},
	{
		ID:         "logos:docker",
		Collection: "logos",
		Name:       "docker",
		Category:   "Logos",
		Keywords:   []string{"docker", "container", "devops", "whale", "kubernetes"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="#0db7ed" d="M13 3h2v2h-2zm-3 0h2v2h-2zm6 3h2v2h-2zm-3 0h2v2h-2zm-3 0h2v2h-2zm-3 0h2v2H7zm9 3h2v2h-2zm-3 0h2v2h-2zm-3 0h2v2h-2zm-3 0h2v2H7zm-3 0h2v2H4zm19.6 2.3c-.3-.2-1.3-.8-2.6-.4c-.4-.9-1.2-1.6-2.1-1.9c-.3.4-.6.8-.7 1.3c-.6-.2-1.3-.2-2-.1c-.4-1.3-1.6-2.2-3.1-2.2H1.9c-.5 0-.9.4-.9.9c0 3.9 1.7 7.5 4.6 9.8c2.1 1.7 4.7 2.6 7.4 2.6c7.7 0 10.9-4.8 11.2-8.6c.1-.4 0-.9-.4-1.2z"/></svg>`,
	},
	{
		ID:         "logos:python",
		Collection: "logos",
		Name:       "python",
		Category:   "Logos",
		Keywords:   []string{"python", "programming", "backend", "ai", "machine-learning", "py"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="#3776ab" d="M11.9 2c-4.4 0-4.1 1.9-4.1 1.9l.1 2h4.2v.6H6.2s-2.9.3-2.9 4.3c0 4 2.5 3.9 2.5 3.9h1.5v-2.1s-.1-2.5 2.5-2.5h4.2s2.4 0 2.4-2.3V4.3s.4-2.3-4.5-2.3zm-2.4 1.3c.5 0 .8.4.8.8s-.4.8-.8.8s-.8-.4-.8-.8s.3-.8.8-.8z"/><path fill="#ffd43b" d="M12.1 22c4.4 0 4.1-1.9 4.1-1.9l-.1-2h-4.2v-.6h5.9s2.9-.3 2.9-4.3c0-4-2.5-3.9-2.5-3.9h-1.5v2.1s.1 2.5-2.5 2.5H10s-2.4 0-2.4 2.3v3.4s-.4 2.3 4.5 2.3zm2.4-1.3c-.5 0-.8-.4-.8-.8s.4-.8.8-.8s.8.4.8.8s-.3.8-.8.8z"/></svg>`,
	},
	{
		ID:         "logos:go",
		Collection: "logos",
		Name:       "go",
		Category:   "Logos",
		Keywords:   []string{"go", "golang", "google", "backend", "language"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="#00add8" d="M1.8 10.4c0-2 1.3-3.6 3.6-3.6c2.4 0 3.7 1.7 3.7 3.7v.5H3.6c.1 1.1.9 1.9 2.1 1.9c.9 0 1.5-.4 1.8-1.1h1.7c-.4 1.7-1.8 2.5-3.5 2.5c-2.3 0-3.9-1.7-3.9-3.9zm3.6-2.2c-1.1 0-1.8.7-1.9 1.6h3.7c0-.9-.7-1.6-1.8-1.6zm7.6 2.2c0-2.2 1.6-3.8 3.9-3.8s3.9 1.6 3.9 3.8c0 2.2-1.6 3.9-3.9 3.9s-3.9-1.7-3.9-3.9zm6.1 0c0-1.3-.9-2.3-2.2-2.3s-2.2 1-2.2 2.3c0 1.3.9 2.3 2.2 2.3s2.2-1 2.2-2.3zm4.5-5.9h-1.8v1.6h1.8V4.5zm0 2.5h-1.8v7.2h1.8V7z"/></svg>`,
	},
	{
		ID:         "logos:aws",
		Collection: "logos",
		Name:       "aws",
		Category:   "Logos",
		Keywords:   []string{"aws", "amazon", "cloud", "hosting", "s3", "lambda", "ec2"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="#ff9900" d="M18.8 17.5c-2.4 1.8-5.8 2.7-8.8 2.7c-4.2 0-8-1.5-10.9-4.1c-.2-.2 0-.5.2-.4c3.1 1.7 6.9 2.7 10.7 2.7c2.7 0 5.7-.7 8.5-2.1c.4-.2.7.2.3.7z"/><path fill="#ff9900" d="M19.9 16.3c-.3-.4-2-.2-3-.1c-.3 0-.3-.2 0-.4c1.9-1.4 5-1 5.3-.6c.3.4-.2 3.5-2 5c-.3.2-.4.1-.3-.1c.4-.9.4-3.3 0-3.8z"/><path fill="currentColor" d="M7 9.8v3c0 .5.3.8.7.8c.3 0 .6-.2.7-.4l1.3-3.4h1.7l-2.2 4.9c-.3.7-1 1.2-1.8 1.2c-1 0-1.6-.7-1.6-1.7V9.8H7zm7.5 4.2h-1.5V9.8h1.5v4.2zm-3.8-3.4l.7 1.7l.7-1.7h1.4l-1.4 3.4h-1.4l-1.4-3.4h1.4z"/></svg>`,
	},
	{
		ID:         "logos:google",
		Collection: "logos",
		Name:       "google",
		Category:   "Logos",
		Keywords:   []string{"google", "alphabet", "search", "cloud", "gsuite"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="#4285f4" d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v4.51h6.6c-.29 1.52-1.14 2.8-2.4 3.68v3.05h3.88c2.27-2.09 3.66-5.17 3.66-9.17z"/><path fill="#34a853" d="M12 24c3.24 0 5.95-1.08 7.93-2.91l-3.88-3.05c-1.08.72-2.45 1.16-4.05 1.16c-3.12 0-5.77-2.1-6.72-4.93H1.24v3.15C3.26 21.36 7.33 24 12 24z"/><path fill="#fbbc05" d="M5.28 14.27a7.2 7.2 0 0 1 0-4.54V6.58H1.24a11.97 11.97 0 0 0 0 10.84l4.04-3.15z"/><path fill="#ea4335" d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0C7.33 0 3.26 2.64 1.24 6.58l4.04 3.15c.95-2.83 3.6-4.98 6.72-4.98z"/></svg>`,
	},
	{
		ID:         "logos:apple",
		Collection: "logos",
		Name:       "apple",
		Category:   "Logos",
		Keywords:   []string{"apple", "mac", "ios", "iphone", "ipad", "macos"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="currentColor" d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47c-1.34.03-1.77-.79-3.29-.79c-1.53 0-2 .77-3.27.82c-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51c1.28-.02 2.5.87 3.29.87c.78 0 2.26-1.07 3.81-.91c.65.03 2.47.26 3.64 1.98c-.09.06-2.17 1.28-2.15 3.81c.03 3.02 2.65 4.03 2.68 4.04c-.03.07-.42 1.44-1.38 2.83M15.97 6.38c.62-.75 1.04-1.8 0.92-2.85c-.9.04-2 .6-2.65 1.35c-.56.64-1.06 1.7-0.93 2.73c1.01.08 2.04-.51 2.66-1.23"/></svg>`,
	},
	{
		ID:         "logos:stripe",
		Collection: "logos",
		Name:       "stripe",
		Category:   "Logos",
		Keywords:   []string{"stripe", "payments", "finance", "billing", "checkout"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="#635bff" d="M13.976 9.15c-2.172-.806-3.356-1.426-3.356-2.409c0-.831.683-1.305 1.901-1.305c2.227 0 4.515.858 6.09 1.631l.89-5.494C17.652.705 14.854 0 11.82 0C5.86 0 1.902 3.12 1.902 8.352c0 6.646 9.153 5.568 9.153 8.428c0 .98-.82 1.408-2.122 1.408c-2.613 0-5.382-1.15-7.142-2.146L.89 21.57C2.96 22.75 6.46 23.6 9.77 23.6c6.236 0 10.37-3.084 10.37-8.428c0-7.05-6.164-6.022-6.164-6.022"/></svg>`,
	},
	{
		ID:         "logos:tailwind",
		Collection: "logos",
		Name:       "tailwind",
		Category:   "Logos",
		Keywords:   []string{"tailwind", "tailwindcss", "css", "styling", "ui"},
		SVG:        `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="#38bdf8" d="M12.001 4.8c-3.2 0-5.2 1.6-6 4.8c1.2-1.6 2.6-2.2 4.2-1.8c.913.228 1.565.89 2.288 1.624C13.666 10.618 15.027 12 18.001 12c3.2 0 5.2-1.6 6-4.8c-1.2 1.6-2.6 2.2-4.2 1.8c-.913-.228-1.565-.89-2.288-1.624C16.335 6.182 14.974 4.8 12.001 4.8zm-6 7.2c-3.2 0-5.2 1.6-6 4.8c1.2-1.6 2.6-2.2 4.2-1.8c.913.228 1.565.89 2.288 1.624C7.666 17.818 9.027 19.2 12.001 19.2c3.2 0 5.2-1.6 6-4.8c-1.2 1.6-2.6 2.2-4.2 1.8c-.913-.228-1.565-.89-2.288-1.624C10.335 13.382 8.974 12 6.001 12z"/></svg>`,
	},
}

// searchOffline searches the embedded catalog for icons matching the query.
func searchOffline(query string, collectionFilter string, limit int) []IconItem {
	query = strings.ToLower(strings.TrimSpace(query))
	tokens := strings.Fields(query)
	var matches []IconItem

	for _, icon := range offlineLibrary {
		if collectionFilter != "" && !strings.EqualFold(icon.Collection, collectionFilter) {
			continue
		}

		matched := false
		if query == "" {
			matched = true
		} else {
			nameLower := strings.ToLower(icon.Name)
			idLower := strings.ToLower(icon.ID)
			catLower := strings.ToLower(icon.Category)

			for _, tok := range tokens {
				if strings.Contains(nameLower, tok) || strings.Contains(idLower, tok) || strings.Contains(catLower, tok) {
					matched = true
					break
				}
				for _, kw := range icon.Keywords {
					if strings.Contains(strings.ToLower(kw), tok) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
		}

		if matched {
			matches = append(matches, IconItem{
				ID:         icon.ID,
				Collection: icon.Collection,
				Name:       icon.Name,
				Category:   icon.Category,
				PreviewURL: "https://api.iconify.design/" + icon.Collection + "/" + icon.Name + ".svg",
			})
			if limit > 0 && len(matches) >= limit {
				break
			}
		}
	}

	return matches
}

// getOfflineIcon retrieves the exact SVG markup from the embedded library if present.
func getOfflineIcon(idOrName string) (OfflineIcon, bool) {
	clean := strings.ToLower(strings.TrimSpace(idOrName))
	// Try matching exact ID first (e.g. "lucide:search")
	for _, item := range offlineLibrary {
		if strings.EqualFold(item.ID, clean) {
			return item, true
		}
	}
	// Try matching prefix/name or name
	if strings.Contains(clean, "/") {
		clean = strings.ReplaceAll(clean, "/", ":")
		for _, item := range offlineLibrary {
			if strings.EqualFold(item.ID, clean) {
				return item, true
			}
		}
	}
	for _, item := range offlineLibrary {
		if strings.EqualFold(item.Name, clean) {
			return item, true
		}
	}
	// Try prefix match in keywords
	for _, item := range offlineLibrary {
		for _, kw := range item.Keywords {
			if strings.EqualFold(kw, clean) {
				return item, true
			}
		}
	}
	return OfflineIcon{}, false
}
