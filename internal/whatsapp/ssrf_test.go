package whatsapp

import (
	"testing"
)

func TestValidateURLAntiSSRF(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid HTTPS URL",
			url:     "https://script.google.com/macros/s/xxx/exec",
			wantErr: false,
		},
		{
			name:    "Invalid HTTP Scheme",
			url:     "http://example.com/image.jpg",
			wantErr: true,
		},
		{
			name:    "Localhost prohibited",
			url:     "https://localhost/admin",
			wantErr: true,
		},
		{
			name:    "127.0.0.1 prohibited",
			url:     "https://127.0.0.1:8080/secret",
			wantErr: true,
		},
		{
			name:    "Empty hostname",
			url:     "https:///path",
			wantErr: true,
		},
		{
			name:    "Non-HTTP scheme",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "Cloud metadata IP 169.254.169.254 prohibited",
			url:     "https://169.254.169.254/latest/meta-data/",
			wantErr: true,
		},
		{
			name:    "Private network 10.x prohibited",
			url:     "https://10.0.0.1/internal.jpg",
			wantErr: true,
		},
		{
			name:    "Private network 192.168.x prohibited",
			url:     "https://192.168.1.1/router.jpg",
			wantErr: true,
		},
		{
			name:    "IPv6 loopback prohibited",
			url:     "https://[::1]/secret",
			wantErr: true,
		},
		{
			name:    "Credentials in URL prohibited",
			url:     "https://user@8.8.8.8/image.jpg",
			wantErr: true,
		},
		{
			name:    "Internal suffix .internal prohibited",
			url:     "https://metadata.google.internal/computeMetadata/v1/",
			wantErr: true,
		},
		{
			name:    "Valid public IP literal over HTTPS",
			url:     "https://8.8.8.8/image.jpg",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateURLAntiSSRF(tc.url)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateURLAntiSSRF(%q) err = %v, wantErr = %v", tc.url, err, tc.wantErr)
			}
		})
	}
}

func TestMediaCacheBounding(t *testing.T) {
	sender := NewSender(nil, nil)
	for i := 0; i < maxMediaCacheEntries+50; i++ {
		sender.setCachedMedia(string(rune(i)), &CachedMedia{URL: "https://8.8.8.8/img.jpg"})
	}
	sender.mediaMu.RLock()
	size := len(sender.mediaCache)
	sender.mediaMu.RUnlock()
	if size > maxMediaCacheEntries {
		t.Errorf("expected mediaCache size <= %d, got %d", maxMediaCacheEntries, size)
	}
}
