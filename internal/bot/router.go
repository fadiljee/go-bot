package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"bot-istimewa/internal/config"
	"bot-istimewa/internal/gsheet"
	"bot-istimewa/internal/locale"
	"bot-istimewa/internal/whatsapp"
)

const (
	StateRoot            = "menuAwal"
	StateMenu1           = "menu1"
	StateMenu1_Sosduk    = "menu11"
	StateMenu1_Ekonomi   = "menu12"
	StateMenu1_Pertanian = "menu13"
	StateMenu1_Sektoral  = "menu14"
	StateMenu1_Search    = "menu15"
	StateMenu2           = "menu2"
	StateMenu3           = "menu3"
	StateMenu4           = "menu4"
	StateMenu5           = "menu5"
	StateMenu6           = "menu6"
	StateMenu6_Pakar     = "menu62"
	StateMenu7           = "menu7"
	StateMenu8           = "menu8"
	StateMenu9           = "menu9"
	StateMenu10          = "menu10"
	StateMenu11_About    = "menu11_about" // Resolves TD-01 (Isolated from menu11 Sosduk)
	StateAdminWaitPass   = "menuAdmin"
	StateAdminLoggedIn   = "menuAdminActive"
)

type BotRouter struct {
	cfg        *config.Config
	sender     *whatsapp.Sender
	gsheet     *gsheet.Client
	sessions   *SessionManager
	handlerMap map[string]func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error
}

func NewBotRouter(cfg *config.Config, sender *whatsapp.Sender, gsheetClient *gsheet.Client) *BotRouter {
	r := &BotRouter{
		cfg:        cfg,
		sender:     sender,
		gsheet:     gsheetClient,
		sessions:   NewSessionManager(cfg.IdleTimeoutMinutes),
		handlerMap: make(map[string]func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error),
	}

	r.registerRoutes()
	return r
}

func (r *BotRouter) Dispatch(ctx context.Context, p *whatsapp.ParsedMessage) {
	if p == nil {
		return
	}

	// Filter broadcast status and groups
	if p.IsGroup || p.JID.Server == "broadcast" || p.JID.Server == "newsletter" {
		return
	}

	// Filter allowed numbers if set
	if len(r.cfg.AllowedNumbers) > 0 {
		senderNum := p.JID.User
		allowed := false
		for _, num := range r.cfg.AllowedNumbers {
			if strings.TrimSpace(num) == senderNum {
				allowed = true
				break
			}
		}
		if !allowed {
			log.Printf("[FILTER] Mengabaikan pesan dari %s (tidak terdaftar di ALLOWED_NUMBERS)", p.JID.String())
			return
		}
	}

	// Get or create session
	session := r.sessions.GetOrCreate(p.JID, func(s *UserSession) {
		r.HandleIdleTimeout(context.Background(), s)
	})

	input := strings.TrimSpace(p.Text)
	inputLower := strings.ToLower(input)

	log.Printf("[ROUTER] Dari: %s | State: %s | Input: '%s'", p.JID.User, session.CurrentState, input)

	// =========================================================
	// 1. GLOBAL KEYWORDS (Overrules any state)
	// =========================================================

	if inputLower == "reset" {
		_ = r.HandleReset(ctx, session, p)
		return
	}

	if input == "100" {
		session.ResetToRoot()
		_ = r.HandleWelcomeFull(ctx, session, p)
		return
	}

	if input == "0" {
		session.ResetToRoot()
		_ = r.HandleWelcomeShort(ctx, session, p)
		return
	}

	if input == "99" {
		_ = r.HandleBack(ctx, session, p)
		return
	}

	if inputLower == strings.ToLower(r.cfg.AdminTriggerKeyword) {
		session.PushState(StateAdminWaitPass)
		_ = r.HandleAdminPrompt(ctx, session, p)
		return
	}

	if session.CurrentState == StateAdminWaitPass {
		_ = r.HandleAdminAuth(ctx, session, p)
		return
	}

	// Mode Admin Aktif: Admin dapat memilih nomor 1-8 untuk detail database atau mengetik menu
	if session.CurrentState == StateAdminLoggedIn {
		if input == "1" || input == "2" || input == "3" || input == "4" || input == "5" || input == "6" || input == "7" || input == "8" {
			key := fmt.Sprintf("menuAdmin%s", input)
			reply := fmt.Sprintf("📂 *Link Database Admin (%s):*\n\n%s\n\n───────────────────────────\n_Ketik angka lain (1-8) atau ketik *menu* untuk melihat semua link._\n_Ketik *0* untuk keluar ke Menu Utama._", input, locale.T(key))
			_ = r.sender.SendTextMessage(ctx, p.JID, reply)
			return
		}

		_ = r.SendAdminDashboard(ctx, session, p)
		return
	}

	// =========================================================
	// 2. STATE-BASED DISPATCH
	// =========================================================

	// Build state-action route key: "STATE:INPUT"
	routeKey := session.CurrentState + ":" + inputLower

	if handler, exists := r.handlerMap[routeKey]; exists {
		if err := handler(ctx, session, p); err != nil {
			log.Printf("❌ [ERROR] Handler error on %s: %v", routeKey, err)
		}
		return
	}

	// Di Menu Awal (StateRoot): Jika input BUKAN nomor menu 1 - 11,
	// anggap sebagai salam / teks pembuka (misal 'tes', 'halo', 'p', dll.) -> Tampilkan Menu Utama
	if session.CurrentState == StateRoot {
		log.Printf("👋 [ROUTER] Pesan pembuka '%s' diterima di menuAwal -> Menampilkan Menu Utama", input)
		if err := r.HandleWelcomeFull(ctx, session, p); err != nil {
			log.Printf("❌ [ERROR] Gagal mengirim Welcome Full: %v", err)
		}
		return
	}

	// Dynamic handler for Keyword Search (StateMenu1_Search)
	if session.CurrentState == StateMenu1_Search {
		if err := r.HandleSearchKeyword(ctx, session, p); err != nil {
			log.Printf("❌ [ERROR] Gagal memproses pencarian keyword: %v", err)
		}
		return
	}

	// Fallback when input not recognized di submenu
	log.Printf("⚠️ [ROUTER] Input '%s' tidak dikenali di state %s -> Mengirim pesan unavailable", input, session.CurrentState)
	if err := r.HandleUnavailable(ctx, session, p); err != nil {
		log.Printf("❌ [ERROR] Gagal mengirim HandleUnavailable: %v", err)
	}
}

func (r *BotRouter) registerRoutes() {
	// Level 1: Root Menu (1 - 11)
	r.handlerMap[StateRoot+":1"] = r.HandleMenu1_LayananData
	r.handlerMap[StateRoot+":2"] = r.HandleMenu2_Indikator
	r.handlerMap[StateRoot+":3"] = r.HandleMenu3_Produk
	r.handlerMap[StateRoot+":4"] = r.HandleMenu4_Video
	r.handlerMap[StateRoot+":5"] = r.HandleMenu5_Infografik
	r.handlerMap[StateRoot+":6"] = r.HandleMenu6_Konsultasi
	r.handlerMap[StateRoot+":7"] = r.HandleMenu7_Berlangganan
	r.handlerMap[StateRoot+":8"] = r.HandleMenu8_Pengaduan
	r.handlerMap[StateRoot+":9"] = r.HandleMenu9_HubDev
	r.handlerMap[StateRoot+":10"] = r.HandleMenu10_FAQ
	r.handlerMap[StateRoot+":11"] = r.HandleMenu11_TentangBPS

	// Level 2: Layanan Data Submenus (1 - 5)
	r.handlerMap[StateMenu1+":1"] = r.HandleLayananSub_Sosduk
	r.handlerMap[StateMenu1+":2"] = r.HandleLayananSub_Ekonomi
	r.handlerMap[StateMenu1+":3"] = r.HandleLayananSub_Pertanian
	r.handlerMap[StateMenu1+":4"] = r.HandleLayananSub_Sektoral
	r.handlerMap[StateMenu1+":5"] = r.HandleLayananSub_SearchPrompt

	// Level 3: Sosduk Topics (1 - 12)
	for i := 1; i <= 12; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu1_Sosduk, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleSosdukTopic(ctx, s, p, idx)
		}
	}

	// Level 3: Ekonomi Topics (1 - 10)
	for i := 1; i <= 10; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu1_Ekonomi, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleEkonomiTopic(ctx, s, p, idx)
		}
	}

	// Level 3: Pertanian Topics (1 - 2)
	for i := 1; i <= 2; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu1_Pertanian, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandlePertanianTopic(ctx, s, p, idx)
		}
	}

	// Level 2: Indikator Strategis Submenus (1 - 7)
	for i := 1; i <= 7; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu2, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleIndikatorDetail(ctx, s, p, idx)
		}
	}

	// Level 2: Produk Statistik (1 - 3)
	for i := 1; i <= 3; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu3, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleProdukDetail(ctx, s, p, idx)
		}
	}

	// Level 2: Video Literasi (1 - 8)
	for i := 1; i <= 8; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu4, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleVideoDetail(ctx, s, p, idx)
		}
	}

	// Level 2: Infografik (1 - 4)
	for i := 1; i <= 4; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu5, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleInfografikDetail(ctx, s, p, idx)
		}
	}

	// Level 2: Konsultasi (1: Umum, 2: Pakar)
	r.handlerMap[StateMenu6+":1"] = r.HandleKonsultasiUmum
	r.handlerMap[StateMenu6+":2"] = r.HandleKonsultasiPakarList

	// Level 3: Pakar Bidang (1 - 8)
	for i := 1; i <= 8; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu6_Pakar, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandlePakarDetail(ctx, s, p, idx)
		}
	}

	// Level 2: FAQ Topics (1 - 10)
	for i := 1; i <= 10; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu10, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleFAQDetail(ctx, s, p, idx)
		}
	}
}
