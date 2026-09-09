package battle

import "sort"

// Go map iteration is randomized. Ranking battles use the same seed for every
// participant, so process effects in numeric ID order before consuming randomness.
func orderedStatuses(src map[int32]*StatusEffect) []*StatusEffect {
 keys:=make([]int,0,len(src));for k:=range src{keys=append(keys,int(k))};sort.Ints(keys)
 out:=make([]*StatusEffect,0,len(keys));for _,k:=range keys{out=append(out,src[int32(k)])};return out
}
func orderedBuffs(src map[int32]*BuffEffect) []*BuffEffect {
 keys:=make([]int,0,len(src));for k:=range src{keys=append(keys,int(k))};sort.Ints(keys)
 out:=make([]*BuffEffect,0,len(keys));for _,k:=range keys{out=append(out,src[int32(k)])};return out
}
