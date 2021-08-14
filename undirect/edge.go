package undirect

type LWWEdge interface {
	GetVertices() (vertices []LWWVertex)
	GetTimestamp() int64
	SetTimestamp(int64) int64
}

type LWWEdgeImpl struct {
	vertices  *[]LWWVertex
	timestamp int64
}

func NewLWWEdgeImpl(vertices []LWWVertex, clock Clock) LWWEdge {
	return &LWWEdgeImpl{
		vertices:  &vertices,
		timestamp: clock.Now().UnixNano(),
	}
}

func (edge *LWWEdgeImpl) GetVertices() (vertices []LWWVertex) {
	return *edge.vertices
}

func (edge *LWWEdgeImpl) GetTimestamp() int64 {
	return edge.timestamp
}

func (edge *LWWEdgeImpl) SetTimestamp(t int64) int64 {
	edge.timestamp = t
	return edge.timestamp
}
