package undirect

import (
	"sort"
)

type Bias int

const (
	Adds    Bias = 0
	Removal Bias = 1
)

type LWWGraph interface {

	// The function check if the vertex is exist in the graph or not.
	// vertex exist when the vertex is:
	//   - in the vertices list
	//   - in vertices list and not in tombstone vertices list
	//   - in both list but timestamp of record from vertices list is greater
	//   - in both list and no time difference but with adds bias
	IsVertexExist(value VertexValue) bool
	// AddVertex check if the record exist, if existed then will udpate the timestamp
	// of existing record to prevent lost of the relations of the edges.
	// If not exist, it will be append to vertices list and the matrix and tombstone matrix
	// wiht null for the diagonal of the matrix.
	AddVertex(value VertexValue) LWWVertex
	// It get the vertex when it exist in terms of LWW aspect. Related to IsVertexExist
	GetVertex(value VertexValue) LWWVertex
	// It remove all of the edges connected and the vertex itself if it exist
	RemoveVertex(value VertexValue)

	// It add the edge when:
	// 	- the vertices exist
	// 	- the vertices are not the same vertex
	// 	then it expends the vertex matix with the connection and the tombstone vertex matrix will null
	AddEdge(v1, v2 LWWVertex) LWWEdge
	// It return the edge of the two, by generate an adjacency vertices list
	// and return the edge that connect with the provided vertices value
	GetEdge(v1, v2 VertexValue) LWWEdge
	// It return the edges that connected with the provided vertex,
	// by generate an adjacency vertices list and return the edges that connect with the provided value
	GetEdges(value VertexValue) []LWWEdge
	// it search through the matrix by the DFS function and get all of the paths between start and end
	GetPaths(start, end VertexValue) [][]VertexValue
	// it update the tombstone if the vertex exist
	RemoveEdgeByVertices(v1, v2 VertexValue)

	// it merge the other graph when the component timestamp is smaller
	Merge(other LWWGraph)
	// get the adjacency vertices of every vertex
	GetAdjacencyVerticesList() map[VertexValue][]VertexValue
	// The function check if the component is exist in the graph or not logically by timestamp
	//
	// component exist when the timestamp is:
	//   - add timestamp is greater
	//   - adds bias when no difference in terms of timestamp
	IsComponentExist(add, remove int64) bool
	// It return the connected vertices, by generate an adjacency vertices list
	// and return the vertices that connect with the provided vertex value
	GetConnectedVertices(value VertexValue) []LWWVertex
	// retrieve the graph bias
	GetBias() Bias
	// retrieve the graph clock
	GetClock() Clock
	// retrieve the graph vertices
	GetVertices() map[VertexValue]LWWVertex
	// retrieve the graph tombstone vertices list
	GetTombstoneVertices() map[VertexValue]LWWVertex
	// retrieve the graph edge matrix
	GetEdgesMatrix() map[VertexValue]map[VertexValue]LWWEdge
	// retrieve the graph edge tombstone matrix
	GetTombstoneEdgesMatrix() map[VertexValue]map[VertexValue]LWWEdge
}

type LWWGraphImpl struct {
	clock                Clock
	bias                 Bias
	vertices             map[VertexValue]LWWVertex
	tombstoneVertices    map[VertexValue]LWWVertex
	edgesMatrix          map[VertexValue]map[VertexValue]LWWEdge
	tombstoneEdgesMatrix map[VertexValue]map[VertexValue]LWWEdge
}

func NewLWWGraph(bias Bias, clockImpl Clock) LWWGraph {
	if bias != Adds && bias != Removal {
		bias = Adds
	}
	if clockImpl == nil {
		clockImpl = &clock{}
	}
	return &LWWGraphImpl{
		clock:                clockImpl,
		bias:                 bias,
		vertices:             make(map[VertexValue]LWWVertex),
		tombstoneVertices:    make(map[VertexValue]LWWVertex),
		edgesMatrix:          make(map[VertexValue]map[VertexValue]LWWEdge),
		tombstoneEdgesMatrix: make(map[VertexValue]map[VertexValue]LWWEdge),
	}
}

func (graph *LWWGraphImpl) AddVertex(value VertexValue) LWWVertex {

	vertex := NewLWWVertex(value, graph.clock)

	if graph.IsVertexExist(value) {
		graph.vertices[vertex.GetValue()].SetTimestamp(graph.clock.Now().UnixNano())
		return graph.GetVertex(value)
	}

	graph.vertices[vertex.GetValue()] = vertex

	for m := range graph.vertices {
		for n := range graph.vertices {
			if v, ok := graph.edgesMatrix[m]; !ok {
				graph.edgesMatrix[m] = map[VertexValue]LWWEdge{n: nil}
			} else if _, ok := v[n]; !ok {
				graph.edgesMatrix[m][n] = nil
			}
			if v, ok := graph.tombstoneEdgesMatrix[m]; !ok {
				graph.tombstoneEdgesMatrix[m] = map[VertexValue]LWWEdge{n: nil}
			} else if _, ok := v[n]; !ok {
				graph.tombstoneEdgesMatrix[m][n] = nil
			}
		}
	}

	return vertex
}

func (graph *LWWGraphImpl) IsVertexExist(value VertexValue) bool {

	v, ok := graph.vertices[value]
	if !ok {
		return false
	}

	tv, ok := graph.tombstoneVertices[value]
	if !ok {
		return true
	}

	switch graph.bias {
	case Removal:
		// when vertex adds timestamp is greater than removal timestamp, it exists
		return v.GetTimestamp() > tv.GetTimestamp()
	default: // Adds
		// when vertex adds timestamp is greater than or equal to removal timestamp, it exists
		return v.GetTimestamp() >= tv.GetTimestamp()
	}
}

func (graph *LWWGraphImpl) IsComponentExist(add, remove int64) bool {

	switch graph.bias {
	case Removal:
		// when vertex adds timestamp is greater than removal timestamp, it exists
		return add > remove
	default: // Adds
		// when vertex adds timestamp is greater than or equal to removal timestamp, it exists
		return add >= remove
	}
}

func (graph *LWWGraphImpl) GetVertex(value VertexValue) LWWVertex {

	if graph.IsVertexExist(value) {
		return graph.vertices[value]
	}

	return nil
}

func (graph *LWWGraphImpl) GetConnectedVertices(value VertexValue) []LWWVertex {

	if !graph.IsVertexExist(value) {
		return nil
	}

	var (
		matrix = graph.GetAdjacencyVerticesList()
		arr    = []LWWVertex{}
	)

	for _, v := range matrix[value] {
		arr = append(arr, graph.GetVertex(v))
	}

	return arr
}

func (graph *LWWGraphImpl) RemoveVertex(value VertexValue) {

	// normally, it is not a metter to append of update the remove set
	// but as the vertex itself might has dependences(edges) in other
	// replicas, it is safer to check is it exist locally and remove
	// locally to make a more reasonable approach for graph use case
	if !graph.IsVertexExist(value) {
		return
	}

	vertices := graph.GetConnectedVertices(value)
	graph.tombstoneVertices[value] = NewLWWVertex(value, graph.clock)

	for i := 0; i < len(vertices); i++ {
		edgeVertices := graph.edgesMatrix[value][vertices[i].GetValue()].GetVertices()
		removeEdge := NewLWWEdgeImpl([]LWWVertex{edgeVertices[0], edgeVertices[1]}, graph.clock)
		graph.tombstoneEdgesMatrix[edgeVertices[0].GetValue()][edgeVertices[1].GetValue()] = removeEdge
		graph.tombstoneEdgesMatrix[edgeVertices[1].GetValue()][edgeVertices[0].GetValue()] = removeEdge
	}

	return
}

func (graph *LWWGraphImpl) AddEdge(v1, v2 LWWVertex) LWWEdge {

	if v1.GetValue().IsEqual(v2.GetValue()) {
		return nil
	}

	if !graph.IsVertexExist(v1.GetValue()) {
		v1 = graph.AddVertex(v1.GetValue())
	}

	if !graph.IsVertexExist(v2.GetValue()) {
		v2 = graph.AddVertex(v2.GetValue())
	}

	edge := NewLWWEdgeImpl([]LWWVertex{v1, v2}, graph.clock)

	graph.edgesMatrix[v1.GetValue()][v2.GetValue()] = edge
	graph.edgesMatrix[v2.GetValue()][v1.GetValue()] = edge

	graph.tombstoneEdgesMatrix[v1.GetValue()][v2.GetValue()] = nil
	graph.tombstoneEdgesMatrix[v2.GetValue()][v1.GetValue()] = nil

	return edge
}

func (graph *LWWGraphImpl) GetEdge(v1, v2 VertexValue) LWWEdge {

	if !graph.IsVertexExist(v1) || !graph.IsVertexExist(v2) {
		return nil
	}

	dict := graph.GetAdjacencyVerticesList()
	if _, ok := dict[v1]; !ok {
		return nil
	}

	if _, ok := dict[v2]; !ok {
		return nil
	}

	return graph.edgesMatrix[v1][v2]
}

func (graph *LWWGraphImpl) GetEdges(value VertexValue) []LWWEdge {

	if !graph.IsVertexExist(value) {
		return nil
	}

	edges := []LWWEdge{}

	dict := graph.GetAdjacencyVerticesList()
	if adj, ok := dict[value]; !ok || len(adj) == 0 {
		return nil
	}

	for i := 0; i < len(dict[value]); i++ {
		edge := graph.edgesMatrix[value][dict[value][i]]
		edges = append(edges, edge)
	}

	return edges
}

func (graph *LWWGraphImpl) GetPaths(v1, v2 VertexValue) [][]VertexValue {
	dfs := graph.NewDFS(v1, v2)
	return dfs.Search()
}

func (graph *LWWGraphImpl) RemoveEdgeByVertices(v1, v2 VertexValue) {

	if v1.IsEqual(v2) {
		return
	}

	vertex1 := graph.GetVertex(v1)
	vertex2 := graph.GetVertex(v2)

	if vertex1 == nil || vertex2 == nil {
		return
	}

	if _, ok := graph.edgesMatrix[v1][v2]; !ok {
		return
	}

	edge := NewLWWEdgeImpl([]LWWVertex{vertex1, vertex2}, graph.clock)

	if te, ok := graph.tombstoneEdgesMatrix[v1][v2]; !ok || te == nil {
		graph.tombstoneEdgesMatrix[v1][v2] = edge
		graph.tombstoneEdgesMatrix[v2][v1] = edge
		return
	}

	if graph.tombstoneEdgesMatrix[v1][v2].GetTimestamp() < edge.GetTimestamp() {
		graph.tombstoneEdgesMatrix[v1][v2] = edge
		graph.tombstoneEdgesMatrix[v2][v1] = edge
		return
	}

	return
}

func (graph *LWWGraphImpl) Merge(other LWWGraph) {
	graph.vertices = mergeVertices(graph.vertices, other.GetVertices())
	graph.tombstoneVertices = mergeVertices(graph.tombstoneVertices, other.GetTombstoneVertices())
	graph.edgesMatrix = mergeEdgesMatrix(graph.edgesMatrix, other.GetEdgesMatrix())
	graph.tombstoneEdgesMatrix = mergeEdgesMatrix(graph.tombstoneEdgesMatrix, other.GetTombstoneEdgesMatrix())
}

func mergeVertices(source, mergeWith map[VertexValue]LWWVertex) map[VertexValue]LWWVertex {

	if mergeWith == nil {
		return source
	}

	for k := range mergeWith {
		if mergeWith[k] == nil {
			continue
		}
		if _, ok := source[k]; !ok || source[k] == nil {
			source[k] = mergeWith[k]
			continue
		}
		if source[k].GetTimestamp() < mergeWith[k].GetTimestamp() {
			source[k] = mergeWith[k]
		}
	}

	return source
}

func mergeEdgesMatrix(source, mergeWith map[VertexValue]map[VertexValue]LWWEdge) map[VertexValue]map[VertexValue]LWWEdge {

	if mergeWith == nil {
		return source
	}

	for m := range mergeWith {
		if mergeWith[m] == nil {
			continue
		}
		for n := range mergeWith[m] {
			if mergeWith[m][n] == nil {
				continue
			}
			if _, ok := source[m]; !ok || source[m] == nil {
				source[m] = mergeWith[m]
				continue
			}
			if _, ok := source[m][n]; !ok || source[m][n] == nil {
				source[m][n] = mergeWith[m][n]
				continue
			}
			if source[m][n].GetTimestamp() < mergeWith[m][n].GetTimestamp() {
				source[m][n] = mergeWith[m][n]
			}
		}
	}

	return source
}

func (graph *LWWGraphImpl) GetAdjacencyVerticesList() map[VertexValue][]VertexValue {

	dict := make(map[VertexValue][]VertexValue)

	for k := range graph.vertices {
		if !graph.IsVertexExist(k) {
			continue
		}
		dict[k] = []VertexValue{}
	}

	if len(dict) == 0 {
		return nil
	}

	for m, v := range graph.edgesMatrix {
		tombstoneVertexM, ok := graph.tombstoneVertices[m]
		if ok && tombstoneVertexM != nil {
			if !graph.IsComponentExist(graph.vertices[m].GetTimestamp(), tombstoneVertexM.GetTimestamp()) {
				continue
			}
		}
		for n, edge := range v {
			if edge == nil {
				continue
			}
			tombstoneVertexN, ok := graph.tombstoneVertices[n]
			if ok && tombstoneVertexN != nil {
				if !graph.IsComponentExist(graph.vertices[n].GetTimestamp(), tombstoneVertexN.GetTimestamp()) {
					continue
				}
			}
			tombstoneEdge, ok := graph.tombstoneEdgesMatrix[m][n]
			if ok && tombstoneEdge != nil {
				if !graph.IsComponentExist(edge.GetTimestamp(), tombstoneEdge.GetTimestamp()) {
					continue
				}
			}
			if _, ok := dict[m]; ok {
				dict[m] = append(dict[m], n)
			}
		}

		sort.Slice(dict[m], func(i, j int) bool {
			return string(dict[m][i]) < string(dict[m][j])
		})
	}

	return dict
}

func (graph *LWWGraphImpl) GetBias() Bias {
	return graph.bias
}

func (graph *LWWGraphImpl) GetClock() Clock {
	return graph.clock
}

func (graph *LWWGraphImpl) GetVertices() map[VertexValue]LWWVertex {
	return graph.vertices
}

func (graph *LWWGraphImpl) GetTombstoneVertices() map[VertexValue]LWWVertex {
	return graph.tombstoneVertices
}

func (graph *LWWGraphImpl) GetEdgesMatrix() map[VertexValue]map[VertexValue]LWWEdge {
	return graph.edgesMatrix
}

func (graph *LWWGraphImpl) GetTombstoneEdgesMatrix() map[VertexValue]map[VertexValue]LWWEdge {
	return graph.tombstoneEdgesMatrix
}
