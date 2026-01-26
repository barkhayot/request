package request

import "testing"

func TestValidateProxy(t *testing.T) {
	type testCase struct {
		name      string
		proxyStr  string
		expectErr bool
	}

	testCases := []testCase{
		{
			name:      "valid HTTP proxy",
			proxyStr:  "http://localhost:8080",
			expectErr: false,
		},
		{
			name:      "valid HTTPS proxy",
			proxyStr:  "https://proxy.example.com:443",
			expectErr: false,
		},
		{
			name:      "invalid proxy URL",
			proxyStr:  "ht!tp://invalid-url",
			expectErr: true,
		},
		{
			name:      "empty proxy string",
			proxyStr:  "",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateProxy(tc.proxyStr)
			if tc.expectErr && err == nil {
				t.Fatalf("expected error for proxy string: %s", tc.proxyStr)
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("unexpected error for proxy string %s: %v", tc.proxyStr, err)
			}
		})
	}
}
