#!/usr/bin/env python3
"""
Generator Dokumentasi Lengkap Sistem & Arsitektur Bot Istimewa v3.4
Badan Pusat Statistik (BPS) Kabupaten Bangka Barat
Format: Dokumen PDF Resmi dengan Diagram Vektor & Flowchart
"""

import os
import sys
from datetime import datetime

# Masukkan lokal lib directory ke sys.path jika diinstall via target
script_dir = os.path.dirname(os.path.abspath(__file__))
lib_dir = os.path.join(script_dir, "lib")
if os.path.exists(lib_dir) and lib_dir not in sys.path:
    sys.path.insert(0, lib_dir)

from reportlab.lib import colors
from reportlab.lib.pagesizes import A4
from reportlab.lib.units import cm, mm
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.platypus import (
    SimpleDocTemplate, Paragraph, Spacer, Table, TableStyle, PageBreak, KeepTogether, HRFlowable
)
from reportlab.pdfgen import canvas
from reportlab.graphics.shapes import Drawing, Rect, String, Line, Polygon, Group, Circle

# ==============================================================================
# WARNA & PALET RESMI BPS
# ==============================================================================
PRIMARY = colors.HexColor("#003366")      # BPS Navy Blue
SECONDARY = colors.HexColor("#0284C7")    # Sky Blue
ACCENT = colors.HexColor("#D97706")       # Gold / Amber
DARK_TEXT = colors.HexColor("#0F172A")    # Slate 900
BODY_TEXT = colors.HexColor("#334155")    # Slate 700
LIGHT_BG = colors.HexColor("#F8FAFC")     # Slate 50
BORDER_COLOR = colors.HexColor("#CBD5E1") # Slate 300
SUCCESS = colors.HexColor("#16A34A")      # Green 600
WARNING = colors.HexColor("#EA580C")      # Orange 600
CARD_BG = colors.HexColor("#F1F5F9")      # Slate 100

# ==============================================================================
# NUMBERED CANVAS (Dua Pass untuk "Halaman X dari Y" & Running Header/Footer)
# ==============================================================================
class NumberedCanvas(canvas.Canvas):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._saved_page_states = []

    def showPage(self):
        self._saved_page_states.append(dict(self.__dict__))
        self._startPage()

    def save(self):
        num_pages = len(self._saved_page_states)
        for state in self._saved_page_states:
            self.__dict__.update(state)
            self.draw_page_decorations(num_pages)
            super().showPage()
        super().save()

    def draw_page_decorations(self, page_count):
        if self._pageNumber == 1:
            return  # Jangan gambar header/footer di cover depan

        self.saveState()
        self.setFont("Helvetica", 8)
        self.setFillColor(colors.HexColor("#64748B"))

        # Running Header
        self.drawString(54, 800, "BPS KABUPATEN BANGKA BARAT  |  Dokumentasi Teknis Bot Istimewa v3.4")
        self.drawRightString(541, 800, "Sistem Chatbot & REST API")
        self.setStrokeColor(BORDER_COLOR)
        self.setLineWidth(0.5)
        self.line(54, 792, 541, 792)

        # Running Footer
        self.line(54, 45, 541, 45)
        self.drawString(54, 32, "Arsitektur Native Go, Web Scraper SQLite, & REST API Terintegrasi")
        page_str = f"Halaman {self._pageNumber} dari {page_count}"
        self.drawRightString(541, 32, page_str)
        self.restoreState()


# ==============================================================================
# PEMBUAT DIAGRAM VEKTOR & FLOWCHART (ReportLab Graphics)
# ==============================================================================
def draw_rounded_box(d, x, y, w, h, fill_col, border_col, title, subtitle="", rx=4, ry=4, title_col=colors.white, sub_col=colors.HexColor("#E2E8F0")):
    d.add(Rect(x, y, w, h, rx=rx, ry=ry, fillColor=fill_col, strokeColor=border_col, strokeWidth=1))
    if subtitle:
        d.add(String(x + w/2, y + h/2 + 3, title, fontName="Helvetica-Bold", fontSize=8.5, textAnchor="middle", fillColor=title_col))
        d.add(String(x + w/2, y + h/2 - 8, subtitle, fontName="Helvetica", fontSize=7, textAnchor="middle", fillColor=sub_col))
    else:
        d.add(String(x + w/2, y + h/2 - 3, title, fontName="Helvetica-Bold", fontSize=8.5, textAnchor="middle", fillColor=title_col))

def draw_arrow(d, x1, y1, x2, y2, color=colors.HexColor("#64748B"), width=1):
    d.add(Line(x1, y1, x2, y2, strokeColor=color, strokeWidth=width))
    # Head arrow
    if x2 > x1 and y1 == y2: # right
        d.add(Polygon([x2, y2, x2-5, y2+3, x2-5, y2-3], fillColor=color, strokeColor=color))
    elif x2 < x1 and y1 == y2: # left
        d.add(Polygon([x2, y2, x2+5, y2+3, x2+5, y2-3], fillColor=color, strokeColor=color))
    elif y2 < y1 and x1 == x2: # down
        d.add(Polygon([x2, y2, x2-3, y2+5, x2+3, y2+5], fillColor=color, strokeColor=color))
    elif y2 > y1 and x1 == x2: # up
        d.add(Polygon([x2, y2, x2-3, y2-5, x2+3, y2-5], fillColor=color, strokeColor=color))


def create_architecture_diagram():
    """Diagram Arsitektur Komponen Global Bot Istimewa v3.4"""
    d = Drawing(490, 200)
    # Background Canvas
    d.add(Rect(0, 0, 490, 200, rx=6, ry=6, fillColor=CARD_BG, strokeColor=BORDER_COLOR, strokeWidth=0.8))
    d.add(String(15, 185, "DIAGRAM ARSITEKTUR GLOBAL SISTEM (ZERO-GAS DEPENDENCY)", fontName="Helvetica-Bold", fontSize=9, fillColor=PRIMARY))

    # Box 1: External User
    draw_rounded_box(d, 15, 120, 95, 45, colors.HexColor("#1E293B"), colors.HexColor("#475569"), "Masyarakat / OPD", "WhatsApp Mobile & Web")
    draw_arrow(d, 110, 142, 135, 142)

    # Box 2: Whatsmeow Connector
    draw_rounded_box(d, 135, 120, 105, 45, SECONDARY, colors.HexColor("#0369A1"), "Whatsmeow Core", "Native Go WebSocket")
    draw_arrow(d, 240, 142, 265, 142)

    # Box 3: Router & State Machine
    draw_rounded_box(d, 265, 115, 115, 55, PRIMARY, colors.HexColor("#0F172A"), "Bot Router & State", "Dispatcher & Limiter")
    
    # Arrow to WhatsApp Internal Group
    draw_arrow(d, 322, 170, 322, 182)
    draw_arrow(d, 322, 182, 400, 182)
    d.add(String(345, 185, "Auto-Forward Tiket", fontName="Helvetica", fontSize=6.5, fillColor=WARNING))

    # Box 4: Internal PST Group
    draw_rounded_box(d, 400, 155, 80, 35, WARNING, colors.HexColor("#C2410C"), "Grup WA PST", "Notifikasi Kantor")

    # Arrow Router to Database
    draw_arrow(d, 322, 115, 322, 85)

    # Box 5: SQLite Database
    draw_rounded_box(d, 265, 35, 115, 50, colors.HexColor("#047857"), colors.HexColor("#065F46"), "SQLite Storage", "scraper.db & whatsmeow.db")

    # Box 6: Scraper Engine
    draw_rounded_box(d, 135, 35, 105, 50, ACCENT, colors.HexColor("#B45309"), "Mesin Scraper", "Next.js Actions & Cron 4h")
    draw_arrow(d, 240, 60, 265, 60)
    d.add(String(243, 64, "Sync", fontName="Helvetica-Bold", fontSize=6.5, fillColor=SUCCESS))

    # Box 7: Portal BPS External
    draw_rounded_box(d, 15, 35, 95, 50, colors.HexColor("#475569"), colors.HexColor("#334155"), "Portal BPS Babar", "bangkabaratkab.bps.go.id")
    draw_arrow(d, 110, 60, 135, 60)
    d.add(String(113, 64, "Scrape", fontName="Helvetica", fontSize=6.5, fillColor=BODY_TEXT))

    # Box 8: REST API Fiber
    draw_rounded_box(d, 400, 35, 80, 85, colors.HexColor("#7C3AED"), colors.HexColor("#6D28D9"), "REST API", "Fiber v2 Port 3000\nPublik & Admin")
    draw_arrow(d, 380, 60, 400, 60)
    draw_arrow(d, 380, 135, 400, 110)
    
    return d


def create_bot_workflow_diagram():
    """Diagram Alur Kerja Interaktif WhatsApp Chatbot (State Machine)"""
    d = Drawing(490, 180)
    d.add(Rect(0, 0, 490, 180, rx=6, ry=6, fillColor=CARD_BG, strokeColor=BORDER_COLOR, strokeWidth=0.8))
    d.add(String(15, 165, "FLOWCHART ALUR DISPATCHER & STATE MACHINE WHATSAPP", fontName="Helvetica-Bold", fontSize=9, fillColor=PRIMARY))

    # 1. Pesan Masuk
    draw_rounded_box(d, 15, 115, 85, 35, colors.HexColor("#334155"), colors.HexColor("#1E293B"), "Pesan Masuk WA", "Sanitasi Input <= 100")
    draw_arrow(d, 100, 132, 120, 132)

    # 2. Rate Limiter
    draw_rounded_box(d, 120, 115, 85, 35, colors.HexColor("#0284C7"), colors.HexColor("#0369A1"), "Rate Limiter", "5 msg / 10s Bucket")
    draw_arrow(d, 205, 132, 225, 132)

    # 3. Decision Admin Keyword
    d.add(Polygon([260, 147, 295, 132, 260, 117, 225, 132], fillColor=ACCENT, strokeColor=colors.HexColor("#B45309"), strokeWidth=1))
    d.add(String(260, 134, "Keyword", fontName="Helvetica-Bold", fontSize=7, textAnchor="middle", fillColor=colors.white))
    d.add(String(260, 125, "Admin?", fontName="Helvetica-Bold", fontSize=7, textAnchor="middle", fillColor=colors.white))

    # Path Ya -> Mode Admin
    draw_arrow(d, 260, 147, 260, 160)
    draw_arrow(d, 260, 160, 310, 160)
    d.add(String(270, 163, "Ya", fontName="Helvetica-Bold", fontSize=7, fillColor=SUCCESS))
    draw_rounded_box(d, 310, 142, 165, 30, WARNING, colors.HexColor("#C2410C"), "Mode Admin (Password Auth)", "Kelola Pakar | Status | Scrape On-Demand")

    # Path Tidak -> Menu Publik
    draw_arrow(d, 295, 132, 320, 132)
    d.add(String(302, 135, "Bukan", fontName="Helvetica", fontSize=7, fillColor=BODY_TEXT))
    draw_rounded_box(d, 320, 115, 155, 32, PRIMARY, colors.HexColor("#0F172A"), "Menu Utama (Root)", "Katalog 11 Layanan Publik")

    # Sub-Menu Branches
    draw_arrow(d, 397, 115, 397, 95)
    
    # 4 Kotak Cabang Menu Utama
    draw_rounded_box(d, 15, 45, 105, 40, colors.HexColor("#059669"), colors.HexColor("#047857"), "Menu 1: Layanan Data", "25 Topik + Direct Tables")
    draw_rounded_box(d, 130, 45, 105, 40, colors.HexColor("#0284C7"), colors.HexColor("#0369A1"), "Menu 2: Indikator", "7 Indikator Strategis Live")
    draw_rounded_box(d, 245, 45, 115, 40, colors.HexColor("#7C3AED"), colors.HexColor("#6D28D9"), "Menu 3: BRS & Publikasi", "Cover Image + Ringkasan")
    draw_rounded_box(d, 370, 45, 105, 40, colors.HexColor("#D97706"), colors.HexColor("#B45309"), "Menu 6: PST & Pakar", "Reservasi Tiket + Direct WA")

    # Connectors
    draw_arrow(d, 397, 95, 67, 95)
    draw_arrow(d, 67, 95, 67, 85)
    draw_arrow(d, 182, 95, 182, 85)
    draw_arrow(d, 302, 95, 302, 85)
    draw_arrow(d, 397, 95, 422, 95)
    draw_arrow(d, 422, 95, 422, 85)

    # Note
    d.add(String(15, 15, "* Seluruh respon menyajikan data dan indikator langsung di WhatsApp (Zero broken links / link not found fix).", fontName="Helvetica-Oblique", fontSize=7, fillColor=BODY_TEXT))
    return d


def create_scraper_workflow_diagram():
    """Diagram Alur Eksekusi Mesin Scraper & Cache Invalidation"""
    d = Drawing(490, 140)
    d.add(Rect(0, 0, 490, 140, rx=6, ry=6, fillColor=CARD_BG, strokeColor=BORDER_COLOR, strokeWidth=0.8))
    d.add(String(15, 125, "FLOWCHART MESIN SCRAPER & SINKRONISASI DATABASE SQLITE", fontName="Helvetica-Bold", fontSize=9, fillColor=PRIMARY))

    # Pemicu (Trigger)
    draw_rounded_box(d, 15, 65, 80, 45, colors.HexColor("#1E293B"), colors.HexColor("#0F172A"), "Pemicu", "• Cron 4-Jam\n• Admin WA\n• REST API")
    draw_arrow(d, 95, 87, 115, 87)

    # Circuit Breaker
    draw_rounded_box(d, 115, 65, 85, 45, ACCENT, colors.HexColor("#B45309"), "Circuit Breaker", "Fail-Safe Monitor\n(Maks 3 Gagal)")
    draw_arrow(d, 200, 87, 220, 87)

    # Fetcher & Next.js Actions
    draw_rounded_box(d, 220, 65, 95, 45, SECONDARY, colors.HexColor("#0369A1"), "Fetcher & Parser", "Polite Delay (3s)\nNext.js Actions")
    draw_arrow(d, 315, 87, 335, 87)

    # Upsert SQLite
    draw_rounded_box(d, 335, 65, 75, 45, colors.HexColor("#059669"), colors.HexColor("#047857"), "SQLite Store", "Insert Baru /\nUpdate Hash")
    draw_arrow(d, 410, 87, 430, 87)

    # Invalidate Cache & Notify
    draw_rounded_box(d, 430, 65, 50, 45, colors.HexColor("#7C3AED"), colors.HexColor("#6D28D9"), "Sync", "Invalidate\n& Rekap")

    # Result Footprint
    d.add(Rect(15, 15, 465, 32, rx=4, ry=4, fillColor=colors.white, strokeColor=BORDER_COLOR, strokeWidth=0.5))
    d.add(String(25, 33, "Keluaran Sinkronisasi:", fontName="Helvetica-Bold", fontSize=7.5, fillColor=DARK_TEXT))
    d.add(String(25, 22, "100 Tabel Statis  |  50 Buku Publikasi  |  34 Berita Resmi Statistik (BRS)  |  49 Berita Daerah  |  8 Kontak Pakar", fontName="Helvetica", fontSize=7, fillColor=BODY_TEXT))

    return d


# ==============================================================================
# FUNGSI UTAMA GENERASI DOKUMEN PDF
# ==============================================================================
def generate_docs_pdf(output_filename="DOKUMENTASI_SISTEM_BOT_ISTIMEWA.pdf"):
    print(f"[*] Memulai kompilasi dokumen PDF: {output_filename} ...")

    doc = SimpleDocTemplate(
        output_filename,
        pagesize=A4,
        leftMargin=54,   # ~1.9 cm
        rightMargin=54,
        topMargin=54,
        bottomMargin=54
    )

    styles = getSampleStyleSheet()

    # Custom Typography Styles
    title_style = ParagraphStyle(
        "CoverTitle",
        parent=styles["Normal"],
        fontName="Helvetica-Bold",
        fontSize=24,
        leading=28,
        textColor=PRIMARY,
        alignment=0,
        spaceAfter=8
    )

    subtitle_style = ParagraphStyle(
        "CoverSubtitle",
        parent=styles["Normal"],
        fontName="Helvetica",
        fontSize=12,
        leading=16,
        textColor=SECONDARY,
        alignment=0,
        spaceAfter=15
    )

    h1_style = ParagraphStyle(
        "Heading1_Custom",
        parent=styles["Heading1"],
        fontName="Helvetica-Bold",
        fontSize=14,
        leading=18,
        textColor=PRIMARY,
        spaceBefore=14,
        spaceAfter=6,
        keepWithNext=True
    )

    h2_style = ParagraphStyle(
        "Heading2_Custom",
        parent=styles["Heading2"],
        fontName="Helvetica-Bold",
        fontSize=10.5,
        leading=14,
        textColor=SECONDARY,
        spaceBefore=10,
        spaceAfter=4,
        keepWithNext=True
    )

    body_style = ParagraphStyle(
        "Body_Custom",
        parent=styles["BodyText"],
        fontName="Helvetica",
        fontSize=8.5,
        leading=12,
        textColor=BODY_TEXT,
        spaceAfter=6
    )

    bullet_style = ParagraphStyle(
        "Bullet_Custom",
        parent=styles["BodyText"],
        fontName="Helvetica",
        fontSize=8.5,
        leading=11.5,
        textColor=BODY_TEXT,
        leftIndent=14,
        spaceAfter=3
    )

    code_style = ParagraphStyle(
        "Code_Custom",
        parent=styles["Normal"],
        fontName="Courier",
        fontSize=7.5,
        leading=10,
        textColor=DARK_TEXT
    )

    callout_style = ParagraphStyle(
        "Callout_Custom",
        parent=styles["Normal"],
        fontName="Helvetica-Oblique",
        fontSize=8,
        leading=11,
        textColor=DARK_TEXT
    )

    table_header_style = ParagraphStyle(
        "TableHeader",
        parent=styles["Normal"],
        fontName="Helvetica-Bold",
        fontSize=8,
        leading=10,
        textColor=colors.white
    )

    table_cell_style = ParagraphStyle(
        "TableCell",
        parent=styles["Normal"],
        fontName="Helvetica",
        fontSize=7.5,
        leading=9.5,
        textColor=BODY_TEXT
    )

    table_cell_bold = ParagraphStyle(
        "TableCellBold",
        parent=styles["Normal"],
        fontName="Helvetica-Bold",
        fontSize=7.5,
        leading=9.5,
        textColor=DARK_TEXT
    )

    story = []

    # ==========================================================================
    # HALAMAN COVER
    # ==========================================================================
    story.append(Spacer(1, 20))
    story.append(Paragraph("BADAN PUSAT STATISTIK KABUPATEN BANGKA BARAT", ParagraphStyle(
        "AgencyHeader", fontName="Helvetica-Bold", fontSize=11, leading=14, textColor=ACCENT
    )))
    story.append(Spacer(1, 10))
    story.append(Paragraph("DOKUMENTASI SISTEM & ARSITEKTUR BOT ISTIMEWA v3.4", title_style))
    story.append(Paragraph("Platform Layanan Data Statistik Terpadu: WhatsApp Chatbot Native Go, Web Scraper SQLite, & REST API Terintegrasi", subtitle_style))
    
    story.append(HRFlowable(width="100%", thickness=2, color=PRIMARY, spaceBefore=4, spaceAfter=15))

    # Info Metadata Box
    meta_data = [
        [Paragraph("<b>Satker / Pemilik Sistem:</b>", table_cell_bold), Paragraph("BPS Kabupaten Bangka Barat (Domain: 1903)", table_cell_style)],
        [Paragraph("<b>Arsitektur Inti:</b>", table_cell_bold), Paragraph("Golang (whatsmeow + fiber/v2), SQLite Storage, Next.js Actions Extractor", table_cell_style)],
        [Paragraph("<b>Status Dependensi:</b>", table_cell_bold), Paragraph("100% Mandiri (Zero-GAS Dependency / Tidak Tergantung Google Sheets)", table_cell_style)],
        [Paragraph("<b>Basis Data Utama:</b>", table_cell_bold), Paragraph("SQLite (storage/scraper.db & storage/whatsmeow.db)", table_cell_style)],
        [Paragraph("<b>Penjadwalan Scraping:</b>", table_cell_bold), Paragraph("Otomatis setiap 4 jam (Cron: 0 */4 * * *) & On-Demand via Admin WA", table_cell_style)],
        [Paragraph("<b>Tanggal Rilis Dokumen:</b>", table_cell_bold), Paragraph(datetime.now().strftime("%d %B %Y, %H:%M WIB"), table_cell_style)],
    ]
    t_meta = Table(meta_data, colWidths=[150, 330])
    t_meta.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,-1), CARD_BG),
        ('BOX', (0,0), (-1,-1), 1, BORDER_COLOR),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('PADDING', (0,0), (-1,-1), 5),
        ('LINEBELOW', (0,0), (-1,-2), 0.5, BORDER_COLOR),
    ]))
    story.append(t_meta)
    story.append(Spacer(1, 15))

    # Ringkasan Eksekutif
    story.append(Paragraph("Ringkasan Eksekutif", h1_style))
    story.append(Paragraph(
        "<b>Bot Istimewa v3.4</b> adalah platform otomasi diseminasi data statistik resmi Badan Pusat Statistik Kabupaten Bangka Barat. "
        "Sistem ini dirancang untuk menggantikan arsitektur lama berbasis Google Apps Script (GAS) yang rentan terhadap batas kuota harian, waktu tunggu eksekusi (timeout), dan tautan portal eksternal yang kedaluwarsa. "
        "Dengan fondasi <b>Golang</b> murni, sistem menyatukan tiga kapabilitas dalam satu kesatuan runtime berbobot ringan (<35 MB RAM): "
        "(1) <i>Chatbot WhatsApp Interaktif</i> dengan penyajian data angka statistik langsung, "
        "(2) <i>Mesin Web Scraper Otomatis</i> dengan basis data SQLite lokal, dan "
        "(3) <i>Layanan REST API Berkecepatan Tinggi</i> untuk integrasi dashboard maupun aplikasi pihak ketiga.",
        body_style
    ))
    story.append(Spacer(1, 10))

    # Diagram Arsitektur Global di Halaman 1
    story.append(Paragraph("Diagram Arsitektur Sistem", h2_style))
    story.append(create_architecture_diagram())
    story.append(PageBreak())

    # ==========================================================================
    # BAB 1: ALUR KERJA CHATBOT WHATSAPP & STATE MACHINE
    # ==========================================================================
    story.append(Paragraph("1. Alur Kerja Chatbot WhatsApp & State Machine", h1_style))
    story.append(Paragraph(
        "Interaksi pengguna dilayani melalui perutean berbasis status (<i>table-driven state machine</i>) yang menjaga konteks percakapan secara terisolasi per nomor telepon pengguna. "
        "Setiap pesan yang masuk melalui koneksi WebSocket whatsmeow melewati serangkaian filter keamanan sebelum dieksekusi oleh modul perutean (<code>BotRouter</code>).",
        body_style
    ))
    story.append(Spacer(1, 6))

    # Flowchart WA
    story.append(create_bot_workflow_diagram())
    story.append(Spacer(1, 10))

    story.append(Paragraph("1.1 Tahapan Pemrosesan Pesan Masuk", h2_style))
    story.append(Paragraph("• <b>Filter Server & Broadcast:</b> Mengabaikan pesan dari channel siaran, grup publik, atau newsletter.", bullet_style))
    story.append(Paragraph("• <b>Sanitasi Input:</b> Teks dibatasi maksimal 100 karakter untuk mencegah serangan Denial of Service / Regular Expression DoS (ReDoS).", bullet_style))
    story.append(Paragraph("• <b>Token-Bucket Rate Limiter:</b> Setiap nomor dibatasi maksimal 5 pesan per jendela waktu 10 detik. Jika melebihi, pesan ditolak selama 60 detik cooldown.", bullet_style))
    story.append(Paragraph("• <b>Pendeteksian Kata Kunci Global:</b> Kata kunci <code>reset</code>, <code>100</code> (menu lengkap), <code>0</code> (menu singkat), dan <code>99</code> (kembali) memotong hierarki state kapan saja.", bullet_style))
    story.append(Paragraph("• <b>Penyajian Data Langsung (Zero Broken Links Fix):</b> Pada versi v3.4, seluruh modul (Menu 1 Layanan Data, Menu 2 Indikator Strategis, dan Menu 3 Produk Statistik) menyajikan angka-angka statistik, deskripsi indikator, dan judul tabel resmi langsung di dalam chat teks WhatsApp tanpa mengarahkan pengguna ke tautan URL luar yang rentan menghasilkan respon HTTP 404 / 400 Bad Request.", bullet_style))
    story.append(Spacer(1, 10))

    # Tabel Katalog Menu WhatsApp
    story.append(Paragraph("1.2 Katalog Layanan Chatbot WhatsApp (11 Menu Utama)", h2_style))
    menu_table_data = [
        [Paragraph("No", table_header_style), Paragraph("Menu Layanan", table_header_style), Paragraph("Modul & Sumber Data", table_header_style), Paragraph("Fitur Unggulan", table_header_style)],
        [Paragraph("1", table_cell_bold), Paragraph("Layanan Data Statistik", table_cell_bold), Paragraph("SQLite (statictable)", table_cell_style), Paragraph("25 Topik (Sosduk, Ekonomi, Pertanian, Sektoral) disajikan langsung.", table_cell_style)],
        [Paragraph("2", table_cell_bold), Paragraph("Indikator Strategis", table_cell_bold), Paragraph("Provider Indikator SQLite", table_cell_style), Paragraph("7 Indikator Utama: Inflasi, Kemiskinan, PDRB, IPM, dll.", table_cell_style)],
        [Paragraph("3", table_cell_bold), Paragraph("Produk Statistik", table_cell_bold), Paragraph("SQLite (BRS & Publikasi)", table_cell_style), Paragraph("Kirim foto cover resmi via SendImageFromURL + ringkasan rilis.", table_cell_style)],
        [Paragraph("4", table_cell_bold), Paragraph("Video Literasi Statistik", table_cell_bold), Paragraph("YouTube PiP In-App", table_cell_style), Paragraph("Putar langsung di aplikasi WA tanpa membebani storage server.", table_cell_style)],
        [Paragraph("5", table_cell_bold), Paragraph("Infografik Statistik", table_cell_bold), Paragraph("SQLite (infographic)", table_cell_style), Paragraph("Katalog infografis tematik visual statistik daerah.", table_cell_style)],
        [Paragraph("6", table_cell_bold), Paragraph("Konsultasi & Kontak Pakar", table_cell_bold), Paragraph("SQLite (bookings & pakar)", table_cell_style), Paragraph("Reservasi tiket konsultasi PST + Direct Chat 8 Pakar Statistik.", table_cell_style)],
        [Paragraph("7", table_cell_bold), Paragraph("Berlangganan Gratis", table_cell_bold), Paragraph("SQLite Langganan", table_cell_style), Paragraph("Pendaftaran broadcast rilis BRS berkala + proteksi unsubscribe.", table_cell_style)],
        [Paragraph("8", table_cell_bold), Paragraph("Pengaduan & Kritik", table_cell_bold), Paragraph("Internal Router", table_cell_style), Paragraph("Formulir masukan pelayanan publik dan pengaduan PST.", table_cell_style)],
        [Paragraph("9", table_cell_bold), Paragraph("Hubungi Developer", table_cell_bold), Paragraph("Internal Router", table_cell_style), Paragraph("Tautan komunikasi teknis langsung ke pengembang bot.", table_cell_style)],
        [Paragraph("10", table_cell_bold), Paragraph("FAQ (Tanya Jawab)", table_cell_bold), Paragraph("Locale Database", table_cell_style), Paragraph("10 Topik FAQ terpopuler seputar layanan BPS Bangka Barat.", table_cell_style)],
        [Paragraph("11", table_cell_bold), Paragraph("Tentang BPS Babar", table_cell_bold), Paragraph("Locale & Video Profil", table_cell_style), Paragraph("Profil Satker 1903, visi misi, tugas fungsi, dan video institusi.", table_cell_style)],
    ]
    t_menu = Table(menu_table_data, colWidths=[25, 120, 125, 215])
    t_menu.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,0), PRIMARY),
        ('BOX', (0,0), (-1,-1), 0.8, BORDER_COLOR),
        ('GRID', (0,0), (-1,-1), 0.5, BORDER_COLOR),
        ('ROWBACKGROUNDS', (0,1), (-1,-1), [colors.white, CARD_BG]),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('PADDING', (0,0), (-1,-1), 4),
    ]))
    story.append(t_menu)
    story.append(PageBreak())

    # ==========================================================================
    # BAB 2: MESIN WEB SCRAPER & SINKRONISASI DATABASE
    # ==========================================================================
    story.append(Paragraph("2. Mesin Web Scraper BPS Bangka Barat & Basis Data", h1_style))
    story.append(Paragraph(
        "Mesin scraper bekerja secara mandiri dan aman mengambil data langsung dari portal resmi <b>bangkabaratkab.bps.go.id</b>. "
        "Modul ini mengekstraksi payload modern <i>Next.js Server Actions</i> tanpa memerlukan browser Chrome / headless yang rakus memori, "
        "sehingga bot dapat berjalan lancar pada VPS berspesifikasi hemat (1 vCPU, 512 MB RAM).",
        body_style
    ))
    story.append(Spacer(1, 6))

    # Flowchart Scraper
    story.append(create_scraper_workflow_diagram())
    story.append(Spacer(1, 10))

    story.append(Paragraph("2.1 Mekanisme Kesopanan & Perlindungan Server (Ethical Scraping)", h2_style))
    story.append(Paragraph("• <b>Circuit Breaker Pattern:</b> Memantau kegagalan beruntun. Jika terjadi 3 kali error berturut-turut, sistem otomatis menghentikan penarikan data sementara (*tripped/open*) selama 60 menit untuk mencegah pembebanan server BPS.", bullet_style))
    story.append(Paragraph("• <b>Jitter & Rate Limiter Internal:</b> Setiap request HTTP disisipi jeda acak (*polite delay*) antara 3.000 ms hingga 4.000 ms (Jitter 1.000 ms).", bullet_style))
    story.append(Paragraph("• <b>User-Agent Transparan:</b> Menggunakan identitas resmi <code>BotIstimewa/3.1 (+https://bangkabaratkab.bps.go.id; kontak: bps1903@bps.go.id)</code> dan patuh penuh terhadap instruksi <code>robots.txt</code>.", bullet_style))
    story.append(Paragraph("• <b>Deduplikasi Konten Berbasis Hash (SHA-256):</b> Mencegah penyimpanan data ganda. Item hanya diperbarui jika hash konten berubah, menghemat I/O disk secara signifikan.", bullet_style))
    story.append(Paragraph("• <b>In-Memory Cache Invalidation:</b> Begitu siklus scraping selesai, memori cache Layanan Data dan Indikator Strategis langsung dibersihkan seketika (<i>evicted</i>) sehingga query pengguna berikutnya menerima data termutakhir.", bullet_style))
    story.append(Spacer(1, 10))

    story.append(Paragraph("2.2 Komposisi Data Hasil Scraping di Database SQLite (Terkini)", h2_style))
    scraper_stat_data = [
        [Paragraph("Kategori Data", table_header_style), Paragraph("Kode Kind", table_header_style), Paragraph("Total Item Tersimpan", table_header_style), Paragraph("Karakteristik Data", table_header_style)],
        [Paragraph("Tabel Statistik Statis", table_cell_bold), Paragraph("<code>statictable</code>", table_cell_style), Paragraph("<b>100 Item</b>", table_cell_bold), Paragraph("Data sektoral, demografi, inflasi, PDRB, dan tabel portal resmi.", table_cell_style)],
        [Paragraph("Buku Publikasi BPS", table_cell_bold), Paragraph("<code>publication</code>", table_cell_style), Paragraph("<b>50 Item</b>", table_cell_bold), Paragraph("Publikasi resmi, Daerah Dalam Angka, cover buku, & nomor katalog.", table_cell_style)],
        [Paragraph("Berita Resmi Statistik (BRS)", table_cell_bold), Paragraph("<code>pressrelease</code>", table_cell_style), Paragraph("<b>34 Item</b>", table_cell_bold), Paragraph("Rilis inflasi bulanan, ketenagakerjaan, kemiskinan, & pertumbuhan ekonomi.", table_cell_style)],
        [Paragraph("Berita Kegiatan Daerah", table_cell_bold), Paragraph("<code>news</code>", table_cell_style), Paragraph("<b>49 Item</b>", table_cell_bold), Paragraph("Artikel dokumentasi sensus, survei lapangan, dan kegiatan Satker 1903.", table_cell_style)],
        [Paragraph("Kontak Narahubung Pakar", table_cell_bold), Paragraph("<code>expert_contacts</code>", table_cell_style), Paragraph("<b>8 Bidang</b>", table_cell_bold), Paragraph("Nomor WA & nama pakar statistik (Sosduk, Ekonomi, Pertanian, dll).", table_cell_style)],
    ]
    t_stat = Table(scraper_stat_data, colWidths=[120, 85, 95, 185])
    t_stat.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,0), SECONDARY),
        ('BOX', (0,0), (-1,-1), 0.8, BORDER_COLOR),
        ('GRID', (0,0), (-1,-1), 0.5, BORDER_COLOR),
        ('ROWBACKGROUNDS', (0,1), (-1,-1), [colors.white, CARD_BG]),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('PADDING', (0,0), (-1,-1), 4),
    ]))
    story.append(t_stat)
    story.append(PageBreak())

    # ==========================================================================
    # BAB 3: DASHBOARD ADMIN & KONTROL LIVE VIA WHATSAPP
    # ==========================================================================
    story.append(Paragraph("3. Dashboard Administrator WhatsApp & Manajemen Kontak Pakar", h1_style))
    story.append(Paragraph(
        "Administrator BPS Kabupaten Bangka Barat dapat mengelola operasional bot dan basis data langsung melalui percakapan WhatsApp "
        "tanpa perlu membuka server atau mengedit file kode program. Dashboard ini dilengkapi proteksi keamanan berlapis.",
        body_style
    ))
    story.append(Spacer(1, 6))

    admin_box_data = [
        [Paragraph("<b>Cara Akses Mode Admin WhatsApp:</b>", table_cell_bold)],
        [Paragraph("1. Ketik kata kunci pemicu: <code>login admin</code> (atau <code>admin 1903</code>, <code>login admin - 1903</code>).", table_cell_style)],
        [Paragraph("2. Masukkan password administrator sesuai konfigurasi <code>ADMIN_PASSWORD</code> (default: <code>supertim1903</code>).", table_cell_style)],
        [Paragraph("3. Dilengkapi sistem proteksi: Jika nomor salah memasukkan password sebanyak 3 kali berturut-turut, nomor tersebut otomatis dikunci (<i>lockout</i>) selama 60 menit.", table_cell_style)],
    ]
    t_admin_box = Table(admin_box_data, colWidths=[485])
    t_admin_box.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,-1), colors.HexColor("#FEF3C7")),
        ('BOX', (0,0), (-1,-1), 1, ACCENT),
        ('PADDING', (0,0), (-1,-1), 5),
    ]))
    story.append(t_admin_box)
    story.append(Spacer(1, 10))

    story.append(Paragraph("3.1 Empat Menu Utama Dashboard Administrator", h2_style))
    story.append(Paragraph(
        "<b>Opsi 1 — Kelola Narahubung Para Pakar (SQLite Database):</b><br/>"
        "Menampilkan daftar narahubung 8 bidang statistik saat ini di database. Admin dapat mengubah nomor kontak atau nama pakar dengan format pesan langsung:<br/>"
        "<code>&lt;nomor_bidang&gt; &lt;no_wa&gt; &lt;nama_pakar&gt;</code> (Contoh: <code>1 08123456789 Bpk. Rahmat, S.Tr.Stat.</code>).<br/>"
        "Data otomatis tersimpan ke tabel <code>expert_contacts</code> di SQLite dan langsung terbarukan di chat WhatsApp publik maupun REST API.",
        body_style
    ))
    story.append(Paragraph(
        "<b>Opsi 2 — Status Mesin Scraper & Database Lokal:</b><br/>"
        "Menampilkan statistik riil isi database SQLite (jumlah tabel statis, publikasi, BRS, dan berita), status circuit breaker, serta riwayat siklus scraping terakhir (waktu mulai, durasi eksekusi, pemicu cron/manual, dan jumlah data baru/diperbarui).",
        body_style
    ))
    story.append(Paragraph(
        "<b>Opsi 3 — Pemicu Scraping On-Demand:</b><br/>"
        "Menjalankan proses scraping publikasi, BRS, dan tabel statis dari portal BPS secara instan di background. "
        "Bot mengirim respon awal konfirmasi dan otomatis mengirimkan pesan laporan rekap hasil ke nomor admin begitu selesai.",
        body_style
    ))
    story.append(Paragraph(
        "<b>Opsi 4 — Tautan Google Sheets (Legacy):</b><br/>"
        "Menyajikan tautan spreadsheet arsip cadangan untuk keperluan verifikasi historis.",
        body_style
    ))
    story.append(Spacer(1, 10))

    # ==========================================================================
    # BAB 4: RESERVASI KONSULTASI PST & FORWARD KE GRUP KANTOR
    # ==========================================================================
    story.append(Paragraph("4. Alur Reservasi Konsultasi PST & Penerusan ke Grup Internal", h1_style))
    story.append(Paragraph(
        "Fitur <b>Pelayanan Statistik Terpadu (PST) Online (Menu 6)</b> memungkinkan masyarakat, mahasiswa, maupun instansi pemerintah daerah (OPD) "
        "mengajukan permohonan konsultasi data langsung dari WhatsApp. Alur kerja reservasi berlangsung otomatis sebagai berikut:",
        body_style
    ))

    pst_steps = [
        [Paragraph("Langkah 1: Pemilihan Layanan", table_cell_bold), Paragraph("Pengguna memilih <b>Menu 6 (Konsultasi PST)</b> lalu memilih opsi <b>1 (Ajukan Konsultasi / Booking)</b>.", table_cell_style)],
        [Paragraph("Langkah 2: Pengisian Formulir", table_cell_bold), Paragraph("Pengguna mengisi data nama lengkap, instansi, nomor WA aktif, topik data yang dibutuhkan, dan rencana tanggal kunjungan.", table_cell_style)],
        [Paragraph("Langkah 3: Penerbitan Tiket Resmi", table_cell_bold), Paragraph("Sistem memvalidasi input dan menerbitkan <b>Nomor Tiket Resmi</b> unik (contoh: <code>#PST-261028-561</code>) yang disimpan ke tabel <code>consultation_bookings</code> di database SQLite.", table_cell_style)],
        [Paragraph("Langkah 4: Auto-Forward ke Grup WA Kantor", table_cell_bold), Paragraph("Bot secara otomatis meneruskan rincian pemohon dan nomor tiket ke grup internal WhatsApp kantor (<code>INTERNAL_PST_GROUP_JID</code>) agar petugas piket PST dapat langsung memverifikasi jadwal.", table_cell_style)],
        [Paragraph("Langkah 5: Auto-Discovery Grup", table_cell_bold), Paragraph("Saat bot dijalankan, terminal mencetak seluruh Group JID yang diikuti bot sehingga administrator dapat menyalin JID yang sesuai ke file <code>.env</code> dengan mudah.", table_cell_style)],
    ]
    t_pst = Table(pst_steps, colWidths=[130, 355])
    t_pst.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,-1), CARD_BG),
        ('BOX', (0,0), (-1,-1), 0.8, BORDER_COLOR),
        ('GRID', (0,0), (-1,-1), 0.5, BORDER_COLOR),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('PADDING', (0,0), (-1,-1), 4.5),
    ]))
    story.append(t_pst)
    story.append(PageBreak())

    # ==========================================================================
    # BAB 5: DOKUMENTASI LENGKAP REST API (MASTER REFERENCE)
    # ==========================================================================
    story.append(Paragraph("5. Spesifikasi Lengkap REST API (Master Reference)", h1_style))
    story.append(Paragraph(
        "Server REST API berjalan menggunakan framework performa tinggi <b>Fiber v2</b> pada port <code>APP_PORT</code> (default: <code>3000</code>). "
        "Seluruh endpoint di bawah <code>/api/*</code> dilindungi oleh header autentikasi <code>X-API-Key: &lt;API_SECRET_KEY&gt;</code>.",
        body_style
    ))
    story.append(Spacer(1, 6))

    api_table_data = [
        [Paragraph("Metode", table_header_style), Paragraph("Endpoint URL", table_header_style), Paragraph("Auth", table_header_style), Paragraph("Deskripsi & Kegunaan", table_header_style)],
        # Health & System
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/health</code>", table_cell_style), Paragraph("Publik", table_cell_style), Paragraph("Health check server, penggunaan memori RAM, status SQLite, dan circuit breaker.", table_cell_style)],
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/qr</code>", table_cell_style), Paragraph("Publik", table_cell_style), Paragraph("Antarmuka web pemindai QR code WhatsApp di browser (otomatis dikunci setelah terhubung).", table_cell_style)],
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/qr</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Mengambil payload QR code WhatsApp dalam format JSON Base64.", table_cell_style)],
        # Indikator Strategis
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/indicators</code><br/>(alias: <code>/api/v1/indikator</code>)", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Katalog lengkap 7 Indikator Strategis (Inflasi, Kemiskinan, PDRB, IPM, dll) format JSON / text.", table_cell_style)],
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/indicators/:id</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Mengambil 1 indikator spesifik berdasarkan ID (1-7) atau slug (<code>?format=text</code>).", table_cell_style)],
        # Layanan Data
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/layanan-data</code><br/>(alias: <code>/api/v1/services/data</code>)", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Mengambil seluruh 4 kategori Layanan Data Statistik (Sosduk, Ekonomi, Pertanian, Sektoral).", table_cell_style)],
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/layanan-data/:category</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Mengambil daftar subtopik di bawah satu kategori tertentu.", table_cell_style)],
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/layanan-data/:category/:topic</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Daftar tabel statis hasil scraping portal BPS untuk topik tertentu (<code>?limit=5</code>).", table_cell_style)],
        # Narahubung Pakar
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/pakar</code><br/>(alias: <code>/api/v1/experts</code>)", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Mengambil daftar 8 bidang narahubung pakar dari database SQLite.", table_cell_style)],
        [Paragraph("POST", table_cell_bold), Paragraph("<code>/api/v1/pakar/:id</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Memperbarui nama dan nomor WhatsApp pakar bidang tertentu (<code>{\"name\":\"...\", \"phone\":\"...\"}</code>).", table_cell_style)],
        # Kontrol Scraper
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/scraper/status</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Status scraper, umur data, jumlah item per kind, circuit breaker, dan jadwal cron.", table_cell_style)],
        [Paragraph("POST", table_cell_bold), Paragraph("<code>/api/v1/scraper/run</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Memicu proses scraping manual on-demand via API (<code>?kind=all</code>).", table_cell_style)],
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/scraper/runs</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Riwayat log audit eksekusi scraping (durasi, jumlah item, error).", table_cell_style)],
        [Paragraph("GET", table_cell_bold), Paragraph("<code>/api/v1/scraper/items</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Mengambil daftar item hasil scraping di database SQLite (<code>?kind=...&limit=20</code>).", table_cell_style)],
        # Broadcast & Cache
        [Paragraph("POST", table_cell_bold), Paragraph("<code>/api/v1/blast</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Mengirim siaran pesan massal (broadcast blast) dengan delay acak anti-banned.", table_cell_style)],
        [Paragraph("POST", table_cell_bold), Paragraph("<code>/api/v1/cache/clear</code>", table_cell_style), Paragraph("Perlu", table_cell_style), Paragraph("Membersihkan in-memory cache data di server.", table_cell_style)],
    ]
    t_api = Table(api_table_data, colWidths=[38, 145, 32, 270])
    t_api.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,0), PRIMARY),
        ('BOX', (0,0), (-1,-1), 0.8, BORDER_COLOR),
        ('GRID', (0,0), (-1,-1), 0.5, BORDER_COLOR),
        ('ROWBACKGROUNDS', (0,1), (-1,-1), [colors.white, CARD_BG]),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('PADDING', (0,0), (-1,-1), 3.5),
    ]))
    story.append(t_api)
    story.append(PageBreak())

    # ==========================================================================
    # BAB 6: SKEMA BASIS DATA SQLITE & KEAMANAN REPOSITORI PUBLIK
    # ==========================================================================
    story.append(Paragraph("6. Skema Basis Data SQLite & Desain Relasi Tabel", h1_style))
    story.append(Paragraph(
        "Seluruh data disimpan secara persisten di dalam basis data SQLite lokal <code>storage/scraper.db</code>. "
        "Desain skema telah dioptimalkan dengan indeks pencarian pada kolom kunci untuk memastikan kecepatan baca/tulis di bawah 5 milidetik.",
        body_style
    ))
    story.append(Spacer(1, 4))

    schema_table_data = [
        [Paragraph("Nama Tabel", table_header_style), Paragraph("Fungsi & Peruntukan", table_header_style), Paragraph("Kolom Utama", table_header_style)],
        [
            Paragraph("<code>scraped_items</code>", table_cell_bold),
            Paragraph("Menyimpan seluruh katalog tabel statistik statis, buku publikasi, rilis BRS, infografis, dan berita daerah.", table_cell_style),
            Paragraph("<code>id, kind, lang, source_key, title, url, release_date, abstract, file_url, content_hash, is_active</code>", code_style)
        ],
        [
            Paragraph("<code>scrape_runs</code>", table_cell_bold),
            Paragraph("Mencatat audit trail setiap eksekusi scraping (cron maupun manual) untuk keperluan pemantauan SLA.", table_cell_style),
            Paragraph("<code>id, trigger, kind, started_at, finished_at, status, requests, items_found, items_new, items_updated</code>", code_style)
        ],
        [
            Paragraph("<code>expert_contacts</code>", table_cell_bold),
            Paragraph("Menyimpan daftar kontak narahubung 8 bidang pakar statistik yang dapat diupdate oleh admin.", table_cell_style),
            Paragraph("<code>id, bidang_id, bidang_slug, bidang_title, pakar_name, phone, is_active, updated_at</code>", code_style)
        ],
        [
            Paragraph("<code>consultation_bookings</code>", table_cell_bold),
            Paragraph("Menyimpan permohonan jadwal konsultasi statistik dan kebutuhan data dari pemohon PST.", table_cell_style),
            Paragraph("<code>id, ticket_number, name, agency, phone, topic, plan_date, status, created_at</code>", code_style)
        ],
        [
            Paragraph("<code>circuit_breaker_state</code>", table_cell_bold),
            Paragraph("Menyimpan status fail-safe circuit breaker mesin scraper.", table_cell_style),
            Paragraph("<code>id, consecutive_failures, open_until, last_failure_reason, updated_at</code>", code_style)
        ],
    ]
    t_schema = Table(schema_table_data, colWidths=[110, 160, 215])
    t_schema.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,0), colors.HexColor("#065F46")),
        ('BOX', (0,0), (-1,-1), 0.8, BORDER_COLOR),
        ('GRID', (0,0), (-1,-1), 0.5, BORDER_COLOR),
        ('ROWBACKGROUNDS', (0,1), (-1,-1), [colors.white, CARD_BG]),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('PADDING', (0,0), (-1,-1), 4),
    ]))
    story.append(t_schema)
    story.append(Spacer(1, 12))

    story.append(Paragraph("7. Standar Keamanan & Perlindungan Repositori Publik", h1_style))
    story.append(Paragraph(
        "Mengingat repositori kode program bersifat terbuka (<i>Public Repository</i>), seluruh prinsip keamanan dan perlindungan privasi data diterapkan secara ketat:",
        body_style
    ))

    sec_rules = [
        [Paragraph("Proteksi Kredensial (.gitignore)", table_cell_bold), Paragraph("File <code>.env</code>, database sesi WhatsApp (<code>storage/whatsmeow.db</code>), dan database scraper (<code>storage/scraper.db</code>) telah di-ignore dari Git dan tidak pernah di-commit ke repositori publik.", table_cell_style)],
        [Paragraph("Pembersihan Folder PRD", table_cell_bold), Paragraph("Folder dokumentasi internal <code>prd/</code> telah di-untrack dan di-ignore dari Git index agar strategi internal kantor tetap aman.", table_cell_style)],
        [Paragraph("Standardisasi Template .env.example", table_cell_bold), Paragraph("File <code>.env.example</code> memuat seluruh 59 variabel konfigurasi yang sinkron dengan <code>.env</code> namun dengan nilai yang dikosongkan (<code>KEY=</code>) sehingga siap pakai oleh pengguna baru.", table_cell_style)],
        [Paragraph("Anti-Flooding & Token Bucket", table_cell_bold), Paragraph("Membatasi pengiriman pesan beruntun dari nomor pengguna yang sama untuk mencegah spam dan pemblokiran akun WhatsApp oleh Meta.", table_cell_style)],
        [Paragraph("Anti-Brute Force Admin", table_cell_bold), Paragraph("Membatasi percobaan login admin maksimal 3 kali sebelum penguncian otomatis (lockout) 60 menit.", table_cell_style)],
    ]
    t_sec = Table(sec_rules, colWidths=[140, 345])
    t_sec.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,-1), CARD_BG),
        ('BOX', (0,0), (-1,-1), 0.8, BORDER_COLOR),
        ('GRID', (0,0), (-1,-1), 0.5, BORDER_COLOR),
        ('VALIGN', (0,0), (-1,-1), 'MIDDLE'),
        ('PADDING', (0,0), (-1,-1), 4.5),
    ]))
    story.append(t_sec)
    story.append(Spacer(1, 15))

    # Tanda Tangan / Penutup Dokumen
    closing_data = [
        [
            Paragraph("<b>Diterbitkan Oleh:</b><br/>Tim Pengembang Sistem Diseminasi Statistik<br/>Badan Pusat Statistik Kabupaten Bangka Barat<br/>Jl. Komplek Perkantoran Terpadu Pemkab Bangka Barat", table_cell_style),
            Paragraph("<b>Verifikasi Sistem:</b><br/>Status: <b>Production Ready (v3.4)</b><br/>Zero Broken Links: <b>Active</b><br/>Scraper Engine: <b>Active (Cron & WA Trigger)</b>", table_cell_style)
        ]
    ]
    t_closing = Table(closing_data, colWidths=[245, 240])
    t_closing.setStyle(TableStyle([
        ('BACKGROUND', (0,0), (-1,-1), colors.HexColor("#EFF6FF")),
        ('BOX', (0,0), (-1,-1), 1, SECONDARY),
        ('PADDING', (0,0), (-1,-1), 6),
    ]))
    story.append(t_closing)

    # Build PDF
    doc.build(story, canvasmaker=NumberedCanvas)
    print(f"[OK] Dokumen PDF berhasil digenerate di: {output_filename}")


if __name__ == "__main__":
    out_file = "DOKUMENTASI_SISTEM_BOT_ISTIMEWA.pdf"
    if len(sys.argv) > 1:
        out_file = sys.argv[1]
    generate_docs_pdf(out_file)
