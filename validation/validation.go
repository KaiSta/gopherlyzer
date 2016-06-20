package validation

import (
	"../types"
	//"bufio"
	"fmt"
	//"os"
)

type machine struct {
	Rs    []types.R
	Trace string
	Succ  bool
	Stop  bool
	Abort bool
}

type nextStr struct {
	ThreadId int
	Rs       []types.R
}

func Rscontains(c types.R, rs []types.R) bool {
	for _, r := range rs {
		if r == c {
			return true
		}
	}
	return false
}

type machine2 struct {
	Threads  []types.R
	PMap     map[types.R]types.R
	StateMap map[types.R]int
	Trace    string
	Succ     bool
	Stop     bool
	Abort    bool
}

type SyncPoint struct {
	T1Id   int
	T2Id   int
	Symbol types.R
}

func (m *machine2) clone() machine2 {
	new := machine2{PMap: m.PMap, Threads: make([]types.R, 0), Trace: m.Trace, StateMap: make(map[types.R]int)}
	for _, t := range m.Threads {
		new.Threads = append(new.Threads, t)
	}
	for k, v := range m.StateMap {
		new.StateMap[k] = v
	}
	return new
}

func (mach *machine2) syncAble() ([]SyncPoint, map[int][]types.R) {
	threadNext := make([]nextStr, 0)
	SelectSp := make(map[int][]types.R)

	for i := range mach.Threads {
		tmp := mach.Threads[i].NextSym()
		cleanedNext := make([]types.R, 0)
		nextEl := nextStr{ThreadId: i}

		for _, x := range tmp {
			switch y := x.(type) {
			case types.SELECTGROUP:
				//fmt.Println("SELECTGROUP")
				xa := SelectSp[i]
				xa = append(xa, y...)
				SelectSp[i] = xa
				//fmt.Println(nextEl.SelectSp)
			default:
				cleanedNext = append(cleanedNext, x)
			}
		}
		nextEl.Rs = cleanedNext
		threadNext = append(threadNext, nextEl)
		//threadNext = append(threadNext, nextStr{i, mach.Threads[i].NextSym()})
	}

	doubleDismiss := make(map[SyncPoint]int)
	loopstopper := make([]SyncPoint, 0)
	syncpoints := make([]SyncPoint, 0)
	for _, n := range threadNext { //for each thread
		for _, r := range n.Rs { //for each symbol that the thread can execute next
			skip := false
			q := r

			if !skip {
				p := mach.PMap[q]
				f := false
				for _, m := range threadNext { //find a partner thread
					if m.ThreadId != n.ThreadId { //shouldn't be itself
						if Rscontains(p, m.Rs) { //if 'next' from thread m contains the partner symbol of thread n
							tmp := SyncPoint{n.ThreadId, m.ThreadId, q}
							f = true
							if _, ok := doubleDismiss[tmp]; !ok {

								//check if it would transfer the thread in a new state:

								divtmp1 := mach.Threads[n.ThreadId].Deriv(q)
								divtmp2 := mach.Threads[m.ThreadId].Deriv(p)

								_, ok1 := mach.StateMap[types.Seq{types.Phi(n.ThreadId), divtmp1}]
								_, ok2 := mach.StateMap[types.Seq{types.Phi(m.ThreadId), divtmp2}]
								if !ok1 && !ok2 {
									syncpoints = append(syncpoints, tmp)
									doubleDismiss[tmp] = 1
									tmp2 := SyncPoint{m.ThreadId, n.ThreadId, p}
									doubleDismiss[tmp2] = 1
								} else {
									loopstopper = append(loopstopper, SyncPoint{n.ThreadId, m.ThreadId, q})
									// // mach.Trace += q.String() + p.String()
									// mach.Threads[n.ThreadId] = divtmp1
									// mach.Threads[m.ThreadId] = divtmp2
									// mach.Succ = true
									// for _, xx := range mach.Threads {
									// 	if !xx.Nullable() {
									// 		mach.Succ = false
									// 		mach.Stop = true
									// 	}
									// }
								}

							}
						}
					}
				}
				if !f {
					syncpoints = append(syncpoints, SyncPoint{n.ThreadId, -1, r})
				}
			}

		}
	}
	count := 0
	for _, s := range syncpoints {
		if s.T2Id == -1 {
			count++
		}
	}

	if count == len(syncpoints) && len(loopstopper) > 0 {
		mach.Threads[loopstopper[0].T1Id] = mach.Threads[loopstopper[0].T1Id].Deriv(loopstopper[0].Symbol)
		p := mach.PMap[loopstopper[0].Symbol]
		mach.Threads[loopstopper[0].T2Id] = mach.Threads[loopstopper[0].T2Id].Deriv(p)
		mach.Trace += loopstopper[0].Symbol.String() + p.String()

		count2 := 0
		for _, t := range mach.Threads {
			if t.Nullable() {
				count2++
			}
		}
		if count2 == len(mach.Threads) {
			mach.Succ = true
		} else {
			mach.Stop = true
		}
		return []SyncPoint{}, SelectSp
	}

	return syncpoints, SelectSp
}
func (m *machine2) sync(sp SyncPoint, selectsp map[int][]types.R) { // syncpoint can contain something like Alt{a, SKIP}, first try what happens if SKIP is added, only if it fails add a
	if sp.T2Id != -1 {
		m.StateMap[types.Seq{types.Phi(sp.T1Id), m.Threads[sp.T1Id]}] = 1
		m.StateMap[types.Seq{types.Phi(sp.T2Id), m.Threads[sp.T2Id]}] = 1

		m.Threads[sp.T1Id] = m.Threads[sp.T1Id].Deriv(sp.Symbol)
		m.Trace += sp.Symbol.String()
		p := m.PMap[sp.Symbol]
		m.Threads[sp.T2Id] = m.Threads[sp.T2Id].Deriv(p)
		m.Trace += p.String()

        //SPECIAL CASE FOR SELECT

        if len(selectsp[sp.T1Id]) > 0 {
            tmp := selectsp[sp.T1Id]
            for _, xx := range tmp {
                if xx != sp.Symbol {
                    pa := m.PMap[xx]
                    
                    for j := range m.Threads {
                        nex := m.Threads[j].NextSym()
                        if Rscontains(pa, nex) {
                            m.Threads[j] = m.Threads[j].Deriv(pa)
                            break
                        }
                    }
                }
            }
        } else if len(selectsp[sp.T2Id]) > 0 {
            tmp := selectsp[sp.T2Id]
            for _, xx := range tmp {
                if xx != p {
                    pa := m.PMap[xx]
                    
                    for j := range m.Threads {
                        nex := m.Threads[j].NextSym()
                        if Rscontains(pa, nex) {
                            m.Threads[j] = m.Threads[j].Deriv(pa)
                            break
                        }
                    }
                }
            }
        }

		count := 0
		for _, t := range m.Threads {
			if t.Nullable() {
				count++
			}
		}
		if count == len(m.Threads) {
			m.Succ = true
		}
	} else {
		if sp.Symbol != types.Eps(1) {
			m.Trace += sp.Symbol.String()
			// m.Threads[sp.T1Id] = m.Threads[sp.T1Id].Deriv(sp.Symbol, sp.T1Id)
			m.Stop = true
		}
	}
}

func run3(initM machine2) {
	//reader := bufio.NewReader(os.Stdin)

	machines := []machine2{initM}

	for {
		for i := range machines {
			if machines[i].Succ || machines[i].Stop || machines[i].Abort {
				continue
			}

			syncpoints, selectSp := machines[i].syncAble()
			//fmt.Println(syncpoints, machines[i].Threads, machines[i].Trace)
			if len(syncpoints) == 0 {
				machines[i].Stop = true
			} else {
				for j := range syncpoints { // for each syncpoint, create a clone (but not for the last one)
					if j < len(syncpoints)-1 {
						clone := machines[i].clone()
						clone.sync(syncpoints[j], selectSp)
						machines = append(machines, clone)
					} else {
						machines[i].sync(syncpoints[j], selectSp)
					}
				}
			}
			//reader.ReadString('\n')
		}
		//reader.ReadString('\n')
		count := 0
		for i := range machines {
			if machines[i].Succ || machines[i].Stop || machines[i].Abort {
				count++
			}
		}
		if count == len(machines) {
			break
		}
	}

	fmt.Println(len(machines))
	doubledismiss := make(map[string]int)
	stopped, succs, aborted := 0, 0, 0
	for i := range machines {
		_, ok := doubledismiss[machines[i].Trace]
		if !ok && !machines[i].Abort {
			fmt.Println(machines[i].Succ, machines[i].Trace, machines[i].Threads)
			doubledismiss[machines[i].Trace] = 1
		}
		if machines[i].Succ {
			succs++
		} else if machines[i].Abort {
			aborted++
		} else if machines[i].Stop {
			stopped++
		}
	}
	fmt.Printf("Succs: %v\nAbort: %v\nStopped: %v\n", succs, aborted, stopped)
	//reader.ReadString('\n')
}

func run(rs []types.R, filter func(rs []types.R) []types.R, partner map[types.R]types.R) {
	//reader := bufio.NewReader(os.Stdin)

	machines := []machine{machine{Rs: rs}}

	for {
		for i := range machines {
			if machines[i].Succ || machines[i].Stop || machines[i].Abort {
				continue
			}

			if len(machines[i].Rs) == 0 {
				machines[i].Stop = true
			} else {
				failedRs := 0
				// fmt.Println(len(machines[i].Rs))
				for r := range machines[i].Rs {
					if machines[i].Rs[r].ContainsPhi() {
						continue
					}
					// succ := false

					next := machines[i].Rs[r].NextSym()
					//     fmt.Println("next:", next, machines[i].Rs)
					//     fmt.Println(machines[i].Trace, machines[i].Rs[r], next)
					next = filter(next)
					//    fmt.Println("next2:", next)

					doubleFilter := make(map[types.R]int)
					for _, n := range next {
						_, ok := doubleFilter[n]
						if !ok {
							doubleFilter[n] = 1
						}
					}
					next = make([]types.R, 0, len(next))
					for k, _ := range doubleFilter {
						next = append(next, k)
					}
					fmt.Println(next)
					for _, vl := range machines[i].Rs {
						if !vl.ContainsPhi() {
							fmt.Println(vl)
						}
					}

					if len(next) == 0 {
						if machines[i].Rs[r].Nullable() {
							machines[i].Succ = true
						} else {
							failedRs++
						}
					} else {
						for _, n := range next {
							clone := machines[i]

							tmp := make([]types.R, 0)
							for k := range clone.Rs {
								tmp = append(tmp, clone.Rs[k].PDeriv(n)...)
							}

							tmp2 := make([]types.R, 0)
							for k := range tmp {
								if !tmp[k].ContainsPhi() {
									tmp2 = append(tmp2, tmp[k].PDeriv(partner[n])...)
								}
								//tmp2 = append(tmp2, tmp[k].PDeriv(partner[n])...)
							}

							for _, t := range tmp2 {
								if t.Nullable() {
									//  succ = true
									clone.Succ = true
								}
								// if !t.ContainsPhi() {
								//     clone.Rs = append(clone.Rs, t)
								// }
							}
							clone.Rs = tmp2
							clone.Trace += n.String() + partner[n].String()
							machines = append(machines, clone)
						}
					}
				}
				if failedRs == len(machines[i].Rs) {
					machines[i].Stop = true
				} else {
					machines[i].Abort = true
				}

				// machines[i].Stop = true
			}
			//	reader.ReadString('\n')
		}

		count := 0

		for i := range machines {
			if machines[i].Succ || machines[i].Stop || machines[i].Abort {
				count++
			}
		}
		if count == len(machines) {
			break
		}
	}

	fmt.Println(len(machines))
	doubledismiss := make(map[string]int)
	stopped, succs, aborted := 0, 0, 0
	for i := range machines {
		//   fmt.Println(machines[i].Rs)
		_, ok := doubledismiss[machines[i].Trace]
		if !ok && !machines[i].Abort {
			fmt.Println(machines[i].Succ, machines[i].Trace, machines[i].Rs[0])
			doubledismiss[machines[i].Trace] = 1
		}
		if machines[i].Succ {
			succs++
		} else if machines[i].Abort {
			aborted++
		} else if machines[i].Stop {
			stopped++
		}
	}
	fmt.Printf("Succs: %v\nAborted: %v\nStopped: %v\n", succs, aborted, stopped)
	//	reader.ReadString('\n')
}

func Run(r types.R, pmap map[types.R]types.R) {
	partnerFilter := func(rs []types.R) (ret []types.R) {
		ret = make([]types.R, 0)
		doubleDismiss := make(map[types.R]int)
		for _, e := range rs {
			if _, ok := doubleDismiss[e]; !ok {
				p := pmap[e]
				for _, r := range rs {
					if r == p {
						ret = append(ret, e)
						doubleDismiss[e] = 1
						doubleDismiss[r] = 1
						break
					}
				}
			}

		}
		return
	}

	run([]types.R{r}, partnerFilter, pmap)
}

func Run2(threads []types.R, pmap map[types.R]types.R) {
	m := machine2{Threads: threads, PMap: pmap, StateMap: make(map[types.R]int)}
	run3(m)
}
