// Japanese grammar follows mch-master/skills_atomic2/front/format.go.
// Input is the battle engine's protobuf JSON, never previously rendered prose.
export type BattleTarget = {
  team: string;
  position?: string;
  paramCondition?: { paramType: string; param: string; calc: string };
};
export type BattleCondition = {
  team: string;
  position: string;
  stateCondition?: string;
  paramCondition?: { param: string; calc: string; rate?: number };
};
export type BattleEffect = {
  target: BattleTarget;
  param: string;
  isDamage?: boolean;
  successRate?: number;
  rate?: { min?: number; max?: number };
  rawStat?: {
    team: string;
    position: string;
    basicStat?: { paramType: string; param: string };
    battleStat?: string;
  };
  computedStat?: string;
  statusEffectType?: string;
  buffEffectType?: string;
  skillId?: number;
  additionalEffect?: {
    rate?: number;
    statusEffectType?: string;
    buffEffectType?: string;
    seriesBonusType?: string;
  };
};
export type BattleSkill = {
  skillId?: number;
  remainingUses?: number;
  trigger?: { triggerRate?: number; conditions: BattleCondition[] };
  effects?: BattleEffect[];
};
export const skillStats: Record<string, string> = {
  HP: "HP",
  PHY: "PHY",
  INT: "INT",
  AGI: "AGI",
  ALL_PARAMS: "PHY/INT/AGI",
  ACTIVE_SKILL_TAKEN_DAMAGE: "受けたダメージ",
  ACTION_ADDED_DAMAGE: "与えたダメージ",
  ACTION_ADDED_HEALING: "与えた回復量",
  CHARGE: "チャージ",
  SHIELD: "シールド",
  INT_HEALING_MODIFIER: "回復係数",
  PHY_HEALING_MODIFIER: "PHY回復係数",
  PHY_INT_HIGHER: "PHY/INTの高い方",
  PHY_INT_LOWER: "PHY/INTの低い方",
  ALL_PARAMS_HIGHEST: "PHY/INT/AGIの最も高い方",
  ALL_PARAMS_LOWEST: "PHY/INT/AGIの最も低い方",
};
export const skillStates: Record<string, string> = {
  ANY_EFFECT: "全ての状態異常",
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
  DECOY: "デコイ",
  RESURRECTION: "復活予約",
  AGI_DAMAGE_REDUCTION: "AGIダメージ軽減",
  HP_DAMAGE_REDUCTION: "HPダメージ軽減",
};
const teams: Record<string, string> = {
  ALLY: "味方",
  ENEMY: "敵",
  BOTH: "敵味方",
};
const positions: Record<string, string> = {
  SELF: "自身",
  TARGET: "対象",
  FIRST: "先頭",
  LAST: "最後尾",
  FRONT: "前衛",
  MIDDLE: "中衛",
  BACK: "後衛",
};
function word(dictionary: Record<string, string>, key: string): string {
  if (!Object.hasOwn(dictionary, key))
    throw Error(`Unsupported battle description key: ${key}`);
  return dictionary[key];
}
const num = (n: number) => (Number.isFinite(n) ? String(n) : "—");
export function describePosition(team: string, position: string): string {
  if (position === "SELF" || position === "TARGET") return positions[position];
  const side = word(teams, team);
  if (position === "ALL") return `${side}全体`;
  if (position === "RANDOM") return `${side}の誰か`;
  if (position === "EXCEPT_SELF") return `自身以外の${side}`;
  return `${word(positions, position)}の${side}`;
}
export function describeStat(param: string, basis = "CURRENT"): string {
  const stat = word(skillStats, param);
  switch (basis) {
    case "CURRENT":
      return stat;
    case "BASE":
      return param === "HP" ? `最大${stat}` : `元の${stat}`;
    case "INCREASE":
      return `${stat}増加量`;
    case "DECREASE":
      return `${stat}減少量`;
    default:
      throw Error(`Unsupported stat basis: ${basis}`);
  }
}
export function describeTarget(t: BattleTarget): string {
  if (t.position) return describePosition(t.team, t.position);
  if (!t.paramCondition) throw Error("Missing battle effect target");
  const p = t.paramCondition,
    stat = describeStat(p.param, p.paramType);
  const value = p.paramType === "CURRENT" || p.paramType === "BASE";
  if (p.calc !== "HIGHEST" && p.calc !== "LOWEST")
    throw Error("Unsupported target comparison");
  return `${stat}が最も${value ? (p.calc === "HIGHEST" ? "高い" : "低い") : p.calc === "HIGHEST" ? "大きい" : "小さい"}${word(teams, t.team)}`;
}
function reference(e: BattleEffect): string {
  if (e.computedStat) return word(skillStats, e.computedStat);
  const r = e.rawStat;
  if (!r) return "";
  if (
    r.battleStat &&
    [
      "ACTIVE_SKILL_TAKEN_DAMAGE",
      "ACTION_ADDED_DAMAGE",
      "ACTION_ADDED_HEALING",
    ].includes(r.battleStat)
  ) {
    const value = word(skillStats, r.battleStat);
    return r.position === "SELF"
      ? value
      : `${describePosition(r.team, r.position)}が${value}`;
  }
  const p = r.basicStat?.param || r.battleStat!;
  const sum =
    (r.position !== "SELF" &&
      r.position !== "TARGET" &&
      (r.position === "ALL" || r.team === "BOTH")) ||
    p === "ALL_PARAMS";
  return `${describePosition(r.team, r.position)}の${describeStat(p, r.basicStat?.paramType || "CURRENT")}${sum ? "合計" : ""}`;
}
type DescriptionOptions = {
  custom?: boolean;
  skillName?: (id: number) => string;
  language?: "ja" | "en";
};
function rate(
  e: BattleEffect,
  additional: number,
  options: DescriptionOptions,
): string {
  const min = e.rate?.min || 0,
    stored = e.rate?.max || 0;
  // Authored skills expose inclusive endpoints; Go compiles them to exclusive max.
  const max = options.custom && stored > min ? stored - 1 : stored;
  const value =
    min === max
      ? num(min + additional)
      : `${num(min + additional)} ~ ${num(max + additional)}`;
  return value + (e.rawStat || e.computedStat ? "%" : "");
}
export function describeBattleEffect(
  e: BattleEffect,
  options: DescriptionOptions = {},
): string {
  if (options.language === "en") return describeEnglishEffect(e, options);
  const target = describeTarget(e.target),
    prob = e.successRate || 0;
  const ref = reference(e),
    value = (ref ? ref + "の" : "") + rate(e, 0, options);
  const stat = [
    "HP",
    "REVIVE",
    "STATUS_EFFECT",
    "BUFF_EFFECT",
    "SKILL",
  ].includes(e.param)
    ? target
    : `${target}の${word(skillStats, e.param)}`;
  const action = e.isDamage ? "ダウン" : "アップ";
  let text: string;
  switch (e.param) {
    case "STATUS_EFFECT":
      text = e.isDamage
        ? `${target}に${num(prob)}%の確率で${word(skillStates, e.statusEffectType!)}を付与`
        : `${target}の${word(skillStates, e.statusEffectType!)}を${num(prob)}%の確率で解除`;
      break;
    case "BUFF_EFFECT":
      text = `${target}に${num(prob)}%の確率で${word(skillStates, e.buffEffectType!)}を付与`;
      break;
    case "SKILL":
      text = `このスキルを${num(prob)}%の確率で[${options.skillName?.(e.skillId!) || `スキル ${e.skillId}`}]に変化`;
      break;
    case "HP":
      text = e.isDamage
        ? `${target}に${value}ダメージ`
        : `${target}のHPを${value}回復`;
      break;
    case "REVIVE":
      text = `${target}を${value}回復した状態で復活`;
      break;
    case "SHIELD":
      text = `${stat}を${value}に更新`;
      break;
    default:
      text = `${stat}を${value}${action}`;
  }
  if (
    !["STATUS_EFFECT", "BUFF_EFFECT", "SKILL"].includes(e.param) &&
    prob !== 100
  )
    text = `${num(prob)}%の確率で、${text}`;
  const a = e.additionalEffect;
  if (a?.seriesBonusType) {
    text += `[${word({ INVERSE: "インバース", DIVERSE: "ダイバース" }, a.seriesBonusType)}(${a.rate || 0})]`;
  } else if (a) {
    const state = word(skillStates, a.statusEffectType || a.buffEffectType!);
    const bonus = rate(e, a.rate || 0, options);
    const ending =
      e.param === "HP"
        ? e.isDamage
          ? "ダメージ"
          : "回復"
        : e.param === "REVIVE"
          ? "回復した状態で復活"
          : e.param === "SHIELD"
            ? "に更新"
            : action;
    if (e.param === "STATUS_EFFECT" || e.param === "BUFF_EFFECT")
      text += `、対象が${state}なら${num(prob + (a.rate || 0))}%の確率で${e.param === "STATUS_EFFECT" && !e.isDamage ? "解除" : "付与"}`;
    else if (e.param === "SKILL")
      text += `、対象が${state}なら${num(prob + (a.rate || 0))}%の確率で変化`;
    else text += `、対象が${state}なら${bonus}${ending}`;
  }
  return text;
}
export function describeBattleTrigger(
  s: BattleSkill,
  language: "ja" | "en" = "ja",
): string {
  if (language === "en") return describeEnglishTrigger(s);
  if (!s.trigger) return "";
  const conditions = s.trigger.conditions;
  // The original converter splits a combined parameter/state row into two clauses.
  const rows: BattleCondition[] = [];
  for (let i = 0; i < conditions.length; i++) {
    const c = { ...conditions[i] },
      next = conditions[i + 1];
    if (
      c.paramCondition &&
      next?.stateCondition &&
      next.team === "BOTH" &&
      next.position === "TARGET"
    ) {
      c.stateCondition = next.stateCondition;
      i++;
    }
    rows.push(c);
  }
  let text = "";
  rows.forEach((c, i) => {
    const subject =
      c.position === "SELF"
        ? "自身が"
        : c.position === "TARGET"
          ? ""
          : c.position === "ALL"
            ? `${word(teams, c.team)}の誰かが`
            : describePosition(c.team, c.position);
    if (c.stateCondition) {
      if (i > 0) text += "かつ";
      if (c.stateCondition === "OPENING") text += "バトル開始時";
      else if (c.stateCondition === "HAS_STATUS_EFFECT")
        text += `${subject}状態異常${i === rows.length - 1 && !c.paramCondition ? "の時" : ""}`;
      else
        text += `${subject}${word({ AFTER_ACTIVE_SKILL: "Active Skillを使用した後", AFTER_ACTIVE_SKILL_TAKEN_DAMAGE: "Active Skillでダメージを受けた後", AFTER_DEATH: "死亡した後" }, c.stateCondition)}`;
    }
    if (c.paramCondition) {
      const p = c.paramCondition;
      if (i > 0 || c.stateCondition) text += "かつ";
      const param = word(skillStats, p.param),
        target = describePosition(c.team, c.position);
      const stat =
        c.position === "TARGET"
          ? param + (p.param === "ALL_PARAMS" ? "合計" : "")
          : `${target}の${param}${c.position === "ALL" || c.team === "BOTH" || p.param === "ALL_PARAMS" ? "合計" : ""}`;
      text += `${stat}が${p.param === "HP" ? "" : "元の値の"}${num(p.rate || 0)}%${word({ UNDER: "未満", OVER: "以上" }, p.calc)}${i === rows.length - 1 ? "の時" : ""}`;
    }
  });
  return `${text}に${num(s.trigger.triggerRate || 0)}%の確率で${(s.remainingUses || 0) > 0 ? `${s.remainingUses}回だけ` : ""}発動`.replaceAll(
    "Active Skill",
    "アクティブスキル",
  );
}
export function describeBattleSkill(
  s: BattleSkill,
  meta: { name: string; effectId?: number },
  options: DescriptionOptions = {},
) {
  const effects: { text: string; rate: number }[] = [];
  const values = (s.effects || []).map((e) => ({
    text: describeBattleEffect(e, options),
    rate: e.successRate || 0,
    identity: JSON.stringify(e),
  }));
  for (let i = 0; i < values.length; i++) {
    const e = values[i];
    let count = 1;
    while (i + 1 < values.length && values[i + 1].identity === e.identity) {
      i++;
      count++;
    }
    effects.push({
      text:
        count > 1
          ? options.language === "en"
            ? `[${count} times]${e.text}`
            : `「${e.text}」を${count}回繰り返す`
          : e.text,
      rate: e.rate,
    });
  }
  return {
    id: s.skillId || 0,
    name: meta.name,
    effectId: meta.effectId || 0,
    condition: describeBattleTrigger(s, options.language),
    effects,
  };
}

// English follows the original master generator, including its terminology.
const enStats: Record<string, string> = {
  HP: "HP",
  PHY: "PHY",
  INT: "INT",
  AGI: "AGI",
  ALL_PARAMS: "PHY/INT/AGI",
  ACTIVE_SKILL_TAKEN_DAMAGE: "damage taken",
  ACTION_ADDED_DAMAGE: "damage dealt",
  ACTION_ADDED_HEALING: "healing dealt by this hero",
  CHARGE: "CHARGE",
  SHIELD: "SHIELD",
  INT_HEALING_MODIFIER: "Healing Coefficient",
  PHY_HEALING_MODIFIER: "PHY Healing Coefficient",
  PHY_INT_HIGHER: "higher PHY/INT",
  PHY_INT_LOWER: "lower PHY/INT",
  ALL_PARAMS_HIGHEST: "highest PHY/INT/AGI",
  ALL_PARAMS_LOWEST: "lowest PHY/INT/AGI",
};
const enStates: Record<string, string> = {
  ANY_EFFECT: "all status effects",
  POISON: "Poison",
  SLEEP: "Sleep",
  CONFUSION: "Confusion",
  FEAR: "Fear",
  BARRIER: "Barrier",
  BLEED: "Bleed",
  STUN: "Stun",
  PROSTRATION: "Prostration",
  CURSE: "Curse",
  CHARM: "Charm",
  BIND: "Bind",
  GRAVITY: "Gravity",
  DECOY: "Decoy",
  RESURRECTION: "Resurrection",
  AGI_DAMAGE_REDUCTION: "AGI damage reduction",
  HP_DAMAGE_REDUCTION: "HP damage reduction",
};
function enPosition(team: string, position: string): string {
  if (position === "SELF") return "this hero";
  if (position === "TARGET") return "target";
  if (position === "ALL")
    return word(
      {
        ALLY: "all allies",
        ENEMY: "all enemies",
        BOTH: "all allies and enemies",
      },
      team,
    );
  if (position === "RANDOM")
    return word(
      { ALLY: "an ally", ENEMY: "an enemy", BOTH: "an ally or enemy" },
      team,
    );
  if (position === "EXCEPT_SELF")
    return `${enPosition(team, "ALL")} except this hero`;
  return `the ${word({ FIRST: "first", LAST: "last", FRONT: "frontline", MIDDLE: "midline", BACK: "backline" }, position)} ${word({ ALLY: "ally", ENEMY: "enemy", BOTH: "ally and enemy" }, team)}`;
}
function enStat(param: string, basis = "CURRENT"): string {
  const stat = word(enStats, param);
  return word(
    {
      CURRENT: "{s}",
      BASE: param === "HP" ? "Max {s}" : "original {s}",
      INCREASE: "{s} increase",
      DECREASE: "{s} decrease",
    },
    basis,
  ).replace("{s}", stat);
}
function enTarget(t: BattleTarget): string {
  if (t.position) return enPosition(t.team, t.position);
  const p = t.paramCondition!;
  return `the ${word({ ALLY: "ally", ENEMY: "enemy", BOTH: "ally or enemy" }, t.team)} with the ${word({ HIGHEST: "highest", LOWEST: "lowest" }, p.calc)} ${enStat(p.param, p.paramType)}`;
}
function enReference(e: BattleEffect): string {
  if (e.computedStat) return word(enStats, e.computedStat);
  const r = e.rawStat;
  if (!r) return "";
  if (
    r.battleStat &&
    [
      "ACTIVE_SKILL_TAKEN_DAMAGE",
      "ACTION_ADDED_DAMAGE",
      "ACTION_ADDED_HEALING",
    ].includes(r.battleStat)
  ) {
    const value = word(enStats, r.battleStat);
    return r.position === "SELF"
      ? value
      : r.battleStat === "ACTION_ADDED_HEALING"
        ? `healing dealt by ${enPosition(r.team, r.position)}`
        : `${value} by ${enPosition(r.team, r.position)}`;
  }
  const param = r.basicStat?.param || r.battleStat!,
    stat = enStat(param, r.basicStat?.paramType || "CURRENT"),
    target = enPosition(r.team, r.position);
  const sum =
    (r.position !== "SELF" &&
      r.position !== "TARGET" &&
      (r.position === "ALL" || r.team === "BOTH")) ||
    param === "ALL_PARAMS";
  return sum ? `the total ${stat} of ${target}` : `${target}'s ${stat}`;
}
function describeEnglishEffect(
  e: BattleEffect,
  options: DescriptionOptions,
): string {
  const target = enTarget(e.target),
    ref = enReference(e),
    amount = rate(e, 0, options),
    value = ref ? `${amount} of ${ref}` : amount,
    prob = e.successRate || 0;
  const plural =
    !!e.target.paramCondition ||
    e.target.position === "ALL" ||
    e.target.team === "BOTH";
  const stat = ![
    "HP",
    "REVIVE",
    "STATUS_EFFECT",
    "BUFF_EFFECT",
    "SKILL",
  ].includes(e.param)
    ? plural
      ? `the ${word(enStats, e.param)} of ${target}`
      : `${target}'s ${word(enStats, e.param)}`
    : target;
  let text: string;
  switch (e.param) {
    case "STATUS_EFFECT":
      text = e.isDamage
        ? `${prob}% chance to inflict ${word(enStates, e.statusEffectType!)} on ${target}.`
        : `${prob}% chance to remove ${word(enStates, e.statusEffectType!)} from ${target}.`;
      break;
    case "BUFF_EFFECT":
      text = `${prob}% chance to inflict ${word(enStates, e.buffEffectType!)} on ${target}.`;
      break;
    case "SKILL":
      text = `${prob}% chance for this skill to become [${options.skillName?.(e.skillId!) || `Skill ${e.skillId}`}].`;
      break;
    case "HP":
      text = e.isDamage
        ? ref
          ? `Deal damage to ${target} equal to ${value}.`
          : `Deal ${value} damage to ${target}.`
        : `Heal ${target} for ${value}${ref ? "" : " HP"}.`;
      break;
    case "REVIVE":
      text = `Revive ${target} with ${value}${ref ? "" : " HP"}.`;
      break;
    case "SHIELD":
      text = `Set ${stat} to ${value} if higher.`;
      break;
    default:
      text = `${e.isDamage ? "Decrease" : "Increase"} ${stat} by ${value}.`;
  }
  if (
    !["STATUS_EFFECT", "BUFF_EFFECT", "SKILL"].includes(e.param) &&
    prob !== 100
  )
    text = `${prob}% chance: ${text}`;
  const a = e.additionalEffect;
  if (a?.seriesBonusType)
    text += `[${word({ INVERSE: "Inverse", DIVERSE: "Diverse" }, a.seriesBonusType)}(${a.rate || 0})]`;
  else if (a) {
    const state = word(enStates, a.statusEffectType || a.buffEffectType!);
    const bonus = ["STATUS_EFFECT", "BUFF_EFFECT", "SKILL"].includes(e.param)
      ? `${prob + (a.rate || 0)}%`
      : rate(e, a.rate || 0, options);
    text += ` ${bonus} if affected by ${state}.`;
  }
  return text.replace(/\s+/g, " ").trim();
}
function describeEnglishTrigger(s: BattleSkill): string {
  if (!s.trigger) return "";
  const rows: BattleCondition[] = [];
  for (let i = 0; i < s.trigger.conditions.length; i++) {
    const c = { ...s.trigger.conditions[i] },
      next = s.trigger.conditions[i + 1];
    if (
      c.paramCondition &&
      next?.stateCondition &&
      next.team === "BOTH" &&
      next.position === "TARGET"
    ) {
      c.stateCondition = next.stateCondition;
      i++;
    }
    rows.push(c);
  }
  let text = "",
    conjunction = "at",
    plural = false;
  rows.forEach((c, i) => {
    if (c.stateCondition) {
      if (c.stateCondition === "OPENING") text += "At the start of battle";
      else {
        const kind =
          c.stateCondition === "HAS_STATUS_EFFECT" ? "when" : "after";
        const join = i > 0 ? (conjunction === kind ? " and" : ",") : "";
        const conj =
          conjunction !== kind ? (kind === "when" ? "If" : "After") : "";
        conjunction = kind;
        const subject =
          c.position === "TARGET"
            ? "it"
            : c.position === "ALL"
              ? enPosition(c.team, "RANDOM")
              : enPosition(c.team, c.position);
        text += `${join} ${conj} ${subject} ${word({ HAS_STATUS_EFFECT: "has a status effect", AFTER_ACTIVE_SKILL: "uses an Active Skill", AFTER_ACTIVE_SKILL_TAKEN_DAMAGE: "takes damage from an Active Skill", AFTER_DEATH: "dies" }, c.stateCondition)}`;
      }
    }
    if (c.paramCondition) {
      const p = c.paramCondition,
        join =
          i > 0 || c.stateCondition
            ? conjunction === "when"
              ? " and"
              : ","
            : "",
        conj = conjunction !== "when" ? "If" : "";
      let stat = word(enStats, p.param);
      const target = enPosition(c.team, c.position);
      if (c.position === "SELF") {
        plural = false;
        stat =
          p.param === "ALL_PARAMS"
            ? `the total ${stat} of ${target}`
            : `${target}'s ${stat}`;
      } else if (c.position === "TARGET")
        stat = `${plural ? "their" : "its"} ${p.param === "ALL_PARAMS" ? "total" : ""} ${stat}`;
      else {
        plural = true;
        stat =
          c.position === "ALL" || c.team === "BOTH" || p.param === "ALL_PARAMS"
            ? `the total ${stat} of ${target}`
            : `${target}'s ${stat}`;
      }
      conjunction = "when";
      text += `${join} ${conj} ${stat} is ${word({ UNDER: "less than", OVER: "over" }, p.calc)} ${p.rate || 0}% ${p.param === "HP" ? "" : "of its original value"}`;
    }
  });
  const uses = s.remainingUses || 0;
  return `${text}, ${s.trigger.triggerRate || 0}% chance to trigger${uses === 1 ? " once" : uses > 1 ? ` max ${uses} times` : ""}.`
    .replace(/\s+/g, " ")
    .trim();
}
