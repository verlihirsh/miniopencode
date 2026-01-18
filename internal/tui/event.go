package tui

type ChunkKind string

const (
	ChunkThinking ChunkKind = "thinking"
	ChunkTool     ChunkKind = "tool"
	ChunkAnswer   ChunkKind = "answer"
	ChunkRaw      ChunkKind = "raw"
	ChunkSkip     ChunkKind = "skip"
	ChunkMeta     ChunkKind = "meta"
)

type Chunk struct {
	Kind      ChunkKind
	Text      string
	PartID    string
	MessageID string
	Complete  bool
}
