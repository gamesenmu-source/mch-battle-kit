package runner

import (
 "os"
 "testing"
 b "mch-light/engine/battle"
)

func TestNoviceProductionSkillsBattle(t *testing.T) {
 file, err := os.Open("../data/skills.ndjson")
 if err != nil { t.Fatal(err) }; defer file.Close()
 skills, err := LoadSkills(file); if err != nil { t.Fatal(err) }
 b.RegisterRepository(&b.CatalogRepository{Skills: skills})
 in := Input{Seed: 321, CatalogVersion: "novice-test", Team: []Fighter{
  {Name: "MCH Warrior", HeroType:10001, HP:282, PHY:181, INT:42, AGI:53, Passive:1017, Actives:[]uint32{2007,2007,3001}},
  {Name: "MCH Tactician", HeroType:10002, HP:282, PHY:160, INT:63, AGI:53, Passive:1018, Actives:[]uint32{2031,2043,3001}},
  {Name: "MCH Artist", HeroType:10003, HP:279, PHY:171, INT:53, AGI:53, Passive:1019, Actives:[]uint32{2094,2007,3001}},
 }}
 result, err := Run(in); if err != nil { t.Fatal(err) }
 if result.Score <= 0 || len(result.Frames)<2 { t.Fatal("novice battle produced no damage or replay") }
 for _, id := range []uint32{10000,10004,11001,15001} {
  in.Team[0].HeroType=id
  if _,err:=Run(in); err==nil {t.Fatalf("unapproved type accepted: %d",id)}
 }
}
