package component

// NetworkPolicy return the struct of NetworkPolicy
type NetworkPolicy struct {
}

// GetName return the name of NetworkPolicy
func (n *NetworkPolicy) GetName() string {
	return "networkpolicy"
}

// Uninstall uninstall NetworkPolicy
func (n *NetworkPolicy) Uninstall() error {
	return nil
}
