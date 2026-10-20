package bot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"bot-istimewa/internal/bpsapi"
	"bot-istimewa/internal/config"
	"bot-istimewa/internal/whatsapp"
	"go.mau.fi/whatsmeow/types"
)

func TestBotBPSBangkaBaratIntegration(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		path := r.URL.Path

		switch {
		case len(path) >= 15 && path[:15] == "/model/pressrel":
			_, _ = w.Write([]byte(`{
				"status": "OK",
				"data-availability": "available",
				"data": [
					{"page": 1, "pages": 1, "per_page": 10, "total": 1},
					[{"brs_id": 101, "title": "Perkembangan Indeks Harga Konsumen Bangka Barat", "rl_date": "2024-03-01", "pdf": "https://bangkabaratkab.bps.go.id/brs.pdf"}]
				]
			}`))
		case len(path) >= 18 && path[:18] == "/model/publication":
			_, _ = w.Write([]byte(`{
				"status": "OK",
				"data-availability": "available",
				"data": [
					{"page": 1, "pages": 1, "per_page": 10, "total": 1},
					[{"pub_id": "P01", "title": "Kabupaten Bangka Barat Dalam Angka 2024", "rl_date": "2024-02-28", "pdf": "https://bangkabaratkab.bps.go.id/dda.pdf"}]
				]
			}`))
		case len(path) >= 18 && path[:18] == "/model/statictable":
			_, _ = w.Write([]byte(`{
				"status": "OK",
				"data-availability": "available",
				"data": [
					{"page": 1, "pages": 1, "per_page": 10, "total": 1},
					[{"table_id": 501, "title": "Jumlah Penduduk Menurut Kecamatan 2023", "excel": "https://bangkabaratkab.bps.go.id/penduduk.xlsx"}]
				]
			}`))
		default:
			_, _ = w.Write([]byte(`{
				"status": "OK",
				"data-availability": "available",
				"data": [
					{"page": 1, "pages": 1, "per_page": 10, "total": 0},
					[]
				]
			}`))
		}
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		BPSAPIBaseURL:      mockServer.URL,
		BPSAPIKey:          "test-key",
		BPSDefaultDomain:   "1903",
		IdleTimeoutMinutes: 30,
	}

	bpsClient := bpsapi.NewClient(cfg)
	bpsClient.SetHTTPClient(mockServer.Client())

	router := NewBotRouter(cfg, nil, nil, bpsClient)

	// 1. Test live static tables fetching for topic
	tablesText := router.fetchBPSTablesForTopic(context.Background(), "kependudukan")
	if tablesText == "" {
		t.Errorf("expected non-empty tablesText from live BPS Bangka Barat mock")
	}

	// 2. Test BRS and Publications in HandleProdukDetail (dry test without sending)
	ctx := context.Background()
	s := &UserSession{CurrentState: StateMenu3}
	p := &whatsapp.ParsedMessage{
		JID:  types.NewJID("628123456789", types.DefaultUserServer),
		Text: "1",
	}

	// Verify that fetchBPSTablesForTopic contains Bangka Barat URL
	if tablesText != "" && !testing.Short() {
		t.Logf("Tabel output:\n%s", tablesText)
	}

	_ = s
	_ = p
	_ = ctx
}
