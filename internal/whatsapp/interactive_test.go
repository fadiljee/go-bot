package whatsapp

import (
	"strings"
	"testing"
)

func TestBuildInteractiveList(t *testing.T) {
	rows := []ListRow{
		{ID: "1", Title: "Opsi 1", Description: "Deskripsi 1"},
		{ID: "2", Title: "Opsi 2", Description: "Deskripsi 2"},
	}
	sections := []ListSection{
		{Title: "Kategori A", Rows: rows},
	}

	msg := BuildInteractiveList("Pilih Salah Satu", "Halo User", "Footer Info", "Buka Pilihan", sections)
	if msg == nil {
		t.Fatal("expected non-nil message")
	}

	viewOnce := msg.GetViewOnceMessage()
	if viewOnce == nil {
		t.Fatal("expected viewOnceMessage wrapper")
	}

	interactive := viewOnce.GetMessage().GetInteractiveMessage()
	if interactive == nil {
		t.Fatal("expected interactiveMessage inside viewOnce")
	}

	if interactive.GetBody().GetText() != "Halo User" {
		t.Errorf("expected body 'Halo User', got '%s'", interactive.GetBody().GetText())
	}

	if interactive.GetFooter().GetText() != "Footer Info" {
		t.Errorf("expected footer 'Footer Info', got '%s'", interactive.GetFooter().GetText())
	}
}

func TestBuildQuickReplyButtons(t *testing.T) {
	buttons := []QuickReplyButton{
		{ID: "0", DisplayText: "Kembali ke Menu Utama"},
		{ID: "99", DisplayText: "Kembali ke Menu Sebelumnya"},
	}

	msg := BuildQuickReplyButtons("Header Judul", "Isi pesan detail...", "Footer BPS", buttons)
	if msg == nil {
		t.Fatal("expected non-nil message")
	}

	viewOnce := msg.GetViewOnceMessage()
	if viewOnce == nil {
		t.Fatal("expected viewOnceMessage wrapper")
	}

	interactive := viewOnce.GetMessage().GetInteractiveMessage()
	if interactive == nil {
		t.Fatal("expected interactiveMessage")
	}

	if interactive.GetHeader().GetTitle() != "Header Judul" {
		t.Errorf("expected header 'Header Judul', got '%s'", interactive.GetHeader().GetTitle())
	}
}

func TestClickToChatLinks(t *testing.T) {
	SetBotNumber("6282164302643")
	defer SetBotNumber("")

	rows := []ListRow{
		{ID: "1", Title: "Layanan Data"},
		{ID: "2", Title: "Indikator"},
	}

	footer := "*99. ↩️ Menu sebelumnya*\n*0. 📑 Menu Awal*"
	body := BuildBodyText("Pilih menu:", rows, footer)

	if !strings.Contains(body, "https://wa.me/6282164302643?text=1") {
		t.Errorf("expected link for option 1, got:\n%s", body)
	}

	if !strings.Contains(body, "https://wa.me/6282164302643?text=99") {
		t.Errorf("expected link for navigation 99, got:\n%s", body)
	}

	if !strings.Contains(body, "https://wa.me/6282164302643?text=0") {
		t.Errorf("expected link for navigation 0, got:\n%s", body)
	}
}
