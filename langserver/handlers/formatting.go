package handlers

import (
	"context"
	"fmt"

	lsctx "github.com/hashicorp/terraform-ls/internal/context"
	"github.com/hashicorp/terraform-ls/internal/hcl"
	ilsp "github.com/hashicorp/terraform-ls/internal/lsp"
	"github.com/hashicorp/terraform-ls/internal/terraform/exec"
	"github.com/hashicorp/terraform-ls/internal/terraform/rootmodule"
	lsp "github.com/sourcegraph/go-lsp"
)

func (h *logHandler) TextDocumentFormatting(ctx context.Context, params lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	var edits []lsp.TextEdit

	fs, err := lsctx.Filesystem(ctx)
	if err != nil {
		return edits, err
	}

	tff, err := lsctx.TerraformFormatterFinder(ctx)
	if err != nil {
		return edits, err
	}

	fh := ilsp.FileHandlerFromDocumentURI(params.TextDocument.URI)
	file, err := fs.GetFile(fh)
	if err != nil {
		return edits, err
	}

	format, err := findTerraformFormatter(ctx, tff, file.Dir())
	if err != nil {
		return edits, err
	}

	formatted, err := format(ctx, file.Text())
	if err != nil {
		return edits, err
	}

	changes := hcl.Diff(file, formatted)

	return ilsp.TextEdits(changes), nil
}

func findTerraformFormatter(ctx context.Context, tff rootmodule.TerraformFormatterFinder, dir string) (exec.Formatter, error) {
	isLoaded, err := tff.IsTerraformLoaded(dir)
	if err != nil {
		if rootmodule.IsRootModuleNotFound(err) {
			return tff.TerraformFormatterForDir(ctx, dir)
		}
		return nil, err
	} else {
		if !isLoaded {
			// TODO: block until it's available <-tff.TerraformLoadingDone()
			return nil, fmt.Errorf("terraform is not available yet for %s", dir)
		}
	}

	return tff.TerraformFormatterForDir(ctx, dir)
}
