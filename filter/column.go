package filter

type Column struct {
	QueryKey string
	DBColumn string
}

func Col(queryKey string, dbColumn ...string) Column {
	column := queryKey
	if len(dbColumn) > 0 && dbColumn[0] != "" {
		column = dbColumn[0]
	}

	return Column{
		QueryKey: queryKey,
		DBColumn: column,
	}
}
