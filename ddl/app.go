package ddl

type Application struct {
	contexts []*BoundedContext
}

func NewApplication(contexts ...*BoundedContext) *Application {
	return &Application{contexts: contexts}
}
