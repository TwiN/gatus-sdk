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

func TestGenerateEndpointKey_Consistency(t *testing.T) {
	// Test that the same input always produces the same output
	group := "api/v1"
	name := "health_check"

	key1 := GenerateEndpointKey(group, name)
	key2 := GenerateEndpointKey(group, name)
	key3 := GenerateEndpointKey(group, name)

	if key1 != key2 || key2 != key3 {
		t.Errorf("GenerateEndpointKey is not consistent: got %q, %q, %q", key1, key2, key3)
	}

	expected := "api-v1_health-check"
	if key1 != expected {
		t.Errorf("GenerateEndpointKey(%q, %q) = %q, want %q", group, name, key1, expected)
	}
}

func TestGenerateEndpointKey_RealWorldExamples(t *testing.T) {
	// Test examples from actual Gatus configurations
	tests := []struct {
		name     string
		group    string
		endpoint string
		expected string
	}{
		{
			name:     "AWS service endpoint",
			group:    "aws/production",
			endpoint: "ec2_instances.us-east-1",
			expected: "aws-production_ec2-instances-us-east-1",
		},
		{
			name:     "Kubernetes namespace and service",
			group:    "k8s/default",
			endpoint: "nginx-ingress",
			expected: "k8s-default_nginx-ingress",
		},
		{
			name:     "Docker container monitoring",
			group:    "docker",
			endpoint: "redis_cache#1",
			expected: "docker_redis-cache-1",
		},
		{
			name:     "Microservice with version",
			group:    "microservices",
			endpoint: "user-api/v2.1",
			expected: "microservices_user-api-v2-1",
		},
		{
			name:     "Database monitoring",
			group:    "databases",
			endpoint: "postgres_main.replica",
			expected: "databases_postgres-main-replica",
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
