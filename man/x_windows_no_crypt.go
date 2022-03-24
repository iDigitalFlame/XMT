//go:build windows && !crypt

package man

const (
	slot   = `\\.\mailslot\`
	prefix = `Global\`
)
