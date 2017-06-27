package epf

// Metadata struct
type Metadata struct {
	Fields     []string
	PrimaryKey []string
	Types      []string
	ExportMode ExportModeType
	TotalItems int
}
