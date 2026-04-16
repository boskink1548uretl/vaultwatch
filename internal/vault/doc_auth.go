// Package vault provides auth method inspection alongside secret, policy,
// token, mount, namespace, lease, and audit device operations.
//
// Auth method listing and retrieval uses the /v1/sys/auth Vault API endpoint.
// Use ListAuthMethods to enumerate all enabled auth backends and GetAuthMethod
// to retrieve a specific backend by its mount path.
package vault
