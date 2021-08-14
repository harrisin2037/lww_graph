package undirect

type DFS struct {
	LWWGraphImpl
	start, end VertexValue
	marked     map[VertexValue]bool
	dict       map[VertexValue][]VertexValue
	paths      []VertexValue
}

func (graph *LWWGraphImpl) NewDFS(start, end VertexValue) *DFS {
	return &DFS{
		*graph,
		start, end,
		make(map[VertexValue]bool),
		graph.GetAdjacencyVerticesList(),
		[]VertexValue{},
	}
}

func (dfs *DFS) Search() [][]VertexValue {

	var (
		current = []VertexValue{dfs.start}
		result  = [][]VertexValue{}
	)

	return *dfs.search(current, &result)
}

func (dfs *DFS) search(current []VertexValue, result *[][]VertexValue) *[][]VertexValue {

	if current[len(current)-1].IsEqual(dfs.end) {
		*result = append(*result, current)
	}

	last := current[len(current)-1]

	for i := 0; i < len(dfs.dict[last]); i++ {

		if exist, ok := dfs.marked[dfs.dict[last][i]]; ok && exist {
			continue
		}

		current = append(current, dfs.dict[last][i])
		dfs.marked[dfs.dict[last][i]] = true

		dfs.search(current, result)
		dfs.marked[dfs.dict[last][i]] = false

		current = current[:len(current)-1]
	}

	return result
}
