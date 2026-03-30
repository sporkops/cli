package spork

import "time"

// Monitor represents an uptime monitor.
type Monitor struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Type            string            `json:"type,omitempty"`
	Target          string            `json:"target,omitempty"`
	Method          string            `json:"method,omitempty"`
	ExpectedStatus  int               `json:"expected_status,omitempty"`
	Interval        int               `json:"interval,omitempty"`
	Timeout         int               `json:"timeout,omitempty"`
	Regions         []string          `json:"regions,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            string            `json:"body,omitempty"`
	Keyword         string            `json:"keyword,omitempty"`
	KeywordType     string            `json:"keyword_type,omitempty"`
	SSLWarnDays     int               `json:"ssl_warn_days,omitempty"`
	AlertChannelIDs []string          `json:"alert_channel_ids,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	Paused          *bool             `json:"paused,omitempty"`
	Status          string            `json:"status,omitempty"`
	LastCheckedAt   string            `json:"last_checked_at,omitempty"`
	CreatedAt       string            `json:"created_at,omitempty"`
	UpdatedAt       string            `json:"updated_at,omitempty"`
}

// Account represents the authenticated user's account.
type Account struct {
	UID              string    `json:"uid"`
	Email            string    `json:"email"`
	Plan             string    `json:"plan"`
	MonitorLimit     int       `json:"monitor_limit"`
	CheckIntervalS   int       `json:"check_interval_s"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	HasPaymentMethod bool      `json:"has_payment_method"`
}

// APIKey represents an API key for programmatic access.
type APIKey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key,omitempty"`
	Prefix     string     `json:"prefix"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// MonitorResult represents a single uptime check result.
type MonitorResult struct {
	ID             string `json:"id"`
	MonitorID      string `json:"monitor_id"`
	Status         string `json:"status"`
	StatusCode     int    `json:"status_code"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	Region         string `json:"region"`
	ErrorMessage   string `json:"error_message,omitempty"`
	CheckedAt      string `json:"checked_at"`
}

// AlertChannel represents an alert notification channel.
type AlertChannel struct {
	ID                 string            `json:"id,omitempty"`
	Name               string            `json:"name"`
	Type               string            `json:"type"`
	Config             map[string]string `json:"config"`
	Verified           bool              `json:"verified,omitempty"`
	Secret             string            `json:"secret,omitempty"`
	LastDeliveryStatus string            `json:"last_delivery_status,omitempty"`
	LastDeliveryAt     string            `json:"last_delivery_at,omitempty"`
	CreatedAt          string            `json:"created_at,omitempty"`
	UpdatedAt          string            `json:"updated_at,omitempty"`
}

// StatusPage represents a public status page.
type StatusPage struct {
	ID                      string            `json:"id,omitempty"`
	Name                    string            `json:"name"`
	Slug                    string            `json:"slug"`
	Components              []StatusComponent `json:"components,omitempty"`
	ComponentGroups         []ComponentGroup  `json:"component_groups,omitempty"`
	CustomDomain            string            `json:"custom_domain,omitempty"`
	DomainStatus            string            `json:"domain_status,omitempty"`
	Theme                   string            `json:"theme,omitempty"`
	AccentColor             string            `json:"accent_color,omitempty"`
	FontFamily              string            `json:"font_family,omitempty"`
	HeaderStyle             string            `json:"header_style,omitempty"`
	LogoURL                 string            `json:"logo_url,omitempty"`
	WebhookURL              string            `json:"webhook_url,omitempty"`
	EmailSubscribersEnabled bool              `json:"email_subscribers_enabled"`
	IsPublic                bool              `json:"is_public"`
	Password                string            `json:"password,omitempty"`
	CreatedAt               string            `json:"created_at,omitempty"`
	UpdatedAt               string            `json:"updated_at,omitempty"`
}

// ComponentGroup organizes components into named sections on the status page.
type ComponentGroup struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Order       int    `json:"order"`
}

// StatusComponent maps a monitor to a display name on a status page.
type StatusComponent struct {
	ID          string `json:"id,omitempty"`
	MonitorID   string `json:"monitor_id"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
	GroupID     string `json:"group_id,omitempty"`
	Order       int    `json:"order"`
}

// Incident represents a status page incident.
type Incident struct {
	ID             string   `json:"id,omitempty"`
	StatusPageID   string   `json:"status_page_id,omitempty"`
	Title          string   `json:"title"`
	Message        string   `json:"message,omitempty"`
	Type           string   `json:"type,omitempty"`
	Status         string   `json:"status,omitempty"`
	Impact         string   `json:"impact,omitempty"`
	ComponentIDs   []string `json:"component_ids,omitempty"`
	StartedAt      string   `json:"started_at,omitempty"`
	ResolvedAt     string   `json:"resolved_at,omitempty"`
	ScheduledStart string   `json:"scheduled_start,omitempty"`
	ScheduledEnd   string   `json:"scheduled_end,omitempty"`
	CreatedAt      string   `json:"created_at,omitempty"`
	UpdatedAt      string   `json:"updated_at,omitempty"`
}

// IncidentUpdate represents a timeline update on an incident.
type IncidentUpdate struct {
	ID         string `json:"id,omitempty"`
	IncidentID string `json:"incident_id,omitempty"`
	Status     string `json:"status,omitempty"`
	Message    string `json:"message,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
}
