package db

import _ "embed"

//go:embed schema.sql
var schemaSQL string

const SchemaVersion = 1
