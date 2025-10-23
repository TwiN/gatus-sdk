package gatussdk

import (
	"strings"
)

// GenerateKey generates a unique key based on a group and name.
// The key format is {group}_{name} where special characters are replaced with '-'.
// Characters replaced: '/', '_', ',', '.', '#', '+', '&' become '-'
// If the group is empty, the key format is _{name}.
//
// Examples:
//   - GenerateKey("core", "blog-home") returns "core_blog-home"
//   - GenerateKey("api/v1", "health_check") returns "api-v1_health-check"
//   - GenerateKey("", "standalone") returns "_standalone"
func GenerateKey(group, name string) string {
	// Replace special characters with '-' in both group and name
	replacer := strings.NewReplacer(
		"/", "-",
		"_", "-",
		",", "-",
		".", "-",
		"#", "-",
		"+", "-",
		"&", "-",
	)
	sanitizedGroup := replacer.Replace(group)
	sanitizedName := replacer.Replace(name)
	// Combine with underscore separator
	if sanitizedGroup == "" {
		return "_" + sanitizedName
	}
	return sanitizedGroup + "_" + sanitizedName
}
