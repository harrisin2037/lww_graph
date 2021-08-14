package undirect

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/go-test/deep"
)

var (
	mockGraphAddAction    = "add"
	mockGraphRemoveAction = "remove"
)

type mockFields struct {
	bias            Bias
	clock           Clock
	verticesPaths   [][]VertexValue
	removedVertices []VertexValue
}

// Check is the graph associative for merge
func TestLWWGraphImpl_Merge_Check_Associative(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")
	D := NewVertexValue("D")
	E := NewVertexValue("E")

	/*
		  A
		 / \
		B	C
	*/
	graphXMock := mockFields{
		bias: Adds,
		verticesPaths: [][]VertexValue{
			{A, B},
			{A, C},
		},
		clock: &testCkock{},
	}

	/*
		  D
		 / \
		C	E
	*/
	graphYMock := mockFields{
		bias: Adds,
		verticesPaths: [][]VertexValue{
			{D, C},
			{D, E},
		},
		clock: &testCkock{},
	}

	/*
		  B
		 / \
		C	D
	*/
	graphZMock := mockFields{
		bias: Adds,
		verticesPaths: [][]VertexValue{
			{B, C},
			{B, D},
		},
		clock: &testCkock{},
	}

	graphXLeft := NewMockGraph(graphXMock)
	graphYLeft := NewMockGraph(graphYMock)
	graphZLeft := NewMockGraph(graphZMock)

	graphXRight := NewMockGraph(graphXMock)
	graphYRight := NewMockGraph(graphYMock)
	graphZRight := NewMockGraph(graphZMock)

	// 1. X' = X U Y
	graphXLeft.Merge(graphYLeft)
	// 2. X'' = X' U Z
	graphXLeft.Merge(graphZLeft)

	// 3. Y' = Y U Z
	graphYRight.Merge(graphZRight)
	// 4. X'' = X U Y'
	graphXRight.Merge(graphYRight)

	gotLeft := graphXLeft.GetAdjacencyVerticesList()
	gotRight := graphXRight.GetAdjacencyVerticesList()

	// check are the results equal
	if !reflect.DeepEqual(gotLeft, gotRight) {
		t.Errorf("LWWGraphImpl.TestLWWGraphImpl_Merge_Check_Associative(): got %v, want %v, diff: %v", gotLeft, gotRight, deep.Equal(gotLeft, gotRight))
	}
}

// Check is the graph commutative for merge
func TestLWWGraphImpl_Merge_Check_Commutative(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")
	D := NewVertexValue("D")
	E := NewVertexValue("E")

	/*
		  A
		 / \
		B	C
	*/
	graphXMock := mockFields{
		bias: Adds,
		verticesPaths: [][]VertexValue{
			{A, B},
			{A, C},
		},
		clock: &testCkock{},
	}

	/*
		  D
		 / \
		C	E
	*/
	graphYMock := mockFields{
		bias: Adds,
		verticesPaths: [][]VertexValue{
			{D, C},
			{D, E},
		},
		clock: &testCkock{},
	}

	graphXLeft := NewMockGraph(graphXMock)
	graphYLeft := NewMockGraph(graphYMock)

	graphXRight := NewMockGraph(graphXMock)
	graphYRight := NewMockGraph(graphYMock)

	// X' = X U Y
	graphXLeft.Merge(graphYLeft)
	// Y' = Y U X
	graphYRight.Merge(graphXRight)

	gotLeft := graphXLeft.GetAdjacencyVerticesList()
	gotRight := graphYRight.GetAdjacencyVerticesList()

	// X' = Y' check are the results equal
	if !reflect.DeepEqual(gotLeft, gotRight) {
		t.Errorf("LWWGraphImpl.TestLWWGraphImpl_Merge_Check_Commutative(): got %v, want %v, diff: %v", gotLeft, gotRight, deep.Equal(gotLeft, gotRight))
	}
}

// Check is the graph Idempotent for merge
func TestLWWGraphImpl_Merge_Check_Idempotent(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")

	/*
		  A
		 / \
		B	C
	*/
	graphXMock := mockFields{
		bias: Adds,
		verticesPaths: [][]VertexValue{
			{A, B},
			{A, C},
		},
		clock: &testCkock{},
	}

	graphX := NewMockGraph(graphXMock)

	graphX1 := NewMockGraph(graphXMock)
	graphX2 := NewMockGraph(graphXMock)

	// X U X
	graphX1.Merge(graphX2)

	gotLeft := graphX.GetAdjacencyVerticesList()
	gotRight := graphX1.GetAdjacencyVerticesList()

	// check are they equal
	if !reflect.DeepEqual(gotLeft, gotRight) {
		t.Errorf("LWWGraphImpl.TestLWWGraphImpl_Merge_Check_Idempotent(): got %v, want %v, diff: %v", gotLeft, gotRight, deep.Equal(gotLeft, gotRight))
	}
}

func NewMockGraph(fields mockFields) LWWGraph {

	graph := NewLWWGraph(fields.bias, fields.clock)

	for i := 0; i < len(fields.verticesPaths); i++ {
		vertices := []LWWVertex{}
		for j := 0; j < len(fields.verticesPaths[i]); j++ {
			vertices = append(vertices, NewLWWVertex(fields.verticesPaths[i][j], graph.GetClock()))
		}
		if len(vertices) > 1 {
			for j := 0; j < len(vertices); j++ {
				if j == 0 {
					continue
				}
				graph.AddEdge(vertices[j-1], vertices[j])
			}
		}
	}

	for i := 0; i < len(fields.removedVertices); i++ {
		graph.RemoveVertex(fields.removedVertices[i])
	}

	return graph
}

func TestNewLWWGraph(t *testing.T) {
	type args struct {
		bias Bias
	}
	tests := []struct {
		name        string
		description string
		args        args
		want        LWWGraph
	}{
		{
			name: "test unknown case",
			args: args{Bias(10)},
			want: NewLWWGraph(Adds, nil),
		},
		{
			name: "test adds case",
			args: args{Adds},
			want: NewLWWGraph(Adds, nil),
		},
		{
			name: "test removal case",
			args: args{Removal},
			want: NewLWWGraph(Removal, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLWWGraph(tt.args.bias, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLWWGraph() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLWWGraphImpl_AddVertex(t *testing.T) {

	t.Run("test add one element", func(t *testing.T) {

		graph := NewLWWGraph(Adds, nil)

		vertexValueA := NewVertexValue("A")

		got := graph.AddVertex(vertexValueA)
		dup := graph.AddVertex(vertexValueA)

		if !reflect.DeepEqual(got, dup) {
			t.Errorf("LWWGraphImpl.AddVertex() fail to return added value = %v, want %v", dup, got)
		}

		vertices := graph.GetVertices()
		if !reflect.DeepEqual(vertices[vertexValueA], got) {
			t.Errorf("LWWGraphImpl.AddVertex() fail, vertices not contain = %v", got)
		}

		tsVertices := graph.GetTombstoneVertices()
		if v, ok := tsVertices[vertexValueA]; ok || v != nil || v == got {
			t.Errorf("LWWGraphImpl.AddVertex() fail, TombstoneVertices contain = %v", got)
		}

		edgeMatrix := graph.GetEdgesMatrix()
		if len(edgeMatrix) == 0 {
			t.Errorf("LWWGraphImpl.AddVertex() fail, edgeMatrix cannot initialize")
		}
		if v, ok := edgeMatrix[vertexValueA]; !ok && v == nil {
			t.Errorf("LWWGraphImpl.AddVertex() fail, edgeMatrix m dimension does not contain = %v", vertexValueA)
		} else if v, ok := edgeMatrix[vertexValueA][vertexValueA]; !ok && v == nil {
			t.Errorf("LWWGraphImpl.AddVertex() fail, edgeMatrix n dimension does not contain = %v", vertexValueA)
		}

		tombstoneEdgeMatrix := graph.GetTombstoneEdgesMatrix()
		if len(tombstoneEdgeMatrix) == 0 {
			t.Errorf("LWWGraphImpl.AddVertex() fail, tombstoneEdgeMatrix cannot initialize")
		}
		if v, ok := tombstoneEdgeMatrix[vertexValueA]; !ok && v == nil {
			t.Errorf("LWWGraphImpl.AddVertex() fail, tombstoneEdgeMatrix m dimension does not contain = %v", vertexValueA)
		} else if v, ok := tombstoneEdgeMatrix[vertexValueA][vertexValueA]; !ok && v == nil {
			t.Errorf("LWWGraphImpl.AddVertex() fail, tombstoneEdgeMatrix n dimension does not contain = %v", vertexValueA)
		}
	})

	t.Run("test add elements for matrix", func(t *testing.T) {

		graph := NewLWWGraph(Adds, nil)
		vertexValuesDict := map[string]VertexValue{}
		verticesDict := map[string]LWWVertex{}

		for i := 0; i < 10; i++ {
			elem := RandStringBytes(1)

			vertexValuesDict[elem] = NewVertexValue(elem)
			verticesDict[elem] = graph.AddVertex(vertexValuesDict[elem])
		}

		count := 0
		length := rand.Intn(10) + 5
		for {
			elem := RandStringBytes(1)
			if _, ok := vertexValuesDict[elem]; ok {
				continue
			}
			vertexValuesDict[elem] = NewVertexValue(elem)
			verticesDict[elem] = graph.AddVertex(vertexValuesDict[elem])
			count++
			if count == length {
				break
			}
		}

		vertices := graph.GetVertices()
		for _, v := range vertexValuesDict {
			_, ok := vertices[v]
			if !ok {
				t.Errorf("LWWGraphImpl.AddVertex() fail, vertices not contain = %v", v)
			}
		}

		tsVertices := graph.GetTombstoneVertices()
		for _, v := range vertexValuesDict {
			tsv, ok := tsVertices[v]
			if ok || tsv != nil {
				t.Errorf("LWWGraphImpl.AddVertex() fail, TombstoneVertices contain = %v", v)
			}
		}

		edgeMatrix := graph.GetEdgesMatrix()
		if len(edgeMatrix) == 0 {
			t.Errorf("LWWGraphImpl.AddVertex() fail, edgeMatrix cannot initialize")
		}

		tombstoneEdgeMatrix := graph.GetEdgesMatrix()
		if len(tombstoneEdgeMatrix) == 0 {
			t.Errorf("LWWGraphImpl.AddVertex() fail, tombstoneEdgeMatrix cannot initialize")
		}
	})

}

func TestVertexValue_IsEqual(t *testing.T) {
	type args struct {
		to VertexValue
	}
	tests := []struct {
		name  string
		value VertexValue
		args  args
		want  bool
	}{
		{
			name:  "is equal",
			value: VertexValue("A"),
			args:  args{VertexValue("A")},
			want:  true,
		},
		{
			name:  "is not equal",
			value: VertexValue("A"),
			args:  args{VertexValue("B")},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsEqual(tt.args.to); got != tt.want {
				t.Errorf("VertexValue.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLWWGraphImpl_IsVertexExist(t *testing.T) {
	type fields struct {
		vertex string
		bias   Bias
		clock  Clock
	}
	type args struct {
		value VertexValue
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test with exist vertex",
			fields: fields{
				"A",
				Adds,
				nil,
			},
			args: args{
				value: NewVertexValue("A"),
			},
			want: true,
		},
		{
			name: "test with vertex not exist",
			fields: fields{
				"B",
				Adds,
				nil,
			},
			args: args{
				value: NewVertexValue("A"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var (
				graph  = NewLWWGraph(tt.fields.bias, tt.fields.clock)
				vertex = NewVertexValue(tt.fields.vertex)
			)

			graph.AddVertex(vertex)

			if got := graph.IsVertexExist(tt.args.value); got != tt.want {
				t.Errorf("LWWGraphImpl.IsVertexExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLWWGraphImpl_IsVertexExist_When_Vertex_Removed(t *testing.T) {
	type fields struct {
		vertex string
		bias   Bias
		clock  Clock
	}
	type args struct {
		value VertexValue
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "test with vertex removed",
			fields: fields{
				"A",
				Adds,
				nil,
			},
			args: args{
				value: NewVertexValue("A"),
			},
			want: false,
		},
		{
			name: "test with vertex remove with bias adds",
			fields: fields{
				"A",
				Adds,
				&testCkock{},
			},
			args: args{
				value: NewVertexValue("A"),
			},
			want: true,
		},
		{
			name: "test with vertex remove with bias removal",
			fields: fields{
				"A",
				Removal,
				&testCkock{},
			},
			args: args{
				value: NewVertexValue("A"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var (
				graph  = NewLWWGraph(tt.fields.bias, tt.fields.clock)
				vertex = NewVertexValue(tt.fields.vertex)
			)

			graph.AddVertex(vertex)

			graph.RemoveVertex(vertex)

			if got := graph.IsVertexExist(tt.args.value); got != tt.want {
				t.Errorf("LWWGraphImpl.IsVertexExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLWWGraphImpl_GetConnectedVertices(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")
	D := NewVertexValue("D")
	E := NewVertexValue("E")

	type args struct {
		value VertexValue
	}
	tests := []struct {
		name   string
		fields mockFields
		args   args
		want   []VertexValue
	}{
		{
			name: "test get connected vertices of empty",
			fields: mockFields{
				verticesPaths: [][]VertexValue{{A}, {B}, {C}, {D}, {E}},
				bias:          Adds,
				clock:         nil,
			},
			args: args{A},
			want: []VertexValue{},
		},
		{
			name: "test get connected vertices",
			fields: mockFields{
				/*
					   A - B - D - E
						\ /  /
						 C -
				*/
				verticesPaths: [][]VertexValue{
					{A, B},
					{A, C},
					{B, C},
					{B, D},
					{C, D},
					{D, E},
				},
				bias:  Adds,
				clock: nil,
			},
			args: args{A},
			want: []VertexValue{B, C},
		},
		{
			name: "test get vertices (removed edge)",
			fields: mockFields{
				/*
					   A - B - D - E
						\ /  /
						 C -
				*/
				verticesPaths: [][]VertexValue{
					{A, B},
					{A, C},
					{B, C},
					{B, D},
					{C, D},
					{D, E},
				},
				removedVertices: []VertexValue{C},
				bias:            Adds,
				clock:           nil,
			},
			args: args{A},
			want: []VertexValue{B},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			graph := NewMockGraph(tt.fields)

			cv := graph.GetConnectedVertices(tt.args.value)

			got := []VertexValue{}
			for i := 0; i < len(cv); i++ {
				got = append(got, cv[i].GetValue())
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LWWGraphImpl.GetConnectedVertices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLWWGraphImpl_RemoveVertex(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")
	D := NewVertexValue("D")
	E := NewVertexValue("E")
	F := NewVertexValue("F")

	type args struct {
		isEdge bool
		values []VertexValue
	}
	tests := []struct {
		name   string
		fields mockFields
		args   args
	}{
		{
			name: "remove vertices without edge",
			fields: mockFields{
				verticesPaths: [][]VertexValue{{A}, {B}, {C}, {D}, {E}, {F}},
				bias:          Adds,
				clock:         nil,
			},
			args: args{
				values: []VertexValue{A, B},
			},
		},
		{
			name: "remove vertices withs edges",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{A, B, C, D},
					{A, B, A, C, D, E},
					{A, B, C, D, E},
					{A, C, A, B, D, E},
					{A, C, B, D, E},
					{A, C, D, E},
				},
				bias:  Adds,
				clock: nil,
			},
			args: args{
				values: []VertexValue{A, B},
				isEdge: true,
			},
		},
		{
			name: "remove vertices without edge with remove bias and same timestamp",
			fields: mockFields{
				verticesPaths: [][]VertexValue{{A}, {B}, {C}, {D}, {E}, {F}},
				bias:          Removal,
				clock:         &testCkock{},
			},
			args: args{
				values: []VertexValue{A, B},
			},
		},
		{
			name: "remove vertices withs edges with remove bias and same timestamp",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{A, B, C, D},
					{A, B, A, C, D, E},
					{A, B, C, D, E},
					{A, C, A, B, D, E},
					{A, C, B, D, E},
					{A, C, D, E},
				},
				bias:  Removal,
				clock: &testCkock{},
			},
			args: args{
				values: []VertexValue{A, B},
				isEdge: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := NewMockGraph(tt.fields)
			for i := 0; i < len(tt.args.values); i++ {
				graph.RemoveVertex(tt.args.values[i])
			}
			for i := 0; i < len(tt.args.values); i++ {
				if got := graph.GetVertex(tt.args.values[i]); got != nil {
					t.Errorf("LWWGraphImpl.RemoveVertex() failed, can access removed vertex with GetVertex, got: %v", got.GetValue())
				}
				if tt.args.isEdge {
					if got := graph.GetEdges(tt.args.values[i]); got != nil {
						edges := map[VertexValue][][]VertexValue{tt.args.values[i]: {}}

						for k := 0; k < len(got); k++ {
							vertices := got[k].GetVertices()
							edges[tt.args.values[i]] = append(edges[tt.args.values[i]], []VertexValue{vertices[0].GetValue(), vertices[1].GetValue()})
						}
						t.Errorf("LWWGraphImpl.RemoveVertex() failed, can access removed vertex edges with GetEdges got: %v", edges)
					}
				}
			}
		})
	}
}

func TestLWWGraphImpl_RemoveVertex_of_Graph_With_Edge(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")
	D := NewVertexValue("D")
	E := NewVertexValue("E")

	type args struct {
		values []VertexValue
	}
	tests := []struct {
		name   string
		fields mockFields
		args   args
	}{
		{
			name: "remove vertices withs edges",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{A, B, C, D},
					{A, B, A, C, D, E},
					{A, B, C, D, E},
					{A, C, A, B, D, E},
					{A, C, B, D, E},
					{A, C, D, E},
				},
				bias:  Adds,
				clock: nil,
			},
			args: args{
				values: []VertexValue{A, B},
			},
		},
		{
			name: "remove vertices withs edges with remove bias and same timestamp",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{A, B, C, D},
					{A, B, A, C, D, E},
					{A, B, C, D, E},
					{A, C, A, B, D, E},
					{A, C, B, D, E},
					{A, C, D, E},
				},
				bias:  Removal,
				clock: &testCkock{},
			},
			args: args{
				values: []VertexValue{A, B},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := NewMockGraph(tt.fields)
			for i := 0; i < len(tt.args.values); i++ {
				graph.RemoveVertex(tt.args.values[i])
			}
			for i := 0; i < len(tt.args.values); i++ {
				if got := graph.GetVertex(tt.args.values[i]); got != nil {
					t.Errorf("LWWGraphImpl.RemoveVertex() failed, can access removed vertex with GetVertex, got: %v", got.GetValue())
				}
				if got := graph.GetEdges(tt.args.values[i]); got != nil {

					edges := map[VertexValue][][]VertexValue{tt.args.values[i]: {}}

					for k := 0; k < len(got); k++ {
						vertices := got[k].GetVertices()
						edges[tt.args.values[i]] = append(edges[tt.args.values[i]], []VertexValue{vertices[0].GetValue(), vertices[1].GetValue()})
					}

					t.Errorf("LWWGraphImpl.RemoveVertex() failed, can access removed vertex edges with GetEdges got: %v", edges)
				}
			}
		})
	}
}

func TestLWWGraphImpl_GetEdges(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")
	D := NewVertexValue("D")
	E := NewVertexValue("E")

	type args struct {
		v1 VertexValue
	}
	tests := []struct {
		name   string
		fields mockFields
		args   args
		want   [][]VertexValue
	}{
		{
			name: "test get edges of empty",
			fields: mockFields{
				verticesPaths: [][]VertexValue{{A}, {B}, {C}, {D}, {E}},
				bias:          Adds,
				clock:         nil,
			},
			args: args{A},
			want: [][]VertexValue{},
		},
		{
			name: "test get edges",
			fields: mockFields{
				/*
					   A - B - D - E
						\ /  /
						 C -
				*/
				verticesPaths: [][]VertexValue{
					{A, B},
					{A, C},
					{B, C},
					{B, D},
					{C, D},
					{D, E},
				},
				bias:  Adds,
				clock: nil,
			},
			args: args{A},
			want: [][]VertexValue{
				{A, B},
				{A, C},
			},
		},
		{
			name: "test get edges (removed edge)",
			fields: mockFields{
				/*
					   A - B - D - E
						\ /  /
						 C -
				*/
				verticesPaths: [][]VertexValue{
					{A, B},
					{A, C},
					{B, C},
					{B, D},
					{C, D},
					{D, E},
				},
				removedVertices: []VertexValue{C},
				bias:            Adds,
				clock:           nil,
			},
			args: args{A},
			want: [][]VertexValue{
				{A, B},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			graph := NewMockGraph(tt.fields)

			got := [][]VertexValue{}

			ge := graph.GetEdges(tt.args.v1)
			for i := 0; i < len(ge); i++ {
				vs := ge[i].GetVertices()
				got = append(got, []VertexValue{vs[0].GetValue(), vs[1].GetValue()})
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LWWGraphImpl.GetPaths() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestLWWGraphImpl_GetPaths(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")
	D := NewVertexValue("D")
	E := NewVertexValue("E")

	type args struct {
		v1 VertexValue
		v2 VertexValue
	}
	tests := []struct {
		name   string
		fields mockFields
		args   args
		want   [][]VertexValue
	}{
		{
			name: "test get path without edges",
			fields: mockFields{
				verticesPaths: [][]VertexValue{{A}, {B}, {C}, {D}, {E}},
				bias:          Adds,
				clock:         nil,
			},
			args: args{
				A, E,
			},
			want: [][]VertexValue{},
		},
		{
			name: "test get path with edges",
			fields: mockFields{
				/*
					   A - B - D - E
						\ /  /
						 C -
				*/
				verticesPaths: [][]VertexValue{
					{A, B},
					{A, C},
					{B, C},
					{B, D},
					{C, D},
					{D, E},
				},
				bias:  Adds,
				clock: nil,
			},
			args: args{
				A, E,
			},
			want: [][]VertexValue{
				{A, B, A, C, D, E},
				{A, B, C, D, E},
				{A, B, D, E},
				{A, C, A, B, D, E},
				{A, C, B, D, E},
				{A, C, D, E},
			},
		},
		{
			name: "test get path with edges (removed edge)",
			fields: mockFields{
				/*
					   A - B - D - E
						\ /  /
						 C -
				*/
				verticesPaths: [][]VertexValue{
					{A, B},
					{A, C},
					{B, C},
					{B, D},
					{C, D},
					{D, E},
				},
				removedVertices: []VertexValue{C},
				bias:            Adds,
				clock:           nil,
			},
			args: args{
				A, E,
			},
			want: [][]VertexValue{
				{A, B, D, E},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			graph := NewMockGraph(tt.fields)

			if got := graph.GetPaths(tt.args.v1, tt.args.v2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LWWGraphImpl.GetPaths() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestLWWGraphImpl_RemoveEdgeByVertices(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")
	D := NewVertexValue("D")
	E := NewVertexValue("E")

	type args struct {
		v1 VertexValue
		v2 VertexValue
	}
	tests := []struct {
		name   string
		fields mockFields
		args   args
		want   map[VertexValue][]VertexValue
	}{
		{
			name: "test remove with edges",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{A, B},
					{A, C},
					{B, C},
					{B, D},
					{C, D},
					{D, E},
				},
				bias:  Adds,
				clock: nil,
			},
			args: args{
				A, B,
			},
			want: map[VertexValue][]VertexValue{
				A: {C},
				B: {C, D},
				C: {A, B, D},
				D: {B, C, E},
				E: {D},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			graph := NewMockGraph(tt.fields)
			graph.RemoveEdgeByVertices(tt.args.v1, tt.args.v2)
			if got := graph.GetAdjacencyVerticesList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LWWGraphImpl.GetAdjacencyVerticesList() = %v, want %v, diff: %v", got, tt.want, deep.Equal(got, tt.want))
			}
		})
	}
}

type mockOperation struct {
	value             VertexValue
	action            string
	duration          time.Duration
	connectedVertices []VertexValue
}

type mockGraphArgument struct {
	bias     Bias
	vertices []mockOperation
}

func NewMockGraphByOperations(args mockGraphArgument) (LWWGraph, testCkock) {

	clock := testCkock{}

	graph := NewLWWGraph(args.bias, &clock)

	var currentTimeline time.Duration = 0

	for i := 0; i < len(args.vertices); i++ {

		vertex := args.vertices[i].value
		vertexCurrentTimeAhead := args.vertices[i].duration - currentTimeline

		clock.AddDuration(vertexCurrentTimeAhead)
		currentTimeline += vertexCurrentTimeAhead

		switch args.vertices[i].action {
		case mockGraphAddAction:
			if args.vertices[i].connectedVertices != nil {
				for _, v := range args.vertices[i].connectedVertices {
					v1 := NewLWWVertex(vertex, graph.GetClock())
					v2 := NewLWWVertex(v, graph.GetClock())
					graph.AddEdge(v1, v2)
				}
			} else {
				graph.AddVertex(vertex)
			}
		case mockGraphRemoveAction:
			if args.vertices[i].connectedVertices != nil {
				for _, v := range args.vertices[i].connectedVertices {
					graph.RemoveEdgeByVertices(vertex, v)
				}
			} else {
				graph.RemoveVertex(vertex)
			}
		}
	}

	return graph, clock
}

func TestLWWGraphImpl_GetAdjacencyVerticesList(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")
	C := NewVertexValue("C")
	D := NewVertexValue("D")
	E := NewVertexValue("E")

	tests := []struct {
		name   string
		fields mockFields
		want   map[VertexValue][]VertexValue
	}{
		{
			name: "test get matrix with no edge",
			fields: mockFields{
				verticesPaths: [][]VertexValue{{A}, {B}, {C}, {D}, {E}},
				bias:          Adds,
				clock:         nil,
			},
			want: nil,
		},
		{
			name: "test get matrix with edges",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{A, B},
					{A, C},
					{B, C},
					{B, D},
					{C, D},
					{D, E},
				},
				bias:  Adds,
				clock: nil,
			},
			want: map[VertexValue][]VertexValue{
				A: {B, C},
				B: {A, C, D},
				C: {A, B, D},
				D: {B, C, E},
				E: {D},
			},
		},
		{
			name: "test get matrix with edges and non exist vertices",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{"a", "b", "c", "d"},
					{"a", "b", "a", "c", "d", "e"},
					{"a", "b", "c", "d", "e"},
					{"a", "c", "a", "b", "d", "e"},
					{"a", "c", "b", "d", "e"},
					{"a", "c", "d", "e"},
				},
				bias:  Adds,
				clock: nil,
			},
			want: map[VertexValue][]VertexValue{
				"a": {"b", "c"},
				"b": {"a", "c", "d"},
				"c": {"a", "b", "d"},
				"d": {"b", "c", "e"},
				"e": {"d"},
			},
		},
		{
			name: "test get matrix with edges and removed vertices",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{A, B, C, D},
					{A, B, A, C, D, E},
					{A, B, C, D, E},
					{A, C, A, B, D, E},
					{A, C, B, D, E},
					{A, C, D, E},
				},
				removedVertices: []VertexValue{A},
				bias:            Adds,
				clock:           nil,
			},
			want: map[VertexValue][]VertexValue{
				B: {C, D},
				C: {B, D},
				D: {B, C, E},
				E: {D},
			},
		},
		{
			name: "test get matrix with edges and removed vertices with adds bias and same timestamp",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{A, B, C, D},
					{A, B, A, C, D, E},
					{A, B, C, D, E},
					{A, C, A, B, D, E},
					{A, C, B, D, E},
					{A, C, D, E},
				},
				removedVertices: []VertexValue{A},
				bias:            Adds,
				clock:           &testCkock{},
			},
			want: map[VertexValue][]VertexValue{
				A: {B, C},
				B: {A, C, D},
				C: {A, B, D},
				D: {B, C, E},
				E: {D},
			},
		},
		{
			name: "test get matrix with edges and removed vertices with removed bias and same timestamp",
			fields: mockFields{
				verticesPaths: [][]VertexValue{
					{A, B, C, D},
					{A, B, A, C, D, E},
					{A, B, C, D, E},
					{A, C, A, B, D, E},
					{A, C, B, D, E},
					{A, C, D, E},
				},
				removedVertices: []VertexValue{A},
				bias:            Removal,
				clock:           &testCkock{},
			},
			want: map[VertexValue][]VertexValue{
				B: {C, D},
				C: {B, D},
				D: {B, C, E},
				E: {D},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			graph := NewMockGraph(tt.fields)
			if got := graph.GetAdjacencyVerticesList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LWWGraphImpl.GetAdjacencyVerticesList() = %v, want %v, diff: %v", got, tt.want, deep.Equal(got, tt.want))
			}
		})
	}
}

func TestLWWGraphImpl_Merge(t *testing.T) {

	A := NewVertexValue("A")
	B := NewVertexValue("B")

	type args struct {
		bias      Bias
		XVertices []mockOperation
		YVertices []mockOperation
	}

	tests := []struct {
		name        string
		description string
		args        args
		want        map[VertexValue][]VertexValue
	}{
		{
			name:        "merge for vertices",
			description: "update of the vertices timestamps is expected",
			args: args{
				bias: Adds,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 3 * time.Minute, nil},
					{B, mockGraphAddAction, 4 * time.Minute, nil},
				},
			},
			want: map[VertexValue][]VertexValue{
				A: {},
				B: {},
			},
		},
		{
			name: "merge with graph with larger timestamp of the removed vertex",
			args: args{
				bias: Adds,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{A, mockGraphRemoveAction, 2 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 3 * time.Minute, nil},
					{B, mockGraphAddAction, 4 * time.Minute, nil},
				},
			},
			want: map[VertexValue][]VertexValue{
				A: {},
				B: {},
			},
		},
		{
			name: "merge with graph with same timestamp of the removed vertex with adds bias",
			args: args{
				bias: Adds,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{A, mockGraphRemoveAction, 2 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 2 * time.Minute, nil},
					{B, mockGraphAddAction, 4 * time.Minute, nil},
				},
			},
			want: map[VertexValue][]VertexValue{
				A: {},
				B: {},
			},
		},
		{
			name: "merge with graph with same timestamp of the removed vertex with removal bias",
			args: args{
				bias: Removal,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{A, mockGraphRemoveAction, 2 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 2 * time.Minute, nil},
					{B, mockGraphAddAction, 4 * time.Minute, nil},
				},
			},
			want: map[VertexValue][]VertexValue{
				B: {},
			},
		},
		{
			name: "check if merge affect the edges",
			args: args{
				bias: Adds,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, []VertexValue{A}},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
				},
			},
			want: map[VertexValue][]VertexValue{
				A: {B},
				B: {A},
			},
		},
		{
			name: "check if merge with removed vertex graph affect the edges",
			args: args{
				bias: Adds,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, []VertexValue{A}},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
					{A, mockGraphRemoveAction, 3 * time.Minute, nil},
				},
			},
			want: map[VertexValue][]VertexValue{
				B: {},
			},
		},
		{
			name: "check if merge with removed vertex(with same timestamp) graph affect the edges (adds bias)",
			args: args{
				bias: Adds,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, []VertexValue{A}},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
					{A, mockGraphRemoveAction, 2 * time.Minute, nil},
				},
			},
			want: map[VertexValue][]VertexValue{
				B: {},
			},
		},
		{
			name: "check if merge with removed vertex(with same timestamp) graph affect the edges (removal bias)",
			args: args{
				bias: Removal,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, []VertexValue{A}},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
					{A, mockGraphRemoveAction, 2 * time.Minute, nil},
				},
			},
			want: map[VertexValue][]VertexValue{
				B: {},
			},
		},
		{
			name: "check with add edge",
			args: args{
				bias: Adds,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
					{A, mockGraphAddAction, 3 * time.Minute, []VertexValue{B}},
				},
			},
			want: map[VertexValue][]VertexValue{
				A: {B},
				B: {A},
			},
		},
		{
			name: "check with removed edge",
			args: args{
				bias: Adds,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, []VertexValue{A}},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, []VertexValue{A}},
					{A, mockGraphRemoveAction, 3 * time.Minute, []VertexValue{B}},
				},
			},
			want: map[VertexValue][]VertexValue{
				A: {},
				B: {},
			},
		},
		{
			name: "check with removed edge with add bias",
			args: args{
				bias: Adds,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, []VertexValue{A}},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, nil},
					{A, mockGraphRemoveAction, 2 * time.Minute, []VertexValue{B}},
				},
			},
			want: map[VertexValue][]VertexValue{
				A: {B},
				B: {A},
			},
		},
		{
			name: "check with removed edge with removal bias",
			args: args{
				bias: Removal,
				XVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, []VertexValue{A}},
				},
				YVertices: []mockOperation{
					{A, mockGraphAddAction, 1 * time.Minute, nil},
					{B, mockGraphAddAction, 2 * time.Minute, []VertexValue{A}},
					{A, mockGraphRemoveAction, 2 * time.Minute, []VertexValue{B}},
				},
			},
			want: map[VertexValue][]VertexValue{
				A: {},
				B: {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			xGraph, xClock := NewMockGraphByOperations(mockGraphArgument{tt.args.bias, tt.args.XVertices})
			yGraph, yClock := NewMockGraphByOperations(mockGraphArgument{tt.args.bias, tt.args.YVertices})

			xClock.SyncWith(&yClock)

			xGraph.Merge(yGraph)

			if got := xGraph.GetAdjacencyVerticesList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LWWGraphImpl.GetAdjacencyVerticesList() = %v, want %v, diff: %v", got, tt.want, deep.Equal(got, tt.want))
			}
		})
	}

}
