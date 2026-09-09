import generated from "../data/custom-skill-templates.json";
import type { CustomEffect, CustomSkill } from "./custom-skills";
import type {
  BattleEffect,
  BattleSkill,
  BattleTarget,
} from "./skill-description";
const templates = generated as unknown as {
  kinds: Record<string, { param: string; isDamage: boolean }>;
  targets: Record<string, BattleTarget>;
  references: Record<string, Pick<BattleEffect, "rawStat" | "computedStat">>;
  triggers: Record<string, BattleSkill["trigger"] | null>;
};
export function compileCustomEffect(e: CustomEffect): BattleEffect {
  const result: BattleEffect = {
    ...templates.kinds[e.kind],
    target: structuredClone(templates.targets[e.target]),
    successRate: e.chance,
  };
  if (e.kind === "status" || e.kind === "cure")
    result.statusEffectType = e.status;
  else if (e.kind === "buff") result.buffEffectType = e.buff;
  else {
    result.rate = { min: e.min, max: e.max > e.min ? e.max + 1 : e.max };
    Object.assign(
      result,
      structuredClone(
        templates.references[`${e.reference}:${e.source}:${e.basis}`],
      ),
    );
  }
  return result;
}
export function compileCustomSkill(s: CustomSkill): BattleSkill {
  const result: BattleSkill = {
    skillId: s.id,
    remainingUses: s.uses,
    effects: s.effects.map(compileCustomEffect),
  };
  const trigger = templates.triggers[s.trigger];
  if (trigger) {
    result.trigger = structuredClone(trigger);
    result.trigger.triggerRate = s.chance;
    if (s.trigger === "lowHp")
      result.trigger.conditions[0].paramCondition!.rate = s.threshold;
  }
  return result;
}
