package wallets

// PreimageProvider is an interface for providing preimages for LN invoices.
type PreimageProvider interface {
	// GetPreimage returns the preimage for the given LN invoice.
	GetPreimage(invoice string) (string, error)
}
