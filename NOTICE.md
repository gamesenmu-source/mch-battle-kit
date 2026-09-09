# 由来と利用条件 / Provenance and rights

戦闘計算および生成済みメッセージ型はMy Crypto HeroesのGo実装を基に、MCH LITE ARENA向けに分離・変更したものです。元の著作権は各権利者に帰属します。本リポジトリに収録するコード・計算データ・画像・音声は、権利者の許諾に基づき、非営利目的に限り利用できます。営利目的での利用を希望する場合は、事前にMCH社に相談し、許諾を得てください。この公開は、MIT等の営利利用を含む包括的な再利用ライセンスを付与するものではありません。

The code, calculation data, images, and audio included in this repository may be used for non-commercial purposes only, with permission from the rights holders. For any commercial use, please consult MCH in advance and obtain permission. Publication does not grant an unrestricted license such as MIT.

権利者の追加許可により、docs/MEDIA.md記載の画像・音声を収録しています。除外指定のコラボ素材、社内文書、元リポジトリの履歴、運営用設定、利用者情報は含めていません。名称の記載はキャラクターや商標の利用許諾を意味しません。

変更範囲: DB依存をローカルの不変スキルカタログに置換。オーラ等の呼び出しを除去。状態・バフの処理順を安定化。3対3入力、カスタムボススキル、リプレイ出力、ローカルCLIを追加。生成型には使用しない互換フィールドが残っています。

Goの外部依存はgo.mod/go.sumに固定し、ソースを同梱していません。各依存のライセンスは各配布元の条件に従います。
