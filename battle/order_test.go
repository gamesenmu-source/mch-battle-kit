package battle
import "testing"
func TestStatusOrderStable(t *testing.T){a,b:=&StatusEffect{},&StatusEffect{};for i:=0;i<100;i++{got:=orderedStatuses(map[int32]*StatusEffect{9:b,2:a});if got[0]!=a||got[1]!=b{t.Fatal("unstable status order")}}}
