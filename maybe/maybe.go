package maybe


/*
module Maybe exposing (Maybe(Just,Nothing), andThen, map, withDefault, oneOf)

{-| This library fills a bunch of important niches in Elm. A `Maybe` can help
you with optional arguments, error handling, and records with optional fields.

# Definition
@docs Maybe

# Common Helpers
@docs map, withDefault, oneOf

# Chaining Maybes
@docs andThen

-}
*/

type Maybe[T any] interface {
	Match() Matcher[T]
	WithDefault(T) T
	Map(func(T)T) Maybe[T]
}


type maybe[T any] struct {
	value T
	tag bool
}

func Just[T any](x T) Maybe[T] {
	return maybe[T]{value: x, tag: true}
}

func Nothing[T any]() Maybe[T] {
	return maybe[T]{tag: false}
}

func (m maybe[T]) Match() Matcher[T] {
	return matcher[T]{m: m}
}

func (m maybe[T]) WithDefault(def T) T {
	if m.tag {
		return m.value
	}
	return def
}

func (m maybe[T]) Map(f func(T) T) Maybe[T] {
	if m.tag {
		return Just(f(m.value))
	}
	return m
}

func AndThen[T, S any](f func(T) Maybe[S], x Maybe[T]) Maybe[S] {
	var v T
	switch m := x.Match(); m {
	case m.Just(&v):
		return f(v)
	case m.Nothing():
	}
	return Nothing[S]()
}

func  Map[T any](f func(T) T, x Maybe[T]) Maybe[T] {
	var v T
	switch m := x.Match(); m {
	case m.Just(&v):
		v = f(v)
		return Just[T](v)
	case m.Nothing():
	}
	return x
}

// --- Matching --------------------------------------------------------------

type Matcher[T any] interface {
	Just(*T) Matcher[T]
	Nothing() Matcher[T]
}

type matcher[T any] struct {
	m maybe[T]
}

func (mm matcher[T]) Just(v *T) Matcher[T] {
	if mm.m.tag {
		*v = mm.m.value
		return mm
	}
	return nil
}

func (mm matcher[T]) Nothing() Matcher[T] {
	if !mm.m.tag {
		return mm
	}
	return nil
}
