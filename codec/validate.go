package codec

// Implemented by types that support structural validation.
//
// Validation is always a separate step from decoding. This allows users to
// decode data that may be structurally invalid, and defer validation until
// they have the full object graph.
type Validatable interface {

	// Validates the receiver's fields and invariants.
	//
	// Returns nil when the value is structurally valid. Implementations
	// should check required fields, mutual exclusion constraints, and
	// value ranges. For composite types, Validate should recurse into
	// nested [Validatable] children.
	Validate() error
}
