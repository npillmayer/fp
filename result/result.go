package result

/*
{-| A `Result` is the result of a computation that may fail. This is a great
way to manage errors in Elm.

# Type and Constructors
@docs Result

# Mapping
@docs map, map2, map3, map4, map5

# Chaining
@docs andThen

# Handling Errors
@docs withDefault, toMaybe, fromMaybe, mapError
-}
*/

type Result[T any] interface {
	Match() Matcher[T]
}

type result[T any] struct {
	value T
	err error
}

func Ok[T any](x T) Result[T] {
	return result[T]{value: x}
}

func Err[T any](err error) Result[T] {
	return result[T]{err: err}
}

func (r result[T]) Match() Matcher[T] {
	return matcher[T]{r: r}
}

// --- Matching --------------------------------------------------------------

type Matcher[T any] interface {
	Ok(*T) Matcher[T]
	Err(*error) Matcher[T]
}

type matcher[T any] struct {
	r result[T]
}

func (rm matcher[T]) Ok(v *T) Matcher[T] {
	if rm.r.err == nil {
		*v = rm.r.value
		return rm
	}
	return nil
}

func (rm matcher[T]) Err(err *error) Matcher[T] {
	if rm.r.err != nil {
		*err = rm.r.err
		return rm
	}
	return nil
}
