# 画像と音声

ファイル一覧・サイズ・SHA-256は `data/media-manifest.json`。画像はPNG、音声はMP3です。元のMCHフロントから保存した素材を、権利者が指定した範囲で収録しています。

- `assets/portraits/hero/{type}.png`: ヒーロー型ID対応の画像。
- `assets/portraits/extension/{type}.png`: エクステ型ID対応の画像。手塚オールスターの1060（ヒョウタンツギ）、2060（ロビタ）、3060（ユニコ）、4060（レオ）、5060（火の鳥）およびPK Alternaの4166（レーヴァテイン）、4167（アルマス）、4168（フェイルノート）は除外。それ以外は権利者が許諾した範囲として収録しています。
- `assets/portraits/enemy/`: 敵画像。敵IDとファイルの対応は `data/enemies.json`。419.png・1044.png・1045.pngに加え、手塚作品の派生画像715.png（ゴースト・猿田彦）・885.png（ゴースト・鉄腕アトム）およびそれらを参照する全敵エントリを除外しています。
- `assets/battle-effects/1.png`–`12.png`: 原作バトルのエフェクト画像。
- `assets/battle-icons/`: PHY/INT/状態異常等のアイコン。
- `assets/battle-backgrounds/raid-100.png`: 通常レイド背景。
- `assets/audio/bgm/pve.mp3`: 通常クエスト。pvp/land/raidは対人戦/ランド/レイド。
- `assets/audio/se/1.mp3`–`5.mp3`: 共通効果音。win/lose/resultは結果用音声。

ヒーロー画像の除外: 2035,4042,4043（手塚作品）、2050,3052,4058（タツノコ）、2051,3053,4059（コブラ）、4061（PK Alterna）。OASYX・ETHEREMON・サトシALPHA CC/OMEGA CCは権利者の明示許可により収録しています。データ内の名称やスキル参照の存在と画像利用許可は別です。

今回未収録: 新規生成のエフェクト、図鑑用アレンジ曲。未収録は使用禁止の断定ではありません。

ブラウザでの音声再生はユーザー操作後に開始し、タブ非表示時は停止してください。このリポジトリに音声プレイヤーは含めていません。
