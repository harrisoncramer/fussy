package unexported

import "go/types"

// surface is the set of named types a caller outside the package can still write down, gathered
// by walking outwards from the exported declarations the sweep is leaving in place.
type surface struct {
	found map[*types.TypeName]bool
	seen  map[types.Type]bool
}

// reachableFrom returns the named types the given declarations put in front of a caller, which
// are the types unexporting would leave an exported signature unable to name.
func reachableFrom(objects []types.Object) map[*types.TypeName]bool {
	s := &surface{found: map[*types.TypeName]bool{}, seen: map[types.Type]bool{}}
	for _, obj := range objects {
		if name, ok := obj.(*types.TypeName); ok {
			s.addNamed(name)
			continue
		}
		s.walk(obj.Type())
	}

	return s.found
}

// addNamed records one named type and descends into as much of it as a caller can see, which is
// its exported methods and either its exported fields or, for anything that is not a struct, the
// whole shape it stands for.
func (s *surface) addNamed(obj *types.TypeName) {
	if obj == nil || s.found[obj] {
		return
	}
	s.found[obj] = true

	named, ok := types.Unalias(obj.Type()).(*types.Named)
	if !ok {
		s.walk(obj.Type())
		return
	}

	for i := range named.NumMethods() {
		if named.Method(i).Exported() {
			s.walk(named.Method(i).Signature())
		}
	}

	structure, ok := named.Underlying().(*types.Struct)
	if !ok {
		s.walk(named.Underlying())
		return
	}

	for i := range structure.NumFields() {
		field := structure.Field(i)
		if field.Exported() || field.Embedded() {
			s.walk(field.Type())
		}
	}
}

// walk descends one type, recording the named types it is built out of, and refuses to descend
// into one it has already been through, since a type parameter constrained by the type it is a
// parameter of walks in a circle forever otherwise.
func (s *surface) walk(typ types.Type) {
	if typ == nil || s.seen[typ] {
		return
	}
	s.seen[typ] = true

	switch shape := typ.(type) {
	case *types.Alias:
		s.addNamed(shape.Obj())
		s.walk(types.Unalias(shape))
	case *types.Named:
		s.addNamed(shape.Obj())
		for i := range shape.TypeArgs().Len() {
			s.walk(shape.TypeArgs().At(i))
		}
	case *types.Pointer:
		s.walk(shape.Elem())
	case *types.Slice:
		s.walk(shape.Elem())
	case *types.Array:
		s.walk(shape.Elem())
	case *types.Chan:
		s.walk(shape.Elem())
	case *types.Map:
		s.walk(shape.Key())
		s.walk(shape.Elem())
	case *types.Signature:
		s.walkSignature(shape)
	case *types.Struct:
		for i := range shape.NumFields() {
			s.walk(shape.Field(i).Type())
		}
	case *types.Interface:
		for i := range shape.NumMethods() {
			s.walk(shape.Method(i).Signature())
		}
		for i := range shape.NumEmbeddeds() {
			s.walk(shape.EmbeddedType(i))
		}
	case *types.TypeParam:
		s.walk(shape.Constraint())
	case *types.Union:
		for i := range shape.Len() {
			s.walk(shape.Term(i).Type())
		}
	}
}

func (s *surface) walkSignature(signature *types.Signature) {
	for _, tuple := range []*types.Tuple{signature.Params(), signature.Results()} {
		for i := range tuple.Len() {
			s.walk(tuple.At(i).Type())
		}
	}

	if params := signature.TypeParams(); params != nil {
		for i := range params.Len() {
			s.walk(params.At(i).Constraint())
		}
	}
}
