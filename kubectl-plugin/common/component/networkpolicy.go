package component

// NetworkPolicy return the struct of NetworkPolicy
type NetworkPolicy struct {
}

func (n *NetworkPolicy) GetName() string {
	return "networkpolicy"
}

// Uninstall uninstall NetworkPolicy
func (n *NetworkPolicy) Uninstall() error {
	return nil
}
