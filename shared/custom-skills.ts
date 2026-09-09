import {
  compileCustomEffect,
  compileCustomSkill,
} from "./custom-skill-mechanics";
import { describeBattleEffect, describeBattleSkill } from "./skill-description";
// The editor labels and replay descriptions share one vocabulary. These keys
// are a bounded data format, never executable code or user supplied formulas.
export const CUSTOM_SKILL_START = 4100000000;
export const kinds = {
  damage: "ダメージ",
  heal: "HPを回復",
  revive: "倒れた味方を復活",
  phyUp: "PHYを上げる",
  phyDown: "PHYを下げる",
  intUp: "INTを上げる",
  intDown: "INTを下げる",
  agiUp: "AGIを上げる",
  agiDown: "AGIを下げる",
  chargeUp: "チャージを増やす",
  chargeDown: "チャージを減らす",
  shield: "シールドを張る",
  status: "状態を付与",
  cure: "状態異常を解除",
  buff: "特殊効果を付与",
} as const;
export const targets = {
  self: "自分",
  enemyFront: "敵の先頭",
  enemyBack: "敵の最後尾",
  enemyAll: "敵全体",
  enemyRandom: "ランダムな敵1体",
  enemyLowestHp: "残りHPが最も低い敵",
  enemyHighestPhy: "PHYが最も高い敵",
  enemyHighestInt: "INTが最も高い敵",
  enemyHighestAgi: "AGIが最も高い敵",
  allyFront: "味方の先頭",
  allyBack: "味方の最後尾",
  allyAll: "味方全体",
  allyOther: "自分以外の味方",
  allyRandom: "ランダムな味方1体",
  allyLowestHp: "残りHPが最も低い味方",
  allyHighestPhy: "PHYが最も高い味方",
  allyHighestInt: "INTが最も高い味方",
  allyHighestAgi: "AGIが最も高い味方",
  allySlotFront: "味方の前衛（位置指定）",
  allySlotMiddle: "味方の中衛（位置指定）",
  allySlotBack: "味方の後衛（位置指定）",
  previous: "このスキルでダメージを与えた相手",
} as const;
export const references = {
  fixed: "固定の数値",
  PHY: "PHY",
  INT: "INT",
  AGI: "AGI",
  HP: "HP",
  CHARGE: "チャージ",
  DAMAGE: "この行動で与えたダメージ",
  HEALING: "この行動で回復したHP",
  TAKEN_DAMAGE: "直前の行動で受けたダメージ",
  HEAL_INT: "回復係数（自身のINTと対象のPHYの平均）",
  HEAL_PHY: "PHY回復係数（自身のPHYと対象のINTの平均）",
} as const;
export const bases = {
  CURRENT: "現在の値",
  BASE: "戦闘開始時の値",
  INCREASE: "増えた分",
  DECREASE: "減った分",
} as const;
export const statuses = {
  POISON: "毒",
  SLEEP: "睡眠",
  CONFUSION: "混乱",
  FEAR: "恐怖",
  BARRIER: "バリア",
  BLEED: "出血",
  STUN: "気絶",
  PROSTRATION: "衰弱",
  CURSE: "呪い",
  CHARM: "魅了",
  BIND: "バインド",
  GRAVITY: "重力",
} as const;
export const buffs = {
  RESURRECTION: "復活予約",
  DECOY: "デコイ（攻撃を引き受ける）",
  AGI_DAMAGE_REDUCTION: "AGIを使ったダメージの軽減",
  HP_DAMAGE_REDUCTION: "HPを使ったダメージの軽減",
} as const;
export const triggers = {
  active: "行動順で使う",
  opening: "バトル開始時",
  damaged: "行動スキルでダメージを受けた後",
  acted: "自分の行動後",
  lowHp: "自分のHPが指定割合を下回った時",
  abnormal: "自分が状態異常の時",
} as const;
export const visuals = {
  1: "物理攻撃",
  2: "魔法攻撃",
  3: "回復",
  4: "強化",
  5: "弱体",
} as const;
export type CustomEffect = {
  kind: keyof typeof kinds;
  target: keyof typeof targets;
  chance: number;
  reference: keyof typeof references;
  source: "self" | "target";
  basis: keyof typeof bases;
  min: number;
  max: number;
  status: keyof typeof statuses | "ANY_EFFECT";
  buff: keyof typeof buffs;
};
export type CustomSkill = {
  id: number;
  name: string;
  trigger: keyof typeof triggers;
  chance: number;
  uses: number;
  threshold: number;
  effectId: number;
  effects: CustomEffect[];
};
export const defaultEffect = (): CustomEffect => ({
  kind: "damage",
  target: "enemyFront",
  chance: 100,
  reference: "PHY",
  source: "self",
  basis: "CURRENT",
  min: 30,
  max: 35,
  status: "POISON",
  buff: "DECOY",
});
export const defaultSkill = (id: number, passive = false): CustomSkill => ({
  id,
  name: "新しいスキル",
  trigger: passive ? "opening" : "active",
  chance: 100,
  uses: passive ? 1 : -1,
  threshold: 50,
  effectId: 1,
  effects: [defaultEffect()],
});
const own = (values: object, v: unknown) =>
  typeof v === "string" && Object.hasOwn(values, v);
const int = (v: unknown, min: number, max: number) =>
  Number.isInteger(v) && Number(v) >= min && Number(v) <= max;
export function parseCustomSkill(value: any): CustomSkill {
  const s = value;
  if (
    !s ||
    !int(s.id, CUSTOM_SKILL_START, CUSTOM_SKILL_START + 11) ||
    typeof s.name !== "string" ||
    !s.name.trim() ||
    s.name.trim().length > 32 ||
    /[\u0000-\u001f\u007f]/.test(s.name)
  )
    throw Error("スキルの名前は1〜32文字で入力してください");
  if (
    !own(triggers, s.trigger) ||
    !int(s.chance, 1, 100) ||
    !(s.uses === -1 || int(s.uses, 1, 20)) ||
    !int(s.threshold, 1, 100) ||
    !int(s.effectId, 1, 5)
  )
    throw Error("発動条件・確率・回数を確認してください");
  if (s.trigger === "active" && s.uses !== -1)
    throw Error("行動スキルは回数制限なしで設定してください");
  if (!Array.isArray(s.effects) || s.effects.length < 1 || s.effects.length > 6)
    throw Error("効果は1〜6個で作ってください");
  const effects = s.effects.map((e: any): CustomEffect => {
    if (
      !e ||
      !own(kinds, e.kind) ||
      !own(targets, e.target) ||
      !int(e.chance, 1, 100) ||
      !own(references, e.reference) ||
      !["self", "target"].includes(e.source) ||
      !own(bases, e.basis)
    )
      throw Error("効果の対象・内容・確率を確認してください");
    if (
      (!own(statuses, e.status) && e.status !== "ANY_EFFECT") ||
      !own(buffs, e.buff)
    )
      throw Error("状態変化を選んでください");
    if (e.kind === "status" && e.status === "ANY_EFFECT")
      throw Error("付与する状態を選んでください");
    if (
      !int(e.min, 0, 1000000) ||
      !int(e.max, e.min, 1000000) ||
      (e.reference !== "fixed" && e.max > 1000)
    )
      throw Error("数値は0〜1,000,000、倍率は0〜1,000%で設定してください");
    if (
      e.kind === "revive" &&
      !["allyAll", "allySlotFront", "allySlotMiddle", "allySlotBack"].includes(
        e.target,
      )
    )
      throw Error("復活の対象は自分・味方から選んでください");
    return {
      kind: e.kind,
      target: e.target,
      chance: e.chance,
      reference: e.reference,
      source: e.source,
      basis: e.basis,
      min: e.min,
      max: e.max,
      status: e.status,
      buff: e.buff,
    };
  });
  return {
    id: s.id,
    name: s.name.trim(),
    trigger: s.trigger,
    chance: s.chance,
    uses: s.uses,
    threshold: s.threshold,
    effectId: s.effectId,
    effects,
  };
}
export function describeEffect(e: CustomEffect): string {
  return describeBattleEffect(compileCustomEffect(e), { custom: true });
}
export function customSkillDetail(s: CustomSkill) {
  const raw = compileCustomSkill(s),
    ja = describeBattleSkill(raw, s, { custom: true }),
    en = describeBattleSkill(raw, s, { custom: true, language: "en" });
  return {
    ...ja,
    en: { name: en.name, condition: en.condition, effects: en.effects },
  };
}
export const customSkillCatalog = (team: { customSkills?: CustomSkill[] }[]) =>
  team.flatMap((f) => (f.customSkills || []).map(customSkillDetail));
