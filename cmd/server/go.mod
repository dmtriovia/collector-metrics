module server

go 1.22.0

replace handlers => ../../internal/handlers
replace middleware => ../../internal/middleware

require handlers v0.0.0-00010101000000-000000000000
require middleware v0.0.0-00010101000000-000000000000