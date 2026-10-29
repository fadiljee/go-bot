package whatsapp

import (
	"encoding/json"
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

	inter := UnpackInteractiveMessage(msg)
	if inter == nil {
		t.Fatal("expected InteractiveMessage within ViewOnce wrapper")
	}

	if inter.GetHeader().GetTitle() != "Pilih Salah Satu" {
		t.Errorf("expected header 'Pilih Salah Satu', got '%s'", inter.GetHeader().GetTitle())
	}

	if inter.GetBody().GetText() != "Halo User" {
		t.Errorf("expected body 'Halo User', got '%s'", inter.GetBody().GetText())
	}

	if inter.GetFooter().GetText() != "Footer Info" {
		t.Errorf("expected footer 'Footer Info', got '%s'", inter.GetFooter().GetText())
	}

	nativeFlow := inter.GetNativeFlowMessage()
	if nativeFlow == nil {
		t.Fatal("expected NativeFlowMessage")
	}

	buttons := nativeFlow.GetButtons()
	if len(buttons) != 1 {
		t.Fatalf("expected 1 single_select button, got %d", len(buttons))
	}

	btn := buttons[0]
	if btn.GetName() != "single_select" {
		t.Errorf("expected button name 'single_select', got '%s'", btn.GetName())
	}

	var params singleSelectParams
	if err := json.Unmarshal([]byte(btn.GetButtonParamsJSON()), &params); err != nil {
		t.Fatalf("failed to unmarshal single_select params: %v", err)
	}

	if params.Title != "Buka Pilihan" {
		t.Errorf("expected dropdown button title 'Buka Pilihan', got '%s'", params.Title)
	}

	if len(params.Sections) != 1 || len(params.Sections[0].Rows) != 2 {
		t.Errorf("expected 1 section with 2 rows, got %+v", params.Sections)
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

	inter := UnpackInteractiveMessage(msg)
	if inter == nil {
		t.Fatal("expected InteractiveMessage within ViewOnce wrapper")
	}

	if inter.GetHeader().GetTitle() != "Header Judul" {
		t.Errorf("expected header 'Header Judul', got '%s'", inter.GetHeader().GetTitle())
	}

	if inter.GetBody().GetText() != "Isi pesan detail..." {
		t.Errorf("expected body 'Isi pesan detail...', got '%s'", inter.GetBody().GetText())
	}

	nativeFlow := inter.GetNativeFlowMessage()
	if nativeFlow == nil {
		t.Fatal("expected NativeFlowMessage")
	}

	nativeButtons := nativeFlow.GetButtons()
	if len(nativeButtons) != 2 {
		t.Fatalf("expected 2 buttons, got %d", len(nativeButtons))
	}

	for i, nb := range nativeButtons {
		if nb.GetName() != "quick_reply" {
			t.Errorf("expected button %d name 'quick_reply', got '%s'", i, nb.GetName())
		}
		var p quickReplyParams
		if err := json.Unmarshal([]byte(nb.GetButtonParamsJSON()), &p); err != nil {
			t.Fatalf("failed to unmarshal quick_reply params %d: %v", i, err)
		}
		if p.ID != buttons[i].ID || p.DisplayText != buttons[i].DisplayText {
			t.Errorf("mismatch button %d: expected %+v, got %+v", i, buttons[i], p)
		}
	}
}

func TestBuildBodyText(t *testing.T) {
	rows := []ListRow{
		{ID: "1", Title: "Layanan Data"},
		{ID: "2", Title: "Indikator"},
	}

	footer := "*99. Menu sebelumnya*\n*0. Menu Awal*"
	body := BuildBodyText("Pilih menu:", rows, footer)

	if !strings.Contains(body, "1️⃣  Layanan Data") {
		t.Errorf("expected option 1, got:\n%s", body)
	}

	if !strings.Contains(body, "2️⃣  Indikator") {
		t.Errorf("expected option 2, got:\n%s", body)
	}

	if !strings.Contains(body, "*0. Menu Awal*") {
		t.Errorf("expected navigation 0 in footer, got:\n%s", body)
	}
}

func TestExtractTextFromMessage(t *testing.T) {
	// 1. From List Message
	rows := []ListRow{
		{ID: "1", Title: "Opsi 1"},
	}
	listMsg := BuildInteractiveList("Judul Header", "Isi pesan body...", "Footer info", "Buka", []ListSection{{Title: "Sec", Rows: rows}})
	extractedList := ExtractTextFromMessage(listMsg)
	if !strings.Contains(extractedList, "*Judul Header*") || !strings.Contains(extractedList, "Isi pesan body...") || !strings.Contains(extractedList, "_Footer info_") {
		t.Errorf("ExtractTextFromMessage list failed: %s", extractedList)
	}

	// 2. From Quick Reply Buttons Message
	buttons := []QuickReplyButton{
		{ID: "0", DisplayText: "Menu Utama"},
	}
	btnMsg := BuildQuickReplyButtons("Header Tombol", "Isi teks...", "Footer BPS", buttons)
	extractedBtn := ExtractTextFromMessage(btnMsg)
	if !strings.Contains(extractedBtn, "*Header Tombol*") || !strings.Contains(extractedBtn, "Isi teks...") || !strings.Contains(extractedBtn, "Ketik *0* untuk Menu Utama") {
		t.Errorf("ExtractTextFromMessage buttons failed: %s", extractedBtn)
	}
}
