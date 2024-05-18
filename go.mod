module github.com/ncotds/nco-qoordinator

go 1.22.1

require (
	github.com/doug-martin/goqu/v9 v9.19.0
	github.com/go-faker/faker/v4 v4.4.1
	github.com/google/uuid v1.6.0
	github.com/minus5/gofreetds v0.0.0-20200826115934-6705a38c49ca
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/minus5/gofreetds v0.0.0-20200826115934-6705a38c49ca => github.com/vitalyshatskikh/gofreetds v0.1.1
