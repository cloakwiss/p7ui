module github.com/cloakwiss/p7ui

go 1.25.3

require (
	github.com/Microsoft/go-winio v0.6.2
	github.com/cloakwiss/project-seven/deserialize v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.2.3
	github.com/sqweek/dialog v0.0.0-20240226140203-065105509627
	github.com/starfederation/datastar-go v1.0.2
)

replace github.com/cloakwiss/project-seven/deserialize => ../deserialize

require (
	github.com/CAFxX/httpcompression v0.0.9 // indirect
	github.com/TheTitanrain/w32 v0.0.0-20180517000239-4f5cfb03fabf // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
