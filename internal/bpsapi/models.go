package bpsapi

import "encoding/json"

// Pagination represents pagination metadata returned in the first element of BPS data array
type Pagination struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
	Count   int `json:"count,omitempty"`
}

// RawBPSResponse represents the root JSON payload returned by webapi.bps.go.id
type RawBPSResponse struct {
	Status           string          `json:"status"`
	DataAvailability string          `json:"data-availability"`
	Message          string          `json:"message,omitempty"`
	Data             json.RawMessage `json:"data,omitempty"`
}

// Subject represents a statistical subject (e.g., Kependudukan, Pertanian)
type Subject struct {
	SubID    any    `json:"sub_id"`
	Title    string `json:"title"`
	SubcatID any    `json:"subcat_id"`
	Subcat   string `json:"subcat"`
	NTabel   any    `json:"ntabel,omitempty"`
}

// SubCat represents a subject category
type SubCat struct {
	SubcatID any    `json:"subcat_id"`
	Title    string `json:"title"`
}

// Variable represents a statistical variable under a subject
type Variable struct {
	VarID   any    `json:"var_id"`
	Title   string `json:"title"`
	SubID   any    `json:"sub_id"`
	SubName string `json:"sub_name,omitempty"`
	Def     string `json:"def,omitempty"`
	Notes   string `json:"notes,omitempty"`
	Unit    string `json:"unit,omitempty"`
	GraphID any    `json:"graph_id,omitempty"`
}

// Publication represents a statistical publication book
type Publication struct {
	PubID   string `json:"pub_id"`
	Title   string `json:"title"`
	RlDate  string `json:"rl_date"`
	Issn    string `json:"issn,omitempty"`
	SchDate string `json:"sch_date,omitempty"`
	Cover   string `json:"cover,omitempty"`
	Pdf     string `json:"pdf,omitempty"`
	Size    string `json:"size,omitempty"`
}

// PressRelease represents an official BPS press release (Berita Resmi Statistik / BRS)
type PressRelease struct {
	BrsID    any    `json:"brs_id"`
	Title    string `json:"title"`
	RlDate   string `json:"rl_date"`
	Abstract string `json:"abstract,omitempty"`
	Pdf      string `json:"pdf,omitempty"`
	Size     string `json:"size,omitempty"`
	Slide    string `json:"slide,omitempty"`
}

// Infographic represents statistical visual infographic
type Infographic struct {
	InfID    any    `json:"inf_id"`
	Title    string `json:"title"`
	Img      string `json:"img"`
	Desc     string `json:"desc,omitempty"`
	Dl       string `json:"dl,omitempty"`
	Category string `json:"category,omitempty"`
}

// StaticTable represents a static table
type StaticTable struct {
	TableID  any    `json:"table_id"`
	Title    string `json:"title"`
	SubjID   any    `json:"subj_id,omitempty"`
	UpdtDate string `json:"updt_date,omitempty"`
	Size     string `json:"size,omitempty"`
	Excel    string `json:"excel,omitempty"`
}

// ListResult contains the parsed outcome for any model
type ListResult[T any] struct {
	Status           string      `json:"status"`
	DataAvailability string      `json:"data_availability"`
	Pagination       *Pagination `json:"pagination,omitempty"`
	Items            []T         `json:"items"`
	IsStale          bool        `json:"is_stale,omitempty"`
}

// ListParams holds query and segment parameters for BPS Web API URL construction
type ListParams struct {
	Model   string `json:"model"`
	Domain  string `json:"domain"`
	Lang    string `json:"lang"`
	Page    int    `json:"page,omitempty"`
	Subject string `json:"subject,omitempty"`
	Subcat  string `json:"subcat,omitempty"`
	Year    string `json:"year,omitempty"`
	Month   string `json:"month,omitempty"`
	Keyword string `json:"keyword,omitempty"`
	Var     string `json:"var,omitempty"`
	Turvar  string `json:"turvar,omitempty"`
	Vervar  string `json:"vervar,omitempty"`
	Th      string `json:"th,omitempty"`
	Turth   string `json:"turth,omitempty"`
}
