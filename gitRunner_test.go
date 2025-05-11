package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTargetDir(t *testing.T) {
	t.Run("ReturnsZiplineeWorkIfSubdirIsDot", func(t *testing.T) {

		// act
		path := getTargetDir(".")

		assert.Equal(t, "/ziplinee-work", path)
	})

	t.Run("ReturnsZiplineeWorkSubdirIfSubdirIsSingleWord", func(t *testing.T) {

		// act
		path := getTargetDir("scripts")

		assert.Equal(t, "/ziplinee-work/scripts", path)
	})

	t.Run("ReturnsZiplineeWorkSubdirIfSubdirIsDotSlashSingleWord", func(t *testing.T) {

		// act
		path := getTargetDir("./scripts")

		assert.Equal(t, "/ziplinee-work/scripts", path)
	})

	t.Run("ReturnsZiplineeWorkSubdirIfSubdirIsMultipleWordsSeparatedBySlash", func(t *testing.T) {

		// act
		path := getTargetDir("scripts/sub")

		assert.Equal(t, "/ziplinee-work/scripts/sub", path)
	})

	t.Run("ReturnsZiplineeWorkSubdirIfSubdirIsDotSlashMultipleWordsSeparatedBySlash", func(t *testing.T) {

		// act
		path := getTargetDir("./scripts/sub")

		assert.Equal(t, "/ziplinee-work/scripts/sub", path)
	})

	t.Run("ReturnsZiplineeWorkSubdirIfSubdirIsAbsolutePath", func(t *testing.T) {

		// act
		path := getTargetDir("/scripts")

		assert.Equal(t, "/ziplinee-work/scripts", path)
	})
}
