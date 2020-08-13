package ddl

type Domain struct {
}

func NewDomain(api *DomainAPI, persistence *DomainPersistenceAPI) *Domain {
	return &Domain{}
}

type DomainAPI struct {
}

func NewDomainAPI(iface func()) *DomainAPI {
	return &DomainAPI{}
}

type DomainPersistenceAPI struct {
}

func NewPersistenceAPI(iface func()) *DomainPersistenceAPI {
	return nil
}
