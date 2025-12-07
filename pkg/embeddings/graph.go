package embeddings

// Graph exposes inbound/outbound links and optional tags for ranking.
type Graph interface {
	Outgoing(id NoteID) []NoteID
	Incoming(id NoteID) []NoteID
	Tags(id NoteID) []string
}
