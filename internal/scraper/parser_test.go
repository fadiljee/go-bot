package scraper

import (
	"net/url"
	"os"
	"testing"
)

func TestParsers_Golden(t *testing.T) {
	selectorsCfg, err := LoadSelectors("../../configs/scraper_selectors.yaml", []Kind{
		KindPressRelease, KindPublication, KindInfographic, KindStaticTable,
	})
	if err != nil {
		t.Fatalf("failed to load selectors: %v", err)
	}

	baseURL, _ := url.Parse("https://bangkabaratkab.bps.go.id/id/pressrelease.html?page=1")

	// 1. Test PressRelease
	t.Run("PressRelease", func(t *testing.T) {
		ks, err := selectorsCfg.ForKind(KindPressRelease)
		if err != nil {
			t.Fatal(err)
		}
		parser := NewGenericKindParser(KindPressRelease, ks)

		listHTML, err := os.ReadFile("testdata/pressrelease_list_p1.html")
		if err != nil {
			t.Fatal(err)
		}
		items, pagination, err := parser.ParseList(listHTML, baseURL)
		if err != nil {
			t.Fatalf("ParseList failed: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
		if !pagination.HasNext {
			t.Errorf("expected pagination.HasNext to be true")
		}
		if items[0].SourceKey != "537" {
			t.Errorf("expected source_key 537, got %s", items[0].SourceKey)
		}

		// Detail
		detailHTML, err := os.ReadFile("testdata/pressrelease_detail_1.html")
		if err != nil {
			t.Fatal(err)
		}
		detailURL, _ := url.Parse(items[0].URL)
		err = parser.ParseDetail(detailHTML, detailURL, &items[0])
		if err != nil {
			t.Fatalf("ParseDetail failed: %v", err)
		}
		if items[0].Abstract == "" {
			t.Errorf("expected abstract to be extracted")
		}
		if items[0].FileURL == "" || items[0].FileType != "pdf" {
			t.Errorf("expected pdf file_url, got %s (%s)", items[0].FileURL, items[0].FileType)
		}
	})

	// 2. Test Publication
	t.Run("Publication", func(t *testing.T) {
		ks, err := selectorsCfg.ForKind(KindPublication)
		if err != nil {
			t.Fatal(err)
		}
		parser := NewGenericKindParser(KindPublication, ks)

		listHTML, err := os.ReadFile("testdata/publication_list_p1.html")
		if err != nil {
			t.Fatal(err)
		}
		pubBase, _ := url.Parse("https://bangkabaratkab.bps.go.id/publication.html")
		items, _, err := parser.ParseList(listHTML, pubBase)
		if err != nil || len(items) != 1 {
			t.Fatalf("expected 1 item, got %d (err: %v)", len(items), err)
		}
		if items[0].SourceKey != "901" {
			t.Errorf("expected source_key 901, got %s", items[0].SourceKey)
		}

		detailHTML, err := os.ReadFile("testdata/publication_detail_1.html")
		if err != nil {
			t.Fatal(err)
		}
		detailURL, _ := url.Parse(items[0].URL)
		err = parser.ParseDetail(detailHTML, detailURL, &items[0])
		if err != nil {
			t.Fatalf("ParseDetail failed: %v", err)
		}
		if items[0].FileSizeText != "14,2 MB" {
			t.Errorf("expected 14,2 MB file size, got %s", items[0].FileSizeText)
		}
	})

	// 3. Test Infographic
	t.Run("Infographic", func(t *testing.T) {
		ks, err := selectorsCfg.ForKind(KindInfographic)
		if err != nil {
			t.Fatal(err)
		}
		parser := NewGenericKindParser(KindInfographic, ks)

		listHTML, err := os.ReadFile("testdata/infographic_list_p1.html")
		if err != nil {
			t.Fatal(err)
		}
		infoBase, _ := url.Parse("https://bangkabaratkab.bps.go.id/infographic.html")
		items, _, err := parser.ParseList(listHTML, infoBase)
		if err != nil || len(items) != 1 {
			t.Fatalf("expected 1 item, got %d (err: %v)", len(items), err)
		}

		detailHTML, err := os.ReadFile("testdata/infographic_detail_1.html")
		if err != nil {
			t.Fatal(err)
		}
		detailURL, _ := url.Parse(items[0].URL)
		err = parser.ParseDetail(detailHTML, detailURL, &items[0])
		if err != nil {
			t.Fatalf("ParseDetail failed: %v", err)
		}
		if items[0].ImageURL == "" {
			t.Errorf("expected image_url in infographic detail")
		}
	})

	// 4. Test StaticTable
	t.Run("StaticTable", func(t *testing.T) {
		ks, err := selectorsCfg.ForKind(KindStaticTable)
		if err != nil {
			t.Fatal(err)
		}
		parser := NewGenericKindParser(KindStaticTable, ks)

		listHTML, err := os.ReadFile("testdata/statictable_list_p1.html")
		if err != nil {
			t.Fatal(err)
		}
		tableBase, _ := url.Parse("https://bangkabaratkab.bps.go.id/statictable.html")
		items, _, err := parser.ParseList(listHTML, tableBase)
		if err != nil || len(items) != 1 {
			t.Fatalf("expected 1 item, got %d (err: %v)", len(items), err)
		}

		detailHTML, err := os.ReadFile("testdata/statictable_detail_1.html")
		if err != nil {
			t.Fatal(err)
		}
		detailURL, _ := url.Parse(items[0].URL)
		err = parser.ParseDetail(detailHTML, detailURL, &items[0])
		if err != nil {
			t.Fatalf("ParseDetail failed: %v", err)
		}
		if items[0].FileType != "xlsx" {
			t.Errorf("expected xlsx file type, got %s", items[0].FileType)
		}
	})
}
