package undirect

type VertexValue string

func (value VertexValue) IsEqual(to VertexValue) bool {
	return value == to
}

func NewVertexValue(v string) VertexValue {
	return VertexValue(v)
}

type LWWVertex interface {
	GetValue() VertexValue
	GetTimestamp() int64
	SetTimestamp(int64) int64
}

type LWWVertexImpl struct {
	value     VertexValue
	timestamp int64
}

func NewLWWVertex(value VertexValue, clock Clock) LWWVertex {
	return &LWWVertexImpl{
		value:     value,
		timestamp: clock.Now().UnixNano(),
	}
}

func (vertex *LWWVertexImpl) GetValue() VertexValue {
	return vertex.value
}

func (vertex *LWWVertexImpl) GetTimestamp() int64 {
	return vertex.timestamp
}

func (vertex *LWWVertexImpl) SetTimestamp(t int64) int64 {
	vertex.timestamp = t
	return vertex.timestamp
}
