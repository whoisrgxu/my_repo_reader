package filters

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Broad text/code extensions
var TextExt = map[string]struct{}{
	// docs & markup
	".txt": {}, ".md": {}, ".mdx": {}, ".rst": {}, ".adoc": {}, ".asciidoc": {},
	".tex": {}, ".bib": {}, ".org": {}, ".textile": {},

	// data / logs
	".csv": {}, ".tsv": {}, ".psv": {}, ".ndjson": {}, ".log": {}, ".properties": {},

	// config / serialization
	".json": {}, ".json5": {}, ".jsonc": {},
	".yaml": {}, ".yml": {}, ".toml": {}, ".ini": {}, ".cfg": {}, ".conf": {}, ".env": {},

	// html/xml/svg
	".html": {}, ".htm": {}, ".xhtml": {}, ".xml": {}, ".xsd": {}, ".xsl": {}, ".xslt": {}, ".dtd": {}, ".svg": {},

	// styles
	".css": {}, ".scss": {}, ".sass": {}, ".less": {}, ".styl": {},

	// web templates
	".ejs": {}, ".pug": {}, ".jade": {}, ".hbs": {}, ".mustache": {}, ".njk": {}, ".twig": {}, ".liquid": {},

	// js/ts ecosystem
	".js": {}, ".mjs": {}, ".cjs": {}, ".jsx": {},
	".ts": {}, ".tsx": {},
	".vue": {}, ".svelte": {}, ".astro": {},

	// go
	".go": {}, ".tmpl": {}, ".mod": {}, ".sum": {},

	// python
	".py": {}, ".pyi": {}, ".pyw": {}, ".pyx": {}, ".pxd": {}, ".pxi": {},

	// ruby
	".rb": {}, ".erb": {}, ".rake": {}, ".gemspec": {},

	// php
	".php": {}, ".phtml": {}, ".php3": {}, ".php4": {}, ".php5": {}, ".php7": {}, ".php8": {},

	// java / groovy / kotlin / scala
	".java": {}, ".jsp": {}, ".groovy": {}, ".gradle": {}, ".gvy": {}, ".gy": {}, ".gsh": {},
	".kt": {}, ".kts": {}, ".ktm": {},
	".scala": {}, ".sc": {}, ".sbt": {},

	// c / c++ / objc / swift
	".c": {}, ".h": {}, ".hpp": {}, ".hh": {}, ".hxx": {}, ".cpp": {}, ".cc": {}, ".cxx": {}, ".ino": {}, ".ipp": {},
	".m": {}, ".mm": {}, ".pch": {},
	".swift": {}, ".xcconfig": {}, ".pbxproj": {}, ".xcscheme": {}, ".xcworkspacedata": {}, ".plist": {}, ".strings": {},

	// .NET / F#
	".cs": {}, ".csx": {}, ".fs": {}, ".fsi": {}, ".fsx": {},

	// rust
	".rs": {}, ".ron": {},

	// haskell / ocaml
	".hs": {}, ".lhs": {}, ".cabal": {},
	".ml": {}, ".mli": {}, ".re": {}, ".rei": {},

	// erlang / elixir
	".erl": {}, ".hrl": {}, ".ex": {}, ".exs": {}, ".eex": {}, ".leex": {}, ".heex": {},

	// lua
	".lua": {}, ".rockspec": {},

	// shells
	".sh": {}, ".bash": {}, ".zsh": {}, ".ksh": {}, ".fish": {}, ".command": {},

	// powershell / batch
	".ps1": {}, ".psm1": {}, ".psd1": {}, ".bat": {}, ".cmd": {},

	// build / tooling
	".cmake": {}, ".ninja": {}, ".bazel": {}, ".bzl": {},

	// infra / IaC
	".tf": {}, ".tfvars": {}, ".hcl": {}, ".cue": {}, ".dhall": {},

	// idl / schema
	".proto": {}, ".thrift": {}, ".avdl": {},

	// query / graph
	".sql": {}, ".psql": {}, ".mysql": {}, ".cql": {}, ".graphql": {}, ".gql": {},

	// diagrams
	".plantuml": {}, ".puml": {}, ".dot": {}, ".gv": {}, ".mermaid": {}, ".mmd": {},

	// data science
	".r": {}, ".R": {}, ".Rmd": {}, ".qmd": {}, ".jl": {},
}

// Well-known text filenames (no extension)
var TextFilenames = map[string]struct{}{
	"Makefile": {}, "CMakeLists.txt": {},
	"Dockerfile": {}, ".dockerignore": {},
	".gitignore": {}, ".gitattributes": {}, ".gitmodules": {},
	".npmrc": {}, ".nvmrc": {}, ".prettierrc": {}, ".eslintignore": {}, ".eslintrc": {},
	"SConstruct": {}, "SConscript": {},
	"BUILD": {}, "BUILD.bazel": {}, "WORKSPACE": {}, "WORKSPACE.bazel": {},
	"Gemfile": {}, "Rakefile": {}, "Vagrantfile": {}, "Procfile": {}, "Jenkinsfile": {},
	"LICENSE": {}, "LICENSE.md": {}, "COPYING": {}, "README": {}, "README.md": {},
	"CHANGELOG": {}, "CHANGELOG.md": {}, "NOTICE": {}, "AUTHORS": {},
}

func hasTextyName(path string) bool {
	base := filepath.Base(path)
	if _, ok := TextFilenames[base]; ok {
		return true
	}
	ext := strings.ToLower(filepath.Ext(base))
	_, ok := TextExt[ext]
	return ok
}

// Robust content sniff
func isProbablyTextFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	const sniff = 8192
	buf := make([]byte, sniff)
	n, _ := f.Read(buf)
	if n == 0 {
		return true // empty counts as text
	}
	s := buf[:n]

	// NUL byte → binary
	if bytes.IndexByte(s, 0x00) != -1 {
		return false
	}

	// MIME hint
	mime := http.DetectContentType(s)
	if strings.HasPrefix(mime, "text/") {
		return true
	}
	switch mime {
	case "application/json", "application/javascript", "application/xml", "image/svg+xml":
		return true
	}

	// UTF-8 → text
	if utf8.Valid(s) {
		return true
	}

	// Printable ASCII heuristic
	printable := 0
	for _, b := range s {
		if b == 9 || b == 10 || b == 13 || (b >= 32 && b <= 126) {
			printable++
		}
	}
	return float64(printable)/float64(len(s)) >= 0.95
}

// Exported helper used by main
func IsTextFile(path string) bool {
	return hasTextyName(path) || isProbablyTextFile(path)
}
