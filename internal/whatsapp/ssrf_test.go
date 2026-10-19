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
