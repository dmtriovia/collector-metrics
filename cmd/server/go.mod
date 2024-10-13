module server

go 1.22.0

replace handlers => ../../internal/handlers

replace middleware => ../../internal/middleware

replace service => ../../internal/service

replace storage => ../../internal/storage

replace models => ../../internal/models

require handlers v0.0.0-00010101000000-000000000000

require middleware v0.0.0-00010101000000-000000000000

require service v0.0.0-00010101000000-000000000000

require (
	github.com/gorilla/mux v1.8.1 // indirect
	storage v0.0.0-00010101000000-000000000000 // indirect
)

require models v0.0.0-00010101000000-000000000000
