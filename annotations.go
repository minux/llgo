// Copyright 2013 The llgo Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package llgo

import (
	"code.google.com/p/go.tools/go/types"
	goimporter "code.google.com/p/go.tools/importer"
	"code.google.com/p/go.tools/ssa"
	"go/ast"
	"go/token"
)

// processAnnotations takes an *ssa.Package and a
// *importer.PackageInfo, and processes all of the
// llgo source annotations attached to each top-level
// function and global variable.
func (c *compiler) processAnnotations(u *unit, pkginfo *goimporter.PackageInfo) {
	members := make(map[types.Object]*LLVMValue, len(u.globals))
	for k, v := range u.globals {
		members[k.(ssa.Member).Object()] = v
	}
	applyAttributes := func(attrs []Attribute, idents ...*ast.Ident) {
		if len(attrs) == 0 {
			return
		}
		for _, ident := range idents {
			if v := members[pkginfo.ObjectOf(ident)]; v != nil {
				for _, attr := range attrs {
					attr.Apply(v)
				}
			}
		}
	}
	for _, f := range pkginfo.Files {
		for _, decl := range f.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				attrs := parseAttributes(decl.Doc)
				applyAttributes(attrs, decl.Name)
			case *ast.GenDecl:
				if decl.Tok != token.VAR {
					continue
				}
				for _, spec := range decl.Specs {
					varspec := spec.(*ast.ValueSpec)
					attrs := parseAttributes(decl.Doc)
					applyAttributes(attrs, varspec.Names...)
				}
			}
		}
	}
}
