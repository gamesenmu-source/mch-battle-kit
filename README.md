# MCH Battle Kit

**AIにこのリポジトリのURLを渡してください。** 最初に [AGENTS.md](AGENTS.md) を読み、サンプルを動かせば、マイクリ軽量版のバトル計算とスキルデータを調べられます。

> このリポジトリを読み、AGENTS.mdに従ってサンプルバトルを実行してください。実際のコードとマスターデータを根拠に、デッキの計算やスキルの仕組みを説明してください。

MCH LITE ARENA向けに分離したGoエンジンです。本家全体や本家の完全再現版ではありません。画像・音楽・ウォレット・課金・ログイン・本番サーバー・個人データは含みません。

## すぐ試す

Go 1.23以降を用意して、リポジトリ直下で実行します。初回のみGo依存パッケージの取得にネットワークを使います。バトル自体はローカルで完結します。

```sh
go run ./cmd/simulate -input examples/battle.json > battle-result.json
go test ./...
```

`battle-result.json` にスコア・勝敗・各行動のHPや状態が出ます。同じ入力・seed・エンジン・マスターで再現できます。入力の `catalogVersion` はラベルです。別バージョンのマスターを自動取得する機能ではありません。

スキル説明は計算用データから生成します。Node.js 24以降なら追加インストール不要です。

```sh
node scripts/describe.mjs 2002 ja
node scripts/describe.mjs 2002 en
```

## 何が入っているか

|場所|内容|
|---|---|
|`battle/`|戦闘計算・状態異常・バフ・ターゲット選択・protobuf型|
|`runner/`|3対3入力、カタログ検証、カスタムボススキル、リプレイ状態|
|`data/skills.ndjson`|計算に使用するスキルのprotobuf JSON。1行1スキル|
|`data/heroes.json`|最大レベルのヒーロー能力・パッシブID|
|`data/extensions.json`|最大レベルの装備加算値・アクティブID|
|`shared/skill-description.ts`|日本語・英語の効果文生成|
|`shared/custom-skills.ts`|自作スキルの選択要素・制約|
|`shared/custom-skill-mechanics.ts`|自作スキルから計算形式への変換|
|`docs/FORMAT.md`|入力・出力・デッキ計算・制限|

AI・開発者は [AGENTS.md](AGENTS.md)、[データ形式](docs/FORMAT.md)、[由来と利用条件](NOTICE.md) を参照してください。

## 範囲と注意

レプリカ・オーラ・本家の一部育成要素は対象外。軽量版のノービス3体は使用可能です。乱数処理の再現性のため一部の処理順を固定しています。本家の全組み合わせとの一致やゲームバランスは保証しません。

本リポジトリは公開コード資料です。第三者の商標・キャラクター・素材の利用権を与えるものではありません。利用許諾は [NOTICE.md](NOTICE.md) を確認してください。
