package idescriptive_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/maratori/idescriptive/pkg/idescriptive"
)

func TestAnalyzer_AllTypes_False(t *testing.T) {
	t.Parallel()

	testdata, err := filepath.Abs("testdata/alltypesfalse")
	require.NoError(t, err)
	analysistest.Run(t, testdata, idescriptive.NewAnalyzer())
}

func TestAnalyzer_AllTypes_True(t *testing.T) {
	t.Parallel()

	testdata, err := filepath.Abs("testdata/alltypestrue")
	require.NoError(t, err)

	analyzer := idescriptive.NewAnalyzer()
	err = analyzer.Flags.Set("all-types", "true")
	require.NoError(t, err)
	analysistest.Run(t, testdata, analyzer)
}
