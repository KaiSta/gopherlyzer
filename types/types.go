package types

import (
	"fmt"
)

var loopCounter map[R]int

type R interface {
	String() string
	Nullable() bool
	IsPhi() bool
	ConcPart() R
	SeqPart() R
	Deriv(r R) R
	PDeriv(r R) []R
	NextSym() []R
	Reduce() R
	ContainsPhi() bool
	Contains(r R) bool
}

type Sym rune
type Star [1]R
type Alt [2]R
type Alt2 [2]R
type SELECTGROUP []R
type Seq [2]R
type Eps int
type Phi int
type Fork [1]R
type Skip []R

func init() {
	loopCounter = make(map[R]int)
}

func (s Sym) Nullable() bool  { return false }
func (s Star) Nullable() bool { return true }
func (a Alt) Nullable() bool  { return a[0].Nullable() || a[1].Nullable() }
func (a Alt2) Nullable() bool  { return a[0].Nullable() || a[1].Nullable() }
func (s Seq) Nullable() bool  { return s[0].Nullable() && s[1].Nullable() }
func (e Eps) Nullable() bool  { return true }
func (p Phi) Nullable() bool  { return false }
func (f Fork) Nullable() bool { return f[0].Nullable() }
func (f Skip) Nullable() bool { return true }
func (f SELECTGROUP) Nullable() bool { return true }

func (s Sym) IsPhi() bool  { return false }
func (s Star) IsPhi() bool { return false }
func (a Alt) IsPhi() bool  { return a[0].IsPhi() && a[1].IsPhi() }
func (a Alt2) IsPhi() bool  { return a[0].IsPhi() && a[1].IsPhi() }
func (s Seq) IsPhi() bool  { return s[0].IsPhi() || s[1].IsPhi() }
func (e Eps) IsPhi() bool  { return false }
func (p Phi) IsPhi() bool  { return true }
func (f Fork) IsPhi() bool { return f[0].IsPhi() }
func (f Skip) IsPhi() bool { return false }
func (f SELECTGROUP) IsPhi() bool { return false }

func (s Sym) ContainsPhi() bool  { return false }
func (s Star) ContainsPhi() bool { return s[0].ContainsPhi() }
func (a Alt) ContainsPhi() bool  { return a[0].ContainsPhi() || a[1].ContainsPhi() }
func (a Alt2) ContainsPhi() bool  { return a[0].ContainsPhi() || a[1].ContainsPhi() }
func (s Seq) ContainsPhi() bool  { return s[0].ContainsPhi() || s[1].ContainsPhi() }
func (e Eps) ContainsPhi() bool  { return false }
func (p Phi) ContainsPhi() bool  { return true }
func (f Fork) ContainsPhi() bool { return f[0].ContainsPhi() }
func (f Skip) ContainsPhi() bool { return false }
func (f SELECTGROUP) ContainsPhi() bool { return false }

func (s Sym) ConcPart() R  { return Phi(1) }
func (s Star) ConcPart() R { return Star{s[0].ConcPart()} }
func (a Alt) ConcPart() R  { return Alt{a[0].ConcPart(), a[1].ConcPart()} }
func (a Alt2) ConcPart() R  { return Alt2{a[0].ConcPart(), a[1].ConcPart()} }
func (s Seq) ConcPart() R  { return Seq{s[0].ConcPart(), s[1].ConcPart()} }
func (e Eps) ConcPart() R  { return e }
func (p Phi) ConcPart() R  { return p }
func (f Fork) ConcPart() R { return f }
func (f Skip) ConcPart() R { return f }
func (f SELECTGROUP) ConcPart() R { return f }

func (s Sym) SeqPart() R  { return s }
func (s Star) SeqPart() R { return Seq{Star{s[0].ConcPart()}, Seq{s[0].SeqPart(), s}} }
func (a Alt) SeqPart() R  { return Alt{a[0].SeqPart(), a[1].SeqPart()} }
func (a Alt2) SeqPart() R  { return Alt{a[0].SeqPart(), a[1].SeqPart()} }
func (s Seq) SeqPart() R {
	return Alt{Seq{s[0].SeqPart(), s[1]},
		Seq{s[0].ConcPart(), s[1].SeqPart()}}
}
func (e Eps) SeqPart() R  { return Phi(1) }
func (p Phi) SeqPart() R  { return p }
func (f Fork) SeqPart() R { return Phi(1) }
func (f Skip) SeqPart() R { return f }
func (f SELECTGROUP) SeqPart() R { return f }

func (s Skip) Deriv(r R) R {
	return s
}
func (s SELECTGROUP) Deriv(r R) R {
	return Phi(1)
}
func (s Sym) Deriv(r R) R {
	if r == s {
		return Eps(1)
	}
	return Phi(1)
}
func (s Star) Deriv(r R) R {
	tmp := s[0].Deriv(r)
	if tmp == Eps(1) {
		return s
	}
	return Seq{tmp, s}
}
func (a Alt) Deriv(r R) R {
	if a[0].Contains(r) && a[1].Contains(r) {
		return Alt{a[0].Deriv(r), a[1].Deriv(r)}
	} else if a[0].Contains(r) {
		return a[0].Deriv(r)
	} else if a[1].Contains(r) {
		return a[1].Deriv(r)
	}

	return Phi(1)
}
func (a Alt2) Deriv(r R) R {
	if a[0].Contains(r) && a[1].Contains(r) {
		return Alt2{a[0].Deriv(r), a[1].Deriv(r)}
	} else if a[0].Contains(r) {
		return a[0].Deriv(r)
	} else if a[1].Contains(r) {
		return a[1].Deriv(r)
	}

	return Phi(1)
}
func (s Seq) Deriv(r R) R {
	//  fmt.Println(s, r)
	if s[0].Contains(r) {
		if s[0].Nullable() {
			if s[1].Contains(r) {
				return Alt{Seq{s[0].Deriv(r), s[1]}, s[1].Deriv(r)}
			} else {
				return Seq{s[0].Deriv(r), s[1]}
			}
		}
		sn := Seq{s[0].Deriv(r), s[1]}
		if sn[0] == Eps(1) {
			return s[1]
		}
		return sn
	}

	if s[0].Nullable() {
		rs := s[1].Deriv(r)
		return rs
	} else {
		return Seq{Phi(1), s[1].Deriv(r)}
	}

	//return Alt{Seq{s[0].Deriv(r), s[1]}, Seq{s[0].concPart(), s[1].Deriv(r)}}
}
func (e Eps) Deriv(r R) R {
	return Phi(1)
}
func (p Phi) Deriv(r R) R {
	return p
}
func (f Fork) Deriv(r R) R {
	tmp := f[0].Deriv(r)
	if tmp == Eps(1) {
		return Eps(1)
	}
	return Fork{tmp}
}

func (s Skip) PDeriv(r R) []R {
	return []R{s}
}
func (s SELECTGROUP) PDeriv(r R) []R {
	return []R{}
}

func (s Sym) PDeriv(r R) []R {
	if r == s {
		return []R{Eps(1)}
	}
	return []R{}
}
func (s Star) PDeriv(r R) (ret []R) {
	ret = append(ret, s[0].PDeriv(r)...)
	for i, el := range ret {
		ret[i] = Seq{el, s}
	}
	return
}
func (a Alt) PDeriv(r R) (ret []R) {
	ret = append(a[0].PDeriv(r), a[1].PDeriv(r)...)
	return
}
func (a Alt2) PDeriv(r R) (ret []R) {
	ret = append(a[0].PDeriv(r), a[1].PDeriv(r)...)
	return
}
func (s Seq) PDeriv(r R) (ret []R) {
	ret = append(ret, s[0].PDeriv(r)...)
	for i, el := range ret {
		ret[i] = Seq{el, s[1]}
	}

	for _, el := range s[1].PDeriv(r) {
		ret = append(ret, Seq{s[0].ConcPart(), el})
	}
	return
}
func (e Eps) PDeriv(r R) []R { return []R{} }
func (p Phi) PDeriv(r R) []R { return []R{} }
func (f Fork) PDeriv(r R) (ret []R) {
	ret = append(ret, f[0].PDeriv(r)...)
	for i, el := range ret {
		ret[i] = Fork{el}
	}
	return
}

func (s Skip) NextSym() []R {
	return []R{s}
}
func (s SELECTGROUP) NextSym() []R {
	return []R{}
}
func (s Sym) NextSym() []R { return []R{s} }
func (s Star) NextSym() []R {
	return s[0].NextSym()
	//return append(s[0].NextSym(), Skip(1))
}
func (a Alt) NextSym() []R { return append(a[0].NextSym(), a[1].NextSym()...) }
func (a Alt2) NextSym() []R { 
    x := a[0].NextSym()
    y := a[1].NextSym()
    var tmp SELECTGROUP = x
    tmp = append(tmp, y...)
    y = append(y, tmp)
    return append(x, y...)
  //  return append(a[0].NextSym(), a[1].NextSym()...) 
}
func (s Seq) NextSym() []R {
	if s[0].ConcPart().IsPhi() {
		return s[0].NextSym()
	}
	return append(s[0].NextSym(), s[1].NextSym()...)
}
func (e Eps) NextSym() []R  { return []R{} }
func (p Phi) NextSym() []R  { return []R{} }
func (f Fork) NextSym() []R { return f[0].NextSym() }

func (s Skip) Reduce() R { return s }
func (s SELECTGROUP) Reduce() R { return s }
func (s Sym) Reduce() R  { return s }
func (e Eps) Reduce() R  { return e }
func (p Phi) Reduce() R  { return p }
func (s Star) Reduce() R {
	x := s[0].Reduce()

	if x != Eps(1) {
		return Star{x}
	}
	return Eps(1)
}
func (a Alt) Reduce() R {
	l := a[0].Reduce()
	r := a[1].Reduce()
	if l == Eps(1) && r == Eps(1) {
		return Eps(1)
	} else if r == Eps(1) {
		return l
	} else if l == Eps(1) {
		return r
	} else {
		return Alt{l, r}
	}
}
func (a Alt2) Reduce() R {
	l := a[0].Reduce()
	r := a[1].Reduce()
	if l == Eps(1) && r == Eps(1) {
		return Eps(1)
	} else if r == Eps(1) {
		return l
	} else if l == Eps(1) {
		return r
	} else {
		return Alt2{l, r}
	}
}
func (s Seq) Reduce() R {
	l := s[0].Reduce()
	r := s[1].Reduce()

	if l == Eps(1) && r == Eps(1) {
		return Eps(1)
	} else if r == Eps(1) {
		return l
	} else if l == Eps(1) {
		return r
	} else {
		return Alt{l, r}
	}
}
func (f Fork) Reduce() R {
	x := f[0].Reduce()
	if x != Eps(1) {
		return Fork{x}
	}
	return Eps(1)
}

func (s Skip) Contains(e R) bool {
	return false
}
func (s SELECTGROUP) Contains(e R) bool {
	return false
}

func (s Sym) Contains(e R) bool {
	switch x := e.(type) {
	case Sym:
		return x == s
	default:
		return x.Contains(s)
	}
}
func (s Star) Contains(e R) bool {
	return s[0].Contains(e)
}
func (a Alt) Contains(e R) bool {
	return a[0].Contains(e) || a[1].Contains(e)
}
func (a Alt2) Contains(e R) bool {
	return a[0].Contains(e) || a[1].Contains(e)
}
func (s Seq) Contains(e R) bool {
	if s[0].Nullable() {
		return s[0].Contains(e) || s[1].Contains(e)
	}
	return s[0].Contains(e)
}
func (a Eps) Contains(e R) bool {
	switch e.(type) {
	case Eps:
		return true
	default:
		return false
	}
}
func (p Phi) Contains(e R) bool {
	return false
}
func (f Fork) Contains(e R) bool {
	return f[0].Contains(e)
}

func (s Skip) String() string {
	return "$SKIP$"
}
func (s SELECTGROUP) String() string {
	return "$SELECTGROUP$"
}
func (s Sym) String() string {
	return (string)(s)
}
func (s Star) String() string {
	return "(" + s[0].String() + ")*"
}
func (a Alt) String() string {
	return "(" + a[0].String() + "+" + a[1].String() + ")"
}
func (a Alt2) String() string {
	return "(" + a[0].String() + "XOR" + a[1].String() + ")"
}
func (e Eps) String() string {
	return "\u03B5"
}
func (s Seq) String() string {
	return "(" + s[0].String() + "." + s[1].String() + ")"
}
func (p Phi) String() string {
	return "\u03C6"
}
func (f Fork) String() string {
	return fmt.Sprintf("Fork(%v)", f[0].String())
}

func Derivs(r R, s string) R {
	for _, c := range s {
		r = r.Deriv(Sym(c))
	}
	return r
}
