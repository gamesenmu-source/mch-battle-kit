import fs from 'node:fs';
import {describeBattleSkill} from '../shared/skill-description.ts';
const id=Number(process.argv[2]??2002),language=process.argv[3]==='en'?'en':'ja';
const skills=fs.readFileSync(new URL('../data/skills.ndjson',import.meta.url),'utf8').trim().split(/\r?\n/).map(JSON.parse);
const skill=skills.find(s=>(s.skillId||0)===id);
if(!skill)throw Error('Unknown skill ID');
const metadata=JSON.parse(fs.readFileSync(new URL('../data/skill-metadata.json',import.meta.url),'utf8'));
const name=n=>language==='en'?metadata[n]?.nameEn:metadata[n]?.name;
console.log(JSON.stringify(describeBattleSkill(skill,{name:name(id)||String(id)},{language,skillName:n=>name(n)||String(n)}),null,2));
