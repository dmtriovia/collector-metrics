module server

go 1.22.0

replace handlers => ../../internal/handlers

replace middleware => ../../internal/middleware

replace service => ../../internal/service

replace storage => ../../internal/storage

require handlers v0.0.0-00010101000000-000000000000

require middleware v0.0.0-00010101000000-000000000000

require service v0.0.0-00010101000000-000000000000 // indirect

require (
	github.com/go-resty/resty/v2 v2.15.3
	github.com/stretchr/testify v1.9.0
	storage v0.0.0-00010101000000-000000000000 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
