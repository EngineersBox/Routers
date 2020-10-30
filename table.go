package routers

// DVRTable ...
type DVRTable map[RouterId]map[RouterId]interface{}

// Row ...
type Row map[RouterId]interface{}

// DVRTable.put ... Place a value at location (i,j)
func (m DVRTable) put(i RouterId, j RouterId, value interface{}) {
	inner, ok := m[i]
	if !ok {
		inner = make(Row, j+1)
	}
	inner[j] = value
	m[i] = inner
}

// DVRTable.putRow ... Place a row of values at location (i,)
func (m DVRTable) putRow(i RouterId, row Row) bool {
	_, ok := m[i]
	if !ok {
		m[i] = row
	}
	return ok
}

// DVRTable.get ... Retrieve a value at location (i,j)
func (m DVRTable) get(i RouterId, j RouterId) interface{} {
	inner, ok := m[i]
	if !ok {
		return nil
	}
	return inner[j]
}

// DVRTable.getRow ... Retrieve a row at location (i,)
func (m DVRTable) getRow(i RouterId) Row {
	inner, ok := m[i]
	if !ok {
		return nil
	}
	return inner
}
