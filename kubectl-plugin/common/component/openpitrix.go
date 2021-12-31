package component

// OpenPitrix return the struct of OpenPitrix
type OpenPitrix struct {
}

// GetName return the name of OpenPitrix
func (p *OpenPitrix) GetName() string {
	return "openpitrix"
}

// Uninstall uninstall OpenPitrix
func (p *OpenPitrix) Uninstall() error {
	return nil
}
