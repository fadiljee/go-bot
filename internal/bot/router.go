package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"bot-istimewa/internal/bpsapi"
	"bot-istimewa/internal/config"
	"bot-istimewa/internal/gsheet"
	"bot-istimewa/internal/locale"
	"bot-istimewa/internal/scraper"
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
	StateMenu4_Sosduk    = "menu4_sosduk"
	StateMenu4_Ekonomi   = "menu4_ekonomi"
	StateMenu5           = "menu5"
	StateMenu6           = "menu6"
	StateMenu6_Booking   = "menu6_booking"
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
	cfg                 *config.Config
	sender              *whatsapp.Sender
	gsheet              *gsheet.Client
	bps                 *bpsapi.Client
	scraper             scraper.DataProvider
	sessions            *SessionManager
	handlerMap          map[string]func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error
	rateLimiter         *RateLimiter
	adminFailedAttempts map[string]int
	adminLockedUntil    map[string]time.Time
	adminMu             sync.Mutex
}

func (r *BotRouter) SetScraper(p scraper.DataProvider) {
	r.scraper = p
}

func NewBotRouter(cfg *config.Config, sender *whatsapp.Sender, gsheetClient *gsheet.Client, bpsClient ...*bpsapi.Client) *BotRouter {
	var bps *bpsapi.Client
	if len(bpsClient) > 0 && bpsClient[0] != nil {
		bps = bpsClient[0]
	} else {
		bps = bpsapi.NewClient(cfg)
	}

	r := &BotRouter{
		cfg:                 cfg,
		sender:              sender,
		gsheet:              gsheetClient,
		bps:                 bps,
		sessions:            NewSessionManager(cfg.IdleTimeoutMinutes),
		handlerMap:          make(map[string]func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error),
		rateLimiter:         NewRateLimiter(cfg.UserRateLimitPer10s, cfg.UserCooldownSeconds),
		adminFailedAttempts: make(map[string]int),
		adminLockedUntil:    make(map[string]time.Time),
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

	// 1. Per-User Token Bucket Rate Limiting (Anti-Flooding / DDoS Protection)
	if !r.rateLimiter.Allow(p.JID.User) {
		log.Printf("⚠️ [RATE LIMIT] Mengabaikan pesan dari %s (dalam cooldown / melebihi kuota pesan)", p.JID.User)
		return
	}

	// Get or create session
	session := r.sessions.GetOrCreate(p.JID, func(s *UserSession) {
		r.HandleIdleTimeout(context.Background(), s)
	})

	// Input sanitization & 100 character limit (Anti-ReDoS / Anti-OOM)
	input := strings.TrimSpace(p.Text)
	runes := []rune(input)
	if len(runes) > 100 {
		input = string(runes[:100])
	}
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

	// Opt-Out / Berhenti Berlangganan (Anti-Spam Report Protection)
	if inputLower == "berhenti_langganan" || inputLower == "unsub" || inputLower == "unsubscribe" || inputLower == "stop" {
		log.Printf("🔕 [UNSUBSCRIBE] Permintaan berhenti berlangganan dari %s", p.JID.User)
		session.ResetToRoot()
		msg := "🔕 *Berhenti Berlangganan Berhasil*\n\nNomor Anda telah dicatat untuk berhenti menerima siaran rilis data berkala BPS Kabupaten Bangka Barat.\n\nTerima kasih telah menggunakan layanan kami 🙏\nKetik *0* untuk kembali ke menu utama."
		_ = r.sender.SendTextMessage(ctx, p.JID, msg)
		return
	}

	// Admin Trigger Keyword Check with Whitelist & Lockout
	if inputLower == strings.ToLower(r.cfg.AdminTriggerKeyword) {
		// Whitelist check: only registered numbers in .env can trigger admin mode
		if !r.isAdminWhitelisted(p.JID.User) {
			log.Printf("🔒 [ADMIN AUDIT] Percobaan login admin dari nomor tidak dikenal: %s (silent reject)", p.JID.User)
			_ = r.HandleUnavailable(ctx, session, p)
			return
		}

		// Lockout check: verify if temporarily locked out due to previous failed attempts
		if locked, until := r.isAdminLockedOut(p.JID.User); locked {
			log.Printf("🔒 [ADMIN AUDIT] Nomor %s mencoba login tetapi masih terkunci hingga %s", p.JID.User, until.Format("15:04:05"))
			msg := fmt.Sprintf("⛔ Nomor Anda sedang dikunci dari percobaan admin hingga %s.\nSilakan coba lagi nanti atau ketik *0* untuk ke menu utama.", until.Format("15:04:05"))
			_ = r.sender.SendTextMessage(ctx, p.JID, msg)
			return
		}

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

	// Dynamic handler for Consultation Booking (StateMenu6_Booking)
	if session.CurrentState == StateMenu6_Booking {
		if err := r.HandleProcessBooking(ctx, session, p); err != nil {
			log.Printf("❌ [ERROR] Gagal memproses booking konsultasi: %v", err)
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

	// Level 3: Video Sub-Menu Sosduk (1 - 5)
	for i := 1; i <= 5; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu4_Sosduk, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleVideoDetail_Sosduk(ctx, s, p, idx)
		}
	}

	// Level 3: Video Sub-Menu Ekonomi (1 - 5)
	for i := 1; i <= 5; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu4_Ekonomi, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleVideoDetail_Ekonomi(ctx, s, p, idx)
		}
	}

	// Level 2: Infografik (1 - 4)
	for i := 1; i <= 4; i++ {
		idx := i
		r.handlerMap[fmt.Sprintf("%s:%d", StateMenu5, idx)] = func(ctx context.Context, s *UserSession, p *whatsapp.ParsedMessage) error {
			return r.HandleInfografikDetail(ctx, s, p, idx)
		}
	}

	// Level 2: Konsultasi (1: Booking Jadwal, 2: Pakar, 3: Info PST & Jam Layanan)
	r.handlerMap[StateMenu6+":1"] = r.HandleMenu6_BookingPrompt
	r.handlerMap[StateMenu6+":2"] = r.HandleKonsultasiPakarList
	r.handlerMap[StateMenu6+":3"] = r.HandleKonsultasiUmum

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

func (r *BotRouter) isAdminWhitelisted(phone string) bool {
	if len(r.cfg.AdminWhitelistNumbers) == 0 {
		return false
	}
	clean := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	if strings.HasPrefix(clean, "08") {
		clean = "62" + clean[1:]
	}
	for _, allowed := range r.cfg.AdminWhitelistNumbers {
		cand := strings.TrimPrefix(strings.TrimSpace(allowed), "+")
		if strings.HasPrefix(cand, "08") {
			cand = "62" + cand[1:]
		}
		if cand == clean {
			return true
		}
	}
	return false
}

func (r *BotRouter) isAdminLockedOut(phone string) (bool, time.Time) {
	r.adminMu.Lock()
	defer r.adminMu.Unlock()

	until, exists := r.adminLockedUntil[phone]
	if !exists {
		return false, time.Time{}
	}
	if time.Now().Before(until) {
		return true, until
	}
	delete(r.adminLockedUntil, phone)
	return false, time.Time{}
}

