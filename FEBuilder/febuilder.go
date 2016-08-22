package febuilder

import (
	"fmt"

	"../algorithm"
	"../simplifier"
	"../types"
)

type state struct {
	alpha         rune
	signMap       map[string]rune
	partnerMap    map[int]*algorithm.HashSet
	closerMap     map[types.R][]types.R
	ReaderMap     map[types.R]int
	expressionMap map[string]types.R
	parseGraph    map[string]simplify.Callgraph2
}

func BuildExpression(parseGraph map[string]simplify.Callgraph2) (types.R, map[types.R]types.R, []types.R, map[types.R][]types.R, map[types.R]int) {
	status := state{'a', make(map[string]rune), make(map[int]*algorithm.HashSet), make(map[types.R][]types.R),
		make(map[types.R]int), make(map[string]types.R), parseGraph}

	threads := make([]types.R, 0)

	for k, v := range status.parseGraph {
		rs := make([]types.R, 0)

		for _, x := range v.ChanOps {
			innerRs := make([]types.R, 0)

			for _, y := range x.Ops {
				if y.Type == "If" {
					innerRs = append(innerRs, status.handleIf2(y))
				} else if y.Type == "Select" {
					//fmt.Println("FEBUILDERINSEL", y)
					innerRs = append(innerRs, status.handleSelect(y))
				} else {
					innerRs = append(innerRs, status.handleOp(y)...)
				}
			}

			if x.Type == "If" {
				ifBody := status.makeSeqs(innerRs)
				if x.Else != nil {
					elseRs := make([]types.R, 0)
					for _, y := range x.Else {
						elseRs = append(elseRs, status.handleOp(y)...)
					}
					elseBody := status.makeSeqs(elseRs)
					innerRs = []types.R{types.Alt{ifBody, elseBody}}
				} else {
					innerRs = []types.R{types.Alt{ifBody, types.Eps(1)}}
				}
			} else if x.Type == "For" {
				innerRs = []types.R{types.Star{status.makeSeqs(innerRs)}}
			} else if x.Type == "Select" {
				innerRs = []types.R{status.handleSelect(x)}
			}
			rs = append(rs, innerRs...)
		}
		status.expressionMap[k] = status.makeSeqs(rs)
	}

	for k, _ := range status.expressionMap {
		//  if tmp := status.parseGraph[k]; tmp.NrRoots > 0 || k  == "main" {
		if status.expressionMap[k] != types.Eps(1) {
			threads = append(threads, status.expressionMap[k])
		}

		//    }
	}

	complete := make([]types.R, 0)
	for k, v := range status.expressionMap {
		if tmp := status.parseGraph[k]; tmp.NrRoots > 0 && k != "main" {
			v = types.Fork{v}
			status.expressionMap[k] = v
		}
		if status.expressionMap[k] != types.Eps(1) {
			complete = append(complete, status.expressionMap[k])
		}
	}

	//sort the expression list so main is the last expression
	for i := range complete {
		switch complete[i].(type) {
		case types.Fork:
		default:
			if complete[i] != types.Eps(1) {
				tmp := complete[i]
				complete[i] = complete[len(complete)-1]
				complete[len(complete)-1] = tmp
			}
		}
	}
	completeFE := status.makeSeqs(complete)

	//build partnermap
	pMap := make(map[types.R]types.R)
	for _, v := range status.partnerMap {
		for j := range v.Iterate() {
			for k := range v.Iterate() {
				if k.(rune) != j.(rune) {
					pMap[types.Sym(k.(rune))] = types.Sym(j.(rune))
					pMap[types.Sym(j.(rune))] = types.Sym(k.(rune))
				}
			}
		}
	}
	return completeFE, pMap, threads, status.closerMap, status.ReaderMap
}

func (s *state) makeSeqs(rs []types.R) types.R {
	if len(rs) == 1 {
		return rs[0]
	} else if len(rs) > 0 {
		seqs1 := rs
		curr := rs
		for len(seqs1) != 1 {
			curr = seqs1
			seqs1 = make([]types.R, 0)
			for i := 0; i < len(curr); i += 2 {
				if i+1 < len(curr) {
					seqs1 = append(seqs1, types.Seq{curr[i], curr[i+1]})
				} else {
					seqs1 = append(seqs1, curr[i])
				}
			}
		}

		return seqs1[0]
	}
	return types.Eps(1)
}

func (s *state) makeAlts(rs []types.R) types.R {
	if len(rs) == 1 {
		return rs[0]
	} else if len(rs) > 0 {
		seqs1 := rs
		curr := rs
		for len(seqs1) != 1 {
			curr = seqs1
			seqs1 = make([]types.R, 0)
			for i := 0; i < len(curr); i += 2 {
				if i+1 < len(curr) {
					seqs1 = append(seqs1, types.Alt{curr[i], curr[i+1]})
				} else {
					seqs1 = append(seqs1, curr[i])
				}
			}
		}

		return seqs1[0]
	}
	return types.Eps(1)
}

func (s *state) makeAlts2(rs []types.R) types.R {
	if len(rs) == 1 {
		return rs[0]
	} else if len(rs) > 0 {
		seqs1 := rs
		curr := rs
		for len(seqs1) != 1 {
			curr = seqs1
			seqs1 = make([]types.R, 0)
			for i := 0; i < len(curr); i += 2 {
				if i+1 < len(curr) {
					seqs1 = append(seqs1, types.Alt2{curr[i], curr[i+1]})
				} else {
					seqs1 = append(seqs1, curr[i])
				}
			}
		}

		return seqs1[0]
	}
	return types.Eps(1)
}

func (st *state) handleOp(op simplify.Operation) (innerRs []types.R) {
	s := fmt.Sprintf("%v%v", op.Pos, op.Op)

	z, ok := st.signMap[s]
	if ok {
		innerRs = append(innerRs, types.Sym(z))
		tmp := st.partnerMap[op.Pos]
		tmp.Insert(z)
		st.partnerMap[op.Pos] = tmp
	} else {
		//fmt.Println(s, string(st.alpha), op)
		st.signMap[s] = st.alpha
		innerRs = append(innerRs, types.Sym(st.alpha))
		tmp, ok := st.partnerMap[op.Pos]
		if op.Op != "#" {
			if !ok {
				tmp = algorithm.NewHashSet()
			}
			tmp.Insert(st.alpha)
			st.partnerMap[op.Pos] = tmp

			if op.Op == "?" {
				st.ReaderMap[types.Sym(st.alpha)] = 1
			}
		} else {
			if ok {
				localAlpha := types.Sym(st.alpha)
				xtmp, ok2 := st.closerMap[localAlpha]
				if !ok2 {
					st.closerMap[localAlpha] = make([]types.R, 0)
				}
				for q := range tmp.Iterate() {
					xtmp = append(xtmp, types.Sym(q.(rune)))
				}
				st.closerMap[localAlpha] = xtmp
			}
		}

		st.alpha = st.alpha + 1
	}
	return
}

func (st *state) handleIf2(op simplify.Operation) types.R {
	ifrs := make([]types.R, 0)
	for _, z := range op.Ops {
		ifrs = append(ifrs, st.handleOp(z)...)
	}
	ifBody := st.makeSeqs(ifrs)

	elsers := make([]types.R, 0)
	for _, z := range op.Else {
		elsers = append(elsers, st.handleOp(z)...)
	}
	elseBody := st.makeSeqs(elsers)

	return types.Alt{ifBody, elseBody}
}

func (st *state) handleSelect(op simplify.Operation) types.R {
	selectCases := make([]types.R, 0)
	for _, y := range op.Ops {
		//fmt.Println("!Y!", y)
		if len(y.Ops) > 0 {
			if len(y.Ops) == 1 {
				tmp := st.handleOp(y)

				//fmt.Println(tmp, y.Ops[0])
				for i := range tmp {
					tmp[i] = types.Seq{tmp[i], st.handleOp(y.Ops[0])[0]}
				}

				selectCases = append(selectCases, tmp...)
			} else {
				rs := make([]types.R, 0)
				for i := range y.Ops {
					rs = append(rs, st.handleOp(y.Ops[i])...)
				}
				selectCases = append(selectCases, st.makeSeqs(rs))
			}

		} else {
			selectCases = append(selectCases, st.handleOp(y)...)
		}

	}
	return st.makeAlts2(selectCases)
}

func (st *state) handleIf(op simplify.Operation) types.R {
	ifrs := make([]types.R, 0)
	for _, z := range op.Ops {
		s := fmt.Sprintf("%v%v", z.Pos, z.Op)
		h, ok := st.signMap[s]
		if ok {
			ifrs = append(ifrs, types.Sym(h))
			tmp := st.partnerMap[z.Pos]
			tmp.Insert(h)
			st.partnerMap[z.Pos] = tmp
		} else {
			st.signMap[s] = st.alpha
			ifrs = append(ifrs, types.Sym(st.alpha))
			tmp, ok := st.partnerMap[z.Pos]
			if !ok {
				tmp = algorithm.NewHashSet()
			}
			tmp.Insert(st.alpha)
			st.partnerMap[z.Pos] = tmp
			st.alpha = st.alpha + 1
		}
	}
	ifBody := st.makeSeqs(ifrs)

	elsers := make([]types.R, 0)
	for _, z := range op.Else {
		s := fmt.Sprintf("%v%v", z.Pos, z.Op)
		h, ok := st.signMap[s]
		if ok {
			elsers = append(elsers, types.Sym(h))
			tmp := st.partnerMap[z.Pos]
			tmp.Insert(h)
			st.partnerMap[z.Pos] = tmp
		} else {
			st.signMap[s] = st.alpha
			// fmt.Println(s, "==", string(st.alpha))
			elsers = append(elsers, types.Sym(st.alpha))
			tmp, ok := st.partnerMap[z.Pos]
			if !ok {
				tmp = algorithm.NewHashSet()
			}
			tmp.Insert(st.alpha)
			st.partnerMap[z.Pos] = tmp
			st.alpha = st.alpha + 1
		}
	}
	elseBody := st.makeSeqs(elsers)
	return types.Alt{ifBody, elseBody}
}
