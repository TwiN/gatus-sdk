package gatussdk

import (
	"strings"
)

// GenerateEndpointKey generates a unique key for an endpoint based on its group and name.
// The key format is {group}_{name} where special characters are replaced with '-'.
// Characters replaced: '/', '_', ',', '.', '#' become '-'.
// If the group is empty, the key format is _{name}.
//
// Examples:
//   - GenerateEndpointKey("core", "blog-home") returns "core_blog-home"
//   - GenerateEndpointKey("api/v1", "health_check") returns "api-v1_health-check"
//   - GenerateEndpointKey("", "standalone") returns "_standalone"
func GenerateEndpointKey(group, name string) string {
	// Replace special characters with '-' in both group and name
	replacer := strings.NewReplacer(
		"/", "-",
		"_", "-",
		",", "-",
		".", "-",
		"#", "-",
	)

	sanitizedGroup := replacer.Replace(group)
	sanitizedName := replacer.Replace(name)

	// Combine with underscore separator
	if sanitizedGroup == "" {
		return "_" + sanitizedName
	}
	return sanitizedGroup + "_" + sanitizedName
}
