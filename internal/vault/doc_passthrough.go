// Package vault provides a client for interacting with HashiCorp Vault.
//
// # Passthrough Secrets
//
// The passthrough (generic) secrets engine stores arbitrary key/value data at
// a given path without versioning. It is the simplest secret backend and is
// useful for storing credentials or configuration that does not require
// rotation history.
//
// Use [Client.GetPassthrough] to read a single secret entry and
// [Client.ListPassthrough] to enumerate keys under a path prefix.
//
// Example:
//
//	entry, err := client.GetPassthrough(ctx, "secret/myapp/db")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(entry.Data["password"])
package vault
