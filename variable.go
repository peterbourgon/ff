package ff

// Variable is the internal representation of an individual variable, wrapping a
// Value and other per-variable metadata.
type Variable struct {
	value Value
	name  string
	def   string
	usage string
	keys  map[SourceType]string
}

// VariableOption sets a per-variable option.
type VariableOption func(*Variable)

// Env instructs the package to take a value from the environment with the given
// env var name.
func Env(name string) VariableOption {
	return func(v *Variable) { v.keys[SourceTypeEnv] = name }
}

// JSON instructs the variable to take a value from a JSON config object with
// the given key.
func JSON(key string) VariableOption {
	return func(v *Variable) { v.keys[SourceTypeJSON] = key }
}
