package team

type Team struct {
	Name         string
	Members      []Member
	Repositories []Repository
}

type Member struct {
	Login string
	Email string
}

type Repository struct {
	Name string
	Permission string
}
