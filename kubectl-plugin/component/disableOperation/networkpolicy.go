package disableOperation

type NetworkPolicy struct {
}

func (n *NetworkPolicy) DeleteRelatedResource() error {
	return nil
}
