package ddl

// Repositories represent a layer which is leaned towards the domain layer, however it is solely focused
// on handling domain objects. Importantly it does not contain any higher business logic. Repositories are
// declared by and belong to the applications domain layer. They never leak details about a concrete
// technical implementation. A repository may be MySQL, PostgreSQL, NoSQL or even a remote service.
// As a consequence, you cannot even leak knowledge about transactions from a higher layer, like a
// request (e.g. an incoming REST-call from the presentation layer) into a repository. If you really ever need
// to guarantee consistency (think of eventual consistency in a clustered NoSQL world), it must
// be implemented at the repository level, firewalled by an appropriate interface contract.
//
//
// It is important to understand, that Repositories is only a compile-time concept which mixes specifications
// and implementations used for code generation. It is not available at runtime. Also, the API is heavily
// inspired by SQL based backends, which limits your design space and therefore causes a kind of
// implementation-leak into the upper layers. But at least, the generated interfaces will make it harder to
// leak unintended implementation specific details (like using sql.NullString).
//
//
// Tips:
//  * Do not expose or rely on SQL auto generated integer ids. Their performance is nice, but it may be very hard to
//    support or migrate to NoSQL databases, especially with eventual consistency. Also, there is always the sweet
//    temptation to expose a guessable id to the outside world, also known as an Insecure Direct Object Reference
//    (IDOR). See OWASP for more details. In most cases, using just a UUID as a primary key may be a good
//    compromise between security, isolation of concerns and performance. Keeping a second "primary" index for
//    UUIDs as proposed by OWASP (August, 2020) hurts usually more than the performance penalty for a 16 byte
//    random primary key in your database engine would ever do.
//  * Never inspect and rely on backend specific errors, which would make your code unportable.
//  * Keep your returned collections and objects small and beware of large datasets in the future.
//  * Design your API with the domain in your mind, just CRUD is usually not the answer.
type Repositories struct {
	name   string
	doc    string
	repos  []*Repository
	tables []*Table
}

// NewRepositories creates a Repositories instance for a specific domain name.
func NewRepositories(domainName string) *Repositories {
	return &Repositories{name: domainName}
}

// Comment describes the purpose and content of the domains repositories.
func (r *Repositories) Comment(doc string) *Repositories {
	r.doc = doc
	return r
}

// AddTables appends implementation specific details about a table. It has no immediate effect for
// the API layer. A table can be used by Repository declaration.
func (r *Repositories) Migrate(tables ...*Table) *Repositories {
	r.tables = append(r.tables, tables...)
	return r
}

// Add appends Repository instances, which are a mixture of generic specification and a bunch of
// implementations. Implementations and the boundary contract are clearly separated in resulting generated code.
func (r *Repositories) Add(repos ...*Repository) *Repositories {
	r.repos = append(r.repos, repos...)
	return r
}
