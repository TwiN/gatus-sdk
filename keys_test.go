package gatussdk

import (
	"testing"
)

func TestGenerateEndpointKey(t *testing.T) {
	tests := []struct {
		name     string
		group    string
		endpoint string
		expected string
	}{
		{
			name:     "normal group and name",
			group:    "core",
			endpoint: "blog-home",
			expected: "core_blog-home",
		},
		{
			name:     "group with forward slash",
			group:    "api/v1",
			endpoint: "health",
			expected: "api-v1_health",
		},
		{
			name:     "name with underscore",
			group:    "services",
			endpoint: "user_service",
			expected: "services_user-service",
		},
		{
			name:     "group and name with special characters",
			group:    "api/v1",
			endpoint: "health_check.test",
			expected: "api-v1_health-check-test",
		},
		{
			name:     "empty group",
			group:    "",
			endpoint: "standalone",
			expected: "_standalone",
		},
		{
			name:     "all special characters in group",
			group:    "test/group_one,two.three#four",
			endpoint: "endpoint",
			expected: "test-group-one-two-three-four_endpoint",
		},
		{
			name:     "all special characters in name",
			group:    "group",
			endpoint: "end/point_a,b.c#d",
			expected: "group_end-point-a-b-c-d",
		},
		{
			name:     "all special characters in both",
			group:    "test/group_one,two.three#four",
			endpoint: "end/point_a,b.c#d",
			expected: "test-group-one-two-three-four_end-point-a-b-c-d",
		},
		{
			name:     "group with comma",
			group:    "region,us-east",
			endpoint: "service",
			expected: "region-us-east_service",
		},
		{
			name:     "name with period",
			group:    "monitoring",
			endpoint: "api.gateway",
			expected: "monitoring_api-gateway",
		},
		{
			name:     "group with hash",
			group:    "env#prod",
			endpoint: "database",
			expected: "env-prod_database",
		},
		{
			name:     "consecutive special characters",
			group:    "api//v1",
			endpoint: "health__check",
			expected: "api--v1_health--check",
		},
		{
			name:     "mixed alphanumeric with special",
			group:    "api-v2/beta_1",
			endpoint: "user.service#2",
			expected: "api-v2-beta-1_user-service-2",
		},
		{
			name:     "only special characters in group",
			group:    "/_,.,#",
			endpoint: "test",
			expected: "------_test",
		},
		{
			name:     "only special characters in name",
			group:    "test",
			endpoint: "/_,.,#",
			expected: "test_------",
		},
		{
			name:     "empty group with special chars in name",
			group:    "",
			endpoint: "standalone/service_1",
			expected: "_standalone-service-1",
		},
		{
			name:     "single character group",
			group:    "a",
			endpoint: "b",
			expected: "a_b",
		},
		{
			name:     "single special character group",
			group:    "/",
			endpoint: "test",
			expected: "-_test",
		},
		{
			name:     "single special character name",
			group:    "test",
			endpoint: "_",
			expected: "test_-",
		},
		{
			name:     "unicode characters (should not be replaced)",
			group:    "日本語",
			endpoint: "テスト",
			expected: "日本語_テスト",
		},
		{
			name:     "mixed unicode and special",
			group:    "日本/語",
			endpoint: "テ_スト",
			expected: "日本-語_テ-スト",
		},
		{
			name:     "numbers with special characters",
			group:    "v1.2.3",
			endpoint: "api_2024",
			expected: "v1-2-3_api-2024",
		},
		{
			name:     "hyphen should not be replaced",
			group:    "us-east-1",
			endpoint: "service-name",
			expected: "us-east-1_service-name",
		},
		{
			name:     "complex real-world example",
			group:    "production/us-east/v2.1",
			endpoint: "payment_gateway.service#1",
			expected: "production-us-east-v2-1_payment-gateway-service-1",
		},
		{
			name:     "empty name (edge case)",
			group:    "group",
			endpoint: "",
			expected: "group_",
		},
		{
			name:     "both empty (edge case)",
			group:    "",
			endpoint: "",
			expected: "_",
		},
		{
			name:     "spaces should not be replaced",
			group:    "my group",
			endpoint: "my endpoint",
			expected: "my group_my endpoint",
		},
		{
			name:     "special at boundaries",
			group:    "/group/",
			endpoint: "_name_",
			expected: "-group-_-name-",
		},
		{
			name:     "group with plus sign",
			group:    "api+v1",
			endpoint: "service",
			expected: "api-v1_service",
		},
		{
			name:     "name with plus sign",
			group:    "services",
			endpoint: "cache+store",
			expected: "services_cache-store",
		},
		{
			name:     "group with ampersand",
			group:    "dev&test",
			endpoint: "endpoint",
			expected: "dev-test_endpoint",
		},
		{
			name:     "name with ampersand",
			group:    "monitoring",
			endpoint: "health&status",
			expected: "monitoring_health-status",
		},
		{
			name:     "both with plus and ampersand",
			group:    "api+v2&beta",
			endpoint: "user+service&test",
			expected: "api-v2-beta_user-service-test",
		},
		{
			name:     "all new special characters",
			group:    "env+prod&region",
			endpoint: "service+api&gateway",
			expected: "env-prod-region_service-api-gateway",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateEndpointKey(tt.group, tt.endpoint)
			if result != tt.expected {
				t.Errorf("GenerateEndpointKey(%q, %q) = %q, want %q",
					tt.group, tt.endpoint, result, tt.expected)
			}
		})
	}
}
