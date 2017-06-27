package epf

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {

	parser, err := NewParser("assets/demo")
	assert.NoError(t, err)
	defer parser.Close()

	md := parser.Metadata()

	assert.Equal(t, 3, len(md.Fields))
	assert.Equal(t, 3, len(md.Types))
	assert.Equal(t, 7, md.TotalItems)
	assert.Equal(t, ExportModeTypeFull, md.ExportMode)
	assert.Contains(t, md.PrimaryKey, "id")

	for index := 1; index <= md.TotalItems; index++ {
		item, err := parser.Read()
		assert.NoError(t, err)

		assert.Contains(t, item, "export_date")
		assert.Contains(t, item, "id")
		assert.Contains(t, item, "name")

		assert.IsType(t, int64(0), item["export_date"])

		assert.Equal(t, int64(1490173201020), item["export_date"].(int64))
		assert.Equal(t, index, item["id"])
		assert.Equal(t, fmt.Sprintf("item %d", index), item["name"])
	}
}
