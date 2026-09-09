package battle

import (
 "context"
 "fmt"
)

// The original engine reads immutable skills from this process-wide catalog.
// Register once at startup; requests must never replace it.
var r Repository
type Repository interface {
 Reload(context.Context) error
 GetBattle(context.Context,uint32)(*Battle,error)
 GetTxBattle(context.Context,uint32)(*Battle,error)
 SaveBattle(context.Context,*Battle)(uint32,error)
 GetSkill(uint32)(*Skill,error)
 SaveSkill(context.Context,*Skill)error
}
func RegisterRepository(repository Repository) { r=repository }
type CatalogRepository struct { Skills map[uint32]*Skill }
func (r *CatalogRepository) GetSkill(id uint32)(*Skill,error) { s,ok:=r.Skills[id]; if !ok{return nil,fmt.Errorf("unknown skill %d",id)}; return s,nil }
func (*CatalogRepository) Reload(context.Context)error{return nil}
func (*CatalogRepository) SaveSkill(context.Context,*Skill)error{return fmt.Errorf("catalog is immutable")}
func (*CatalogRepository) GetBattle(context.Context,uint32)(*Battle,error){return nil,fmt.Errorf("results belong to D1")}
func (*CatalogRepository) GetTxBattle(context.Context,uint32)(*Battle,error){return nil,fmt.Errorf("results belong to D1")}
func (*CatalogRepository) SaveBattle(context.Context,*Battle)(uint32,error){return 0,fmt.Errorf("results belong to D1")}
