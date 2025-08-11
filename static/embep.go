package static

import (
	"embed"
)

//go:embed css/*
var Css embed.FS

//go:embed img/*
var Img embed.FS

//go:embed js/*
var Js embed.FS

//go:embed views/index.html views/payment.html
var Views embed.FS
