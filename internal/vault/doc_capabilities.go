// Package vault provides capabilities checking for Vault tokens.
//
// The capabilities module wraps the /v1/sys/capabilities-self endpoint,
// allowing callers to determine what operations the current token is
// permitted to perform on specific secret paths.
//
// Example usage:
//
//	caps, err := client.GetCapabilities(ctx, []string{"secret/data/myapp"})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(caps["secret/data/myapp"]) // ["read", "list"]
//
// HasCapability provides a convenience wrapper for single-path, single-cap checks.
package vault
